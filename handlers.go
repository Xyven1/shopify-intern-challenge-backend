package main

import (
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

func Edit(w http.ResponseWriter, r *http.Request) {
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
	fmt.Fprintf(w, "Updated Record ID is %d", affect)
}
