package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go_final_project/pkg/db"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, ok := decodeTask(w, r)
	if !ok {
		return
	}

	if err := checkTask(&task, time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeDBError(w, "добавить задачу", err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": strconv.FormatInt(id, 10)})
}

// checkTask проверяет поля задачи и приводит дату к актуальной:
// пустая дата заменяется на сегодня, прошедшая — на сегодня
// или на следующую дату по правилу повторения.
func checkTask(task *db.Task, now time.Time) error {
	if strings.TrimSpace(task.Title) == "" {
		return errors.New("не указан заголовок задачи")
	}
	return checkDate(task, now)
}

func checkDate(task *db.Task, now time.Time) error {
	today := now.Format(DateFormat)
	if task.Date == "" {
		task.Date = today
	}

	t, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return errors.New("дата должна быть в формате " + DateFormat)
	}

	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	if afterNow(now, t) {
		if task.Repeat == "" {
			task.Date = today
		} else {
			task.Date = next
		}
	}
	return nil
}
