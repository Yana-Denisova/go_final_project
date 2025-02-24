package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

var key = "TODO_PORT"
var webDir = "./web"
var DB *sql.DB

func main() {
	checkDatabse()

	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	DB = db
	defer DB.Close()

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.Handle("/css/style.css", http.FileServer(http.Dir(webDir)))
	http.Handle("/js/scripts.min.js", http.FileServer(http.Dir(webDir)))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/task/done", doneTaskHandler)
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
