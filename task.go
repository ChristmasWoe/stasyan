package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"
)

type Task struct {
	Id   string
	Name string
	// ProjectId string
	// LabelId   string
	Status bool
}

type ProjectTaskRelation struct {
	TaskId    string
	ProjectId string
	Id        string
}

type LabelTaskRelation struct {
	TaskId  string
	LabelId string
	Id      string
}

func getTask(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	r.ParseMultipartForm(0)
	tid := r.FormValue("task_id")
	var ts Task

	//get task by task id
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM task WHERE id = '%s';", tid))
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		rows.Scan(&ts.Name, &ts.Status, &ts.Id)
	}

	//get task's project by task id
	rows, err = db.Query(fmt.Sprintf("SELECT * FROM projecttask WHERE task_id = '%s';", tid))
	if err != nil {
		log.Fatal(err)
	}

	var pid string

	for rows.Next() {
		var ptr ProjectTaskRelation
		rows.Scan(&ptr.TaskId, &ptr.ProjectId, &ptr.Id)
		pid = ptr.ProjectId
	}
	//get lables id

	lidsArray := make([]string, 0)

	rows, err = db.Query(fmt.Sprintf("SELECT * FROM labeltask WHERE task_id = '%s';", tid))
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var ltr LabelTaskRelation
		rows.Scan(&ltr.TaskId, &ltr.LabelId, &ltr.Id)

		lidsArray = append(lidsArray, ltr.LabelId)
	}

	taskBytes, _ := json.MarshalIndent(struct {
		Id        string
		Name      string
		ProjectId string
		Labels    []string
		Status    bool
	}{Id: tid,
		Name:      ts.Name,
		ProjectId: pid,
		Labels:    lidsArray,
		Status:    ts.Status,
	}, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.Write(taskBytes)

	defer rows.Close()
	defer db.Close()

}

func getTasks(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	r.ParseMultipartForm(0)

	if r.FormValue("project_id") != "" { // by project
		rows, err := db.Query(fmt.Sprintf("SELECT * FROM projecttask WHERE project_id = '%s';", r.FormValue("project_id")))
		if err != nil {
			log.Fatal(err)
		}

		tids := make([]string, 0)

		for rows.Next() {
			var ptr ProjectTaskRelation
			rows.Scan(&ptr.TaskId, &ptr.ProjectId, &ptr.Id)
			tids = append(tids, ptr.TaskId)
		}
		rows, err = db.Query(fmt.Sprintf("SELECT * FROM task WHERE id IN (%s)", "'"+strings.Join(tids, "', '")+"'"))
		if err != nil {
			log.Fatal(err)
		}

		tasks := make([]Task, 0)

		for rows.Next() {
			var ts Task
			rows.Scan(&ts.Name, &ts.Status, &ts.Id)
			tasks = append(tasks, ts)
		}

		tasksBytes, _ := json.MarshalIndent(tasks, "", "\t")
		w.Header().Set("Content-Type", "application/json")
		w.Write(tasksBytes)

		defer rows.Close()
		defer db.Close()

	} else { //by label
		rows, err := db.Query(fmt.Sprintf("SELECT * FROM labeltask WHERE label_id = '%s';", r.FormValue("label_id")))
		if err != nil {
			log.Fatal(err)
		}

		tids := make([]string, 0)

		for rows.Next() {
			var ltr LabelTaskRelation
			rows.Scan(&ltr.TaskId, &ltr.LabelId, &ltr.Id)

			tids = append(tids, ltr.TaskId)
		}

		rows, err = db.Query(fmt.Sprintf("SELECT * FROM task WHERE id IN (%s)", "'"+strings.Join(tids, "', '")+"'"))
		if err != nil {
			log.Fatal(err)
		}

		tasks := make([]Task, 0)

		for rows.Next() {
			var ts Task
			rows.Scan(&ts.Name, &ts.Status, &ts.Id)
			tasks = append(tasks, ts)
		}

		tasksBytes, _ := json.MarshalIndent(tasks, "", "\t")
		w.Header().Set("Content-Type", "application/json")
		w.Write(tasksBytes)

		defer rows.Close()
		defer db.Close()
	}
}

func createTask(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	var ts Task
	r.ParseMultipartForm(0)

	ts.Name = r.FormValue("name")
	projectId := r.FormValue("project_id")
	labelIds := r.FormValue("labels")
	ts.Status = true
	ts.Id = uuid.NewV4().String()

	sqlStatement := `INSERT INTO task (name, status, id) VALUES ($1, $2, $3)`
	_, err := db.Exec(sqlStatement, ts.Name, ts.Status, ts.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	sqlStatement = `INSERT INTO projecttask (task_id, project_id, id) VALUES ($1, $2, $3)`
	_, err = db.Exec(sqlStatement, ts.Id, projectId, uuid.NewV4().String())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	if labelIds != "" {
		labels := strings.Split(labelIds, ";")
		for _, lid := range labels {
			sqlStatement = `INSERT INTO labeltask (task_id, label_id, id) VALUES ($1, $2, $3)`
			_, err = db.Exec(sqlStatement, ts.Id, lid, uuid.NewV4().String())
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				panic(err)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	tsBytes, _ := json.MarshalIndent(ts, "", "\t")
	w.Write(tsBytes)
	defer db.Close()
}

func editTask(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	r.ParseMultipartForm(0)
	var ts Task
	ts.Id = r.FormValue("id")
	ts.Name = r.FormValue("name")
	if r.FormValue("status") == "true" {
		ts.Status = true
	} else {
		ts.Status = false
	}
	newProjectId := r.FormValue("project_id")
	newLabelIds := r.FormValue("labels")

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM projecttask WHERE task_id = '%s';", ts.Id))
	if err != nil {
		log.Fatal(err)
	}

	ptrs := make([]ProjectTaskRelation, 0)

	for rows.Next() {
		var ptr ProjectTaskRelation
		rows.Scan(&ptr.TaskId, &ptr.ProjectId, &ptr.Id)

		ptrs = append(ptrs, ptr)
	}

	if ptrs[0].ProjectId != newProjectId {
		sqlStatement := `UPDATE projecttask SET task_id = $1, project_id = $2 WHERE id = $3;`
		_, err := db.Exec(sqlStatement, ts.Id, newProjectId, ptrs[0].Id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			panic(err)
		}
	}

	rows, err = db.Query(fmt.Sprintf("SELECT * FROM labeltask WHERE task_id = '%s';", ts.Id))
	if err != nil {
		log.Fatal(err)
	}

	ltrs := make([]LabelTaskRelation, 0)
	oldLabelIdsArray := make([]string, 0)
	for rows.Next() {
		var ltr LabelTaskRelation
		rows.Scan(&ltr.TaskId, &ltr.LabelId, &ltr.Id)

		ltrs = append(ltrs, ltr)
		oldLabelIdsArray = append(oldLabelIdsArray, ltr.LabelId)
	}

	oldLabelIds := strings.Join(oldLabelIdsArray[:], ";")

	if oldLabelIds != newLabelIds {
		for _, l := range ltrs {
			sqlStatement := `DELETE FROM labeltask WHERE id = $1 ;`
			_, err = db.Exec(sqlStatement, l.Id)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				panic(err)
			}
		}

		labels := strings.Split(newLabelIds, ";")
		for _, lid := range labels {
			sqlStatement := `INSERT INTO labeltask (task_id, label_id, id) VALUES ($1, $2, $3);`
			_, err = db.Exec(sqlStatement, ts.Id, lid, uuid.NewV4().String())
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				panic(err)
			}
		}
	}

	sqlStatement := `UPDATE task SET name = $1, status = $2 WHERE id = $3;`
	_, err = db.Exec(sqlStatement, ts.Name, strconv.FormatBool(ts.Status), ts.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	tsBytes, _ := json.MarshalIndent(ts, "", "\t")
	w.Write(tsBytes)
	defer db.Close()
}

func tickTask(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	r.ParseMultipartForm(0)
	var ts Task
	ts.Id = r.FormValue("id")
	if r.FormValue("status") == "true" {
		ts.Status = true
	} else {
		ts.Status = false
	}

	sqlStatement := `UPDATE task SET status = $1 WHERE id = $2;`
	_, err := db.Exec(sqlStatement, strconv.FormatBool(ts.Status), ts.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	tsBytes, _ := json.MarshalIndent(ts, "", "\t")
	w.Write(tsBytes)
	defer db.Close()
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	r.ParseMultipartForm(0)
	tid := r.FormValue("task_id")

	_, err := db.Query(fmt.Sprintf("DELETE FROM projecttask WHERE task_id = '%s';", tid))
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Query(fmt.Sprintf("DELETE FROM labeltask WHERE task_id = '%s';", tid))
	if err != nil {
		log.Fatal(err)
	}

	sqlStatement := `DELETE FROM task WHERE id = $1;`
	_, err = db.Exec(sqlStatement, tid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	tsBytes, _ := json.MarshalIndent(true, "", "\t")
	w.Write(tsBytes)
	defer db.Close()
}
