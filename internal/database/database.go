package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"final-project/internal/cache"
	moduls "final-project/internal/moduls"

	_ "github.com/mattn/go-sqlite3"
)

// DB представляет структуру базы данных с кэшем
type DB struct {
	*sql.DB
	cache *cache.Cache
	mu    sync.RWMutex
}

var (
	dbInstance *DB
	once       sync.Once
)

// InitDatabase инициализирует подключение к базе данных
func InitDatabase() *DB {
	once.Do(func() {
		dbFile := os.Getenv("TODO_DBFILE")
		if dbFile == "" {
			log.Fatal("Не задан файл базы данных")
		}

		db, err := sql.Open("sqlite3", dbFile)
		if err != nil {
			log.Fatalf("Ошибка открытия базы данных: %v", err)
		}

		// Настройка пула соединений
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		// Создание индексов
		if err := createIndexes(db); err != nil {
			log.Printf("Ошибка создания индексов: %v", err)
		}

		// Создание кэша
		dbInstance = &DB{
			DB:    db,
			cache: cache.NewCache(),
		}
	})

	return dbInstance
}

// createIndexes создает необходимые индексы
func createIndexes(db *sql.DB) error {
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date)`,
		`CREATE INDEX IF NOT EXISTS idx_title ON scheduler(title)`,
		`CREATE INDEX IF NOT EXISTS idx_repeat ON scheduler(repeat)`,
	}

	// Создание индексов
	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return fmt.Errorf("ошибка создания индекса: %v", err)
		}
	}
	return nil
}

// ReadTask читает задачи с использованием кэша
func (db *DB) ReadTask(date string) ([]moduls.Scheduler, error) {
	cacheKey := fmt.Sprintf("tasks_%s", date)

	// Проверяем кэш
	if cached, ok := db.cache.Get(cacheKey); ok {
		return cached.([]moduls.Scheduler), nil
	}

	// Если нет в кэше, читаем из БД
	var tasks []moduls.Scheduler
	var rows *sql.Rows
	var err error

	// Проверяем, есть ли дата в запросе
	if date != "" {
		rows, err = db.Query(`
			SELECT id, date, title, comment, repeat 
			FROM scheduler 
			WHERE date = ? 
			ORDER BY date
		`, date)
	} else {
		rows, err = db.Query(`
			SELECT id, date, title, comment, repeat 
			FROM scheduler 
			ORDER BY date
		`)
	}

	// Если есть ошибка, возвращаем ее
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к базе данных: %w", err)
	}
	defer rows.Close()

	// Сканируем строки
	for rows.Next() {
		var task moduls.Scheduler
		// Сканируем строки
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		tasks = append(tasks, task)
	}

	// Сохраняем в кэш на 5 минут
	db.cache.Set(cacheKey, tasks, 5*time.Minute)

	return tasks, nil
}

// Create добавляет новую задачу с инвалидацией кэша
func (db *DB) Create(task *moduls.Scheduler) (int, error) {
	result, err := db.Exec(`
		INSERT INTO scheduler (date, title, comment, repeat) 
		VALUES (?, ?, ?, ?)
	`, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	// Получаем ID последней вставленной строки
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Инвалидируем кэш
	db.invalidateCache()
	return int(id), nil
}

// Update обновляет задачу с инвалидацией кэша
func (db *DB) Update(task *moduls.Scheduler) error {
	result, err := db.Exec(`
		UPDATE scheduler 
		SET date = ?, title = ?, comment = ?, repeat = ? 
		WHERE id = ?
	`, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	// Получаем количество затронутых строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Если строк нет, возвращаем ошибку
	if rowsAffected == 0 {
		return errors.New("задача не найдена")
	}

	// Инвалидируем кэш
	db.invalidateCache()
	return nil
}

// Delete удаляет задачу с инвалидацией кэша
func (db *DB) Delete(id string) error {
	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return err
	}

	// Получаем количество затронутых строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("задача не найдена")
	}

	// Инвалидируем кэш
	db.invalidateCache()
	return nil
}

// invalidateCache очищает кэш
func (db *DB) invalidateCache() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.cache = cache.NewCache()
}

// Функция для проверки соединения с базой данных
func TestDatabaseConnection(db *sql.DB) error {
	// Выполняем простой запрос для создания таблицы, если она не существует
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT NOT NULL,
            title TEXT NOT NULL,
            comment TEXT,
            repeat TEXT(128)
        );
        CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
    `)
	if err != nil {
		log.Printf("Ошибка выполнения тестового запроса: %v", err)
		return fmt.Errorf("ошибка выполнения тестового запроса: %w", err)
	}
	log.Println("Тестовый запрос к базе данных выполнен успешно")
	return nil
}

// GetpoID получает задачу по ID
func (db *DB) GetpoID(id string) (moduls.Scheduler, error) {
	if db == nil {
		return moduls.Scheduler{}, errors.New("database not initialized")
	}
	var task moduls.Scheduler
	log.Printf("Получение задачи с ID: %s", id)

	// Используем ? placeholders для безопасного выполнения запроса
	row := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Задача с ID %s не найдена", id)
			return moduls.Scheduler{}, fmt.Errorf("задача с ID %s не найдена", id)
		}
		log.Printf("Ошибка при получении задачи: %v", err)
		return moduls.Scheduler{}, fmt.Errorf("ошибка при получении задачи: %v", err)
	}
	log.Printf("Задача найдена: %+v", task)
	return task, nil
}

// SearchDate ищет задачи по дате
func (db *DB) SearchDate(date string) ([]moduls.Scheduler, error) {
	log.Printf("SearchDate вызван с параметром: %s", date)
	query := `
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
        WHERE date = ? 
        ORDER BY date ASC
    `

	rows, err := db.Query(query, date)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer rows.Close()

	var tasks []moduls.Scheduler
	for rows.Next() {
		var task moduls.Scheduler
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("ошибка сканирования: %w", err)
		}
		tasks = append(tasks, task)
	}

	// Всегда возвращаем пустой слайс, если нет результатов
	if tasks == nil {
		return []moduls.Scheduler{}, nil
	}

	return tasks, nil
}

// Searchtitl ищет задачи по названию
func (db *DB) Searchtitl(search string) ([]moduls.Scheduler, error) {
	log.Printf("Searchtitl вызван с параметром: %s", search)

	query := `
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
        WHERE title LIKE ? 
        ORDER BY date ASC
    `
	rows, err := db.Query(query, "%"+search+"%") // Используем LIKE для поиска по названию
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer rows.Close()

	var tasks []moduls.Scheduler
	for rows.Next() {
		var task moduls.Scheduler
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("ошибка сканирования: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
