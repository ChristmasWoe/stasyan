package main

import (
	"encoding/json"
	"log"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

type Project struct {
	Id          string
	Name        string
	Color       string
	Description string
}

func getProjects(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	rows, err := db.Query("SELECT * FROM project")
	if err != nil {
		log.Fatal(err)
	}

	prs := make([]Project, 0)

	for rows.Next() {
		var pr Project
		rows.Scan(&pr.Name, &pr.Color, &pr.Description, &pr.Id)

		prs = append(prs, pr)
	}

	prsBytes, _ := json.MarshalIndent(prs, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.Write(prsBytes)

	defer rows.Close()
	defer db.Close()
}

func createProject(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	var pr Project
	r.ParseMultipartForm(0)

	pr.Name = r.FormValue("name")
	pr.Description = r.FormValue("description")
	pr.Color = r.FormValue("color")
	pr.Id = uuid.NewV4().String()

	sqlStatement := `INSERT INTO project (name, color, description, id) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(sqlStatement, pr.Name, pr.Color, pr.Description, pr.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	prBytes, _ := json.MarshalIndent(pr, "", "\t")
	w.Write(prBytes)
	defer db.Close()
}

func editProject(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	r.ParseMultipartForm(0)
	var pr Project
	pr.Id = r.FormValue("id")
	pr.Name = r.FormValue("name")
	pr.Description = r.FormValue("description")
	pr.Color = r.FormValue("color")
	sqlStatement := `UPDATE project SET name = $1, description = $2, color = $3 WHERE id = $4;`
	_, err := db.Exec(sqlStatement, pr.Name, pr.Description, pr.Color, pr.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	prBytes, _ := json.MarshalIndent(pr, "", "\t")
	w.Write(prBytes)
	defer db.Close()
}
