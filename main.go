package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var key = "TODO_PORT"
var webDir = "./web"
var DB *sql.DB

type Task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title" binding:"required"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func checkDatabse() {
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}

	if install {
		db, err := sql.Open("sqlite", "scheduler.db")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()

		createTable := "CREATE TABLE scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, date INTEGER NOT NULL, title TEXT NOT NULL, comment TEXT, repeat varchar(128) NOT NULL);"
		createIndex := "CREATE INDEX dateindex ON scheduler (date);"

		_, err = db.Exec(createTable)
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = db.Exec(createIndex)
		if err != nil {
			fmt.Println(err)
			return
		}

		install = false
	}
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	t, err := time.Parse("20060102", date)
	if err != nil {
		return "", errors.New("date не может быть преобразовано в корректную дату")
	}

	if repeat == "" {
		return "", errors.New("не задан формат повторения")
	}

	if repeat == "y" {
		newDate := t.AddDate(1, 0, 0)
		for newDate.Format("20060102") <= now.Format("20060102") {
			newDate = newDate.AddDate(1, 0, 0)
		}
		newDateStr := newDate.Format("20060102")
		return newDateStr, nil
	}

	parts := strings.Split(repeat, " ")
	if len(parts) < 2 {
		return "", errors.New("некорректный формат повторения")
	}

	letter := parts[0]
	dayNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Println("Ошибка преобразования:", err)
		return "", errors.New("ошибка преобразования")
	}

	if letter == "d" {
		if dayNumber > 400 {
			return "", errors.New("максимально допустимое число днй равно 400")
		}
		newDate := t.AddDate(0, 0, dayNumber)
		for newDate.Format("20060102") <= now.Format("20060102") {
			newDate = newDate.AddDate(0, 0, dayNumber)
		}
		newDateStr := newDate.Format("20060102")
		return newDateStr, nil
	} else {
		return "", errors.New("некорректный формат повторения")
	}
}

func AddTask(date string, title string, comment string, repeat string) (int64, error) {
	res, err := DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", date),
		sql.Named("title", title),
		sql.Named("comment", comment),
		sql.Named("repeat", repeat))
	if err != nil {
		return 0, errors.New("ошибка записи в БД")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetTaskById(id int) (Task, error) {
	var task Task
	row := DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id", sql.Named("id", id))
	err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return task, errors.New("задача с таким id не найдена")
	}
	return task, nil
}
func UpdateTask(id int, date string, title string, comment string, repeat string) error {
	_, err := DB.Exec("UPDATE scheduler SET date=:date, title=:title, comment=:comment, repeat=:repeat WHERE id=:id",
		sql.Named("id", id),
		sql.Named("date", date),
		sql.Named("title", title),
		sql.Named("comment", comment),
		sql.Named("repeat", repeat))
	if err != nil {
		return errors.New("ошибка записи в БД")
	}
	return nil
}

func GetTasks() ([]Task, error) {
	rows, err := DB.Query("SELECT id, date, title, comment, repeat FROM scheduler order by date limit 50")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		task := Task{}

		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := GetTasks()

	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = make([]Task, 0)
	}

	resp, err := json.Marshal(map[string]any{"tasks": tasks})
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	taskId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, errorResponse("неверный номер задачи"), http.StatusBadRequest)
		return
	}
	task, err := GetTaskById(taskId)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(task)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	timeNow, err := time.Parse("20060102", now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")
	text, err := NextDate(timeNow, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, text)
}

func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		http.Error(w, errorResponse("название задачи обязательно для заполнения"), http.StatusBadRequest)
		return
	}

	if task.Date == "" {
		dateNow := time.Now()
		task.Date = dateNow.Format("20060102")
	} else {
		dateNow, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, errorResponse("некорректный формат даты"), http.StatusBadRequest)
			return
		}
		if dateNow.Format("20060102") < time.Now().Format("20060102") {
			if task.Repeat == "" {
				date := time.Now()
				task.Date = date.Format("20060102")
			} else {
				date, err := NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
					return
				}
				task.Date = date
			}
		}
	}
	id, err := AddTask(task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(map[string]any{"id": id})
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	taskId, err := strconv.Atoi(task.Id)
	if err != nil {
		http.Error(w, errorResponse("неверный номер задачи"), http.StatusBadRequest)
		return
	}

	_, err = GetTaskById(taskId)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, errorResponse("название задачи обязательно для заполнения"), http.StatusBadRequest)
		return
	}

	if task.Date == "" {
		dateNow := time.Now()
		task.Date = dateNow.Format("20060102")
	} else {
		dateNow, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, errorResponse("некорректный формат даты"), http.StatusBadRequest)
			return
		}
		if dateNow.Format("20060102") < time.Now().Format("20060102") {
			if task.Repeat == "" {
				date := time.Now()
				task.Date = date.Format("20060102")
			} else {
				date, err := NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
					return
				}
				task.Date = date
			}
		}
	}
	err = UpdateTask(taskId, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPost:
		AddTaskHandler(w, r)
	case http.MethodDelete:
		fmt.Fprintln(w, "DELETE запрос обработан")
	case http.MethodPut:
		UpdateTaskHandler(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func errorResponse(text string) string {
	message, _ := json.Marshal(map[string]any{"error": text})
	return string(message)
}

func main() {
	//checkDatabse()

	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	DB = db
	defer db.Close()

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.Handle("/css/style.css", http.FileServer(http.Dir(webDir)))
	http.Handle("/js/scripts.min.js", http.FileServer(http.Dir(webDir)))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/tasks", getTasksHandler)

	if value, exists := os.LookupEnv(key); exists {
		if err := http.ListenAndServe(value, nil); err != nil {
			fmt.Printf("ошибка запуска сервера: %s\n", err.Error())
			return
		}
	} else {
		if err := http.ListenAndServe(":7540", nil); err != nil {
			fmt.Printf("ошибка запуска сервера: %s\n", err.Error())
			return
		}
	}
}
