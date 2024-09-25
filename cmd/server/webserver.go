package main

import (
	"log"
	"net/http"
	"os"

	"final-project/internal/config"
	"final-project/internal/database"
	"final-project/internal/tasks"
	"final-project/internal/token"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	defaultPort = "7540"
	webDir      = "./web"
)

func main() {
	// Загрузка конфигурации
	_, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Получение порта из переменной окружения
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}
	// Инициализация базы данных
	db := database.InitDatabase()

	// Создаем новый роутер Chi
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Обслуживаем статические файлы
	fileServer := http.FileServer(http.Dir(webDir))
	r.Handle("/*", http.StripPrefix("/", fileServer))

	// Обработчик запроса /api/nextdate
	r.Get("/api/nextdate", token.Auth(tasks.NextDateHandler))
	// Обработчик запроса /api/task
	r.Method("GET", "/api/task", token.Auth(tasks.TaskHandler))
	r.Method("POST", "/api/task", token.Auth(tasks.TaskHandler))
	r.Method("PUT", "/api/task", token.Auth(tasks.TaskHandler))
	r.Method("DELETE", "/api/task", token.Auth(tasks.TaskHandler))
	// Обработчик запроса /api/search
	r.Get("/api/search", func(w http.ResponseWriter, r *http.Request) {
		tasks.SearchHandler(w, r, db)
	})
	// Обработчик запроса /api/sign
	r.Post("/api/sign", token.Auth(tasks.HandleSign))

	// Добавляем обработчик для /api/tasks
	r.Get("/api/tasks", token.Auth(tasks.TaskHandler))
	// Добавляем обработчик для /api/task/done
	handleTaskDoneWrapper := func(w http.ResponseWriter, r *http.Request) {
		tasks.HandleTaskDone(w, r, db)
	}
	// Используем обертку вместо оригинальной функции
	r.Put("/api/task/done", token.Auth(handleTaskDoneWrapper))

	if os.Getenv("TODO_JWT_SECRET") == "" {
		log.Fatal("TODO_JWT_SECRET не установлен в переменных окружения")
	}

	// Запуск сервера на указанном порту
	log.Printf("Запуск сервера на порту %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
