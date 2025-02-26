package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title" binding:"required"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func checkDatabase() {

	appPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")

	var install bool
	_, err = os.Stat(dbFile)
	fmt.Println(err)
	if err != nil {
		install = true
	}

	if install {
		db, err := sql.Open("sqlite", "scheduler.db")
		if err != nil {
			log.Fatal(err)
			return
		}
		defer db.Close()

		createTable := "CREATE TABLE scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, date INTEGER NOT NULL, title TEXT NOT NULL, comment TEXT, repeat varchar(128) NOT NULL);"
		createIndex := "CREATE INDEX dateindex ON scheduler (date);"

		_, err = db.Exec(createTable)
		if err != nil {
			log.Fatal(err)
			return
		}
		_, err = db.Exec(createIndex)
		if err != nil {
			log.Fatal(err)
			return
		}
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

func DeleteTask(id int) error {
	_, err := DB.Exec("DELETE FROM scheduler WHERE id=:id", sql.Named("id", id))
	if err != nil {
		return errors.New("ошибка удаления из БД")
	}
	return nil
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
