package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3" // Импорт драйвера SQLite
)

// Определение типа Task
type Scheduler struct {
	ID      int
	Date    string
	Title   string
	Comment string
	Repeat  string
}

// Функция для инициализации базы данных
func InitDatabase() *sql.DB {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		log.Fatal("Не задан файл базы данных")
	}
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}
	return db
}

// Функция для получения задач с возможностью фильтрации
func GetTasks(db *sql.DB, titleFilter string, dateFilter string) ([]Scheduler, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE 1=1`
	args := []interface{}{}

	// Фильтрация по заголовку
	if titleFilter != "" {
		query += " AND title LIKE ?"
		args = append(args, "%"+titleFilter+"%")
	}

	// Фильтрация по дате
	if dateFilter != "" {
		query += " AND date = ?"
		args = append(args, dateFilter)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения задач: %w", err)
	}
	defer rows.Close()

	var tasks []Scheduler
	for rows.Next() {
		var task Scheduler
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("ошибка сканирования задачи: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return tasks, nil
}

// Функция для получения задачи по ID
func GetpoID(db *sql.DB, id int) (Scheduler, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	row := db.QueryRow(query, id)
	var task Scheduler
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return Scheduler{}, fmt.Errorf("ошибка получения задачи: %w", err)
	}
	return task, nil
}

// Функция для создания задачи
func Create(db *sql.DB, task *Scheduler) error {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("ошибка подготовки запроса: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("ошибка получения ID новой задачи: %w", err)
	}

	task.ID = int(id)
	return nil
}

// Функция для обновления задачи
func Update(db *sql.DB, task *Scheduler) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	stmt, err := db.Prepare(query) //подготовка запроса
	if err != nil {
		return fmt.Errorf("ошибка подготовки запроса: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	return nil
}

// Функция для удаления задачи
func Delete(db *sql.DB, id int) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("ошибка подготовки запроса: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	return nil
}
