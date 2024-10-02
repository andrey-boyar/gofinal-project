package tasks

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"final-project/internal/database"
	"final-project/internal/nextdate"
)

const DP = "20060102" //формат даты

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
		// Обработка POST-запроса для добавления задачи
		handleTaskPost(w, r, db)
	case http.MethodGet:
		// Обработка GET-запроса для получения информации о задаче
		handleTaskGet(w, r, db)
	case http.MethodPut:
		// Обработка PUT-запроса для обновления задачи
		handleTaskPut(w, r, db)
	case http.MethodDelete:
		// Обработка DELETE-запроса для удаления задачи
		handleTaskDelete(w, r, db)
	default:
		http.Error(w, "недопустимый метод", http.StatusMethodNotAllowed)
	}
}

// Функция для добавления задачи
func handleTaskPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var taskData Task //структура для хранения данных задачи

	// Декодирование JSON тела запроса
	response, err := json.Marshal(map[string]int{"id": taskData.ID})
	if err != nil {
		setErrorResponse(w, "Ошибка кодирования JSON", err)
		return
	}
	//w.Write(response)
	if _, err := w.Write(response); err != nil {
		log.Printf("Ошибка при отправке ответа: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	} //отправка ответа

	// Установка даты по умолчанию или проверка формата даты
	if len(taskData.Date) == 0 {
		taskData.Date = time.Now().Format(DP)              //установка даты по умолчанию
		log.Printf("Дата по умолчанию: %s", taskData.Date) //логирование даты по умолчанию
	} else {
		date, err := time.Parse(DP, taskData.Date) //парсинг даты
		if err != nil {
			setErrorResponse(w, "Неверный формат даты", err)
			return
		}
		//установка даты по умолчанию
		if date.Before(time.Now().Truncate(24 * time.Hour)) {
			taskData.Date = time.Now().Format(DP)
		}
	}

	// Проверка заголовка задачи
	if len(taskData.Title) == 0 {
		setErrorResponse(w, "Недопустимый заголовок", errors.New("заголовок пустой"))
		return
	}

	// Проверка формата повтора
	//if len(taskData.Repeat) > 0 {
	//if _, err := nextdate.NextDate(time.Now(), taskData.Date, taskData.Repeat); err != nil {
	//setErrorResponse(w, "Недопустимый формат повтора", errors.New("нет такого формата"))
	//return
	//}	//}
	if taskData.Repeat != "" {
		if _, err := nextdate.NextDate(time.Now(), taskData.Date, taskData.Repeat); err != nil {
			setErrorResponse(w, "Неверный формат повтора", err)
			return
		}
	}

	// Добавление задачи в базу данных
	taskID, err := insertTask(db, taskData)
	if err != nil {
		setErrorResponse(w, "Не удалось создать задачу", err)
		return
	}

	// Возвращение ID созданной задачи
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]int64{"id": taskID}); err != nil {
		setErrorResponse(w, "Не удалось закодировать ответ", err)
		return
	}
	log.Printf("Добавлена задача с id=%d", taskID) //логирование добавления задачи
}

// handleTaskGet обрабатывает GET-запрос для получения информации о задаче.
func handleTaskGet(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Получение ID задачи из URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		// Если ID не указан, возвращаем список задач
		SearchHandler(w, r, db)
		return
	}

	// Преобразование ID в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("неверный формат ID: %s", err), http.StatusBadRequest)
		return
	}

	// Запрос информации о задаче из базы данных
	var task Task
	err = db.QueryRow(`
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE id = ?
	`, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		} else {
			setErrorResponse(w, "ошибка получения задачи", err)
		}
		return
	}

	// Возвращаем JSON-ответ с информацией о задаче
	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//получение задачи
	tasks, err := database.GetTasks(db, map[string]string{"id": idStr})
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка получения задач: %s"}`, err), http.StatusInternalServerError)
		return
	}
	//отправка ответа
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks}); err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

// handleTaskPut обрабатывает PUT-запрос для обновления задачи.
func handleTaskPut(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Десериализация JSON-запроса
	var task database.Scheduler
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		setErrorResponse(w, "ошибка десериализации JSON", err)
		return
	}

	// Проверка обязательного поля ID
	if task.ID == 0 {
		setErrorResponse(w, "Не указан ID задачи", errors.New("ID пустой"))
		return
	}
	// Проверка обязательного поля Title
	if task.Title == "" {
		setErrorResponse(w, "Не указан заголовок задачи", errors.New("заголовок пустой"))
		return
	}

	// Установка даты по умолчанию или проверка формата даты
	if len(task.Date) == 0 {
		task.Date = time.Now().Format("20060102")
	} else {
		date, err := time.Parse("20060102", task.Date)
		if err != nil {
			setErrorResponse(w, "неверный формат даты", err)
			return
		}
		//установка даты по умолчанию
		if date.Before(time.Now()) {
			task.Date = time.Now().Format("20060102")
		}
	}

	// Проверка формата повтора
	if len(task.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			setErrorResponse(w, "Неверный формат повтора", err)
			return
		}
	}

	// Обновление задачи в базе данных
	_, err = db.Exec(`
		UPDATE scheduler
		SET date = ?, title = ?, comment = ?, repeat = ?
		WHERE id = ?
	`, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка обновления задачи: %s", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем пустой JSON-ответ
	jsonError(w, http.StatusOK)
	//обновление задачи
	err = database.Update(db, &database.Scheduler{})
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка обновления задачи: %s"}`, err), http.StatusInternalServerError)
		return
	}
	//отправка ответа
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

// handleTaskDelete обрабатывает DELETE-запрос для удаления задачи.
func handleTaskDelete(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Получение ID задачи из URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "отсутствует ID задачи", http.StatusBadRequest)
		return
	}

	// Преобразование ID в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("неверный формат ID: %s", err), http.StatusBadRequest)
		return
	}

	// Удаление задачи из базы данных
	result, err := db.Exec(`
		DELETE FROM scheduler
		WHERE id = ?
	`, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка удаления задачи: %s", err), http.StatusInternalServerError)
		return
	}

	// Проверка, сколько строк было удалено
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка получения количества удаленных строк: %s", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем JSON-ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if rowsAffected > 0 {
		// Отправляем пустой JSON, если задача удалена
		fmt.Fprint(w, "{}")
	} else {
		// Возвращаем ошибку, если задача не найдена
		http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
	}
	//удаление задачи
	err = database.Delete(db, id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка удаления задачи: %s"}`, err), http.StatusInternalServerError)
		return
	}
	//отправка ответа
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

// HandleTaskDone обрабатывает PUT-запрос для отметки задачи как выполненной.
func HandleTaskDone(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Получение ID задачи из URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "отсутствует ID задачи", http.StatusBadRequest)
		return
	}

	// Преобразование ID в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("неверный формат ID: %s", err), http.StatusBadRequest)
		return
	}

	// Обновление задачи в базе данных
	_, err = db.Exec(`
		UPDATE scheduler
		SET completed = 1
		WHERE id = ?
	`, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка обновления задачи: %s", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем пустой JSON-ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "{}")
	//обновление задачи
	_, err = db.Exec("UPDATE scheduler SET completed = 1 WHERE id = ?", id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка обновления задачи: %s"}`, err), http.StatusInternalServerError)
		return
	}
	//отправка ответа
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

// fetchTasks получает задачи из базы данных.
func fetchTasks(db *sql.DB, search string, dateStr string) ([]Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= ?`
	args := []interface{}{time.Now().Format(DP)} //добавление даты в запрос

	if search != "" {
		query += " AND (title LIKE ? OR comment LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	if dateStr != "" {
		date, err := time.Parse(DP, dateStr)
		if err != nil {
			return nil, fmt.Errorf("неверный формат даты: %w", err)
		}
		query += " AND date = ?"
		args = append(args, date.Format(DP)) //добавление даты в запрос
	}
	//условное ограничение на количество задач
	query += " ORDER BY date ASC LIMIT 50"
	//запрос на поиск задач
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения задач: %w", err)
	}
	defer rows.Close()

	var tasks []Task //Создание массива задач
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

// Функция для поиска задач
func SearchHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	dateStr := r.URL.Query().Get("date")

	tasks, err := fetchTasks(db, search, dateStr)
	if err != nil {
		setErrorResponse(w, "Ошибка получения задач", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks}); err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	log.Printf("Получено %d задач", len(tasks))
}

// jsonError отправляет JSON-ответ с ошибкой.
func jsonError(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	//fmt.Fprintf(w, `{"error": "%s"}`, http.StatusText(status))
}

// setErrorResponse отправляет JSON-ответ с ошибкой.
func setErrorResponse(w http.ResponseWriter, message string, err error) {
	errorMsg := fmt.Sprintf("%s: %v", message, err)
	w.WriteHeader(http.StatusBadRequest)
	//json.NewEncoder(w).Encode(map[string]string{"error": errorMsg})
	if err := json.NewEncoder(w).Encode(map[string]string{"error": errorMsg}); err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	errorResponse := Errors{
		Errors: fmt.Errorf("%s: %w", message, err).Error(),
	}

	// Сериализация ответа об ошибке
	errorData, marshalErr := json.Marshal(errorResponse)
	if marshalErr != nil {
		// Если возникла ошибка при маршалинге, отправляем простую ошибку
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)

	// Пишем ответ
	_, writeErr := w.Write(errorData)
	if writeErr != nil {
		// Если возникла ошибка при записи ответа, логируем её
		fmt.Printf("Error writing response: %v\n", writeErr)
	}
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
