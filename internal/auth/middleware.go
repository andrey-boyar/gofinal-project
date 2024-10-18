package auth

import (
	"log"
	"net/http"

	//"os"
	"bytes"
	//"github.com/golang-jwt/jwt/v5"
	//"github.com/joho/godotenv"
)

type responseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	// w.body.Write(b)
	if _, err := w.body.Write(b); err != nil {
		log.Printf("Ошибка при записи тела ответа: %v", err)
	}
	log.Printf("Ответ: %s", w.body.String())
	return w.ResponseWriter.Write(b)
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lrw := &responseWriter{ResponseWriter: w, body: &bytes.Buffer{}}
		log.Printf("Запрос: %s", r.URL.Path)
		next.ServeHTTP(lrw, r)
		log.Printf("Ответ: %s", lrw.body.String())
	}
}
