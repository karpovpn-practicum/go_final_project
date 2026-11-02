package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	maxDayInterval = 400
	// Сколько лет вперёд ищем дату для правил w и m. Худший реальный случай —
	// 29 февраля: между високосными годами бывает до 8 лет (2096 → 2104).
	maxSearchYears = 10
)

// NextDate возвращает следующую после now дату задачи, начинающейся с dstart
// и повторяющейся по правилу repeat.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("некорректная дата %q", dstart)
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("не указано правило повторения")
	}

	var next time.Time
	switch parts[0] {
	case "d":
		next, err = nextByDays(now, date, parts[1:])
	case "y":
		next, err = nextByYear(now, date, parts[1:])
	case "w":
		next, err = nextByWeekdays(now, date, parts[1:])
	case "m":
		next, err = nextByMonthDays(now, date, parts[1:])
	default:
		err = errors.New("неизвестный тип правила")
	}
	if err != nil {
		return "", fmt.Errorf("правило повторения %q: %w", repeat, err)
	}
	return next.Format(DateFormat), nil
}

// nextByDays обрабатывает правило "d <число>".
func nextByDays(now, date time.Time, args []string) (time.Time, error) {
	if len(args) != 1 {
		return time.Time{}, errors.New("нужно указать один интервал в днях")
	}
	interval, err := strconv.Atoi(args[0])
	if err != nil || interval < 1 || interval > maxDayInterval {
		return time.Time{}, fmt.Errorf("интервал должен быть числом от 1 до %d", maxDayInterval)
	}
	for {
		date = date.AddDate(0, 0, interval)
		if afterNow(date, now) {
			return date, nil
		}
	}
}

// nextByYear обрабатывает правило "y".
func nextByYear(now, date time.Time, args []string) (time.Time, error) {
	if len(args) != 0 {
		return time.Time{}, errors.New("правило не принимает параметров")
	}
	for {
		date = date.AddDate(1, 0, 0)
		if afterNow(date, now) {
			return date, nil
		}
	}
}

// nextByWeekdays обрабатывает правило "w <дни недели через запятую>".
func nextByWeekdays(now, date time.Time, args []string) (time.Time, error) {
	if len(args) != 1 {
		return time.Time{}, errors.New("нужно указать дни недели")
	}
	var weekdays [8]bool
	err := parseList(args[0], func(n int) bool { return n >= 1 && n <= 7 }, func(n int) {
		weekdays[n] = true
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("дни недели: %w", err)
	}

	return nextMatching(now, date, func(d time.Time) bool {
		wd := int(d.Weekday())
		if wd == 0 {
			wd = 7
		}
		return weekdays[wd]
	})
}

// nextByMonthDays обрабатывает правило "m <дни месяца> [месяцы]".
func nextByMonthDays(now, date time.Time, args []string) (time.Time, error) {
	if len(args) != 1 && len(args) != 2 {
		return time.Time{}, errors.New("нужно указать дни месяца и, опционально, месяцы")
	}

	var days [32]bool
	var lastDay, preLastDay bool
	err := parseList(args[0], func(n int) bool { return (n >= 1 && n <= 31) || n == -1 || n == -2 }, func(n int) {
		switch n {
		case -1:
			lastDay = true
		case -2:
			preLastDay = true
		default:
			days[n] = true
		}
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("дни месяца: %w", err)
	}

	var months [13]bool
	if len(args) == 2 {
		err = parseList(args[1], func(n int) bool { return n >= 1 && n <= 12 }, func(n int) {
			months[n] = true
		})
		if err != nil {
			return time.Time{}, fmt.Errorf("месяцы: %w", err)
		}
	} else {
		for i := range months {
			months[i] = true
		}
	}

	return nextMatching(now, date, func(d time.Time) bool {
		if !months[d.Month()] {
			return false
		}
		last := daysInMonth(d)
		return days[d.Day()] ||
			(lastDay && d.Day() == last) ||
			(preLastDay && d.Day() == last-1)
	})
}

// nextMatching перебирает дни после max(date, now) и возвращает первый,
// для которого match вернул true.
func nextMatching(now, date time.Time, match func(time.Time) bool) (time.Time, error) {
	start := dayOf(date)
	if today := dayOf(now); today.After(start) {
		start = today
	}
	limit := start.AddDate(maxSearchYears, 0, 0)
	for d := start.AddDate(0, 0, 1); !d.After(limit); d = d.AddDate(0, 0, 1) {
		if match(d) {
			return d, nil
		}
	}
	return time.Time{}, errors.New("нет подходящей даты")
}

// parseList разбирает список чисел через запятую, проверяет каждое
// через valid и передаёт в set.
func parseList(s string, valid func(int) bool, set func(int)) error {
	for _, item := range strings.Split(s, ",") {
		n, err := strconv.Atoi(item)
		if err != nil || !valid(n) {
			return fmt.Errorf("недопустимое значение %q", item)
		}
		set(n)
	}
	return nil
}

// afterNow сообщает, что дата date позже дня now, без учёта времени суток.
func afterNow(date, now time.Time) bool {
	return dayOf(date).After(dayOf(now))
}

// dayOf отбрасывает время суток и часовой пояс, оставляя только дату.
func dayOf(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func daysInMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if s := r.FormValue("now"); s != "" {
		t, err := time.Parse(DateFormat, s)
		if err != nil {
			http.Error(w, fmt.Sprintf("некорректная дата now %q", s), http.StatusBadRequest)
			return
		}
		now = t
	}

	next, err := NextDate(now, r.FormValue("date"), r.FormValue("repeat"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(next))
}
