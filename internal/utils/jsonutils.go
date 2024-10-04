package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// DecodeJSON декодирует JSON из тела запроса.
func DecodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return SendError(w, "Неверный Content-Type", http.StatusUnsupportedMediaType)
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(v); err != nil {
		return SendError(w, "Ошибка декодирования JSON", http.StatusBadRequest)
	}
	return nil
}

// SendJSON отправляет JSON-ответ клиенту.
func SendJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		SendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}

// SendError отправляет JSON-ответ с ошибкой.
func SendError(w http.ResponseWriter, message string, status int) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(map[string]string{"error": message})
}
