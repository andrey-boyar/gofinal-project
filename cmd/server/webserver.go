package main

import (
	"log"
	"net/http"
	"os"

	"final-project/internal/auth"
	"final-project/internal/config"
	"final-project/internal/database"
	"final-project/internal/tasks"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	defaultPort = "7540"
	webDir      = "./web"
)

// Функция для установки кодировки UTF-8
//func setUTF8Middleware(next http.Handler) http.Handler {
//return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "application/json; charset=utf-8")
//	next.ServeHTTP(w, r)	}) }

func main() {
	// Загрузка конфигурации
	_, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Получение порта из переменной окружения
	port := os.Getenv("TODO_PORT")
	if port == "" { //если нет порта, то используем порт по умолчанию
		port = defaultPort
	}
	// Инициализация базы данных
	db := database.InitDatabase() // Функция для получения соединения с БД
	defer db.Close()              // Закрываем соединение с БД после завершения работы

	// Создаем новый роутер Chi
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)    // Логирование запросов
	r.Use(middleware.Recoverer) // Обработка паник
	//r.Use(setUTF8Middleware)    // Установка кодировки UTF-8

	// Обслуживаем статические файлы
	fileServer := http.FileServer(http.Dir(webDir))
	r.Handle("/*", http.StripPrefix("/", fileServer)) // Обработка статических файлов

	// Проверка наличия секретного ключа для JWT
	jwtSecret := os.Getenv("TODO_JWT_SECRET")
	userAuth := jwtSecret != ""

	/// проверка тестового окружения
	isTestEnv := os.Getenv("TEST_ENV") == "true"
	if os.Getenv("TEST_ENV") == "true" {
		log.Println("запуск в тестовом окружении")
	}

	// Обработчик запроса /api/nextdate
	if isTestEnv {
		r.Get("/api/nextdate", tasks.NextDateHandler)
	} else {
		r.Get("/api/nextdate", auth.Auth(tasks.NextDateHandler))
	}
	// Обработчик запроса /api/task
	if userAuth {
		r.Method("GET", "/api/task", auth.Auth(func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) }))
		r.Method("POST", "/api/task", auth.Auth(func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) }))
		r.Method("PUT", "/api/task", auth.Auth(func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) }))
		r.Method("DELETE", "/api/task", auth.Auth(func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) }))
	} else {
		r.Method("GET", "/api/task", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) }))
		r.Method("POST", "/api/task", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) }))
		r.Method("PUT", "/api/task", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) }))
		r.Method("DELETE", "/api/task", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) }))
	}
	// Обработчик запроса /api/search
	SearchHandler := func(w http.ResponseWriter, r *http.Request) {
		tasks.SearchHandler(w, r, db)
	}
	// Обработчик запроса /api/search
	if userAuth {
		r.Get("/api/search", auth.Auth(SearchHandler))
	} else {
		r.Get("/api/search", SearchHandler)
	}
	// Обработчик запроса /api/sign
	if userAuth {
		r.Post("/api/signin", auth.Auth(auth.HandleSign))
	} else {
		r.Post("/api/signin", auth.HandleSign)
	}

	// Добавляем обработчик для /api/tasks
	if userAuth {
		r.Get("/api/tasks", auth.Auth(func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) }))
	} else {
		r.Get("/api/tasks", func(w http.ResponseWriter, r *http.Request) { tasks.TaskHandler(w, r, db) })
	}

	// Добавляем обработчик для /api/task/done
	handleTaskDoneWrapper := func(w http.ResponseWriter, r *http.Request) {
		tasks.HandleTaskDone(w, r, db)
	}
	// Используем обертку вместо оригинальной функции
	if userAuth {
		r.Put("/api/task/done", auth.Auth(handleTaskDoneWrapper))
	} else {
		r.Put("/api/task/done", handleTaskDoneWrapper)
	}
	if os.Getenv("TODO_JWT_SECRET") == "" {
		log.Fatal("TODO_JWT_SECRET не установлен в переменных окружения")
	}

	// Запуск сервера на указанном порту
	log.Printf("Запуск сервера на порту %s", port)
	log.Printf("База данных: %v", db != nil)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
