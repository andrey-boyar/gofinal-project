package utils

import (
	"log"
	"net/http"
)

// HandleError обрабатывает ошибки и отправляет ответ клиенту
func HandleError(w http.ResponseWriter, message string, err error, status int) {
	log.Printf("%s: %v", message, err)
	SendError(w, message, status)
}
