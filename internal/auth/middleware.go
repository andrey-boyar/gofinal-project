package auth

import (
	"net/http"

	"bytes"
)

// Структура для записи ответа
type responseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

// Записывает тело ответа в буфер
func (w *responseWriter) Write(b []byte) (int, error) {
	// w.body.Write(b)
	if _, err := w.body.Write(b); err != nil {
		//log.Printf("Ошибка при записи тела ответа: %v", err)
	}
	//log.Printf("Ответ: %s", w.body.String())
	return w.ResponseWriter.Write(b)
}

// Логирование запросов
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lrw := &responseWriter{ResponseWriter: w, body: &bytes.Buffer{}}
		//log.Printf("Запрос: %s", r.URL.Path)
		next.ServeHTTP(lrw, r)
		//log.Printf("Ответ: %s", lrw.body.String())
	}
}
