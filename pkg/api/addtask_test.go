package api

import (
	"testing"
	"time"

	"go_final_project/pkg/db"
)

func TestCheckTask(t *testing.T) {
	now := time.Date(2024, 1, 26, 15, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		task     db.Task
		wantDate string
		wantErr  bool
	}{
		{"пустая дата — сегодня", db.Task{Title: "a"}, "20240126", false},
		{"сегодня с повтором не сдвигается", db.Task{Title: "a", Date: "20240126", Repeat: "d 1"}, "20240126", false},
		{"будущая дата не меняется", db.Task{Title: "a", Date: "20240301", Repeat: "y"}, "20240301", false},
		{"прошлая без повтора — сегодня", db.Task{Title: "a", Date: "20231220"}, "20240126", false},
		{"прошлая с повтором — следующая", db.Task{Title: "a", Date: "20240108", Repeat: "d 10"}, "20240128", false},
		{"без заголовка", db.Task{Date: "20240126"}, "", true},
		{"заголовок из пробелов", db.Task{Title: "  "}, "", true},
		{"неверный формат даты", db.Task{Title: "a", Date: "26.01.2024"}, "", true},
		{"несуществующая дата", db.Task{Title: "a", Date: "20240192"}, "", true},
		{"неверное правило", db.Task{Title: "a", Date: "20240212", Repeat: "ooops"}, "", true},
		{"неверное правило у будущей даты", db.Task{Title: "a", Date: "20250101", Repeat: "w"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := tt.task
			err := checkTask(&task, now)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ожидалась ошибка, дата %q", task.Date)
				}
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if task.Date != tt.wantDate {
				t.Errorf("дата %q, ожидалась %q", task.Date, tt.wantDate)
			}
		})
	}
}
