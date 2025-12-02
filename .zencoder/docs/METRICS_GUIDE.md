# 📊 Руководство по просмотру метрик

## Обзор

В проект добавлены комплексные метрики для мониторинга:
- **HTTP запросы** (количество, время отклика)
- **gRPC запросы** (количество, время отклика)  
- **Ошибки** (HTTP и gRPC)
- **Обработанные задачи** (создание, обновление, удаление)
- **Подключения к БД** (активные, всего в пуле)

## Запуск приложения с метриками

### 1️⃣ Запустить Docker Compose (БД, RabbitMQ, Prometheus, Grafana)

```bash
docker-compose up -d
```

Проверить статус:
```bash
docker-compose ps
```

### 2️⃣ Запустить приложение

```bash
./task-service
```

Или если нужно пересобрать:
```bash
go build -o task-service ./cmd/server && ./task-service
```

## 📈 Просмотр метрик

### Вариант 1: Prometheus Dashboard (сырые метрики)

1. Откройте http://localhost:9090
2. Перейдите на вкладку "Graph"
3. Введите название метрики в поле поиска

**Доступные метрики:**

| Метрика | Описание | Тип |
|---------|---------|-----|
| `http_requests_total` | Всего HTTP запросов | Counter |
| `http_request_duration_seconds` | Длительность HTTP запроса | Histogram |
| `http_errors_total` | Всего HTTP ошибок | Counter |
| `grpc_requests_total` | Всего gRPC запросов | Counter |
| `grpc_request_duration_seconds` | Длительность gRPC запроса | Histogram |
| `grpc_errors_total` | Всего gRPC ошибок | Counter |
| `tasks_created_total` | Всего создано задач | Counter |
| `tasks_updated_total` | Всего обновлено задач | Counter |
| `tasks_deleted_total` | Всего удалено задач | Counter |
| `tasks_processed_total` | Всего обработано задач | Counter |
| `db_connections_active` | Активные подключения к БД | Gauge |
| `db_connections_pool_total` | Всего подключений в пуле | Gauge |
| `db_errors_total` | Всего ошибок БД | Counter |
| `db_query_duration_seconds` | Длительность запроса БД | Histogram |

**Примеры запросов:**

```promql
# Количество HTTP запросов в секунду за 5 минут
rate(http_requests_total[5m])

# Среднее время обработки HTTP запроса
http_request_duration_seconds_sum / http_request_duration_seconds_count

# Количество ошибок HTTP за 5 минут
rate(http_errors_total[5m])

# Активные подключения к БД
db_connections_active

# Общее количество создано, обновлено, удалено задач
tasks_created_total
tasks_updated_total  
tasks_deleted_total

# Ошибки БД
rate(db_errors_total[5m])
```

### Вариант 2: Grafana Dashboard (красивые графики)

1. Откройте http://localhost:3000
2. Введите учетные данные (см. `.env`):
   - Username: `admin` (или значение `ADMIN_NAME`)
   - Password: смотрите переменную `ADMIN_PASSWORD` в `.env`

3. **Добавить источник данных Prometheus:**
   - Нажмите "Add data source"
   - Выберите "Prometheus"
   - URL: `http://prometheus:9090`
   - Нажмите "Save & test"

4. **Создать Dashboard:**
   - Нажмите "New" → "Dashboard"
   - Нажмите "Add panel"
   - В поле "Metrics" введите название метрики
   - Нажмите "Apply"

**Рекомендуемые графики:**

```
# 1. HTTP трафик
Panel: rate(http_requests_total[5m])
Legend: HTTP requests/sec

# 2. gRPC трафик  
Panel: rate(grpc_requests_total[5m])
Legend: gRPC requests/sec

# 3. Ошибки
Panel: rate(http_errors_total[5m]) + rate(grpc_errors_total[5m])
Legend: Errors/sec

# 4. Обработанные задачи
Panel: rate(tasks_processed_total[5m])
Legend: Tasks processed/sec

# 5. Подключения БД
Panel: db_connections_active
Legend: Active DB connections

# 6. Среднее время ответа HTTP
Panel: histogram_quantile(0.95, http_request_duration_seconds)
Legend: p95 response time
```

### Вариант 3: REST API метрик (json формат)

```bash
curl http://localhost:8080/metrics
```

Это вернет все метрики в Prometheus текстовом формате.

### Вариант 4: pprof Профилировщик (CPU, Память)

1. Откройте http://localhost:6060/debug/pprof/

**Доступные профили:**

- `/debug/pprof/heap` - Снимок памяти
- `/debug/pprof/goroutine` - Все горутины
- `/debug/pprof/profile` - CPU профиль (30 сек)
- `/debug/pprof/allocs` - История аллокаций

**Пример получения CPU профиля:**
```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

**Пример профиля памяти:**
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

## 🔄 Как работают метрики

### HTTP метрики
- Каждый запрос к REST API (8080) отслеживается middleware
- Автоматически считается время обработки
- Ошибки (HTTP 4xx, 5xx) отмечаются отдельно

### gRPC метрики  
- Каждый запрос к gRPC серверу (9090) отслеживается interceptor'ом
- Работает для unary и stream запросов
- Ошибки отмечаются по статус-коду

### Задачи
- `TasksCreated` - инкрементируется в `CreateTask()`
- `TasksUpdated` - инкрементируется в `UpdateTask()`
- `TasksDeleted` - инкрементируется в `DeleteTask()`
- `TasksProcessed` - общий счетчик всех операций

### База данных
- `DBConnections` - обновляется каждые 5 сек (активные соединения)
- `DBConnectionsPool` - обновляется каждые 5 сек (всего в пуле)
- `DBErrors` - инкрементируется при ошибках БД

## 🧪 Тестирование метрик

Генерируйте нагрузку и смотрите как растут метрики:

```bash
# Генерация HTTP запросов
while true; do
  curl -s http://localhost:8080/v1/tasks | jq .
  sleep 1
done
```

Или используйте `grpcurl`:
```bash
grpcurl -plaintext localhost:9090 list
```

## 📝 Логирование ошибок

Все ошибки БД автоматически отслеживаются метриками:
- Счетчик `db_errors_total` инкрементируется
- Ошибка также логируется в stdout

## ⚙️ Настройка

### Интервал сбора метрик (Prometheus)
Отредактируйте `prometheus.yml`:
```yaml
scrape_interval: 10s  # По умолчанию 10 секунд
```

### Интервал обновления метрик БД
В `internal/infrastructure/metrics/db_metrics.go`:
```go
ticker := time.NewTicker(5 * time.Second)  // Менять это значение
```

## 🐛 Решение проблем

### Метрики не появляются в Prometheus
1. Проверьте, что приложение запущено: http://localhost:8080/metrics
2. Проверьте prometheus.yml валиден
3. Перезагрузите Docker контейнер: `docker-compose restart prometheus`

### Grafana не видит Prometheus
1. Проверьте что Prometheus запущен: http://localhost:9090
2. В Grafana проверьте data source: `http://prometheus:9090` (не localhost!)
3. Нажмите "Test" в настройках data source

### БД метрики показывают 0
1. Проверьте что БД подключена
2. БД метрики обновляются каждые 5 секунд - подождите немного
3. Проверьте логи приложения

## 📚 Дальнейшее развитие

Можно добавить:
- Custom метрики для бизнес-логики
- Alert'ы в Prometheus/Grafana
- Service monitoring на уровне OS
- Tracing (Jaeger/Zipkin)
- Custom dashboards в Grafana