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
		handleTaskPost(w, r)
	//case http.MethodGet:
	//search(w, r, db)
	case http.MethodPut:
		handleTaskPut(w, r, db)
	case http.MethodDelete:
		handleTaskDelete(w, r)
	case http.MethodGet:
		id := r.URL.Query().Get("id")
		if id != "" {
			// Получение конкретной задачи по ID
			task, err := database.GetpoID(id)
			if err != nil {
				utils.SendError(w, "Ошибка при получении задачи", err)
				return
			}
			utils.SendJSON(w, http.StatusOK, task)
		} else {
			// Поиск задач
			search(w, r, db)
		}
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// Функция для добавления задачи
func handleTaskPost(w http.ResponseWriter, r *http.Request) {
	log.Println("saving task started")
	defer log.Println("saving task finished")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	//(w http.ResponseWriter, r *http.Request) {
	// Проверка Content-Type
	var taskData moduls.Scheduler
	if err := utils.DecodeJSON(w, r, &taskData); err != nil {
		log.Printf("Ошибка при декодировании JSON: %v", err)
		utils.SendError(w, "Ошибка при декодировании JSON", err)
		return
	}
	log.Printf("входные данные: %+v", taskData)
	// Декодирование JSON тела запроса
	//if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
	//utils.SendError(w, "Ошибка декодирования JSON", err)
	//	return
	//}
	// Установка даты по умолчанию или проверка формата даты
	if len(taskData.Date) == 0 {
		taskData.Date = time.Now().Format(utils.DateFormat)
	} else {
		//проверка формата даты
		date, err := time.Parse(utils.DateFormat, taskData.Date)
		if err != nil {
			utils.SendError(w, "Неверный формат даты", err)
			return
		}
		// если дата в прошлом, то ставим текущую дату
		if date.Before(time.Now()) {
			taskData.Date = time.Now().Format(utils.DateFormat)
		}
	}
	///if taskData.Title == "" {
	//utils.SendError(w, "Не указан заголовок задачи", nil)
	//return
	//}
	// Проверка заголовка задачи
	if len(taskData.Title) == 0 {
		utils.SendError(w, "Не указан заголовок задачи", nil)
		return
	}
	// Проверка формата повтора
	if len(taskData.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), taskData.Date, taskData.Repeat); err != nil {
			utils.SendError(w, "Неверный формат повтора", err)
			return
		}
	}
	// Добавление задачи в базу данных
	taskId, err := database.Create(&taskData)
	if err != nil {
		utils.SendError(w, "Ошибка создания задачи", err)
		return
	}
	// Возвращение ID созданной задачи
	utils.SendJSON(w, http.StatusCreated, moduls.TaskId{Id: int(taskId)})
	//json.NewEncoder(w).Encode(moduls.TaskId{Id: int(taskId)})
	log.Printf("Added task with id=%d", taskId)
}

// функция для поиска задач
func search(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	tasks, err := GetTasks(db, search, "")
	if err != nil {
		utils.SendError(w, "Ошибка получения задачи", err)
		return
	}
	//tasks, err = database.ReadTask(db)
	//if err != nil {
	//	utils.SendError(w, "Ошибка получения задач", err)
	//	return
	//}
	// Отправка ответа
	//json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
	utils.SendJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})

	log.Printf("Read %d tasks", len(tasks))
}

// handleTaskPut обновляет задачу
func handleTaskPut(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task moduls.Scheduler

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.SendError(w, "JSON deserialization error", err)
		return
	}
	if err := validateTaskInput(&task); err != nil {
		utils.SendError(w, "Invalid input", err)
		return
	}
	//проверка id
	if len(task.ID) == 0 {
		utils.SendError(w, "invalid id", nil)
		return
	}
	//проверка id на число
	if _, err := strconv.Atoi(task.ID); err != nil {
		utils.SendError(w, "invalid id", err)
		return
	}
	//проверка даты
	parseDate, err := time.Parse(utils.DateFormat, task.Date)
	if err != nil {
		utils.SendError(w, "invalid date format", err)
		return
	}
	//если дата в прошлом, то ставим текущую дату
	if parseDate.Before(time.Now()) {
		// как в создании задачи
		task.Date = time.Now().Format(utils.DateFormat)
	}
	//проверка заголовка
	if len(task.Title) == 0 {
		utils.SendError(w, "invalid title", nil)
		return
	}
	//проверка формата повтора
	if len(task.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			utils.SendError(w, "invalid repeat format", err)
			return
		}
	}
	//обновление задачи
	updatedTask, err := database.Update(db, &task)
	if err != nil {
		utils.SendError(w, "failed to update task", err)
		return
	}
	utils.SendJSON(w, http.StatusOK, updatedTask)
	utils.SendJSON(w, http.StatusOK, task)
}

/*w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedTask); err != nil {
		utils.SendError(w, "failed to encode response", err)
		return
	}
	log.Printf("Updated task with id=%s", task.ID)
}*/

// GetTaskByID получает задачу по ID
func GetTaskByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	task, err := database.GetpoID(id)
	if err != nil {
		utils.SendError(w, "failed to get task by id", err)
		return
	}
	// Отправка ответа
	utils.SendJSON(w, http.StatusOK, task)
	log.Printf("Read task with id=%s", id)
}

// UpdateTask обновляет задачу в базе данных
func UpdateTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task moduls.Scheduler
	//декодирование JSON тела запроса
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.SendError(w, "JSON deserialization error", err)
		return
	}
	//проверка id
	if len(task.ID) == 0 {
		utils.SendError(w, "invalid id", nil)
		return
	}
	//проверка id на число
	if _, err := strconv.Atoi(task.ID); err != nil {
		utils.SendError(w, "invalid id", err)
		return
	}
	//проверка даты
	parseDate, err := time.Parse(utils.DateFormat, task.Date)
	if err != nil {
		utils.SendError(w, "invalid date format", err)
		return
	}
	//если дата в прошлом, то ставим текущую дату
	if parseDate.Before(time.Now()) {
		// как в создании задачи
		task.Date = time.Now().Format(utils.DateFormat)

	}
	//проверка заголовка
	if len(task.Title) == 0 {
		utils.SendError(w, "invalid title", nil)
		return
	}
	//проверка формата повтора
	if len(task.Repeat) > 0 {
		if _, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			utils.SendError(w, "invalid repeat format", err)
			return
		}
	}
	//обновление задачи
	_, err = database.Update(db, &task)
	if err != nil {
		utils.SendError(w, "failed to update task", err)
		return
	}
	utils.SendJSON(w, http.StatusOK, task)
}

/*utils.SendJSON(w, http.StatusOK, task)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		utils.SendError(w, "failed to encode response", err)
		return
	}
}*/

// HandleTaskDone обрабатывает запрос на выполнение задачи
func HandleTaskDone(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("API: Завершение задачи")
	//w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		log.Println("API: Не передан id задачи")
		w.WriteHeader(http.StatusBadRequest)
		utils.SendError(w, "Не передан id задачи", nil)
		return
	}

	idTask, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("Неверный формат id задачи: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		utils.SendError(w, "Неверный формат id задачи", err)
		return
	}

	task, err := database.GetpoID(strconv.Itoa(idTask))
	if err != nil {
		log.Printf("Ошибка при получении задачи: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendError(w, "Ошибка при получении задачи", err)
		return
	}

	if task.Repeat != "" {
		log.Printf("Задача повторяется: %s", task.Repeat)
		baseDate, _ := time.Parse("20060102", task.Date)
		nextDateTask, err := nextdate.NextDate(baseDate, task.Date, task.Repeat)
		if err != nil {
			log.Printf("Ошибка при получении следующей даты задачи: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendError(w, "Ошибка при получении следующей даты задачи", err)
			return
		}
		task.Date = nextDateTask
		_, err = database.Update(db, &task)
		if err != nil {
			log.Printf("Ошибка при обновлении задачи: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendError(w, "Ошибка при обновлении задачи", err)
			return
		}
	} else {
		log.Println("Задача не повторяется, удаляем задачу")
		err = database.Delete(strconv.Itoa(idTask))
		if err != nil {
			log.Printf("Ошибка при удалении задачи: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.SendError(w, "Ошибка при удалении задачи", err)
			return
		}
	}
	// Отправляем JSON-ответ
	//w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleTaskDelete удаляет задачу
func handleTaskDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	//удаление задачи
	err := database.Delete(id)
	if err != nil {
		utils.SendError(w, "failed to delete task", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	/*w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct{}{}); err != nil {
		utils.SendError(w, "failed to encode response", err)
		return
	}*/
	log.Printf("Deleted task with id=%s", id)
}

// функция для поиска задач в базе данных заголовка или комментария
// func searchTask(db *sql.DB, search string) ([]moduls.Scheduler, error) {
//
//	if len(search) > 0 {
//		if date, err := time.Parse("02.01.2006", search); err == nil {
//
// return database.SearchDate(db, date.Format(utils.DateFormat))
//
//	}
//
// return database.Searchtitl(db, search)
//
//	}
//
// /	return database.ReadTask(db)
// }
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

/*if task.Repeat == "" {
		// Если у задачи нет повторения, удаляем её
		if err := handleTaskDelete(db, id); err != nil {
			utils.HandleError(w, "Ошибка удаления задачи", err, http.StatusInternalServerError)
			return
		}
		log.Printf("Задача с ID %s успешно удалена", id)
	} else {
		// Если у задачи есть повторение, создаём новую задачу и удаляем старую
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			utils.HandleError(w, "Ошибка вычисления следующей даты", err, http.StatusInternalServerError)
			return
		}
		log.Printf("Следующая дата: %s", nextDate)
		newTask := moduls.Scheduler{
			ID:      "",
			Date:    nextDate,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		}

		//newTask := *task // Создаем копию задачи
		newTask.ID = "" // Сбрасываем ID для создания новой задачи
		newTask.Date = nextDate
		_, err = insertTask(db, newTask)
		if err != nil {
			log.Printf("Ошибка создания новой повторяющейся задачи: %v", err)
			utils.SendError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//if err := Create(db, &newTask); err != nil {
		//	utils.HandleError(w, "Ошибка создания новой повторяющейся задачи", err, http.StatusInternalServerError)
		//	return
		//}

		if err := handleTaskDelete(db, id); err != nil {
			if err == sql.ErrNoRows {
				utils.SendError(w, "Задача не найдена", http.StatusNotFound)
			} else {
				utils.SendError(w, "Ошибка удаления задачи", http.StatusInternalServerError)
			}
			return
		}
	}
	log.Printf("Задача с ID %s успешно выполнена", id)
	w.WriteHeader(http.StatusOK)
	//utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Задача выполнена"})
}
// HandleTaskDone обрабатывает запрос на выполнение задачи
func HandleTaskDone(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	task, err := database.GetpoID(id)
	if id == "" {
		utils.SendError(w, "Не указан ID задачи", nil)
		return
	}

	if err != nil {
		//if err.Error() == "task not found" {
		//	utils.SendError(w, "Задача не найдена")
		//} else {
		utils.SendError(w, "Ошибка при получении задачи", err)
		//}
		return
	}
	//если нет повтора, то удаляем задачу
	if task.Repeat == "" {
		err = database.Delete(id)
		if err != nil {
			utils.SendError(w, "failed to delete task", err)
			return
		}
		log.Printf("task with id=%s was deleted", task.ID)
	} //else {

	log.Printf("HandleTaskDone: current task date: %s, repeat: %s", task.Date, task.Repeat)
	//получаем следующую дату
	nextDate, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat)
	log.Printf("HandleTaskDone: next date calculated: %s", nextDate)
	if err != nil {
		utils.SendError(w, "failed to get next date", err)
		return
	}
	task.Date = nextDate
	// Обновляем задачу с новой датой
	updatedTask, err := database.Update(db, &task)
	if err != nil {
		utils.SendError(w, "failed to update task", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTask)

	//utils.SendJSON(w, http.StatusOK, updatedTask)
	//}
	// Отправка ответа
	//utils.SendJSON(w, http.StatusOK, task)
	//if err := json.NewEncoder(w).Encode(struct{}{}); err != nil {
	//utils.SendError(w, "failed to encode response", err)
	//return
	//}
	log.Printf("Updated task with id=%s", task.ID)
}
// Функция для создания задачи
func Create(db *sql.DB, task *moduls.Scheduler) error {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	stmt, err := db.Prepare(query)
	log.Printf("Создание задачи с ID: %s", task.ID)
	if err != nil {
		log.Printf("Ошибка подготовки запроса: %v", err)
		return fmt.Errorf("ошибка подготовки запроса: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("ошибка получения ID новой задачи: %w", err)
	}

	task.ID = strconv.FormatInt(id, 10)
	log.Printf("Добавлена задача с id=%s", task.ID)
	return nil
}

/*if task.Repeat == "" {
		if err := handleTaskDelete(db, task.ID); err != nil {
			log.Printf("Ошибка удаления задачи: %v", err)
			utils.SendError(w, "Ошибка удаления задачи", http.StatusInternalServerError)
			return
		}
		log.Printf("Задача с id=%s была удалена", task.ID)
	} else {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			log.Printf("Ошибка вычисления следующей даты: %v", err)
			utils.SendError(w, "Ошибка обработки повторяющейся задачи", http.StatusInternalServerError)
			return
		}
		task.Date = nextDate
		if err := database.Update(db, &task); err != nil {
			log.Printf("Ошибка обновления задачи: %v", err)
			utils.SendError(w, "Ошибка обновления задачи", http.StatusInternalServerError)
			return
		}
		log.Printf("Обновлена задача с id=%s", task.ID)
	}
	w.WriteHeader(http.StatusOK)
	//utils.SendError(w, "ok", http.StatusOK)
}

/*log.Printf("Получен запрос к HandleTaskDone: %s", r.URL.String())
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.SendError(w, "ID задачи не указан", http.StatusBadRequest)
		return
	}

	task, err := GetTaskByID(db, id)
	if err != nil {
		log.Printf("Ошибка при получении задачи: %v", err)
		utils.SendError(w, "Задача не найдена", http.StatusNotFound)
		return
	}
	if task == nil {
		utils.SendError(w, "Задача не найдена", http.StatusNotFound)
		return
	}
	if task.Repeat != "" {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			utils.SendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		task.Date = nextDate
		if err := UpdateTask(db, task); err != nil {
			utils.SendError(w, "Ошибка обновления задачи", http.StatusInternalServerError)
			return
		}
	} else {
		if err := handleTaskDelete(db, id); err != nil {
			utils.SendError(w, "Ошибка удаления задачи", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func handleTaskDelete(db *sql.DB, id string) error {
	//_, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	//return err}
	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	log.Printf("Удалена задача с id=%s", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("задача с id=%s не найдена", id) //sql.ErrNoRows
	}

	return nil
}

// SearchHandler обрабатывает запрос на поиск задач
func SearchHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	tasks, err := GetTasks(db, r.URL.Query().Get("search"), r.URL.Query().Get("date"))
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Если задач нет, отправляем пустой массив вместо null
	if tasks == nil {
		tasks = []moduls.Scheduler{}
	}
	//w.WriteHeader(http.StatusOK)
	utils.SendJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
	//w.Header().Set("Content-Type", "application/json")
}

// insertTask добавляет задачу в базу данных.
func insertTask(db *sql.DB, task moduls.Scheduler) (int64, error) {
	if task.Repeat != "" {
		if err := ValidateRepeatFormat(task.Repeat); err != nil {
			return 0, fmt.Errorf("неверный формат повтора: %v", err)
		}
	}
	result, err := db.Exec(`
        INSERT INTO scheduler (date, title, comment, repeat)
        VALUES (?, ?, ?, ?)
    `, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		log.Printf("Ошибка добавления задачи: %v", err)
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			log.Printf("MySQL error number: %d", mysqlErr.Number)
			switch mysqlErr.Number {
			case 1062:
				return 0, fmt.Errorf("задача с таким заголовком уже существует")
			case 1406:
				return 0, fmt.Errorf("слишком длинное значение для одного из полей")
			}
		}
		return 0, fmt.Errorf("ошибка при добавлении задачи в базу данных")
		//return 0, err
	}
	id, err := result.LastInsertId()
	//log.Printf("ID вставленной задачи: %d", id)
	if err != nil {
		log.Printf("Ошибка при получении ID вставленной задачи: %v", err)
		return 0, err
	}
	log.Printf("Задача успешно вставлена с ID: %d", id)
	return id, nil
}*/

// Функция для получения задач с возможностью фильтрации
func GetTasks(db *sql.DB, titleFilter string, dateFilter string) ([]moduls.Scheduler, error) {
	log.Printf("Получение задач с фильтрами: title=%s, date=%s", titleFilter, dateFilter)
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE 1=1`
	args := []interface{}{}

	// Фильтрация по заголовку
	if titleFilter != "" {
		log.Printf("Фильтрация по заголовку: %s", titleFilter)
		query += " AND title LIKE ?"
		args = append(args, "%"+titleFilter+"%")
		//query += " AND (title LIKE ? OR comment LIKE ?)"
		//args = append(args, "%"+titleFilter+"%", "%"+titleFilter+"%")
	}

	// Фильтрация по дате
	if dateFilter != "" {
		log.Printf("Фильтрация по дате: %s", dateFilter)
		query += " AND date = ?"
		args = append(args, dateFilter)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		return nil, fmt.Errorf("ошибка получения задач: %w", err)
	}
	defer rows.Close()

	var tasks []moduls.Scheduler
	for rows.Next() {
		var task moduls.Scheduler
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			log.Printf("Ошибка сканирования задачи: %v", err)
			return nil, fmt.Errorf("ошибка сканирования задачи: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при итерации по строкам: %v", err)
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}
	log.Printf("Получено задач: %d", len(tasks))

	// Добавим проверку на пустой результат
	if len(tasks) == 0 {
		return []moduls.Scheduler{}, nil // Возвращаем пустой слайс вместо nil
	}
	return tasks, nil
}
