# Используем образ Go для сборки
FROM golang:1.22 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы 
COPY . .

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=linux go build -o final-project cmd/server/webserver.go

# Используем минимальный образ для запуска
FROM ubuntu:latest

# Устанавливаем SQLite
RUN apt-get update && apt-get install -y sqlite3

# Копируем собранное приложение из предыдущего этапа
COPY --from=builder /app/final-project /app/final-project

# Копируем директорию web
COPY web /app/web

# Копируем .env файл
COPY .env /app/.env

# Устанавливаем рабочую директорию
WORKDIR /app

# Определяем переменные окружения
ENV TODO_PORT=7540
ENV TODO_DBFILE=./scheduler.db
ENV TODO_PASSWORD="12345"

# Открываем порт
EXPOSE 7540

# Запускаем приложение
CMD ["./final-project"]
