package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

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

func main() {
	checkDatabse()

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.Handle("/css/style.css", http.FileServer(http.Dir(webDir)))
	http.Handle("/js/scripts.min.js", http.FileServer(http.Dir(webDir)))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(webDir)))

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
