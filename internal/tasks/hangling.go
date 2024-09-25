package tasks

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	//"github.com/dgrijalva/jwt-go"
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

type Credentials struct {
	Password string `json:"password"`
}

type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

func HandleSign(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Неверный запрос", http.StatusBadRequest)
		return
	}

	expectedPassword := os.Getenv("TODO_PASSWORD")
	if creds.Password != expectedPassword {
		http.Error(w, `{"error": "Неверный пароль"}`, http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(8 * time.Hour)
	claims := &Claims{
		PasswordHash: creds.Password, // Можно использовать хэш пароля
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Ошибка создания токена", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
