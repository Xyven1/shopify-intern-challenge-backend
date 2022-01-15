package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	r := mux.NewRouter()
	db, err := sqlx.Open("sqlite3", "./database.db")
	checkErr(err)
	query, err := db.Prepare("INSERT INTO event_history (action, item_uid, item_previous, description) VALUES (?, ?, ?, ?)")
	checkErr(err)

	env := &Env{
		DB: db,
		PushHistory: func(action string, item_uid int64, item_previous string, description string) {
			_, err := query.Exec(action, item_uid, item_previous, description)
			checkErr(err)
		},
	}

	r.HandleFunc("/", Index)
	r.Handle("/create", Handler{env, Create}).Methods("POST")
	r.Handle("/read", Handler{env, Read}).Methods("GET")
	r.Handle("/update", Handler{env, Update}).Methods("POST")
	r.Handle("/update/{option}", Handler{env, Update}).Methods("POST")
	r.Handle("/delete", Handler{env, Delete}).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", r))
}
