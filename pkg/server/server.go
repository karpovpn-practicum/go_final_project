package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"go_final_project/pkg/api"
)

const (
	defaultPort = 7540
	webDir      = "./web"
)

// port возвращает порт из переменной окружения TODO_PORT или порт по умолчанию.
func port() int {
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 && p < 65536 {
			return p
		}
		log.Printf("некорректное значение TODO_PORT=%q, используется порт %d", envPort, defaultPort)
	}
	return defaultPort
}

// Run запускает веб-сервер, который отдаёт файлы фронтенда.
func Run() error {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	api.Init(mux)

	addr := fmt.Sprintf(":%d", port())
	log.Printf("сервер запущен на %s", addr)
	return http.ListenAndServe(addr, mux)
}
