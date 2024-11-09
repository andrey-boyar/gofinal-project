package tasks

import (
	"database/sql"
	"encoding/json"
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
			return
		}
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// GetTasksHandler получает задачи или все задачи, если фильтры не указаны
func GetTasksHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	log.Printf("GetTasksHandler вызван с параметром search: %s", search)

	// 1. Сначала проверяем пустой поиск
	if search == "" {
		tasks, err := database.ReadTask(db, "")
		if err != nil {
			log.Printf("Ошибка при получении всех задач: %v", err)
			utils.SendError(w, "Ошибка при получении задач", http.StatusInternalServerError)
			return
		}
		utils.SendJSON(w, http.StatusOK, map[string]interface{}{
			"tasks": tasks,
		})
		return
	}

	// 2. Затем проверяем, является ли поиск датой
	if isDateFormat(search) {
		formattedDate := convertDateFormat(search)
		tasks, err := database.SearchDate(db, formattedDate)
		if err != nil {
			log.Printf("Ошибка при поиске по дате: %v", err)
			utils.SendError(w, "Ошибка при поиске по дате", http.StatusInternalServerError)
			return
		}
		utils.SendJSON(w, http.StatusOK, map[string]interface{}{
			"tasks": tasks,
		})
		return
	}
	// 3. Если это не дата - значит это текстовый поиск
	tasks, err := database.Searchtitl(db, search)
	if err != nil {
		log.Printf("Ошибка при поиске по названию: %v", err)
		utils.SendError(w, "Ошибка при поиске по названию", http.StatusInternalServerError)
		return
	}
	utils.SendJSON(w, http.StatusOK, map[string]interface{}{
		"tasks": tasks,
	})
}

// Проверка формата даты
func isDateFormat(s string) bool {
	_, err := time.Parse("02.01.2006", s)
	//return err == nil
	if err != nil {
		log.Printf("Строка '%s' не является датой: %v", s, err)
		return false
	}
	log.Printf("Строка '%s' является корректной датой", s)
	return true
}

// Конвертация формата даты
func convertDateFormat(date string) string {
	t, err := time.Parse("02.01.2006", date)
	//return t.Format("20060102")
	if err != nil {
		log.Printf("Ошибка конвертации даты '%s': %v", date, err)
		return ""
	}
	formatted := t.Format("20060102")
	log.Printf("Дата '%s' преобразована в '%s'", date, formatted)
	return formatted
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
