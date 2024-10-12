package database

import (
	"database/sql"
	"errors"
	moduls "final-project/internal/moduls"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // Импорт драйвера SQLite
)

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
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Fatalf("Файл базы данных не существует: %s", dbFile)
	}
	if db != nil {
		return db
	}
	// Настройка пула соединений
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Проверка соединения с базой данных
	if err := db.Ping(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	log.Printf("Успешное подключение к базе данных: %s", dbFile)
	return db
}

// Функция для проверки соединения с базой данных
func TestDatabaseConnection(db *sql.DB) error {
	// Выполняем простой запрос
	_, err := db.Exec("SELECT 1")
	if err != nil {
		log.Printf("Ошибка выполнения тестового запроса: %v", err)
		return fmt.Errorf("ошибка выполнения тестового запроса: %w", err)
	}
	log.Println("Тестовый запрос к базе данных выполнен успешно")
	return nil
}

// SearchTasksByTitle ищет задачи по названию
func Searchtitl(db *sql.DB, search string) ([]moduls.Scheduler, error) {
	var tasks []moduls.Scheduler

	//запрос на поиск задач
	query := `SELECT id, date, title, comment, repeat 
		FROM scheduler 
		WHERE title LIKE :search OR comment LIKE :search 
		ORDER BY date 
		LIMIT 50
	`
	//подставляем поисковый запрос
	search = fmt.Sprintf("%%%s%%", search)
	rows, err := db.Query(query, sql.Named("search", search))

	//проверка на ошибки
	if err != nil {
		return []moduls.Scheduler{}, err
	}
	defer rows.Close()

	//считываем задачи построчно
	for rows.Next() {
		var task moduls.Scheduler
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return []moduls.Scheduler{}, err
		}
		tasks = append(tasks, task)
	}

	//проверка на ошибки
	if err := rows.Err(); err != nil {
		return []moduls.Scheduler{}, err
	}

	//проверка на пустой массив
	if tasks == nil {
		tasks = []moduls.Scheduler{}
	}
	return tasks, nil
}

// SearchTasksByDate ищет задачи по дате
func SearchDate(db *sql.DB, date string) ([]moduls.Scheduler, error) {
	var tasks []moduls.Scheduler

	//считываем задачи по дате
	rows, err := db.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = :date LIMIT 50",
		sql.Named("date", date))
	if err != nil {
		return []moduls.Scheduler{}, err
	}
	defer rows.Close()

	//считываем задачи построчно
	for rows.Next() {
		var task moduls.Scheduler
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return []moduls.Scheduler{}, err
		}
		tasks = append(tasks, task)
	}

	//проверка на ошибки
	if err := rows.Err(); err != nil {
		return []moduls.Scheduler{}, err
	}

	//проверка на пустой массив
	if tasks == nil {
		tasks = []moduls.Scheduler{}
	}
	return tasks, nil

}

// Функция для получения задачи по ID
func GetpoID(id string) (moduls.Scheduler, error) {
	var task moduls.Scheduler
	db := InitDatabase()
	defer db.Close()
	//считываем задачу по id
	//row := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id",
	//	sql.Named("id", id))
	//if err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
	//	return moduls.Scheduler{}, err
	//}
	row := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err == sql.ErrNoRows {
		return moduls.Scheduler{}, fmt.Errorf("задача с ID %s не найдена", id)
	}
	if err != nil {
		return moduls.Scheduler{}, fmt.Errorf("ошибка при получении задачи: %v", err)
	}
	return task, nil
}

// Функция для создания задачи
func Create(task *moduls.Scheduler) (int, error) {
	db := InitDatabase()
	if db == nil {
		return 0, errors.New("database not initialized")
	}
	defer db.Close()

	// Вставляем задачу в таблицу
	result, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		return 0, err
	}

	// Получаем ID последней вставленной записи
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Функция для обновления задачи
func Update(db *sql.DB, task *moduls.Scheduler) (moduls.Scheduler, error) {
	//var updatedTask moduls.Scheduler

	// Обновляем задачу в таблице
	result, err := db.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
		sql.Named("id", task.ID),
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		return moduls.Scheduler{}, err
	}

	// Получаем количество измененных строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return moduls.Scheduler{}, err
	}

	// Проверка на успешное обновление
	if rowsAffected == 0 {
		return moduls.Scheduler{}, errors.New("failed to update")
	}

	return *task, nil
}

// Функция для удаления задачи
func Delete(id string) error {
	db := InitDatabase()
	defer db.Close()

	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("ошибка при удалении задачи: %v", err)
	}

	// Получаем количество удаленных строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при получении количества затронутых строк: %v", err)
	}

	// Проверка на успешное удаление
	if rowsAffected == 0 {
		return nil //errors.New("failed to delete")
	}

	//возвращаем удаленную задачу
	return nil
}

// ReadTask читает все задачи из базы данных
func ReadTask(db *sql.DB) ([]moduls.Scheduler, error) {
	var tasks []moduls.Scheduler
	//считываем все задачи
	rows, err := db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date")
	if err != nil {
		return []moduls.Scheduler{}, err
	}
	defer rows.Close()
	//считываем задачи построчно
	for rows.Next() {
		var task moduls.Scheduler
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return []moduls.Scheduler{}, err
		}
		tasks = append(tasks, task)
	}
	//проверка на ошибки
	if err := rows.Err(); err != nil {
		return []moduls.Scheduler{}, err
	}
	//проверка на пустой массив
	if tasks == nil {
		tasks = []moduls.Scheduler{}
	}
	return tasks, nil
}
