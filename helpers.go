package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func makeStructJSON(w http.ResponseWriter, rows *sql.Rows) error {

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	count := len(columns)
	values := make([]string, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	masterData := make([]interface{}, 0)

	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return err
		}
		rowData := make(map[string]interface{})
		for i, v := range values {
			x := v
			if nx, ok := strconv.ParseFloat(string(x), 64); ok == nil {
				rowData[columns[i]] = nx
			} else if b, ok := strconv.ParseBool(string(x)); ok == nil {
				rowData[columns[i]] = b
			} else if fmt.Sprintf("%T", string(x)) == "string" {
				rowData[columns[i]] = string(x)
			} else {
				fmt.Printf("Failed on if for type %T of %v\n", x, x)
			}
		}
		masterData = append(masterData, rowData)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(masterData)

	if err != nil {
		return err
	}
	return err
}
