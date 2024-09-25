package token

import (
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtKey []byte

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Ошибка загрузки .env файла")
	}
	jwtKey = []byte(os.Getenv("TODO_JWT_SECRET"))
	if len(jwtKey) == 0 {
		log.Println("JWT ключ не установлен в переменных окружения")
	}
}

type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
				return
			}

			tokenStr := cookie.Value
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})

			if err != nil || !token.Valid || claims.PasswordHash != pass {
				http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
