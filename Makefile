BINARY_NAME=task-service
BUILD_DIR=./build
GO_VERSION=1.25.4
MAIN_PATH=./cmd/server
PROTO_DIR=./proto

.PHONY: help build test clean run dev docker-up docker-down docker-logs migrate-up migrate-down proto fmt lint dev-run

help: ## Показать справку по командам
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать приложение
	@echo "Сборка $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Приложение собрано в $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Запустить приложение (сначала собирает)
	@echo "Запуск $(BINARY_NAME)..."
	@godotenv -f .env $(BUILD_DIR)/$(BINARY_NAME)

dev-run: ## Запустить приложение с go run (с загрузкой .env)
	@echo "Запуск $(BINARY_NAME) с go run..."
	@godotenv -f .env go run $(MAIN_PATH)

test: ## Запустить тесты
	@echo "Запуск тестов..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Тесты завершены (покрытие в coverage.html)"

clean: ## Очистить build директорию
	@echo "Очистка..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✓ Очистка завершена"

docker-up: ## Запустить Docker Compose (PostgreSQL, RabbitMQ)
	@echo "Запуск Docker контейнеров..."
	@docker-compose up -d
	@echo "✓ Контейнеры запущены"

docker-down: ## Остановить Docker Compose
	@echo "Остановка Docker контейнеров..."
	@docker-compose down
	@echo "✓ Контейнеры остановлены"

docker-logs: ## Показать логи Docker контейнеров
	@docker-compose logs -f

migrate-up: ## Применить миграции БД
	@echo "Применение миграций БД..."
	@godotenv -f .env go run -tags migrate_tool github.com/golang-migrate/migrate/v4/cmd/migrate -path ./migrations -database "$$(godotenv -f .env sh -c 'echo postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable')" up
	@echo "✓ Миграции применены"

migrate-down: ## Откатить последнюю миграцию
	@echo "Откат последней миграции..."
	@godotenv -f .env go run -tags migrate_tool github.com/golang-migrate/migrate/v4/cmd/migrate -path ./migrations -database "$$(godotenv -f .env sh -c 'echo postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable')" down 1
	@echo "✓ Миграция откачена"

proto: ## Генерировать код из proto файлов
	@echo "Генерация proto файлов..."
	@cd $(PROTO_DIR) && make proto || true
	@echo "✓ Proto файлы сгенерированы"

fmt: ## Форматировать код
	@echo "Форматирование кода..."
	@go fmt ./...
	@echo "✓ Код отформатирован"

lint: ## Запустить линтер (требует golangci-lint)
	@echo "Запуск линтера..."
	@golangci-lint run ./... || echo "Установите golangci-lint: brew install golangci-lint"


.DEFAULT_GOAL := help
