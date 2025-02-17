package main

import (
	"database/sql"
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
		if now.Year() > t.Year() {
			newDate := time.Date(
				now.Year(),
				t.Month(),
				t.Day(),
				t.Hour(),
				t.Minute(),
				t.Second(),
				t.Nanosecond(),
				t.Location(),
			)
			newDateStr := newDate.Format("20060102")
			return newDateStr, nil
		}
		newDate := t.AddDate(1, 0, 0)
		newDateStr := newDate.Format("20060102")
		return newDateStr, nil
	}

	parts := strings.Split(repeat, " ")
	if len(parts) < 2 {
		return "", errors.New("некорректный формат повторения")
	}

	letter := parts[0]
	dayNumber, err := strconv.Atoi(parts[1])
	fmt.Println("Буква:", letter)
	fmt.Println("Число:", dayNumber)
	if err != nil {
		fmt.Println("Ошибка преобразования:", err)
		return "", errors.New("ошибка преобразования")
	}

	if letter == "d" {
		if dayNumber > 400 {
			return "", errors.New("максимально допустимое число днй равно 400")
		}
		newDate := t.AddDate(0, 0, dayNumber)
		for newDate.Before(now) {
			newDate = newDate.AddDate(0, 0, dayNumber)
		}
		newDateStr := newDate.Format("20060102")
		return newDateStr, nil
	} else {
		fmt.Println("Error:", err)
		return "", errors.New("некорректный формат повторения")
	}
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

func main() {
	checkDatabse()
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.Handle("/css/style.css", http.FileServer(http.Dir(webDir)))
	http.Handle("/js/scripts.min.js", http.FileServer(http.Dir(webDir)))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", nextDateHandler)

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
