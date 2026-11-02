package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"go_final_project/pkg/db"
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
	}
}

// decodeTask читает задачу из тела запроса. При ошибке сам отвечает клиенту
// и возвращает false.
func decodeTask(w http.ResponseWriter, r *http.Request) (db.Task, bool) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON: "+err.Error())
		return task, false
	}
	return task, true
}

// checkID проверяет, что идентификатор задачи указан и является числом.
func checkID(id string) error {
	if id == "" {
		return errors.New("не указан идентификатор задачи")
	}
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		return errors.New("некорректный идентификатор задачи")
	}
	return nil
}

// writeDBError отвечает клиенту на ошибку из пакета db: 404 для
// отсутствующей задачи, 500 для остальных с записью подробностей в лог.
func writeDBError(w http.ResponseWriter, action string, err error) {
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	log.Printf("%s: %v", action, err)
	writeError(w, http.StatusInternalServerError, "не удалось "+action)
}
