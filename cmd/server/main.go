package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/St1cky1/task-service/internal/api"
	"github.com/St1cky1/task-service/internal/models"
	"github.com/St1cky1/task-service/internal/rabbitmq"
	"github.com/St1cky1/task-service/internal/repo"
	"github.com/St1cky1/task-service/internal/service"
	"github.com/St1cky1/task-service/internal/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	rabbitMQURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"))
	// Запускаем миграции
	if err := runMigrations(dbURL); err != nil {
		log.Fatal("❌ Ошибка миграций:", err)
	}

	// Подключаемся к БД
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("❌ Ошибка подключения к БД:", err)
	}
	defer db.Close()

	// Проверяем соединение с БД
	if err := db.Ping(context.Background()); err != nil {
		log.Fatal("❌ Не удалось подключиться к БД:", err)
	}
	fmt.Println("✅ Подключение к БД установлено")

	// Подключаемся к RabbitMQ
	rabbitMQ, err := rabbitmq.NewRabbitMQClient(rabbitMQURL)
	if err != nil {
		log.Fatal("❌ Ошибка подключения к RabbitMQ:", err)
	}
	defer rabbitMQ.Close()
	fmt.Println("✅ Подключение к RabbitMQ установлено")

	// Инициализируем репозитории
	userRepo := repo.NewUserRepository(db)
	taskRepo := repo.NewTaskRepository(db)
	taskAuditRepo := repo.NewTaskAuditRepository(db)

	// Инициализируем сервисы
	taskService := service.NewTaskService(taskRepo, userRepo, taskAuditRepo, rabbitMQ)

	// Запускаем воркер для обработки аудит-сообщений
	auditWorker := worker.NewAuditWorker(rabbitMQ, taskAuditRepo)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	go func() {
		fmt.Println("Запуск Audit Worker...")
		auditWorker.Start(workerCtx)
	}()

	// Запускаем непрерывную генерацию задач
	taskGenCtx, taskGenCancel := context.WithCancel(context.Background())
	defer taskGenCancel()
	go continuousTaskGeneration(taskGenCtx, taskService, userRepo)

	// Запускаем HTTP сервер
	go startHTTPServer(taskService)

	fmt.Println("Сервисный слой с RabbitMQ работает!")
	fmt.Println("HTTP API доступен на http://localhost:8080")
	fmt.Println("RabbitMQ Management: http://localhost:15672")
	fmt.Println("Audit Worker запущен и ожидает сообщения...")
	fmt.Println("Непрерывная генерация задач запущена...")
	fmt.Println("Для остановки нажмите Ctrl+C")

	// Ждем сигнал завершения
	waitForShutdown(workerCancel, taskGenCancel)
}

// Непрерывная генерация задач
func continuousTaskGeneration(ctx context.Context, taskService *service.TaskService, userRepo *repo.UserRepository) {
	// Сначала создаем тестового пользователя
	user := createOrGetTestUser(ctx, userRepo)
	if user == nil {
		log.Println("❌ Не удалось создать тестового пользователя")
		return
	}

	taskCounter := 0
	statuses := []models.TaskStatus{
		models.StatusPending,
		models.StatusInProgres,
		models.StatusCompleted,
		models.StatusCancelled,
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Генерация задач остановлена")
			return
		case <-time.After(5 * time.Second): // Генерируем задачу каждые 5 секунд
			taskCounter++

			// Случайный статус
			status := statuses[taskCounter%len(statuses)]

			// Создаем задачу
			taskReq := &models.CreateTaskRequest{
				Title:       fmt.Sprintf("Авто-задача #%d", taskCounter),
				Description: fmt.Sprintf("Сгенерирована автоматически в %s", time.Now().Format("15:04:05")),
				Status:      status,
				OwnerId:     user.ID,
			}

			task, err := taskService.CreateTask(ctx, taskReq, user.ID)
			if err != nil {
				log.Printf("❌ Ошибка создания авто-задачи: %v", err)
				continue
			}

			fmt.Printf("✅ Создана авто-задача: ID=%d, Title=%s, Status=%s\n",
				task.ID, task.Title, task.Status)

			// Случайно обновляем или удаляем каждую 3-ю задачу
			if taskCounter%3 == 0 {
				// Обновляем задачу
				updateReq := models.UpdateTaskRequest{
					Title:  fmt.Sprintf("обновленная задача #%d", taskCounter),
					Status: models.StatusCompleted,
				}

				updatedTask, err := taskService.UpdateTask(ctx, task.ID, user.ID, &updateReq)
				if err != nil {
					log.Printf("❌ Ошибка обновления авто-задачи: %v", err)
				} else {
					fmt.Printf("Обновлена авто-задача: %s (%s)\n", updatedTask.Title, updatedTask.Status)
				}
			}

			// Удаляем каждую 5-ю задачу
			if taskCounter%5 == 0 {
				err = taskService.DeleteTask(ctx, task.ID, user.ID)
				if err != nil {
					log.Printf("❌ Ошибка удаления авто-задачи: %v", err)
				} else {
					fmt.Printf("Удалена авто-задача: ID=%d\n", task.ID)
				}
			}

			// Показываем статистику каждые 10 задач
			if taskCounter%10 == 0 {
				tasks, err := taskService.ListTasks(ctx, user.ID, "")
				if err != nil {
					log.Printf("❌ Ошибка получения списка задач: %v", err)
				} else {
					fmt.Printf("Статистика: создано %d задач, в БД: %d задач\n",
						taskCounter, len(tasks))
				}
			}
		}
	}
}

// Создает или получает тестового пользователя
func createOrGetTestUser(ctx context.Context, userRepo *repo.UserRepository) *models.User {
	// Пробуем получить пользователя с ID=1
	user, err := userRepo.GetById(ctx, 1)
	if err != nil {
		log.Printf("❌ Ошибка получения пользователя: %v", err)
		return nil
	}

	if user != nil {
		fmt.Printf("✅ Найден существующий пользователь: ID=%d, Name=%s\n", user.ID, user.Name)
		return user
	}

	// Создаем нового пользователя
	userReq := &models.CreateUserRequest{Name: "Auto-Generated User"}
	user, err = userRepo.Create(ctx, userReq)
	if err != nil {
		log.Printf("❌ Ошибка создания пользователя: %v", err)
		return nil
	}

	fmt.Printf("✅ Создан новый пользователь: ID=%d, Name=%s\n", user.ID, user.Name)
	return user
}

func startHTTPServer(taskService *service.TaskService) {
	router := api.NewRouter(taskService)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println(" Запуск HTTP сервера на порту 8080...")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Ошибка HTTP сервера: %v", err)
	}
}

func waitForShutdown(workerCancel context.CancelFunc, taskGenCancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Ожидаем сигнал завершения (Ctrl+C)...")
	<-sigChan

	fmt.Println("👋 Завершение работы...")

	// Останавливаем воркер и генератор задач
	workerCancel()
	taskGenCancel()

	// Даем время для graceful shutdown
	time.Sleep(2 * time.Second)
	fmt.Println("✅ Приложение завершено корректно")
}

func runMigrations(dbURL string) error {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return fmt.Errorf("ошибка создания мигратора: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка выполнения миграций: %w", err)
	}

	fmt.Println("✅ Миграции выполнены успешно")
	return nil
}
