# 🚀 Быстрый старт: Просмотр метрик

## 1 минута: Запуск

```bash
# Терминал 1: Docker
docker-compose up -d

# Терминал 2: Приложение
go build -o task-service ./cmd/server && ./task-service
```

## 2 минуты: Просмотр сырых метрик

Откройте в браузере:
```
http://localhost:9090
```

Введите в поле поиска:
- `http_requests_total` - количество HTTP запросов
- `grpc_requests_total` - количество gRPC запросов  
- `tasks_processed_total` - обработано задач
- `db_connections_active` - активные подключения к БД

## 3 минуты: Красивые графики (Grafana)

1. Откройте http://localhost:3000
2. Username: `admin`
3. Password: посмотрите в `.env` значение `ADMIN_PASSWORD`

**После входа:**
1. Нажмите "Add data source"
2. Выберите "Prometheus"  
3. URL: `http://prometheus:9090`
4. "Save & test"
5. Создайте новый Dashboard и добавляйте панели

## 🎯 Основные метрики

| Порт | Назначение |
|------|-----------|
| **8080** | REST API + `/metrics` endpoint |
| **9090** | gRPC сервер + Prometheus |
| **3000** | Grafana dashboard |
| **6060** | pprof профилировщик |
| **5432** | PostgreSQL БД |
| **5672** | RabbitMQ |
| **15672** | RabbitMQ Management UI |

## 📊 Примеры PromQL запросов

```promql
# HTTP запросы в секунду (за 5 минут)
rate(http_requests_total[5m])

# Ошибки в секунду
rate(http_errors_total[5m]) + rate(grpc_errors_total[5m])

# 95-й перцентиль времени ответа
histogram_quantile(0.95, http_request_duration_seconds)

# Активные подключения к БД
db_connections_active

# Всего обработано задач
tasks_processed_total
```

## 🔍 Весь путь данных метрик

```
Приложение (metrics.go)
    ↓
pprof Профилировщик (http://localhost:6060/debug/pprof/)
    ↓
REST API /metrics endpoint (http://localhost:8080/metrics)
    ↓
Prometheus (http://localhost:9090)
    ↓
Grafana Dashboard (http://localhost:3000)
```

## ⚡ Добавленный функционал

✅ **HTTP middleware** - отслеживание всех REST запросов  
✅ **gRPC interceptors** - отслеживание gRPC вызовов  
✅ **Task metrics** - счетчики создания/обновления/удаления  
✅ **DB metrics** - мониторинг пула соединений и ошибок  
✅ **Error tracking** - автоматический подсчет ошибок  
✅ **Duration histograms** - распределение времени обработки  

## 🆘 Если метрики не видны

1. Метрики собираются по мере работы приложения - подождите несколько минут
2. Проверьте что все контейнеры запущены: `docker-compose ps`
3. Проверьте доступность приложения: `curl http://localhost:8080/metrics`
4. В Prometheus используйте `http://prometheus:9090` а не `localhost`

## 📖 Подробное руководство

Смотрите `METRICS_GUIDE.md` для детального описания всех метрик и примеров.