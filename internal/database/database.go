package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // Импорт драйвера SQLite
)

// Определение типа Task
type Scheduler struct {
	ID          int
	Name        string
	Description string
	Completed   bool
	Date        string
	Title       string
	Comment     string
	Repeat      string
}

// Функция для инициализации базы данных
func InitDatabase() *sql.DB {
	// Получение пути к файлу базы данных из переменной окружения
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		// Если переменная окружения не задана, используем путь по умолчанию
		appPath, err := os.Executable()
		if err != nil {
			log.Fatalf("Ошибка получения пути к файлу: %v", err)
		}
		dbFile = filepath.Join(filepath.Dir(appPath), "scheduler.db")
	}

	// Проверка существования файла базы данных
	_, err := os.Stat(dbFile)
	var install bool
	if err != nil {
		install = true
		dbFile = "file:scheduler.db?cache=shared&mode=rwc"
	}

	// Открытие подключения к базе данных
	//db, err := sql.Open("sqlite3", dbFile)
	db, err := sql.Open("sqlite3", "file:scheduler.db?cache=shared&mode=rwc")
	//db, err := sql.Open("sqlite3", dbFile+"?_foreign_keys=on&_journal_mode=WAL&_synchronous=NORMAL&_charset=utf8")
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}

	// Проверка соединения с базой данных
	if err := db.Ping(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Если файл базы данных отсутствует, создаем таблицу
	if install {
		// Создаем таблицу scheduler
		_, err = db.Exec(`
			CREATE TABLE scheduler (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				date TEXT NOT NULL,
				title TEXT NOT NULL,
				comment TEXT,
				repeat TEXT
			)
		`)
		if err != nil {
			log.Fatalf("Ошибка создания таблицы: %v", err)
		}

		// Создаем индекс по полю date
		_, err = db.Exec(`
			CREATE INDEX date_idx ON scheduler (date)
		`)
		if err != nil {
			log.Fatalf("Ошибка создания индекса: %v", err)
		}

		fmt.Println("База данных создана")
	} else {
		fmt.Println("База данных уже существует")
	}
	return db
}

// Функция для закрытия соединения с базой данных
func CloseDb(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("Ошибка закрытия соединения с базой данных: %v", err)
	}
}

// Функция для получения задач
func GetTasks(db *sql.DB, filters map[string]string) ([]Scheduler, error) {
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE 1=1"
	var args []interface{}

	// Добавляем фильтры в запрос
	if date, ok := filters["date"]; ok {
		query += " AND date = ?"
		args = append(args, date)
	}
	if search, ok := filters["search"]; ok {
		query += " AND (title LIKE ? OR comment LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	// Добавляем сортировку и лимит
	query += " ORDER BY date ASC LIMIT 50"

	// Подготовка запроса
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка подготовки запроса: %w", err)
	}
	defer stmt.Close()

	// Выполнение запроса
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var tasks []Scheduler
	for rows.Next() {
		var t Scheduler
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}
	return tasks, nil
}

// Функция для получения задачи по ID
func GetpoID(db *sql.DB, id int) (Scheduler, error) {
	query := `SELECT id, date, title, comment, repeat 
        FROM scheduler 
        WHERE id = ?`
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
	stmt, err := db.Prepare(query)
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

// Функция для поиска задач
func Search(db *sql.DB, search string) ([]Scheduler, error) {
	query := `SELECT id, date, title, comment, repeat 
        FROM scheduler 
        WHERE title LIKE ? OR comment LIKE ? 
        ORDER BY date ASC
        LIMIT 50`
	search = "%" + search + "%"
	rows, err := db.Query(query, search, search)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var tasks []Scheduler
	for rows.Next() {
		var task Scheduler
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	return tasks, nil
}

// Функция для поиска задач по дате
func SearchDate(db *sql.DB, date string) ([]Scheduler, error) {
	query := `SELECT id, date, title, comment, repeat 
        FROM scheduler 
        WHERE date = ? 
        ORDER BY date ASC
        LIMIT 10`
	rows, err := db.Query(query, date)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var tasks []Scheduler
	for rows.Next() {
		var task Scheduler
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	return tasks, nil
}
