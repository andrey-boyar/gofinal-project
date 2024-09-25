package config

import (
	//"encoding/json"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string `json:"port"`
	DBFile    string `json:"db_file"`
	JWTSecret string `json:"jwt_secret"`
	Password  string `json:"password"`
}

func LoadConfig(filename string) (*Config, error) {
	err := godotenv.Load(filename)
	if err != nil {
		return nil, err
	}

	config := &Config{
		Port:      os.Getenv("TODO_PORT"),
		DBFile:    os.Getenv("TODO_DBFILE"),
		JWTSecret: os.Getenv("TODO_JWT_SECRET"),
		Password:  os.Getenv("TODO_PASSWORD"),
	}

	return config, nil
}
