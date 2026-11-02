package api

import (
	"net/http"
	"time"

	"go_final_project/pkg/db"
)

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if err := checkID(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeDBError(w, "получить задачу", err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, ok := decodeTask(w, r)
	if !ok {
		return
	}
	if err := checkID(task.ID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := checkTask(&task, time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		writeDBError(w, "обновить задачу", err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
