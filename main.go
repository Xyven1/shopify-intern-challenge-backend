package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
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
	db, err := sql.Open("sqlite3", "./database.db")
	checkErr(err)
	query, err := db.Prepare("INSERT INTO event_history (action, item_uid, description) VALUES (?, ?, ?)")
	checkErr(err)

	env := &Env{
		DB:           db,
		HistoryQuery: query,
	}

	r.HandleFunc("/", Index)
	r.Handle("/create", Handler{env, Create}).Methods("POST")
	r.Handle("/read", Handler{env, Read}).Methods("GET")
	r.Handle("/update", Handler{env, Update}).Methods("POST")
	r.Handle("/update/{option}", Handler{env, Update}).Methods("POST")
	r.Handle("/delete", Handler{env, Delete}).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", r))
}
