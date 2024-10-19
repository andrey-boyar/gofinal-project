package tasks

import (
<<<<<<< HEAD
	"log"
=======
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
	"net/http"
	"time"

	"final-project/internal/nextdate"
	"final-project/internal/utils"
)

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметр "now" из запроса и парсим его
	now, err := time.Parse(utils.DateFormat, r.FormValue("now"))
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Получаем параметры "date" и "repeat" из запроса
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Вычисляем следующую дату с помощью функции NextDate
	nextDate, err := nextdate.NextDate(now, date, repeat)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Возвращаем результат в ответе
<<<<<<< HEAD
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(nextDate))

	if err != nil {
		log.Printf("writing tasks data error: %v", err)
	}
=======
	w.Write([]byte(nextDate))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(map[string]string{"next_date": nextDate})
	// utils.SendError(w, http.StatusOK, map[string]string{"next_date": nextDate})
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
}
