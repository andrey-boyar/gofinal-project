package auth

import (
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtKey []byte //ключ для токена

// функция для загрузки переменных окружения
func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Ошибка загрузки .env файла")
	}
	jwtKey = []byte(os.Getenv("TODO_JWT_SECRET")) //получение ключа из переменных окружения
	if len(jwtKey) == 0 {
		log.Println("JWT ключ не установлен в переменных окружения")
	}
}

// тип для хранения данных токена
//type Claims struct {
//PasswordHash string `json:"password_hash"`
///jwt.RegisteredClaims
//}

// функция для аутентификации
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, является ли запрос к /api/nextdate
		if r.URL.Path == "/api/nextdate" {
			next.ServeHTTP(w, r)
			return
		}
		pass := os.Getenv("TODO_PASSWORD") //получение пароля из переменных окружения
		if len(pass) > 0 {
			cookie, err := r.Cookie("token") //получение токена из запроса
			if err != nil {
				http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
				return
			}

			tokenStr := cookie.Value //получение токена из куки
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})
			if err != nil || !token.Valid || claims.PasswordHash != pass {
				http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r) //вызов следующего обработчика
	})
}
