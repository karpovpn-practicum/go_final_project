package db

import (
	"database/sql"
	"errors"
	"strings"
)

// ErrNotFound возвращается, если задачи с указанным идентификатором нет.
var ErrNotFound = errors.New("задача не найдена")

// Task — задача планировщика.
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет задачу в таблицу scheduler и возвращает её идентификатор.
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetTask возвращает задачу по идентификатору.
func GetTask(id string) (*Task, error) {
	var t Task
	err := db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id`,
		sql.Named("id", id)).
		Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTask сохраняет все поля задачи с идентификатором task.ID.
func UpdateTask(task *Task) error {
	res, err := db.Exec(`UPDATE scheduler
		SET date = :date, title = :title, comment = :comment, repeat = :repeat
		WHERE id = :id`,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.ID))
	if err != nil {
		return err
	}
	return checkAffected(res)
}

// UpdateDate меняет дату задачи с идентификатором id.
func UpdateDate(next string, id string) error {
	res, err := db.Exec(`UPDATE scheduler SET date = :date WHERE id = :id`,
		sql.Named("date", next),
		sql.Named("id", id))
	if err != nil {
		return err
	}
	return checkAffected(res)
}

// DeleteTask удаляет задачу с идентификатором id.
func DeleteTask(id string) error {
	res, err := db.Exec(`DELETE FROM scheduler WHERE id = :id`, sql.Named("id", id))
	if err != nil {
		return err
	}
	return checkAffected(res)
}

// checkAffected возвращает ErrNotFound, если запрос не затронул ни одной строки.
func checkAffected(res sql.Result) error {
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

// Tasks возвращает ближайшие задачи, отсортированные по дате.
func Tasks(limit int) ([]*Task, error) {
	return queryTasks(`SELECT id, date, title, comment, repeat FROM scheduler
		ORDER BY date, id LIMIT :limit`,
		sql.Named("limit", limit))
}

// TasksByDate возвращает задачи на дату date в формате 20060102.
func TasksByDate(date string, limit int) ([]*Task, error) {
	return queryTasks(`SELECT id, date, title, comment, repeat FROM scheduler
		WHERE date = :date ORDER BY id LIMIT :limit`,
		sql.Named("date", date),
		sql.Named("limit", limit))
}

// SearchTasks возвращает задачи, в заголовке или комментарии которых
// встречается подстрока text. Символы % и _ в text ищутся буквально.
func SearchTasks(text string, limit int) ([]*Task, error) {
	return queryTasks(`SELECT id, date, title, comment, repeat FROM scheduler
		WHERE title LIKE :search ESCAPE '\' OR comment LIKE :search ESCAPE '\'
		ORDER BY date, id LIMIT :limit`,
		sql.Named("search", "%"+likeEscaper.Replace(text)+"%"),
		sql.Named("limit", limit))
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func queryTasks(query string, args ...any) ([]*Task, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*Task, 0)
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, rows.Err()
}
