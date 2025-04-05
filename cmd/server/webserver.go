package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

type Server struct {
	router *chi.Mux
	db     *sql.DB
	port   string
}

func NewServer() (*Server, error) {
	// Загрузка конфигурации
	_, err := config.LoadConfig(".env")
	if err != nil {
		return nil, err
	}

	// Инициализация базы данных
	db := database.InitDatabase()
	if err := database.TestDatabaseConnection(db.DB); err != nil {
		return nil, err
	}

	// Создание роутера
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Настройка маршрутов
	router.SetupRouter(r, db)

	// Получение порта
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}

	return &Server{
		router: r,
		db:     db.DB,
		port:   port,
	}, nil
}

func (s *Server) Start() error {
	// Настройка статических файлов с кэшированием
	fs := http.FileServer(http.Dir(webDir))
	fileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Добавляем заголовки кэширования
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Header().Set("Expires", time.Now().AddDate(1, 0, 0).Format(time.RFC1123))
		fs.ServeHTTP(w, r)
	})
	s.router.Mount("/", fileServer)

	// Настройка graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера
	go func() {
		log.Printf("Запуск сервера на порту %s", s.port)
		if err := http.ListenAndServe(":"+s.port, s.router); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	<-stop
	log.Println("Получен сигнал завершения работы")
	return s.Shutdown()
}

// Закрытие базы данных
func (s *Server) Shutdown() error {
	if s.db != nil {
		s.db.Close()
	}
	return nil
}

// Основная функция
func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Ошибка инициализации сервера: %v", err)
	}

	// Запуск сервера
	if err := server.Start(); err != nil {
		log.Fatalf("Ошибка работы сервера: %v", err)
	}
}
