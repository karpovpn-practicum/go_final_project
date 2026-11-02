package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// writeJSON сериализует data в JSON и отправляет его с кодом status.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("запись JSON-ответа: %v", err)
	}
}

// writeError отправляет JSON вида {"error": "<msg>"}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
