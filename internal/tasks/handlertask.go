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
		id := r.URL.Query().Get("id")
		if id != "" {
			task, err := database.GetpoID(db, id)
			if err != nil {
				utils.SendError(w, err.Error(), http.StatusNotFound)
				return
			}
			utils.SendJSON(w, http.StatusOK, task)
		} else {
			utils.SendError(w, "Invalid request", http.StatusBadRequest) // Возвращаем ошибку, если id не указан
			//GetTasksHandler(w, r, db)                                    // Вызов GetTasksHandler для обработки запросов с фильтрами
			return
		}
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// GetTasksHandler получает задачи или все задачи, если фильтры не указаны
func GetTasksHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	dateFilter := r.URL.Query().Get("date")

	var tasks []moduls.Scheduler
	var err error

	// Если фильтры пустые, получаем все задачи
	if search == "" && dateFilter == "" {
		tasks, err = database.ReadTask(db, "") // Получаем все задачи
	} else {
		// Получаем задачи с фильтрами
		tasks, err = GetTasks(db, search, dateFilter)
	}

	if err != nil {
		log.Printf("Ошибка при получении задач: %v", err)
		utils.SendError(w, "Ошибка при получении задач", http.StatusInternalServerError)
		return
	}

	// Если задач нет, возвращаем пустой массив
	if tasks == nil {
		tasks = []moduls.Scheduler{}
	}

	// Устанавливаем заголовок и отправляем ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	utils.SendJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
}

// Функция для добавления задачи
func handleTaskPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
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
}

// функция для поиска задач
func SearchTasks(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	date := r.URL.Query().Get("date")

	// поиск задач
	tasks, err := searchDate(db, search, date)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// если задач нет, то возвращаем пустой слайс
	if tasks == nil {
		tasks = []moduls.Scheduler{}
	}

	// Устанавливаем заголовок и отправляем ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	utils.SendJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
}

// получение задач по дате или поиску
func searchDate(db *sql.DB, search string, dateFilter string) ([]moduls.Scheduler, error) {
	var tasks []moduls.Scheduler
	// Проверка на пустые фильтры
	if search == "" && dateFilter == "" {
		//return tasks, nil // Возвращаем пустой срез, если оба фильтра пустые
		return database.ReadTask(db, "") // Получаем все задачи
		//return []moduls.Scheduler{}, nil
	}
	if len(search) > 0 {
		log.Printf("Поиск задач для даты/поиска: %s", search)
		// Проверка формата даты
		if dateFilter != "" {
			formats := []string{"02.01.2006", "20060102", utils.DateFormat, utils.DateFormatDB}
			var validDate bool
			for _, format := range formats {
				if _, err := time.Parse(format, dateFilter); err == nil {
					validDate = true
					break
				}
			}
			if !validDate {
				return nil, fmt.Errorf("неверный формат даты: %s", dateFilter)
			}
		}
	}
	var err error
	//var tasks []moduls.Scheduler

	if search != "" {
		tasks, err = database.Searchtitl(db, search) // Поиск задач по заголовку
		if err != nil {
			return nil, err
		}
	}

	if dateFilter != "" {
		dateTasks, err := database.SearchDate(db, dateFilter) // Поиск задач по дате
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, dateTasks...)
	}
	// Удаление дубликатов, если необходимо
	tasks = removeDuplicates(tasks)

	// Проверка на наличие задач
	//if len(tasks) == 0 {
	//return nil, fmt.Errorf("задачи не найдены для заданных фильтров")
	//}
	return tasks, nil
}

// Функция для удаления дубликатов из среза задач
func removeDuplicates(tasks []moduls.Scheduler) []moduls.Scheduler {
	seen := make(map[string]struct{}) // Изменяем тип ключа на string
	result := []moduls.Scheduler{}

	for _, task := range tasks {
		if _, ok := seen[task.ID]; !ok {
			seen[task.ID] = struct{}{}
			result = append(result, task)
		}
	}
	return result
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

	// Возвращаем пустой ответ
	if err := json.NewEncoder(w).Encode(struct{}{}); err != nil {
		utils.SendError(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

// GetTasks получает задачи с фильтрами
func GetTasks(db *sql.DB, titleFilter string, dateFilter string) ([]moduls.Scheduler, error) {
	log.Printf("Получение задач с фильтрами: title=%s, date=%s", titleFilter, dateFilter)

	// Если фильтры пустые, возвращаем все задачи
	if titleFilter == "" && dateFilter == "" {
		return database.ReadTask(db, "") // Получаем все задачи
	}

	// Если есть хотя бы один фильтр, вызываем searchDate
	tasks, err := searchDate(db, titleFilter, dateFilter)
	if err != nil {
		return nil, err
	}

	return tasks, nil // Возвращаем результат searchDate
	// Если есть хотя бы один фильтр, вызываем searchDate
	//return searchDate(db, titleFilter, dateFilter)
}
