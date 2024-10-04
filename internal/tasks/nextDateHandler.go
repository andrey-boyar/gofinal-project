package tasks

import (
	"final-project/internal/utils"
	"fmt"
	"log"
	"net/http"
	"time"
)

// NextDateHandler обрабатывает GET-запросы для вычисления следующей даты задачи.
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	log.Printf("Получен запрос к /api/nextdate: %s", r.URL.String())

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		utils.SendError(w, `{"error": "Required parameters are missing"}`, http.StatusBadRequest)
		return
	}

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		utils.SendError(w, fmt.Sprintf(`{"error": "Incorrect date format 'now': %s"}`, err), http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		utils.SendError(w, fmt.Sprintf(`{"error": %s"}`, err), http.StatusBadRequest)
		return
	}

	if _, err := w.Write([]byte(nextDate)); err != nil {
		log.Printf("Error when sending a reply: %v", err)
		utils.SendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
