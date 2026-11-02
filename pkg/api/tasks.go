package api

import (
	"net/http"
	"strings"
	"time"

	"go_final_project/pkg/db"
)

const (
	tasksLimit   = 50
	searchFormat = "02.01.2006"
)

type tasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	var (
		tasks []*db.Task
		err   error
	)

	search := strings.TrimSpace(r.FormValue("search"))
	if search == "" {
		tasks, err = db.Tasks(tasksLimit)
	} else if date, perr := time.Parse(searchFormat, search); perr == nil {
		tasks, err = db.TasksByDate(date.Format(DateFormat), tasksLimit)
	} else {
		tasks, err = db.SearchTasks(search, tasksLimit)
	}
	if err != nil {
		writeDBError(w, "получить список задач", err)
		return
	}

	writeJSON(w, http.StatusOK, tasksResp{Tasks: tasks})
}
