package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"final-project/internal/nextdate"
)

// NextDateHandler обрабатывает GET-запросы для вычисления следующей даты задачи.
// Принимает параметры: now, date и repeat.
// Возвращает JSON с полем next_date или ошибку.
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Установка заголовка Content-Type для JSON
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Получаем параметры запроса
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")
	//nowStr := r.FormValue("now")
	//dateStr := r.FormValue("date")
	//repeat := r.FormValue("repeat")

	// Проверка наличия всех необходимых параметров
	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, `{"error": "Отсутствуют необходимые параметры"}`, http.StatusBadRequest)
		return
	}

	// Преобразуем строку даты в объект времени
	now, err := time.Parse("20060102", nowStr)
	// fmt.Println(now)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "неверный формат даты 'now': %s"}`, err), http.StatusBadRequest)
		return
	}

	// Вызываем функцию NextDate
	nextDate, err := nextdate.NextDate(now, dateStr, repeat)
	// nextDate, err := time.Parse("20060102", "20060102")
	//	fmt.Println(nextDate)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "ошибка: %s"}`, err), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"next_date": nextDate})
	// Возвращаем результат в формате JSON
	// response := map[string]string{"next_date": nextDate}
	// json.NewEncoder(w).Encode(response)
	// fmt.Fprintf(w, "%s", nextDate)
}
