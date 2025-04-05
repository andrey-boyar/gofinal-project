package auth

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Структура для записи ответа
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

// Write перехватывает запись ответа
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteHeader перехватывает установку статуса ответа
func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Структура для логирования запроса
type RequestLog struct {
	Method     string        `json:"method"`
	Path       string        `json:"path"`
	Status     int           `json:"status"`
	Duration   time.Duration `json:"duration"`
	RequestID  string        `json:"request_id"`
	UserAgent  string        `json:"user_agent"`
	RemoteAddr string        `json:"remote_addr"`
	Body       interface{}   `json:"body,omitempty"`
}

// Логирование запросов
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем буфер для тела ответа
		lrw := &responseWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
		}

		// Добавляем request_id в заголовки
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		r.Header.Set("X-Request-ID", requestID)

		// Обрабатываем запрос
		next.ServeHTTP(lrw, r)

		// Формируем лог
		reqLog := RequestLog{
			Method:     r.Method,
			Path:       r.URL.Path,
			Status:     lrw.statusCode,
			Duration:   time.Since(start),
			RequestID:  requestID,
			UserAgent:  r.UserAgent(),
			RemoteAddr: r.RemoteAddr,
		}

		// Логируем запрос
		logJSON, _ := json.Marshal(reqLog)
		log.Printf("Request: %s", string(logJSON))
	})
}

// generateRequestID генерирует уникальный ID запроса
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString генерирует случайную строку заданной длины
func randomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// AuthMiddleware проверяет аутентификацию
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем проверку для публичных эндпоинтов
		if isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем токен
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// TODO: Добавить проверку JWT токена
		// if !validateToken(token) {
		//     http.Error(w, "Invalid token", http.StatusUnauthorized)
		//     return
		// }

		next.ServeHTTP(w, r)
	})
}

// isPublicEndpoint проверяет, является ли эндпоинт публичным
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/api/health",
		"/api/login",
		"/api/register",
	}

	for _, pp := range publicPaths {
		if path == pp {
			return true
		}
	}
	return false
}
