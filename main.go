package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Task struct {
	Id       int
	Detail   string
	Pic      string
	Deadline string
	Isdone   bool
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func dbConnect() (db *sql.DB) {
	cfg := mysql.Config{
		User:                 "simpletasks",
		Passwd:               "simple123",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "simpletasks",
		AllowNativePasswords: true,
	}
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	return db
}

var tmpl, err = template.ParseGlob("views/*")

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	db := dbConnect()
	rows, err := db.Query("select * from tasks")
	checkErr(err)

	task := Task{}
	tasks := []Task{}

	for rows.Next() {
		var id int
		var detail, pic, deadline string
		var isdone bool
		err = rows.Scan(&id, &detail, &pic, &deadline, &isdone)
		checkErr(err)
		task.Id = id
		task.Detail = detail
		task.Pic = pic
		task.Deadline = deadline
		task.Isdone = isdone
		tasks = append(tasks, task)

	}
	err = tmpl.ExecuteTemplate(w, "index", tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer db.Close()
}

func handlerNew(w http.ResponseWriter, r *http.Request) {
	err = tmpl.ExecuteTemplate(w, "new", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlerInsert(w http.ResponseWriter, r *http.Request) {
	db := dbConnect()
	if r.Method == "POST" {
		detail := r.FormValue("detail")
		pic := r.FormValue("pic")
		deadline := r.FormValue("deadline")
		isdone := r.FormValue("is_done")
		insForm, err := db.Prepare("INSERT INTO tasks(detail, pic, deadline, is_done) VALUES(?,?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(detail, pic, deadline, isdone)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func handlerEdit(w http.ResponseWriter, r *http.Request) {
	db := dbConnect()
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM tasks WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}
	task := Task{}
	for selDB.Next() {
		var id int
		var detail, pic, deadline string
		var isdone bool
		err = selDB.Scan(&id, &detail, &pic, &deadline, &isdone)
		if err != nil {
			panic(err.Error())
		}
		task.Id = id
		task.Detail = detail
		task.Pic = pic
		task.Deadline = deadline
		task.Isdone = isdone
	}
	tmpl.ExecuteTemplate(w, "edit", task)
	defer db.Close()
}

func main() {

	http.HandleFunc("/", handlerIndex)
	http.HandleFunc("/index", handlerIndex)
	http.HandleFunc("/new", handlerNew)
	http.HandleFunc("/insert", handlerInsert)
	http.HandleFunc("/edit", handlerEdit)

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("assets"))))

	var address = "localhost:8080"
	fmt.Printf("server started at %s\n", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}
