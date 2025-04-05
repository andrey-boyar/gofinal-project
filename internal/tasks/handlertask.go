package tasks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"final-project/internal/utils"
)

// Task структура для хранения информации о задаче.
type Task struct {
	ID      int    `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Errors структура для ответа с ошибкой
type Errors struct {
	Errors string `json:"error"`
}

// TaskHandler обрабатывает запросы к /api/task.
func TaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Установка заголовка Content-Type для JSON
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Обработка запросов в зависимости от метода
	switch r.Method {
	case http.MethodPost:
		handleTaskPost(w, r, db)
	case http.MethodGet:
		handleTaskGet(w, r, db)
	case http.MethodPut:
		handleTaskPut(w, r, db)
	case http.MethodDelete:
		handleTaskDelete(w, r, db)
	default:
		http.Error(w, "недопустимый метод", http.StatusMethodNotAllowed)
	}
}

// Функция для добавления задачи
func handleTaskPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var taskData Task // структура для хранения данных задачи

	// Декодирование JSON тела запроса
	if err := utils.DecodeJSON(w, r, &taskData); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		utils.SendError(w, "Ошибка декодирования JSON", http.StatusBadRequest)
		return
	}

	// Установка даты по умолчанию или проверка формата даты
	if len(taskData.Date) == 0 {
		taskData.Date = time.Now().Format("20060102") // установка даты по умолчанию
		log.Printf("Дата по умолчанию: %s", taskData.Date)
	} else {
		date, err := time.Parse("20060102", taskData.Date) // парсинг даты
		if err != nil {
			log.Printf("Неверный формат даты: %v", err)
			utils.SendError(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}
		if date.Before(time.Now().Truncate(24 * time.Hour)) {
			taskData.Date = time.Now().Format("20060102")
		}
	}

	// Проверка заголовка задачи
	if len(taskData.Title) == 0 {
		utils.SendError(w, "Недопустимый заголовок", http.StatusBadRequest)
		return
	}

	// Проверка формата повтора
	if taskData.Repeat != "" {
		if _, err := NextDate(time.Now(), taskData.Date, taskData.Repeat); err != nil {
			log.Printf("Неверный формат повтора: %v", err)
			utils.SendError(w, "Неверный формат повтора", http.StatusBadRequest)
			return
		}
	}

	// Добавление задачи в базу данных
	taskID, err := insertTask(db, taskData)
	if err != nil {
		log.Printf("Не удалось создать задачу: %v", err)
		utils.SendError(w, "Не удалось создать задачу", http.StatusBadRequest)
		return
	}

	// Возвращение ID созданной задачи
	utils.SendJSON(w, http.StatusCreated, map[string]int64{"id": taskID})
	log.Printf("Добавлена задача с id=%d", taskID)
}

// handleTaskGet обрабатывает GET-запрос для получения информации о задаче.
func handleTaskGet(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Получение ID задачи из URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		utils.SendError(w, "отсутствует ID задачи", http.StatusBadRequest)
		SearchHandler(w, r, db)
		return
	}

	// Преобразование ID в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, fmt.Sprintf("неверный формат ID: %s", err), http.StatusBadRequest)
		return
	}

	// Запрос информации о задаче из базы данных
	var task Task
	err = db.QueryRow(`
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE id = ?`, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.SendError(w, "Задача не найдена", http.StatusNotFound)
		} else {
			log.Printf("Ошибка получения задачи: %v", err)
			utils.SendError(w, "ошибка получения задачи", http.StatusBadRequest)
		}
		return
	}

	// Возвращаем JSON-ответ с информацией о задаче
	utils.SendJSON(w, http.StatusOK, task)
	log.Printf("Получена задача с id=%d", task.ID)
}

// handleTaskPut обрабатывает PUT-запрос для обновления задачи.
func handleTaskPut(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Printf("Ошибка десериализации JSON: %v", err)
		utils.SendError(w, "ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	// Проверка обязательного поля ID
	if task.ID == 0 {
		utils.SendError(w, "Не указан ID задачи", http.StatusBadRequest)
		return
	}
	// Проверка обязательного поля Title
	if task.Title == "" {
		utils.SendError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	// Установка даты по умолчанию или проверка формата даты
	if len(task.Date) == 0 {
		task.Date = time.Now().Format("20060102")
	} else {
		date, err := time.Parse("20060102", task.Date)
		if err != nil {
			log.Printf("Неверный формат даты: %v", err)
			utils.SendError(w, "неверный формат даты", http.StatusBadRequest)
			return
		}
		if date.Before(time.Now()) {
			task.Date = time.Now().Format("20060102")
		}
	}

	// Проверка формата повтора
	if len(task.Repeat) > 0 {
		if _, err := NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			log.Printf("Неверный формат повтора: %v", err)
			utils.SendError(w, "Неверный формат повтора", http.StatusBadRequest)
			return
		}
	}

	// Обновление задачи в базе данных
	_, err = db.Exec(`
		UPDATE scheduler
		SET date = ?, title = ?, comment = ?, repeat = ?
		WHERE id = ?`, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		log.Printf("Ошибка обновления задачи: %v", err)
		utils.SendError(w, fmt.Sprintf("ошибка обновления задачи: %s", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем пустой JSON-ответ
	utils.SendJSON(w, http.StatusOK, map[string]interface{}{})
	log.Printf("Обновлена задача с id=%d", task.ID)
}

// HandleTaskDone обрабатывает PUT-запрос для отметки задачи как выполненной.
func HandleTaskDone(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Получение ID задачи из URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		utils.SendError(w, "отсутствует ID задачи", http.StatusBadRequest)
		return
	}

	// Преобразование ID в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, fmt.Sprintf("неверный формат ID: %s", err), http.StatusBadRequest)
		return
	}

	// Обновление задачи в базе данных
	_, err = db.Exec(`
		UPDATE scheduler
		SET completed = 1
		WHERE id = ?
	`, id)
	if err != nil {
		utils.SendError(w, fmt.Sprintf("ошибка обновления задачи: %s", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	utils.SendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	log.Printf("Задача с id=%d отмечена как выполненная", id)
}

// handleTaskDelete обрабатывает DELETE-запрос для удаления задачи.
func handleTaskDelete(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Получение ID задачи из URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		utils.SendError(w, "отсутствует ID задачи", http.StatusBadRequest)
		return
	}

	// Преобразование ID в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, fmt.Sprintf("неверный формат ID: %s", err), http.StatusBadRequest)
		return
	}

	// Удаление задачи из базы данных
	result, err := db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		log.Printf("Ошибка удаления задачи: %s", err)
		utils.SendError(w, fmt.Sprintf("ошибка удаления задачи: %s", err), http.StatusInternalServerError)
		return
	}

	// Проверка, сколько строк было удалено
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка получения количества удаленных строк: %s", err)
		utils.SendError(w, fmt.Sprintf("ошибка получения количества удаленных строк: %s", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем пустой JSON, если задача удалена
	if rowsAffected > 0 {
		utils.SendJSON(w, http.StatusOK, map[string]interface{}{}) // Успешное удаление
		log.Printf("Задача с id=%d удалена", id)
	} else {
		utils.SendError(w, "Задача не найдена", http.StatusNotFound) // Задача не найдена
	}
}

// fetchTasks получает задачи из базы данных.
func fetchTasks(db *sql.DB, search string, dateStr string) ([]Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= ?`
	args := []interface{}{time.Now().Format("20060102")} // добавление даты в запрос

	if search != "" {
		query += " AND (title LIKE ? OR comment LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	if dateStr != "" {
		date, err := time.Parse("20060102", dateStr)
		if err != nil {
			return nil, fmt.Errorf("неверный формат даты: %w", err)
		}
		query += " AND date = ?"
		args = append(args, date.Format("20060102")) // добавление даты в запрос
	}
	// условное ограничение на количество задач
	query += " ORDER BY date ASC LIMIT 50"
	// запрос на поиск задач
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения задач: %w", err)
	}
	defer rows.Close()

	var tasks []Task // Создание массива задач
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования задачи: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// / Функция для поиска задач
func SearchHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	dateStr := r.URL.Query().Get("date")

	tasks, err := fetchTasks(db, search, dateStr)
	if err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		utils.SendError(w, "Ошибка получения задач", http.StatusBadRequest)
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
	log.Printf("Получено %d задач", len(tasks))
}

// insertTask добавляет задачу в базу данных.
func insertTask(db *sql.DB, task Task) (int64, error) {
	result, err := db.Exec(`
        INSERT INTO scheduler (date, title, comment, repeat)
        VALUES (?, ?, ?, ?)
    `, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
