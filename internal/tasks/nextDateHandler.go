package tasks

import (
	"encoding/json"
	"final-project/internal/nextdate"
	"final-project/internal/utils"
	"net/http"
	"time"
)

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметр "now" из запроса и парсим его
	now, err := time.Parse(utils.DateFormat, r.FormValue("now"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Получаем параметры "date" и "repeat" из запроса
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Вычисляем следующую дату с помощью функции NextDate
	nextDate, err := nextdate.NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Возвращаем результат в ответе
	/*w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(nextDate))

	if err != nil {
		log.Printf("writing tasks data error: %v", err)
	}*/
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"next_date": nextDate})
	//utils.SendError(w, http.StatusOK, map[string]string{"next_date": nextDate})
}

// NextDateHandler обрабатывает GET-запросы для вычисления следующей даты задачи
/*func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	log.Printf("Получен запрос к /api/nextdate: %s", r.URL.String())

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		utils.SendError(w, `{"error": "Required parameters are missing"}`, http.StatusBadRequest)
		log.Printf("Необходимые параметры отсутствуют")
		return
	}

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		utils.SendError(w, fmt.Sprintf(`{"error": "Incorrect date format 'now': %s"}`, err), http.StatusBadRequest)
		log.Printf("Ошибка в формате даты: %v", err)
		return
	}

	_, err = time.Parse("20060102", dateStr)
	if err != nil {
		utils.SendError(w, fmt.Sprintf(`{"error": "Incorrect date format 'date': %s"}`, err), http.StatusBadRequest)
		log.Printf("Ошибка в формате даты: %v", err)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}*/
