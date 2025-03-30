package config

import (
	"fmt"
	"os"

	"final-project/internal/moduls"

	"github.com/joho/godotenv"
)

// Функция загрузки конфигурации
func LoadConfig(filename string) (*moduls.Config, error) {
	err := godotenv.Load(filename) // загрузка переменных окружения
	if err != nil {
		return nil, err
	}
	// Конфигурация
	config := &moduls.Config{
		Port:   os.Getenv("TODO_PORT"),
		DBFile: os.Getenv("TODO_DBFILE"),
		// JWTSecret: os.Getenv("TODO_JWT_SECRET"),
		// Password: os.Getenv("TODO_PASSWORD"),
	}

	// Проверка обязательных полей
	if config.Port == "" || config.DBFile == "" { //|| config.Password == "" {
		return nil, fmt.Errorf("отсутствуют обязательные переменные окружения")
	}
	return config, nil
}
