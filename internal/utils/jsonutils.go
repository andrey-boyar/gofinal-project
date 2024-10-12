package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// DecodeJSON декодирует JSON из тела запроса.
func DecodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if r.Header.Get("Content-Type") != "application/json" {
		SendError(w, "Неверный Content-Type", nil)
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(v); err != nil {
		//log.Printf("err: %v", err)
		SendError(w, "Ошибка декодирования JSON", err)
		return err
	}
	return nil
}

// SendJSON отправляет JSON-ответ клиенту.
func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Ошибка при сериализации JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

//json.NewEncoder(w).Encode(data)

func SendError(w http.ResponseWriter, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	errorMessage := message
	if err != nil {
		errorMessage += ": " + err.Error()
	}
	json.NewEncoder(w).Encode(map[string]string{"error": errorMessage})
}

// SendError отправляет JSON-ответ с ошибкой
/*func SendError(w http.ResponseWriter, s string, err error) {
	//формируем структуру с ошибкой
	error := moduls.Errors{
		Errors: fmt.Errorf("%s", s).Error()}
	//формируем ответ
	errorData, _ := json.Marshal(error)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	// Пишем ответ
	_, writeErr := w.Write(errorData)
	if writeErr != nil {
		http.Error(w, fmt.Errorf("error: %w", writeErr).Error(), http.StatusBadRequest)
	}
	//return nil
}*/

// HandleError обрабатывает ошибки и отправляет ответ клиенту
//func HandleError(w http.ResponseWriter, message string, err error) {
//log.Printf("%s: %v", message, err)
//errorMessage := fmt.Sprintf("%s: %v", message, err)
//	SendError(w, errorMessage)
//}
