package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Create(env *Env, w http.ResponseWriter, r *http.Request) error {
	name := r.FormValue("name")
	ammount := r.FormValue("ammount")
	if name == "" {
		return StatusError{http.StatusBadRequest, fmt.Errorf("please enter a name")}
	} else if ammount == "" {
		return StatusError{http.StatusBadRequest, fmt.Errorf("please enter an ammount")}
	}
	stmt, err := env.DB.Prepare("INSERT INTO inventory (name,ammount) VALUES(?,?)")
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	res, err := stmt.Exec(name, ammount)
	if err != nil {
		return StatusError{http.StatusBadRequest, err}
	}
	id, err := res.LastInsertId()
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	fmt.Fprintf(w, "New Record ID is %d", id)
	r.URL.RawQuery = r.URL.RawQuery + "&uid=" + fmt.Sprintf("%d", id)
	env.HistoryQuery.Exec("create", id, "created new record")
	return nil
}

func Read(env *Env, w http.ResponseWriter, r *http.Request) error {
	minAmt := r.FormValue("minAmt")
	maxAmt := r.FormValue("maxAmt")
	var rows *sql.Rows
	var err error
	switch {
	case minAmt == "" && maxAmt == "":
		rows, err = env.DB.Query("SELECT * FROM inventory")
	case minAmt != "" && maxAmt == "":
		rows, err = env.DB.Query("SELECT * FROM inventory WHERE ammount>=?", minAmt)
	case minAmt == "" && maxAmt != "":
		rows, err = env.DB.Query("SELECT * FROM inventory WHERE ammount<=?", maxAmt)
	case minAmt != "" && maxAmt != "":
		rows, err = env.DB.Query("SELECT * FROM inventory WHERE ammount BETWEEN ? AND ?", minAmt, maxAmt)
	}
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	makeStructJSON(w, rows)
	return nil
}

func Update(env *Env, w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	options := vars["option"]
	//parse form values
	uid := r.FormValue("uid")
	name := r.FormValue("name")
	ammount := r.FormValue("ammount")
	if uid == "" {
		return StatusError{http.StatusBadRequest, fmt.Errorf("please enter a uid")}
	}
	//prepare query
	query := "UPDATE inventory SET "
	if name != "" {
		query += "name=$1"
	}
	switch options {
	case "":
		// query.a"UPDATE inventory SET ammount=? WHERE uid=?"
		query += ",ammount=$2"
	case "increment":
		query += ",ammount=ammount+$2"
	}
	query += " WHERE uid=$3"
	stmt, err := env.DB.Prepare(query)
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	//execute query
	res, err := stmt.Exec(name, ammount, uid)
	if err != nil {
		return StatusError{http.StatusBadRequest, err}
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	if affect > 0 {
		fmt.Fprintf(w, "Deleted Record ID is %d", affect)
	} else {
		fmt.Fprintf(w, "No record found")
	}
	env.HistoryQuery.Exec("create", uid, "created new record")
	return nil
}

func Delete(env *Env, w http.ResponseWriter, r *http.Request) error {
	uid := r.FormValue("uid")
	if uid == "" {
		return StatusError{http.StatusBadRequest, fmt.Errorf("please enter a uid")}
	}
	stmt, err := env.DB.Prepare("DELETE FROM inventory WHERE uid=?")
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	res, err := stmt.Exec(uid)
	if err != nil {
		return StatusError{http.StatusBadRequest, err}
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	if affect > 0 {
		fmt.Fprintf(w, "Deleted Record ID is %d", affect)
		env.HistoryQuery.Exec("delete", uid, "created new record")
	} else {
		fmt.Fprintf(w, "No record found")
	}
	return nil
}
