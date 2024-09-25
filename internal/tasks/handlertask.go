package tasks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"final-project/internal/database"
	"final-project/internal/nextdate"
)

// Task структура для хранения информации о задаче.
type Task struct {
	ID      int    `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// TaskHandler обрабатывает запросы к /api/task.
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	db := database.InitDatabase() // Функция для получения соединения с БД
	defer db.Close()              // Закрываем соединение с БД после завершения работы
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

// handleTaskPost обрабатывает POST-запрос для добавления задачи.
func handleTaskPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Десериализация JSON-запроса
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка десериализации JSON: %s"}`, err), http.StatusBadRequest)
		return
	}

	// Проверка обязательного поля Title
	if task.Title == "" {
		http.Error(w, `{"error": "Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	// Установка даты по умолчанию или проверка формата даты
	if len(task.Date) == 0 {
		task.Date = time.Now().Format("20060102")
	} else {
		date, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "неверный формат даты: %s"}`, err), http.StatusBadRequest)
			return
		}

		if date.Before(time.Now()) {
			task.Date = time.Now().Format("20060102")
		}
	}

	// Проверка формата повтора
	if len(task.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "неверный формат повтора: %s"}`, err), http.StatusBadRequest)
			return
		}
	}

	// Вставка новой задачи в базу данных
	result, err := db.Exec(`
		INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (?, ?, ?, ?)
	`, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка вставки задачи: %s"}`, err), http.StatusInternalServerError)
		return
	}

	// Получение ID новой записи
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка получения ID: %s"}`, err), http.StatusInternalServerError)
		return
	}

	// Возвращаем JSON-ответ с ID новой записи
	fmt.Fprintf(w, `{"id": %d}`, id)

	err = database.Create(db, &database.Scheduler{})
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка создания задачи: %s"}`, err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
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
			http.Error(w, fmt.Sprintf("ошибка получения задачи: %s", err), http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем JSON-ответ с информацией о задаче
	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tasks, err := database.GetTasks(db, map[string]string{"id": idStr})
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка получения задач: %s"}`, err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

// handleTaskPut обрабатывает PUT-запрос для обновления задачи.
func handleTaskPut(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Десериализация JSON-запроса
	var task database.Scheduler
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка десериализации JSON: %s"}`, err), http.StatusBadRequest)
		return
	}

	// Проверка обязательного поля ID
	if task.ID == 0 {
		http.Error(w, `{"error": "Не указан ID задачи"}`, http.StatusBadRequest)
		return
	}
	// Проверка обязательного поля Title
	if task.Title == "" {
		http.Error(w, `{"error": "Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	// Установка даты по умолчанию или проверка формата даты
	if len(task.Date) == 0 {
		task.Date = time.Now().Format("20060102")
	} else {
		date, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "неверный формат даты: %s"}`, err), http.StatusBadRequest)
			return
		}

		if date.Before(time.Now()) {
			task.Date = time.Now().Format("20060102")
		}
	}

	// Проверка формата повтора
	if len(task.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "неверный формат повтора: %s"}`, err), http.StatusBadRequest)
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

	err = database.Update(db, &database.Scheduler{})
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка обновления задачи: %s"}`, err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
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

	err = database.Delete(db, id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка удаления задачи: %s"}`, err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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

	_, err = db.Exec("UPDATE scheduler SET completed = 1 WHERE id = ?", id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка обновления задачи: %s"}`, err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

}

// SearchHandler обрабатывает GET-запрос для получения списка ближайших задач.
func SearchHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Проверка метода запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	// Получение параметров запроса
	search := r.URL.Query().Get("search")
	dateStr := r.URL.Query().Get("date")

	// Создание SQL-запроса
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= ?`
	args := []interface{}{time.Now().Format("20060102")}

	// Добавление условия поиска по `search`
	if search != "" {
		query += " AND (title LIKE ? OR comment LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	// Добавление условия выбора по `date`
	if dateStr != "" {
		// Преобразуем дату в формат 20060102
		date, err := time.Parse("20060102", dateStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("неверный формат даты: %s", err), http.StatusBadRequest)
			return
		}
		query += " AND date = ?"
		args = append(args, date.Format("20060102"))
	}

	// Добавьте условие сортировки и ограничение на количество задач (LIMIT 50)
	query += " ORDER BY date ASC LIMIT 50"

	// Запрос задач из базы данных
	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка получения задач: %s", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Создание списка задач
	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			http.Error(w, fmt.Sprintf("ошибка сканирования задачи: %s", err), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	// Отправка JSON-ответа с списком задач
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// jsonError отправляет JSON-ответ с ошибкой.
func jsonError(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error": "%s"}`, http.StatusText(status))
}
