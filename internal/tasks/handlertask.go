package tasks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	//"fmt"
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
		id := r.URL.Query().Get("id")
		if id != "" {
			task, err := database.GetpoID(id)
			if err != nil {
				if strings.Contains(err.Error(), "не найдена") {
					utils.SendError(w, err.Error(), http.StatusNotFound)
				} else {
					utils.SendError(w, "Ошибка при получении задачи", http.StatusInternalServerError)
				}
				return
			}
			utils.SendJSON(w, http.StatusOK, task)
		} else {
			search(w, r, db)
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
	log.Printf("входные данные: %+v", taskData)
	// Установка даты по умолчанию или проверка формата даты
	if len(taskData.Date) == 0 {
		taskData.Date = time.Now().Format(utils.DateFormat)
	} else {
		// проверка формата даты
		date, err := time.Parse(utils.DateFormat, taskData.Date)
		if err != nil {
			utils.SendError(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}
		// если дата в прошлом, то ставим текущую дату
		if date.Before(time.Now()) {
			taskData.Date = time.Now().Format(utils.DateFormat)
		}
	}

	// Проверка заголовка задачи
	if len(taskData.Title) == 0 {
		utils.SendError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}
	// Проверка формата повтора
	if len(taskData.Repeat) > 0 {
		log.Printf("Ошибка валидации формата повтора: %v", taskData.Repeat)
		if _, err := nextdate.NextDate(time.Now(), taskData.Date, taskData.Repeat); err != nil {
			utils.SendError(w, "Неверный формат повтора", http.StatusBadRequest)
			return
		}
		if err := nextdate.ValidateRepeatFormat(taskData.Repeat); err != nil {
			utils.SendError(w, fmt.Sprintf("Неверный формат повтора: %v", err), http.StatusBadRequest)
			return
		}
		taskData.Repeat = nextdate.NormalizeRepeatFormat(taskData.Repeat)
		if err := nextdate.ValidateRepeatFormat(taskData.Repeat); err != nil {
			log.Printf("Ошибка валидации формата повтора: %v", err)
			utils.SendError(w, fmt.Sprintf("Неверный формат повтора: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Добавление задачи в базу данных
	taskId, err := database.InsertTask(db, taskData)
	if err != nil {
		utils.SendError(w, "Ошибка создания задачи", http.StatusInternalServerError)
		return
	}
	// Возвращение ID созданной задачи
	log.Printf("Added task with id=%d", taskId)
	utils.SendJSON(w, http.StatusCreated, map[string]interface{}{
		"id": taskId,
	})
}

// функция для поиска задач
func search(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	date := r.URL.Query().Get("date")

	tasks, err := GetTasks(db, search, date)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []moduls.Scheduler{}
	}
	// Отправка ответа
	utils.SendJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
	log.Printf("Read %d tasks", len(tasks))
}

// handleTaskPut обновляет задачу
func handleTaskPut(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task moduls.Scheduler

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.SendError(w, "JSON deserialization error", http.StatusBadRequest)
		return
	}
	if err := validateTaskInput(&task); err != nil {
		utils.SendError(w, "Invalid input", http.StatusBadRequest)
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
	// проверка формата повтора
	if len(task.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			utils.SendError(w, "invalid repeat format", http.StatusBadRequest)
			return
		}
	}
	// обновление задачи
	updatedTask, err := database.Update(db, &task)
	if err != nil {
		utils.SendError(w, "failed to update task", http.StatusInternalServerError)
		return
	}
	utils.SendJSON(w, http.StatusOK, updatedTask)
	// utils.SendJSON(w, http.StatusOK, task)
}

// GetTaskByID получает задачу по ID
func GetTaskByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	task, err := database.GetpoID(id)
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
	if id == "" {
		log.Println("API: Не передан id задачи")
		w.WriteHeader(http.StatusBadRequest)
		utils.SendError(w, "Не передан id задачи", http.StatusBadRequest)
		return
	}

	idTask, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("Неверный формат id задачи: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		utils.SendError(w, "Неверный формат id задачи", http.StatusBadRequest)
		return
	}

	task, err := database.GetpoID(strconv.Itoa(idTask))
	if err != nil {
		log.Printf("Ошибка при получении задачи: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendError(w, "Ошибка при получении задачи", http.StatusInternalServerError)
		return
	}

	if task.Repeat != "" {
		log.Printf("Задача повторяется: %s", task.Repeat)
		baseDate, _ := time.Parse("20060102", task.Date)
		nextDateTask, err := nextdate.NextDate(baseDate, task.Date, task.Repeat)
		log.Printf("HandleTaskDone: next date calculated: %s", nextDateTask)
		if err != nil {
			log.Printf("Ошибка при получении следующей даты задачи: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendError(w, "Ошибка при получении следующей даты задачи", http.StatusInternalServerError)
			return
		}
		if len(task.Repeat) > 0 {
			task.Repeat = nextdate.NormalizeRepeatFormat(task.Repeat)
			if err := nextdate.ValidateRepeatFormat(task.Repeat); err != nil {
				log.Printf("Ошибка валидации формата повтора: %v", err)
				utils.SendError(w, fmt.Sprintf("Неверный формат повтора: %v", err), http.StatusBadRequest)
				return
			}
		}
		task.Date = nextDateTask
		_, err = database.Update(db, &task)
		if err != nil {
			log.Printf("Ошибка при обновлении задачи: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendError(w, "Ошибка при обновлении задачи", http.StatusInternalServerError)
			return
		}
	} else {
		log.Println("Задача не повторяется, удаляем задачу")
		err = database.Delete(strconv.Itoa(idTask))
		if err != nil {
			log.Printf("Ошибка при удалении задачи: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendError(w, "Ошибка при удалении задачи", http.StatusInternalServerError)
			return
		}
	}
	// Отправляем JSON-ответ
	w.WriteHeader(http.StatusOK)
}

// handleTaskDelete удаляет задачу
func handleTaskDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	// удаление задачи
	err := database.Delete(id)
	if err != nil {
		utils.SendError(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	log.Printf("Deleted task with id=%s", id)
}

// функция для валидации задачи
func validateTaskInput(task *moduls.Scheduler) error {
	if task.ID == "" {
		return fmt.Errorf("ID задачи не может быть пустым")
	}
	if task.Title == "" {
		return fmt.Errorf("заголовок задачи не может быть пустым")
	}
	if _, err := time.Parse(utils.DateFormat, task.Date); err != nil {
		return fmt.Errorf("неверный формат даты")
	}
	if task.Repeat != "" {
		if err := nextdate.ValidateRepeatFormat(task.Repeat); err != nil {
			return err
		}
	}
	return nil
}

func GetTasks(db *sql.DB, titleFilter string, dateFilter string) ([]moduls.Scheduler, error) {
	log.Printf("Получение задач с фильтрами: title=%s, date=%s", titleFilter, dateFilter)

	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE 1=1"
	var args []interface{}

	if titleFilter != "" {
		query += " AND (title LIKE ? OR comment LIKE ?)"
		args = append(args, "%"+titleFilter+"%", "%"+titleFilter+"%")
	}

	if dateFilter != "" {
		query += " AND date = ?"
		args = append(args, dateFilter)
	}

	query += " ORDER BY date"

	rows, err := db.Query(query, args...)
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
