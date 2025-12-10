package api

import (
	"net/http"

	"finalsprint/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	const limit = 50

	search := r.URL.Query().Get("search")

	tasks, err := db.Tasks(limit, search)
	if err != nil {
		writeError(w, "Error retrieving tasks from the database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}

	writeJSON(w, TasksResp{
		Tasks: tasks,
	}, http.StatusOK)
}
