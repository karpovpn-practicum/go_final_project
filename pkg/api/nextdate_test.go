package api

import (
	"testing"
	"time"
)

func TestNextDate(t *testing.T) {
	now := time.Date(2024, 1, 26, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		dstart, repeat, want string
	}{
		// Примеры из задания.
		{"20240229", "y", "20250301"},
		{"20240113", "d 7", "20240127"},
		{"20240116", "m 16,5", "20240205"},
		{"20240201", "m -1,18", "20240218"},
		// Дата в будущем: d и y всё равно делают шаг.
		{"20240228", "d 1", "20240229"},
		{"20240301", "y", "20250301"},
		// 29 февраля — ближайшее только через 4 года, 30 февраля не бывает.
		{"20240301", "m 29 2", "20280229"},
		{"20240126", "w 5", "20240202"},
		// Ошибки.
		{"20240126", "", ""},
		{"20240126", "l 34", ""},
		{"20240126", "d", ""},
		{"20240126", "d 0", ""},
		{"20240126", "d 405", ""},
		{"20240126", "d 7 7", ""},
		{"20240126", "y 1", ""},
		{"20240126", "w", ""},
		{"20240126", "w 8", ""},
		{"20240126", "w 1,", ""},
		{"20240126", "m 4,40", ""},
		{"20240126", "m 5 1,13", ""},
		{"20240126", "m -3", ""},
		{"20240126", "m 30 2", ""},
		{"20240126", "m 1 2 3", ""},
		{"2024013", "d 1", ""},
		{"20240230", "d 1", ""},
	}

	for _, tt := range tests {
		got, err := NextDate(now, tt.dstart, tt.repeat)
		if tt.want == "" {
			if err == nil {
				t.Errorf("NextDate(%q, %q) = %q, ожидалась ошибка", tt.dstart, tt.repeat, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("NextDate(%q, %q): неожиданная ошибка: %v", tt.dstart, tt.repeat, err)
			continue
		}
		if got != tt.want {
			t.Errorf("NextDate(%q, %q) = %q, ожидалось %q", tt.dstart, tt.repeat, got, tt.want)
		}
	}
}

func TestNextDateIgnoresTimeOfDay(t *testing.T) {
	// Вечер 26 января по Москве — это всё ещё 26-е, задача на 27-е должна остаться 27-м.
	msk := time.FixedZone("MSK", 3*60*60)
	now := time.Date(2024, 1, 26, 23, 30, 0, 0, msk)

	got, err := NextDate(now, "20240120", "d 7")
	if err != nil {
		t.Fatal(err)
	}
	if got != "20240127" {
		t.Errorf("got %q, want 20240127", got)
	}
}
