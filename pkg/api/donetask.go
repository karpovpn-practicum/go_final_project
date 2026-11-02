package api

import (
	"net/http"
	"time"

	"go_final_project/pkg/db"
)

// doneTaskHandler отмечает задачу выполненной: одноразовую удаляет,
// периодическую переносит на следующую дату по правилу повторения.
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeDBError(w, "удалить задачу", err)
			return
		}
		writeJSON(w, http.StatusOK, struct{}{})
		return
	}

	next, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := db.UpdateDate(next, id); err != nil {
		writeDBError(w, "обновить дату задачи", err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if err := checkID(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeDBError(w, "удалить задачу", err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
