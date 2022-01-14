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

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func eventHistory(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)
		statusCode := lrw.statusCode
		if statusCode == http.StatusOK {
			db := dbConn()
			defer db.Close()
			if r.Method == "POST" {
				query, err := db.Prepare("INSERT INTO event_history (action, item_uid, description) VALUES (?, ?, ?)")
				checkErr(err)
				switch r.URL.Path {
				case "/create":
					query.Exec("create", "", r.FormValue("uid"))
				case "/update":
					if vars := mux.Vars(r); vars["option"] == "increment" {
						query.Exec("update", r.FormValue("uid"), "increment")
					}
					query.Exec("update", r.FormValue("uid"), "")
				case "/delete":
					query.Exec("delete", r.FormValue("uid"), "")
				}
			}
		}
	})
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", Index)
	r.HandleFunc("/create", Create).Methods("POST")
	r.HandleFunc("/read", Read).Methods("GET")
	r.HandleFunc("/update", Update).Methods("POST")
	r.HandleFunc("/update/{option}", Update).Methods("POST")
	r.HandleFunc("/delete", Delete).Methods("POST")
	r.Use(eventHistory)
	log.Fatal(http.ListenAndServe(":8080", r))
}
