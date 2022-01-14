package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func dbConn() (db *sql.DB) {
	db, err := sql.Open("sqlite3", "./database.db")
	checkErr(err)
	return db
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/create", Create).Methods("POST")
	r.HandleFunc("/read", Read).Methods("GET")
	r.HandleFunc("/update", Update).Methods("POST")
	r.HandleFunc("/update/{option}", Update).Methods("POST")
	r.HandleFunc("/delete", Delete).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", r))
}
