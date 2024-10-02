# Установка переменных
APP_NAME = scheduler-app
DOCKER_IMAGE = $(APP_NAME)
DOCKER_CONTAINER = $(APP_NAME)-container
PORT = 7540
DB_FILE = ./scheduler.db

# Установка переменных окружения
export TODO_PORT = $(PORT)
export TODO_DBFILE = $(DB_FILE)
export TODO_PASSWORD = 123456789
export TODO_JWT_SECRET = 1qaz2wsx3edc4rfv5tgb6yhn7ujm8ik9ol0p
# Правило по умолчанию
.PHONY: all
all: build run

# Сборка проекта
.PHONY: build
build:
	go build -o $(APP_NAME) cmd/server/webserver.go

# Запуск проекта
.PHONY: run
run:
	./$(APP_NAME)

# Запуск тестов
.PHONY: test
test:
	go test -v ./...

# Очистка скомпилированных файлов
.PHONY: clean
clean:
	rm -f $(APP_NAME)

# Сборка Docker-образа
.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Запуск Docker-контейнера
.PHONY: docker-run
docker-run:
	docker run -d --name $(DOCKER_CONTAINER) -p $(PORT):$(PORT) -v $(PWD)/$(DB_FILE):/scheduler.db -e TODO_PASSWORD=$(TODO_PASSWORD) $(DOCKER_IMAGE)

# Остановка Docker-контейнера
.PHONY: docker-stop
docker-stop:
	docker stop $(DOCKER_CONTAINER) && docker rm $(DOCKER_CONTAINER)

# Очистка Docker-образов и контейнеров
.PHONY: docker-clean
docker-clean:
	docker rmi $(DOCKER_IMAGE)
	docker system prune -f