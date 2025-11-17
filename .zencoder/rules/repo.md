---
description: Repository Information Overview
alwaysApply: true
---

# Task Service - Информация о репозитории

## Резюме
Task Service - это микросервис на Go для управления задачами с поддержкой gRPC и REST API. Сервис автоматически генерирует задачи, работает с PostgreSQL базой данных и использует RabbitMQ для асинхронной обработки событий (аудита).

## Структура проекта
```
task-service/
├── cmd/server/           - Entry point приложения
├── internal/
│   ├── api/              - API слой (gRPC, REST handlers, router)
│   ├── service/          - Бизнес-логика сервиса
│   ├── repo/             - Слой доступа к данным
│   ├── client/           - Клиенты БД и RabbitMQ
│   ├── models/           - Структуры данных
│   └── worker/           - Worker для асинхронной обработки
├── migrations/           - SQL миграции БД
├── configs/              - Конфигурация
├── docker-compose.yaml   - Docker Compose для локальной разработки
└── protoc/               - Protoc бинарник и зависимости
```

## Язык и Runtime
**Язык**: Go
**Версия**: 1.25.4
**Система сборки**: go build
**Пакетный менеджер**: Go Modules (go.mod/go.sum)

## Основные зависимости
- **chi/v5** (v5.0.8) - HTTP маршрутизатор
- **grpc** (v1.76.0) - gRPC framework
- **grpc-gateway/v2** (v2.27.3) - REST proxy для gRPC
- **pgx/v5** (v5.7.6) - PostgreSQL драйвер
- **amqp091-go** (v1.10.0) - RabbitMQ клиент
- **golang-migrate** (v4.19.0) - Миграции БД
- **protobuf** (v1.36.10) - Protocol Buffers

## Сборка и установка
```bash
# Загрузка зависимостей
go mod download

# Сборка приложения
go build -o task-service ./cmd/server

# Запуск приложения
./task-service
```

## Docker
**Docker Compose файл**: `docker-compose.yaml`
**Сервисы**:
- **PostgreSQL 18.0** - База данных (порт 5432)
- **RabbitMQ 3.13** - Message broker с Management UI (портов 5672 и 15672)

**Запуск**:
```bash
docker-compose up -d
```

**Переменные окружения** (из .env):
- DB_USER, DB_PASSWORD, DB_NAME, DB_PORT
- RABBITMQ_USER, RABBITMQ_PASSWORD, RABBITMQ_PORT

## Миграции БД
**Расположение**: `migrations/`
**Инструмент**: golang-migrate

Миграции:
- `001_create_user_table.{up,down}.sql` - Таблица пользователей
- `002_create_task_table.{up,down}.sql` - Таблица задач
- `003_create_task_audit_table.{up,down}.sql` - Таблица аудита задач

## API
**Формат**: Protocol Buffers (Proto3)
**Главный файл**: `internal/api/proto/task_service.proto`

**Поддерживаемые протоколы**:
- gRPC (на основе Protobuf)
- REST (через grpc-gateway)

**Компоненты API**:
- `internal/api/grpc/server.go` - gRPC сервер
- `internal/api/handlers/task_handler.go` - REST обработчики
- `internal/api/router.go` - Маршруты приложения

## Архитектура
- **Worker**: `internal/worker/audit_worker.go` - Асинхронный worker для обработки событий
- **Service**: `internal/service/task_service.go` - Бизнес-логика
- **Repository**: Слой доступа к данным (task_repo, user_repo, task_audit_repo)
- **Client**: Клиенты для PostgreSQL и RabbitMQ
- **Models**: Структуры данных (User, Task, TaskAudit)

## Точка входа
- **Main приложение**: `cmd/server/main.go` - инициализирует БД, запускает gRPC сервер, обработчики и worker'ов
