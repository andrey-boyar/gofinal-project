package utils

import (
	"encoding/json"
	moduls "final-project/internal/moduls"
	"fmt"
	"log"
	"net/http"
)

// DecodeJSON декодирует JSON из тела запроса.
func DecodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if r.Header.Get("Content-Type") != "application/json" {
		SendError(w, "Неверный Content-Type", http.StatusBadRequest)
		return fmt.Errorf("неверный Content-Type")
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(v); err != nil {
		SendError(w, "Ошибка декодирования JSON", http.StatusBadRequest)
		return fmt.Errorf("ошибка декодирования JSON: %w", err)
	}
	return nil
}

// SendJSON отправляет JSON-ответ клиенту.
func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	// Если data это слайс Scheduler, оборачиваем его в TaskResponse
	if tasks, ok := data.([]moduls.Scheduler); ok {
		response := moduls.SchedulerList{
			Tasks: tasks,
		}
		data = response
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		SendError(w, "Ошибка кодирования JSON", http.StatusInternalServerError)
		return
	}
}

// SendError отправляет ошибку клиенту.
func SendError(w http.ResponseWriter, message string, statusCode int) {
	log.Printf("Отправка ошибки: %s (код: %d)", message, statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
