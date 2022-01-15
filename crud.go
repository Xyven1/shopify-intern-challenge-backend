package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type Item struct {
	Uid     int64
	Name    *string `schema:"name"`
	Ammount *int64  `schema:"ammount"`
}

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
	env.PushHistory("create", id, "", "created new record")
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
	//parse form values
	err := r.ParseForm()
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	var decoder = schema.NewDecoder()
	var newItem Item
	err = decoder.Decode(&newItem, r.PostForm)
	if err != nil {
		return StatusError{http.StatusBadRequest, err}
	}
	//save old values
	json, err := jsonFromUid(env, newItem.Uid)
	if err != nil {
		return err
	}
	//prepare query
	params := []interface{}{}
	values := []string{}

	if newItem.Name != nil {
		params = append(params, newItem.Name)
		values = append(values, "name=?")
	}
	if newItem.Ammount != nil {
		params = append(params, newItem.Ammount)
		if mux.Vars(r)["option"] == "increment" {
			values = append(values, "ammount=ammount+?")
		} else {
			values = append(values, "ammount=?")
		}
	}
	if len(params) == 0 {
		return StatusError{http.StatusBadRequest, fmt.Errorf("please enter at least new data for at least one column")}
	}
	params = append(params, newItem.Uid)
	query := "UPDATE inventory SET " + strings.Join(values, ",") + " WHERE uid=?"
	//execute query
	_, err = env.DB.Exec(query, params...)
	if err != nil {
		return StatusError{http.StatusBadRequest, err}
	}
	//push history
	env.PushHistory("update", newItem.Uid, json, "created new record")
	return nil
}

func jsonFromUid(env *Env, uid int64) (string, error) {
	var item Item
	err := env.DB.QueryRowx("SELECT * FROM inventory WHERE uid=?", uid).StructScan(&item)
	if err != nil {
		return "", StatusError{http.StatusBadRequest, err}
	}
	json, err := json.Marshal(item)
	if err != nil {
		return "", StatusError{http.StatusInternalServerError, err}
	}
	return string(json), nil
}

func Delete(env *Env, w http.ResponseWriter, r *http.Request) error {
	uid, err := strconv.ParseInt(r.FormValue("uid"), 10, 64)
	if err != nil {
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
		env.PushHistory("delete", uid, "", "created new record")
	} else {
		fmt.Fprintf(w, "No record found")
	}
	return nil
}
