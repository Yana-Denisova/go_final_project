package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

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
	successResponse(resp, w)
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

	successResponse(resp, w)
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
	successResponse(resp, w)

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
	successResponse([]byte("{}"), w)
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	taskId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, errorResponse("неверный номер задачи"), http.StatusBadRequest)
		return
	}
	_, err = GetTaskById(taskId)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}
	err = DeleteTask(taskId)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}
	successResponse([]byte("{}"), w)
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPost:
		AddTaskHandler(w, r)
	case http.MethodPut:
		UpdateTaskHandler(w, r)
	case http.MethodDelete:
		DeleteTaskHandler(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

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

	if task.Repeat == "" {
		err = DeleteTask(taskId)
		if err != nil {
			http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
			return
		}
		successResponse([]byte("{}"), w)
		return
	}

	newDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	err = UpdateTask(taskId, newDate, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}
	successResponse([]byte("{}"), w)
}

func errorResponse(text string) string {
	message, _ := json.Marshal(map[string]any{"error": text})
	return string(message)
}

func successResponse(response []byte, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
