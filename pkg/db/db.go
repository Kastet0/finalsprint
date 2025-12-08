package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const (
	schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT,
    repeat VARCHAR(128) DEFAULT ''
);
`
	dateFormat = "20060102"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	dbConn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	if install {
		log.Printf("Creating new db: %s", dbFile)

		if _, err := dbConn.Exec(schema); err != nil {
			dbConn.Close()
			return fmt.Errorf("failed to create schema: %v", err)
		}

		log.Println("Database created")
	} else {
		log.Println("Database exists")
	}

	DB = dbConn

	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func AddTask(task *Task) (int64, error) {
	if DB == nil {
		return 0, sql.ErrConnDone
	}

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`

	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	return id, err
}

func Tasks(limit int, search string) ([]*Task, error) {
	if DB == nil {
		return nil, sql.ErrConnDone
	}

	var query string
	var args []any

	dateToFind, err := time.Parse("02.01.2006", search)

	if search == "" {
		query = `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`
		args = []any{limit}
	} else if err == nil {
		dateStr := dateToFind.Format(dateFormat)
		query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?`
		args = []any{dateStr, limit}
	} else {
		searchPattern := "%" + search + "%"
		query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`
		args = []any{searchPattern, searchPattern, limit}
	}
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %v", err)
	}
	defer rows.Close()

	var tasks []*Task

	for rows.Next() {
		task := &Task{}
		var id int64
		var date, title, comment, repeat string

		err = rows.Scan(&id, &date, &title, &comment, &repeat)
		if err != nil {
			log.Printf("error scaning row: %v", err)
			continue
		}

		task.ID = fmt.Sprintf("%d", id)
		task.Date = date
		task.Title = title
		task.Comment = comment
		task.Repeat = repeat

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %v", err)
	}

	return tasks, nil
}

func GetTask(id string) (*Task, error) {
	if DB == nil {
		return nil, sql.ErrConnDone
	}

	task := &Task{}
	var dbID int64
	var date, title, comment, repeat string

	row := DB.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id)

	err := row.Scan(&dbID, &date, &title, &comment, &repeat)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ID %snot found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("scan error %v", err)
	}
	task.ID = fmt.Sprintf("%d", dbID)
	task.Date = date
	task.Title = title
	task.Comment = comment
	task.Repeat = repeat

	return task, nil
}

func UpdateTask(task *Task) error {
	if DB == nil {
		return sql.ErrConnDone
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`

	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)

	if err != nil {
		return fmt.Errorf("runtime error UPDATE: %v", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error check RowsAffected: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("ID %s not found", task.ID)
	}
	if count > 1 {
		log.Printf("updated: %d tasks for ID %s", count, task.ID)
	}

	return nil
}

func DeleteTask(id string) error {
	if DB == nil {
		return sql.ErrConnDone
	}

	res, err := DB.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("runtime error DELETE: %v", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error check RowsAffected: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("ID %s not found", id)
	}
	return nil
}

func UpdateDate(id string, nextDate string) error {
	if DB == nil {
		return sql.ErrConnDone
	}
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := DB.Exec(query, nextDate, id)

	if err != nil {
		return fmt.Errorf("runtime error UPDATE DATE: %v", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error check RowsAffectes: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("ID %s not found", id)
	}

	return nil
}
