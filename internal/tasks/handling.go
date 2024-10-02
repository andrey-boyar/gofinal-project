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

var jwtKey []byte //ключ для jwt
// функция для загрузки переменных окружения
func init() {
	err := godotenv.Load(".env") //загрузка переменных окружения
	if err != nil {
		log.Println("Ошибка загрузки .env файла")
	}
	jwtKey = []byte(os.Getenv("TODO_JWT_SECRET")) //получение ключа из переменных окружения
	if len(jwtKey) == 0 {
		log.Println("JWT ключ не установлен в переменных окружения")
	}
}

// структура для хранения пароля
type Credentials struct {
	Password string `json:"password"`
}

// структура для хранения токена
type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

// функция для обработки запроса на вход
func HandleSign(w http.ResponseWriter, r *http.Request) {
	var creds Credentials                         //структура для хранения пароля
	err := json.NewDecoder(r.Body).Decode(&creds) //декодирование тела запроса
	if err != nil {
		http.Error(w, "Неверный запрос", http.StatusBadRequest)
		return
	}

	expectedPassword := os.Getenv("TODO_PASSWORD") //получение пароля из переменных окружения
	//log.Printf("установленный пароль: %s", expectedPassword)
	//log.Printf("Полученный пароль: %s", creds.Password)
	if creds.Password != expectedPassword {
		http.Error(w, `{"error": "Неверный пароль"}`, http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(8 * time.Hour) //время действия токена
	claims := &Claims{
		PasswordHash: creds.Password, // Можно использовать хэш пароля
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) //создание токена
	tokenString, err := token.SignedString(jwtKey)             //подпись токена
	if err != nil {
		http.Error(w, "Ошибка создания токена", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{ //установка куки
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	//json.NewEncoder(w).Encode(map[string]string{"token": tokenString}) //отправка токена
	if err := json.NewEncoder(w).Encode(map[string]string{"token": tokenString}); err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}
