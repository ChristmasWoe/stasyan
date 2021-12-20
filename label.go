package main

import (
	"encoding/json"
	"log"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

type Label struct {
	Id          string
	Name        string
	Color       string
	Description string
}

func getLabels(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	rows, err := db.Query("SELECT * FROM label")
	if err != nil {
		log.Fatal(err)
	}

	lbs := make([]Label, 0)

	for rows.Next() {
		var lb Label
		rows.Scan(&lb.Name, &lb.Color, &lb.Description, &lb.Id)

		lbs = append(lbs, lb)
	}

	lbsBytes, _ := json.MarshalIndent(lbs, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.Write(lbsBytes)

	defer rows.Close()
	defer db.Close()
}

func createLabel(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	var lb Label
	r.ParseMultipartForm(0)

	lb.Name = r.FormValue("name")
	lb.Description = r.FormValue("description")
	lb.Color = r.FormValue("color")
	lb.Id = uuid.NewV4().String()

	sqlStatement := `INSERT INTO label (name, color, description, id) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(sqlStatement, lb.Name, lb.Color, lb.Description, lb.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	lbBytes, _ := json.MarshalIndent(lb, "", "\t")
	w.Write(lbBytes)
	defer db.Close()
}

func editLabel(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	r.ParseMultipartForm(0)
	var lb Project
	lb.Id = r.FormValue("id")
	lb.Name = r.FormValue("name")
	lb.Description = r.FormValue("description")
	lb.Color = r.FormValue("color")
	sqlStatement := `UPDATE label SET name = $1, description = $2, color = $3 WHERE id = $4;`
	_, err := db.Exec(sqlStatement, lb.Name, lb.Description, lb.Color, lb.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	lbBytes, _ := json.MarshalIndent(lb, "", "\t")
	w.Write(lbBytes)
	defer db.Close()
}
