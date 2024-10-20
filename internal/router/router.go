package router

import (
	"database/sql"
	"final-project/internal/auth"
	"final-project/internal/tasks"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

// SetupRouter настраивает маршруты для API
func SetupRouter(r *chi.Mux, db *sql.DB) {
	//r := chi.NewRouter()
	// Добавляем глобальные middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth.LoggingMiddleware(next.ServeHTTP).ServeHTTP(w, r)
		})
	})

	// Маршруты для API
	r.Route("/api", func(r chi.Router) {
		r.Get("/nextdate", tasks.NextDateHandler)

		r.Route("/task", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) })
			r.Post("/", func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) })
			r.Post("/done", func(w http.ResponseWriter, r *http.Request) { tasks.HandleTaskDone(w, r, db) })
		})
		r.Get("/tasks", func(w http.ResponseWriter, r *http.Request) { tasks.GetTasksHandler(w, r, db) })
	})
	// Добавляем HealthCheckHandler
	r.Get("/api/health", HealthCheckHandler(db))
}

// HealthCheckHandler возвращает функцию-обработчик для проверки работоспособности
func HealthCheckHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Проверка соединения с базой данных")
		// Проверяем соединение с базой данных
		err := db.Ping()
		log.Printf("Ошибка проверки соединения с базой данных: %v", err)
		if err != nil {
			log.Printf("Ошибка проверки соединения с базой данных: %v", err)
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}

		// Если всё в порядке, отправляем успешный ответ
		w.WriteHeader(http.StatusOK)
		log.Println("База данных подключена")
		w.Write([]byte("OK"))
	}
}
