# Планировщик задач (Task Scheduler)

## Описание проекта

Планировщик задач - это веб-приложение для эффективного управления задачами и повышения продуктивности. Приложение позволяет создавать, редактировать, удалять и отслеживать выполнение задач с поддержкой повторяющихся событий.

### Основные возможности

- ✅ Создание, редактирование и удаление задач
- ✅ Поддержка повторяющихся задач (ежедневно, еженедельно, ежемесячно)
- ✅ Поиск задач по дате и названию
- ✅ Отметка задач как выполненных
- ✅ Автоматическое планирование следующей даты для повторяющихся задач
- ✅ Кэширование для повышения производительности
- ✅ Аутентификация пользователей
- ✅ RESTful API для интеграции с другими системами

## Технический стек

### Бэкенд
- **Язык программирования**: Go 1.21+
- **База данных**: SQLite3
- **Фреймворк**: net/http (стандартная библиотека Go)
- **Маршрутизация**: Chi Router
- **Аутентификация**: JWT
- **Кэширование**: In-memory кэш с TTL
- **Логирование**: Структурированные логи в JSON формате

### Фронтенд
- **HTML5/CSS3**: Современный адаптивный дизайн
- **JavaScript**: Взаимодействие с API
- **Bootstrap**: Компоненты интерфейса

### Инфраструктура
- **Контейнеризация**: Docker
- **Управление задачами**: Task (современная альтернатива Make)
- **Автоперезагрузка**: Air (для разработки)
- **Линтинг**: golangci-lint

## Быстрый старт

### Предварительные требования
- Go 1.21 или выше
- Git
- Docker (опционально)
- Task (опционально, для управления задачами)

### Локальный запуск

1. **Клонирование репозитория**:
   ```bash
   git clone https://github.com/yourusername/task-scheduler.git
   cd task-scheduler
   ```

2. **Настройка окружения**:
   ```bash
   # Копирование примера конфигурации
   cp .env.example .env
   
   # Редактирование .env файла при необходимости
   ```

3. **Установка зависимостей**:
   ```bash
   go mod tidy
   ```

4. **Запуск приложения**:
   ```bash
   # Стандартный запуск
   go run cmd/server/webserver.go
   
   # Или с использованием Task
   task
   ```

5. **Доступ к приложению**:
   Откройте браузер и перейдите по адресу `http://localhost:7540`

### Запуск с использованием Task

Task - это современная альтернатива Make, которая упрощает управление задачами в проекте.

1. **Установка Task**:
   - Windows: `choco install go-task`
   - macOS: `brew install go-task`
   - Linux: `sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d`

2. **Основные команды**:
   ```bash
   # Запуск сервера в режиме разработки
   task
   
   # Сборка проекта
   task build
   
   # Запуск тестов
   task test
   
   # Проверка кода линтером
   task lint
   
   # Запуск в режиме разработки с автоперезагрузкой
   task dev
   ```

### Запуск в Docker

1. **Сборка Docker-образа**:
   ```bash
   task docker-build
   # или
   docker build -t scheduler .
   ```

2. **Запуск контейнера**:
   ```bash
   task docker-run
   # или
   docker run -p 7540:7540 -v $(pwd)/scheduler.db:/app/scheduler.db scheduler
   ```

## API Endpoints

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| GET | /api/tasks | Получить список всех задач |
| GET | /api/task?id={id} | Получить задачу по ID |
| POST | /api/task | Создать новую задачу |
| PUT | /api/task | Обновить существующую задачу |
| DELETE | /api/task?id={id} | Удалить задачу |
| POST | /api/task/done?id={id} | Отметить задачу как выполненную |
| GET | /api/nextdate?date={date}&repeat={repeat} | Получить следующую дату для повторяющейся задачи |
| GET | /api/health | Проверка работоспособности сервера |

## Структура проекта

```
.
├── cmd/                  # Точки входа приложения
│   └── server/           # Веб-сервер
├── internal/             # Внутренние пакеты
│   ├── auth/             # Аутентификация и авторизация
│   ├── cache/            # Кэширование
│   ├── config/           # Конфигурация
│   ├── database/         # Работа с базой данных
│   ├── moduls/           # Модели данных
│   ├── nextdate/         # Логика расчета следующей даты
│   ├── router/           # Маршрутизация
│   ├── tasks/            # Обработчики задач
│   └── utils/            # Утилиты
├── web/                  # Статические файлы
├── tests/                # Тесты
├── .env.example          # Пример конфигурации
├── .air.toml             # Конфигурация Air
├── Dockerfile            # Конфигурация Docker
├── Taskfile.yml          # Задачи для Task
└── README.md             # Документация
```

## Тестирование

### Запуск тестов
```bash
# Запуск всех тестов
task test

# Запуск тестов с покрытием
go test -cover ./...
```

### Настройка тестового окружения
1. Создайте файл `.env.test` на основе `.env.example`
2. Установите тестовые параметры в файле
3. Запустите тесты с указанием тестового окружения:
   ```bash
   TODO_ENV=test go test ./tests
   ```

## Производительность

Приложение оптимизировано для высокой производительности:
- Кэширование часто запрашиваемых данных
- Индексы в базе данных для быстрого поиска
- Пул соединений с базой данных
- Оптимизированные SQL-запросы
- Кэширование статических файлов

## Безопасность

- Аутентификация с использованием JWT
- Защита от SQL-инъекций
- Валидация входных данных
- Безопасное хранение паролей
- Защита от CSRF-атак

## Лицензия

MIT

## Авторы

- Ваше имя - [ваш@email.com](mailto:ваш@email.com)

## Благодарности

- [Chi Router](https://github.com/go-chi/chi) - легковесный маршрутизатор для Go
- [SQLite](https://www.sqlite.org/) - встраиваемая база данных
- [Task](https://taskfile.dev/) - современная альтернатива Make
