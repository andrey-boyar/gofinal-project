package tasks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"final-project/internal/database"
	"final-project/internal/moduls"
	"final-project/internal/nextdate"
	"final-project/internal/utils"
)

// TaskHandler обрабатывает запросы к /api/task.
func TaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	switch r.Method {
	case http.MethodPost:
		handleTaskPost(w, r, db)
	case http.MethodPut:
		handleTaskPut(w, r, db)
	case http.MethodDelete:
		handleTaskDelete(w, r)
	case http.MethodGet:
		//SearchTasks(w, r, db)
		id := r.URL.Query().Get("id")
		if id != "" {
			task, err := database.GetpoID(db, id)
			if err != nil {
				utils.SendError(w, err.Error(), http.StatusNotFound)
				return
			}
			utils.SendJSON(w, http.StatusOK, task)
		} else {
			SearchTasks(w, r, db)
		}
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// GetTasksHandler получает задачи
func GetTasksHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	// titleFilter := r.URL.Query().Get("title")
	dateFilter := r.URL.Query().Get("date")

	// получение задач
	tasks, err := GetTasks(db, search, dateFilter)
	if err != nil {
		log.Printf("Ошибка при получении задач: %v", err)
		utils.SendError(w, "Ошибка при получении задач", http.StatusInternalServerError)
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
}

// Функция для добавления задачи
func handleTaskPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// Проверка Content-Type
	var taskData moduls.Scheduler
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		utils.SendError(w, "Ошибка при декодировании JSON", http.StatusBadRequest)
		return
	}
	// Установка даты по умолчанию или проверка формата даты
	if len(taskData.Date) == 0 {
		taskData.Date = time.Now().Format(utils.DateFormat)
	} else {
		date, err := time.Parse(utils.DateFormat, taskData.Date)
		if err != nil {
			utils.SendError(w, "bad data format", http.StatusBadRequest)
			return
		}

		if date.Before(time.Now()) {
			taskData.Date = time.Now().Format(utils.DateFormat)
		}
	}
	// Проверка заголовка задачи
	if len(taskData.Title) == 0 {
		utils.SendError(w, "invalid title", http.StatusBadRequest)
		return
	}
	// Проверка формата повтора
	if len(taskData.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), taskData.Date, taskData.Repeat); err != nil {
			utils.SendError(w, "invalid repeat format", http.StatusBadRequest)
			return
		}
	}
	// Добавление задачи в базу данных
	taskId, err := database.Create(db, &taskData)
	if err != nil {
		utils.SendError(w, "failed to create task", http.StatusInternalServerError)
		return
	}
	// Возвращение ID созданной задачи
	utils.SendJSON(w, http.StatusCreated, map[string]interface{}{"id": taskId})
	// Преобразуем taskId в строку и возвращаем её напрямую
	//utils.SendJSON(w, http.StatusCreated, strconv.Itoa(taskId))
	log.Println("Added task with id=%d", taskId)
}

// функция для поиска задач
func SearchTasks(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	//date := r.URL.Query().Get("date")

	// поиск задач
	tasks, err := searchDate(db, search)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// если задач нет, то возвращаем пустой слайс
	if tasks == nil {
		tasks = []moduls.Scheduler{}
	}
	// Отправка ответа
	utils.SendJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
	log.Printf("Read %d tasks", len(tasks))
}

// получение задач по дате или поиску
func searchDate(db *sql.DB, search string) ([]moduls.Scheduler, error) {
	if len(search) > 0 {
		if date, err := time.Parse("02.01.2006", search); err == nil {
			return database.ReadTask(db, date.Format(utils.DateFormat))
		}
		return database.Searchtitl(db, search)
	}
	return database.SearchDate(db, search)
}

// handleTaskPut обновляет задачу
func handleTaskPut(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task moduls.Scheduler

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.SendError(w, "JSON deserialization error", http.StatusBadRequest)
		return
	}
	// проверка id
	if len(task.ID) == 0 {
		utils.SendError(w, "invalid id", http.StatusBadRequest)
		return
	}
	// проверка id на число
	if _, err := strconv.Atoi(task.ID); err != nil {
		utils.SendError(w, "invalid id", http.StatusBadRequest)
		return
	}
	// проверка даты
	parseDate, err := time.Parse(utils.DateFormat, task.Date)
	if err != nil {
		utils.SendError(w, "invalid date format", http.StatusBadRequest)
		return
	}
	if parseDate.Before(time.Now()) {
		// как в создании задачи
		task.Date = time.Now().Format(utils.DateFormat)
	}
	// проверка заголовка
	if len(task.Title) == 0 {
		utils.SendError(w, "invalid title", http.StatusBadRequest)
		return
	}
	// проверка формата повтора
	if len(task.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			utils.SendError(w, "invalid repeat format", http.StatusBadRequest)
			return
		}
	}

	_, err = database.Update(db, &task) // обновление задачи
	if err != nil {
		utils.SendError(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	utils.SendJSON(w, http.StatusOK, task)
	//if err := json.NewEncoder(w).Encode(task); err != nil {
	//utils.SendError(w, "failed to encode response", http.StatusInternalServerError)
	//	return
	//}
}

// GetTaskByID получает задачу по ID
func GetTaskByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	task, err := database.GetpoID(db, id)
	if err != nil {
		utils.SendError(w, "failed to get task by id", http.StatusInternalServerError)
		return
	}
	// Отправка ответа
	utils.SendJSON(w, http.StatusOK, task)
	log.Printf("Read task with id=%s", id)
}

// UpdateTask обновляет задачу в базе данных
func UpdateTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task moduls.Scheduler
	// декодирование JSON тела запроса
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.SendError(w, "JSON deserialization error", http.StatusBadRequest)
		return
	}
	// проверка id
	if len(task.ID) == 0 {
		utils.SendError(w, "invalid id", http.StatusBadRequest)
		return
	}
	// проверка id на число
	if _, err := strconv.Atoi(task.ID); err != nil {
		utils.SendError(w, "invalid id", http.StatusBadRequest)
		return
	}
	// проверка даты
	parseDate, err := time.Parse(utils.DateFormat, task.Date)
	if err != nil {
		utils.SendError(w, "invalid date format", http.StatusBadRequest)
		return
	}
	// если дата в прошлом, то ставим текущую дату
	if parseDate.Before(time.Now()) {
		// как в создании задачи
		task.Date = time.Now().Format(utils.DateFormat)
	}
	// проверка заголовка
	if len(task.Title) == 0 {
		utils.SendError(w, "invalid title", http.StatusBadRequest)
		return
	}
	if task.Repeat != "" { // если повтор есть, то получаем следующую дату
		nextDate, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			utils.SendError(w, fmt.Sprintf("Неверный формат повтора: %v", err), http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	}
	// проверка формата повтора
	if len(task.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			utils.SendError(w, "invalid repeat format", http.StatusBadRequest)
			return
		}
	}
	// обновление задачи
	_, err = database.Update(db, &task)
	if err != nil {
		utils.SendError(w, "failed to update task", http.StatusInternalServerError)
		return
	}
	utils.SendJSON(w, http.StatusOK, task)
}

// HandleTaskDone обрабатывает запрос на выполнение задачи
func HandleTaskDone(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("API: Завершение задачи")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	task, err := database.GetpoID(db, id)
	if err != nil {
		utils.SendError(w, "failed to get task by id", http.StatusInternalServerError)
		return
	}

	if task.Repeat == "" {
		err = database.Delete(task.ID)
		if err != nil {
			utils.SendError(w, "failed to delete task", http.StatusInternalServerError)
			return
		}
		//log.Println(fmt.Sprintf("task with id=%s was deleted", task.ID))
	} else {
		task.Date, err = nextdate.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			utils.SendError(w, "failed to get next date", http.StatusInternalServerError)
			return
		}
		// Обновляем задачу с новой датой
		_, err = database.Update(db, &task)
		if err != nil {
			utils.SendError(w, "failed to update task", http.StatusInternalServerError)
			return
		}
	}

	if err := json.NewEncoder(w).Encode(struct{}{}); err != nil {
		utils.SendError(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
	//log.Println(fmt.Sprintf("Updated task with id=%s", task.ID))
}

// handleTaskDelete удаляет задачу
func handleTaskDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.SendError(w, "ID задачи не указан", http.StatusBadRequest)
		return
	}
	// удаление задачи
	err := database.Delete(id)
	if err != nil {
		utils.SendError(w, "failed to delete task", http.StatusInternalServerError)
		return
	}
	// отправка ответа
	w.WriteHeader(http.StatusNoContent)
	log.Printf("Deleted task with id=%s", id)
}

// GetTasks получает задачи с фильтрами
func GetTasks(db *sql.DB, titleFilter string, dateFilter string) ([]moduls.Scheduler, error) {
	log.Printf("Получение задач с фильтрами: title=%s, date=%s", titleFilter, dateFilter)

	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE 1=1"
	var args []interface{}

	if titleFilter != "" { // если titleFilter не пустой, то добавляем в запрос
		query += " AND (title LIKE ? OR comment LIKE ?)"
		args = append(args, "%"+titleFilter+"%", "%"+titleFilter+"%")
	}

	if dateFilter != "" { // если dateFilter не пустой, то добавляем в запрос
		query += " AND date = ?"
		args = append(args, dateFilter)
	}

	query += " ORDER BY date"

	rows, err := db.Query(query, args...) // выполнение запроса
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var tasks []moduls.Scheduler
	for rows.Next() {
		var task moduls.Scheduler
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("ошибка сканирования задачи: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	if len(tasks) == 0 {
		log.Printf("Задачи не найдены для фильтров: title=%s, date=%s", titleFilter, dateFilter)
		return []moduls.Scheduler{}, nil // Возвращаем пустой слайс вместо nil
	}

	log.Printf("Получено задач: %d", len(tasks))
	return tasks, nil
}
