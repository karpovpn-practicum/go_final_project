package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE scheduler (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    date    CHAR(8) NOT NULL DEFAULT "",
    title   VARCHAR(256) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat  VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX scheduler_date ON scheduler (date);
`

var db *sql.DB

// Init открывает базу данных dbFile. Если файла нет, создаёт его
// вместе с таблицей scheduler и индексом по дате.
func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("проверка файла БД %s: %w", dbFile, err)
	}
	install := err != nil

	conn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("открытие БД %s: %w", dbFile, err)
	}
	if err = conn.Ping(); err != nil {
		conn.Close()
		return fmt.Errorf("подключение к БД %s: %w", dbFile, err)
	}

	if install {
		if _, err = conn.Exec(schema); err != nil {
			conn.Close()
			os.Remove(dbFile)
			return fmt.Errorf("создание схемы БД: %w", err)
		}
	}

	db = conn
	return nil
}

// Close закрывает соединение с базой данных.
func Close() error {
	if db == nil {
		return nil
	}
	return db.Close()
}
