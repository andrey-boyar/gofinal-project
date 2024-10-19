package main

import (
	"log"
	"net/http"
	"os"

	"final-project/internal/config"
	"final-project/internal/database"

	//"final-project/internal/moduls"
	"final-project/internal/router"

	"github.com/go-chi/chi"
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
	// Проверяем, установлен ли пароль в переменной окружения TODO_PASSWORD
	//cfg := moduls.Config{
	//	Password: os.Getenv("TODO_PASSWORD"),
	//		Port:     defaultPort,
	//}
	//if cfg.Password == "" {
	//	log.Fatal("TODO_PASSWORD environment variable is required")
	//}

	// Инициализация базы данных
	db := database.InitDatabase()
	defer db.Close()

	// Выполняем тестовый запрос
	if err := database.TestDatabaseConnection(db); err != nil {
		log.Fatalf("Ошибка при выполнении тестового запроса к базе данных: %v", err)
	}

	// Создаем новый роутер
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Настраиваем роутер
	router.SetupRouter(r, db)

	// Обслуживаем статические файлы // Путь к директории с веб-файлами
	webServer := http.FileServer(http.Dir(webDir))
	// r.Handle("/*", http.StripPrefix("/", webServer))
	r.Mount("/", webServer)

	// Запуск сервера на указанном порту
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("Запуск сервера на порту %s", port)
	log.Printf("База данных: %v", db != nil)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
