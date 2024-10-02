package config

import (
	//"encoding/json"
	"os"

	"github.com/joho/godotenv"
)

// структура конфигурации
type Config struct {
	Port      string `json:"port"`
	DBFile    string `json:"db_file"`
	JWTSecret string `json:"jwt_secret"`
	Password  string `json:"password"`
	TestEnv   string `json:"test_env"`
}

// Функция загрузки конфигурации
func LoadConfig(filename string) (*Config, error) {
	err := godotenv.Load(filename) //загрузка переменных окружения
	if err != nil {
		return nil, err
	}
	//Конфигурация
	config := &Config{
		Port:      os.Getenv("TODO_PORT"),
		DBFile:    os.Getenv("TODO_DBFILE"),
		JWTSecret: os.Getenv("TODO_JWT_SECRET"),
		Password:  os.Getenv("TODO_PASSWORD"),
	}

	return config, nil
}
