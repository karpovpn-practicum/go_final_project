package main

import (
	"log"
	"os"
	// Встроенная база часовых поясов: в Docker-образе нет /usr/share/zoneinfo,
	// без неё TZ=Europe/Moscow игнорируется и «сегодня» считается по UTC.
	_ "time/tzdata"

	"go_final_project/pkg/db"
	"go_final_project/pkg/server"
)

const defaultDBFile = "scheduler.db"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dbFile := defaultDBFile
	if envFile := os.Getenv("TODO_DBFILE"); envFile != "" {
		dbFile = envFile
	}

	if err := db.Init(dbFile); err != nil {
		return err
	}
	defer db.Close()

	return server.Run()
}
