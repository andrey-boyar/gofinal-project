package router

import (
	"final-project/internal/auth"
	"final-project/internal/database"
	"final-project/internal/tasks"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	apiPrefix = "/api"
	taskPath  = "/task"
	tasksPath = "/tasks"
)

// SetupRouter настраивает маршруты для API
func SetupRouter(r *chi.Mux, db *database.DB) {
	// Добавляем глобальные middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(auth.LoggingMiddleware)
	r.Use(auth.AuthMiddleware)
	r.Use(middleware.Recoverer)

	// API маршруты
	r.Route(apiPrefix, func(r chi.Router) {
		// Маршруты для задач
		r.Route(taskPath, func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) })
			r.Post("/", func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) })
			r.Post("/done", func(w http.ResponseWriter, r *http.Request) { tasks.HandleTaskDone(w, r, db) })
		})

		// Дополнительные маршруты
		r.Get("/nextdate", tasks.NextDateHandler)
		r.Get("/tasks", func(w http.ResponseWriter, r *http.Request) { tasks.GetTasksHandler(w, r, db) })
		r.Get("/health", HealthCheckHandler(db))
	})
}

// HealthCheckHandler возвращает функцию-обработчик для проверки работоспособности
func HealthCheckHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			log.Printf("Ошибка проверки соединения с базой данных: %v", err)
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
