package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Create(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	defer db.Close()
	if r.Method == "POST" {
		name := r.FormValue("name")
		ammount := r.FormValue("ammount")
		if name == "" {
			fmt.Fprintf(w, "Please enter a name")
			return
		} else if ammount == "" {
			fmt.Fprintf(w, "Please enter an ammount")
			return
		}
		stmt, err := db.Prepare("INSERT INTO inventory (name,ammount) VALUES(?,?)")
		checkErr(err)
		res, err := stmt.Exec(name, ammount)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
		id, err := res.LastInsertId()
		checkErr(err)
		fmt.Fprintf(w, "New Record ID is %d", id)
	}
}

func Read(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	defer db.Close()
	minAmt := r.FormValue("minAmt")
	maxAmt := r.FormValue("maxAmt")
	var rows *sql.Rows
	var err error
	switch {
	case minAmt == "" && maxAmt == "":
		rows, err = db.Query("SELECT * FROM inventory")
	case minAmt != "" && maxAmt == "":
		rows, err = db.Query("SELECT * FROM inventory WHERE ammount>=?", minAmt)
	case minAmt == "" && maxAmt != "":
		rows, err = db.Query("SELECT * FROM inventory WHERE ammount<=?", maxAmt)
	case minAmt != "" && maxAmt != "":
		rows, err = db.Query("SELECT * FROM inventory WHERE ammount BETWEEN ? AND ?", minAmt, maxAmt)
	}
	checkErr(err)
	makeStructJSON(w, rows)
}

func Update(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	defer db.Close()

	vars := mux.Vars(r)
	options := vars["option"]
	//parse form values
	name := r.FormValue("name")
	ammount := r.FormValue("ammount")
	if name == "" {
		fmt.Fprintf(w, "Please enter a name")
		return
	} else if ammount == "" {
		fmt.Fprintf(w, "Please enter an ammount")
		return
	}
	//prepare query
	var query string
	switch options {
	case "":
		query = "UPDATE inventory SET ammount=? WHERE name=?"
	case "increment":
		query = "UPDATE inventory SET ammount=ammount+? WHERE name=?"
	}
	stmt, err := db.Prepare(query)
	checkErr(err)
	//execute query
	res, err := stmt.Exec(ammount, name)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	affect, err := res.RowsAffected()
	checkErr(err)
	if affect > 0 {
		fmt.Fprintf(w, "Deleted Record ID is %d", affect)
	} else {
		fmt.Fprintf(w, "No record found")
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	defer db.Close()

	name := r.FormValue("name")
	if name == "" {
		fmt.Fprintf(w, "Please enter a name")
		return
	}
	stmt, err := db.Prepare("DELETE FROM inventory WHERE name=?")
	checkErr(err)
	res, err := stmt.Exec(name)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	affect, err := res.RowsAffected()
	checkErr(err)
	if affect > 0 {
		fmt.Fprintf(w, "Deleted Record ID is %d", affect)
	} else {
		fmt.Fprintf(w, "No record found")
	}
}
