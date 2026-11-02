package api

import "net/http"

// DateFormat — формат хранения и передачи дат задач.
const DateFormat = "20060102"

// Init регистрирует обработчики API.
func Init(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/nextdate", nextDateHandler)
	mux.HandleFunc("POST /api/signin", signinHandler)
	mux.HandleFunc("/api/task", auth(taskHandler))
	mux.HandleFunc("POST /api/task/done", auth(doneTaskHandler))
	mux.HandleFunc("GET /api/tasks", auth(tasksHandler))
}
