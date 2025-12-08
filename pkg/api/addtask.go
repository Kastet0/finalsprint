package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"finalsprint/pkg/db"
	"finalsprint/pkg/nextdate"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, "JSON deserialization error: "+err.Error())
		return
	}

	if len(task.Title) == 0 {
		writeError(w, "Task title not specified")
		return
	}

	if err := checkDate(&task); err != nil {
		writeError(w, err.Error())
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeError(w, "Error adding task to database: "+err.Error())
		return
	}

	writeJSON(w, IDResponse{ID: strconv.FormatInt(id, 10)}, http.StatusOK)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		writeError(w, "id is empty")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == fmt.Sprintf("ID task %s dont found", id) {
			writeError(w, "task not found")
			return
		}
		writeError(w, "error get task from DB:"+err.Error())
		return
	}
	writeJSON(w, task, http.StatusOK)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, "JSON deserialization error"+err.Error())
		return
	}

	if len(task.ID) == 0 {
		writeError(w, "no task ID specified for update")
		return
	}

	if len(task.Title) == 0 {
		writeError(w, "title is empty")
		return
	}

	if err := checkDate(&task); err != nil {
		writeError(w, err.Error())
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		if err.Error() == fmt.Sprintf("ID task %s not found", task.ID) {
			writeError(w, "task not found")
			return
		}
		writeError(w, "error update task in DB"+err.Error())
		return
	}
	writeJSON(w, EmptyResponse{}, http.StatusOK)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		writeError(w, "id is empty")
		return
	}

	if err := db.DeleteTask(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, "task not found")
			return
		}
		writeError(w, "error DELETE task in DB:"+err.Error())
		return
	}
	writeJSON(w, EmptyResponse{}, http.StatusOK)
}

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed")
		return
	}

	id := r.URL.Query().Get("id")

	if id == "" {
		writeError(w, "id is empty")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, "task nor found")
			return
		}
		writeError(w, "error GET task in DB: "+err.Error())
		return
	}

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, "task not found")
				return
			}
			writeError(w, "error DELETE task in DB:"+err.Error())
			return
		}
	} else {

		taskTime, err := time.Parse(dateFormat, task.Date)
		if err != nil {
			writeError(w, "internal error: failed to parse task date: "+err.Error())
			return
		}

		nextDateStr, err := nextdate.NextDate(taskTime, task.Date, task.Repeat)
		if err != nil {
			writeError(w, "error calculate nextdate: "+err.Error())
			return
		}

		if err := db.UpdateDate(id, nextDateStr); err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, "task not found")
				return
			}
			writeError(w, "error UPDATE task in DB: "+err.Error())
			return
		}
	}

	writeJSON(w, EmptyResponse{}, http.StatusOK)

}

func writeJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, errMsg string) {
	writeJSON(w, ErrorResponse{Error: errMsg}, http.StatusBadRequest)
}

func checkDate(task *db.Task) error {
	now := time.Now()
	taskDate := task.Date

	if len(taskDate) == 0 {
		task.Date = now.Format("20060102")
		return nil
	}

	t, err := time.Parse(dateFormat, taskDate)
	if err != nil {
		return errors.New("date is presented in a format other than 20060102")
	}

	nextDateStr := ""
	if len(task.Repeat) > 0 {
		nextDateStr, err = nextdate.NextDate(now, taskDate, task.Repeat)
		if err != nil {
			return errors.New("repetition rule is specified in the wrong format")
		}
	}

	if nextdate.AfterNow(now, t) {
		if len(task.Repeat) == 0 {
			task.Date = now.Format("20060102")
		} else {
			task.Date = nextDateStr
		}
	}

	return nil
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type IDResponse struct {
	ID string `json:"id"`
}

type EmptyResponse struct{}
