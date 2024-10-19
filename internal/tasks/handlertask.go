package tasks

import (
	"database/sql"
	"encoding/json"
	"fmt"

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
<<<<<<< HEAD

=======
		if _, err := time.Parse("20060102", taskData.Date); err != nil {
			utils.SendError(w, "Неверный формат даты. Используйте YYYYMMDD", http.StatusBadRequest)
			return
		}
		// если дата в прошлом, то ставим текущую дату
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
		if date.Before(time.Now()) {
			taskData.Date = time.Now().Format(utils.DateFormat)
		}
		if taskData.Repeat != "" {
			if strings.HasPrefix(taskData.Repeat, "y") && !strings.Contains(taskData.Repeat, ".") {
				utils.SendError(w, "Для ежегодного повтора необходимо указать дату в формате 'y MM.DD'", http.StatusBadRequest)
				return
			}
		}
	}
	// if strings.HasPrefix(taskData.Repeat, "y") {
	if taskData.Repeat == "y" || (strings.HasPrefix(taskData.Repeat, "y") && !strings.Contains(taskData.Repeat, ".")) {
		// Если дата не указана, используем дату из поля Date
		date, err := utils.ParseDate(taskData.Date)
		if err != nil {
			utils.SendError(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}
		taskData.Repeat = fmt.Sprintf("y %02d.%02d", date.Month(), date.Day())
		log.Printf("Автоматически дополнен формат повтора: %s", taskData.Repeat)

	}

	// Валидация формата повтора
	if len(taskData.Repeat) > 0 {
		if err := nextdate.ValidateRepeatFormat(taskData.Repeat); err != nil {
			log.Printf("Ошибка валидации формата повтора: %v", err)
			utils.SendError(w, fmt.Sprintf("Неверный формат повтора: %v", err), http.StatusBadRequest)
			return
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
	//if err := json.NewEncoder(w).Encode(moduls.TaskId{Id: taskId}); err != nil {
	//utils.SendError(w, "failed to encode response", http.StatusInternalServerError)
	//return
	//}
	log.Println("Added task with id=%d", taskId)
}

// функция для поиска задач
func SearchTasks(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	//date := r.URL.Query().Get("date")

	tasks, err := searchDate(db, search)
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
	if len(task.ID) == 0 {
		utils.SendError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := strconv.Atoi(task.ID); err != nil {
		utils.SendError(w, "invalid id", http.StatusBadRequest)
		return
	}
	parseDate, err := time.Parse(utils.DateFormat, task.Date)
	if err != nil {
		utils.SendError(w, "invalid date format", http.StatusBadRequest)
		return
	}
	if parseDate.Before(time.Now()) {
		// как в создании задачи
		task.Date = time.Now().Format(utils.DateFormat)
	}
	if len(task.Title) == 0 {
		utils.SendError(w, "invalid title", http.StatusBadRequest)
		return
	}
	if len(task.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			utils.SendError(w, "invalid repeat format", http.StatusBadRequest)
			return
		}
	}

	_, err = database.Update(db, &task)
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
	if task.Repeat != "" {
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

<<<<<<< HEAD
	if task.Repeat == "" {
		err = database.Delete(task.ID)
=======
	task, err := database.GetpoID(db, strconv.Itoa(idTask))
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
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
		if err != nil {
			utils.SendError(w, "failed to delete task", http.StatusInternalServerError)
			return
		}
<<<<<<< HEAD
		//log.Println(fmt.Sprintf("task with id=%s was deleted", task.ID))
	} else {
		task.Date, err = nextdate.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			utils.SendError(w, "failed to get next date", http.StatusInternalServerError)
			return
=======
		nextDate, _ := time.Parse("20060102", nextDateTask)
		if nextDate.Before(time.Now()) {
			utils.SendError(w, "Следующая дата задачи находится в прошлом", http.StatusBadRequest)
			return
		}
		if len(task.Repeat) > 0 {
			task.Repeat = nextdate.NormalizeRepeatFormat(task.Repeat)
			if err := nextdate.ValidateRepeatFormat(task.Repeat); err != nil {
				log.Printf("Ошибка валидации формата повтора: %v", err)
				utils.SendError(w, fmt.Sprintf("Неверный формат повтора: %v", err), http.StatusBadRequest)
				return
			}
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
		}
		// Обновляем задачу с новой датой
		_, err = database.Update(db, &task)
		if err != nil {
			utils.SendError(w, "failed to update task", http.StatusInternalServerError)
			return
		}
	}

	//utils.SendJSON(w, http.StatusOK, task)
	//utils.SendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
	//w.Header().Set("Content-Type", "application/json")
	//w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	//utils.SendJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	w.WriteHeader(http.StatusNoContent)
	log.Printf("Deleted task with id=%s", id)
}

// функция для валидации задачи
/*func validateTaskInput(task *moduls.Scheduler) error {
	if task.ID == "" {
		return fmt.Errorf("ID задачи не может быть пустым")
	}
	if task.Title == "" {
		return fmt.Errorf("заголовок задачи не может быть пустым")
	}
	taskDate, err := time.Parse(utils.DateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("неверный формат даты")
	}
	if taskDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return fmt.Errorf("дата не может быть меньше сегодняшней")
	}
	if task.Repeat != "" {
		if err := nextdate.ValidateRepeatFormat(task.Repeat); err != nil {
			return err
		}
	}
	return nil
}*/

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
