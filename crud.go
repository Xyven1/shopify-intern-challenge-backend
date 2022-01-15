package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type Event struct {
	Uid          int64     `db:"uid"`
	Action       string    `db:"action"`
	ItemUid      int64     `db:"item_uid"`
	ItemPrevious *Item     `db:"item_previous"`
	Timestamp    time.Time `db:"timestamp"`
	Comment      string    `db:"comment"`
}

type Item struct {
	Uid     int64  `json:"uid"`
	Name    string `json:"name"`
	Ammount int64  `json:"ammount"`
}

func (e *Item) Scan(src interface{}) error {
	var value string
	switch src.(type) {
	case string:
		value = src.(string)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
	if value == "" {
		return nil
	}
	*e = Item{}
	err := json.Unmarshal([]byte(value), e)
	if err != nil {
		return err
	}
	return nil
}

type ItemJson Item

type CreateReq struct {
	Name    string `schema:"name"`
	Ammount int64  `schema:"ammount"`
}

type ReadReq struct {
	MinAmmount *int64 `schema:"minAmmount"`
	MaxAmmount *int64 `schema:"maxAmmount"`
}

type UpdateReq struct {
	Uid     int64   `schema:"uid"`
	Name    *string `schema:"name"`
	Ammount *int64  `schema:"ammount"`
}

type DeleteReq struct {
	Uid     int64
	Comment *string `schema:"comment"`
}

type UndoReq struct {
	Uid     *int64 `schema:"uid"`
	ItemUid *int64 `schema:"item_uid"`
}

func Create(env *Env, w http.ResponseWriter, r *http.Request) error {
	var req CreateReq
	err := parseFormData(r, &req)
	if err != nil {
		return err
	}
	res, err := env.DB.Exec("INSERT INTO inventory (name, ammount) VALUES (?, ?)", req.Name, req.Ammount)
	if err != nil {
		return StatusError{http.StatusBadRequest, err}
	}
	id, err := res.LastInsertId()
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	fmt.Fprintf(w, "New Record ID is %d", id)
	env.PushHistory("create", id, "", "")
	return nil
}

func Read(env *Env, w http.ResponseWriter, r *http.Request) error {
	var req ReadReq
	err := parseFormData(r, &req)
	if err != nil {
		return err
	}
	items := []ItemJson{}
	queries := []string{}
	params := []interface{}{}
	if req.MaxAmmount != nil {
		queries = append(queries, " WHERE ammount<=? ")
		params = append(params, *req.MaxAmmount)
	}
	if req.MinAmmount != nil {
		queries = append(queries, " WHERE ammount>=? ")
		params = append(params, *req.MinAmmount)
	}
	err = env.DB.Select(&items, "SELECT * FROM inventory"+strings.Join(queries, "AND"), params...)
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	json.NewEncoder(w).Encode(items)
	return nil
}

func Update(env *Env, w http.ResponseWriter, r *http.Request) error {
	var req UpdateReq
	err := parseFormData(r, &req)
	if err != nil {
		return err
	}
	//save old values
	json, err := jsonFromUid(env, req.Uid)
	if err != nil {
		return err
	}
	//prepare query
	params := []interface{}{}
	values := []string{}

	if req.Name != nil {
		params = append(params, req.Name)
		values = append(values, "name=?")
	}
	if req.Ammount != nil {
		params = append(params, req.Ammount)
		if mux.Vars(r)["option"] == "increment" {
			values = append(values, "ammount=ammount+?")
		} else {
			values = append(values, "ammount=?")
		}
	}
	if len(params) == 0 {
		return StatusError{http.StatusBadRequest, fmt.Errorf("please enter at least new data for at least one column")}
	}
	params = append(params, req.Uid)
	//execute query
	_, err = env.DB.Exec("UPDATE inventory SET "+strings.Join(values, ",")+" WHERE uid=?", params...)
	if err != nil {
		return StatusError{http.StatusBadRequest, err}
	}
	//push history
	env.PushHistory("update", req.Uid, json, "")
	return nil
}

func Delete(env *Env, w http.ResponseWriter, r *http.Request) error {
	var req DeleteReq
	err := parseFormData(r, &req)
	if err != nil {
		return err
	}
	//save old values
	json, err := jsonFromUid(env, req.Uid)
	if err != nil {
		return err
	}
	res, err := env.DB.Exec("DELETE FROM inventory WHERE uid=?", req.Uid)
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	if _, err := res.RowsAffected(); err != nil {
		return StatusError{http.StatusBadRequest, fmt.Errorf("no record found with uid %d", req.Uid)}
	}
	comment := ""
	if req.Comment != nil {
		comment = *req.Comment
	}
	env.PushHistory("delete", req.Uid, json, comment)
	return nil
}

func Undo(env *Env, w http.ResponseWriter, r *http.Request) error {
	var req UndoReq
	err := parseFormData(r, &req)
	if err != nil {
		return err
	}
	//restore old values
	var event Event
	if req.ItemUid != nil {
		//undo last action for specific item
		err := env.DB.Get(&event, "SELECT * FROM event_history WHERE item_uid=? ORDER BY uid DESC LIMIT 1", req.ItemUid)
		if err != nil {
			return StatusError{http.StatusBadRequest, fmt.Errorf("no history found for item with uid %d", req.Uid)}
		}
	} else if req.Uid != nil {
		//undo specific action in history
		err := env.DB.Get(&event, "SELECT * FROM event_history WHERE uid=?", req.Uid)
		if err != nil {
			return StatusError{http.StatusBadRequest, err}
		}

	} else {
		//undo last action
		err := env.DB.Get(&event, "SELECT * FROM event_history ORDER BY uid DESC LIMIT 1")
		if err != nil {
			return StatusError{http.StatusBadRequest, fmt.Errorf("no history found")}
		}
	}
	switch event.Action {
	case "delete":
		_, err := env.DB.Exec("INSERT INTO inventory (uid, name, ammount) VALUES (?, ?, ?)", event.ItemUid, event.ItemPrevious.Name, event.ItemPrevious.Ammount)
		if err != nil {
			return StatusError{http.StatusInternalServerError, err}
		}
	case "update":
		_, err := env.DB.Exec("UPDATE inventory SET name=?, ammount=? WHERE uid=?", event.ItemPrevious.Name, event.ItemPrevious.Ammount, event.ItemUid)
		if err != nil {
			return StatusError{http.StatusInternalServerError, err}
		}
	default:
		return StatusError{http.StatusBadRequest, fmt.Errorf("unknown action %s", event.Action)}
	}
	env.DB.Exec("DELETE FROM event_history WHERE uid=?", event.Uid)
	//execute query
	return nil
}

func parseFormData(r *http.Request, item interface{}) error {
	err := r.ParseForm()
	if err != nil {
		return StatusError{http.StatusInternalServerError, err}
	}
	var decoder = schema.NewDecoder()
	err = decoder.Decode(item, r.PostForm)
	if err != nil {
		return StatusError{http.StatusBadRequest, err}
	}
	return nil
}

func jsonFromUid(env *Env, uid int64) (string, error) {
	var item ItemJson
	err := env.DB.Get(&item, "SELECT * FROM inventory WHERE uid=?", uid)
	if err != nil {
		return "", StatusError{http.StatusBadRequest, err}
	}
	json, err := json.Marshal(item)
	if err != nil {
		return "", StatusError{http.StatusInternalServerError, err}
	}
	return string(json), nil
}
