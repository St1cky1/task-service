# Профилирование сервиса с использованием pprof

После запуска сервиса, pprof сервер будет доступен на `http://localhost:6060/debug/pprof/`

## Доступные профилировщики

### 1. CPU Профилирование (30 секунд)
```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

### 2. Memory Профилирование
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

### 3. Goroutine Профилирование
```bash
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### 4. Allocation Профилирование (памяти при выделении)
```bash
go tool pprof http://localhost:6060/debug/pprof/allocs
```

### 5. Mutex Профилирование (блокировки)
```bash
go tool pprof http://localhost:6060/debug/pprof/mutex
```

### 6. Block Профилирование (события синхронизации)
```bash
go tool pprof http://localhost:6060/debug/pprof/block
```

## Интерактивные команды в pprof

После запуска команды pprof, вы получите интерактивное меню. Полезные команды:

- `top10` - показать топ-10 функций по использованию ресурсов
- `web` - открыть граф в браузере (требует `graphviz`)
- `list <function>` - показать исходный код функции с аннотациями
- `help` - справка по командам

## Web UI (графический интерфейс)

Если у вас установлен Graphviz, можно использовать веб-интерфейс:

```bash
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap
```

Это откроет интерактивный графический интерфейс в браузере на `http://localhost:8081`

## Сохранение профилей

### Сохранить CPU профиль
```bash
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

### Сохранить Memory профиль
```bash
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

## Быстрая проверка

Посмотреть количество горутин:
```bash
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

Посмотреть текущую статистику памяти:
```bash
curl http://localhost:6060/debug/pprof/heap?debug=1
```

## Пример анализа

1. Запустите сервис
2. Дайте ему работать несколько минут под нагрузкой
3. Запустите профилирование:
   ```bash
   go tool pprof http://localhost:6060/debug/pprof/heap
   ```
4. В интерактивном меню введите `top10` для анализа
5. Используйте `list <function>` для детального изучения

## Примечания

- **cpu** профиль требует активной нагрузки для получения значимых результатов
- **heap** профиль показывает текущее использование памяти
- **goroutine** полезен для поиска утечек горутин
- Для web UI необходимо установить Graphviz: `brew install graphviz`