package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	grpcapi "github.com/St1cky1/task-service/internal/api/grpc"
	"github.com/St1cky1/task-service/internal/entity"
	"github.com/St1cky1/task-service/internal/infrastructure/auth"
	"github.com/St1cky1/task-service/internal/infrastructure/client"
	"github.com/St1cky1/task-service/internal/infrastructure/metrics"
	"github.com/St1cky1/task-service/internal/infrastructure/worker"
	"github.com/St1cky1/task-service/internal/repository"
	"github.com/St1cky1/task-service/internal/usecase"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var wg sync.WaitGroup

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

	// Инициализируем мониторинг метрик БД
	metrics.InitDBMetrics(db)

	// Подключаемся к RabbitMQ
	rabbitMQ, err := client.NewRabbitMQClient(rabbitMQURL)
	if err != nil {
		log.Fatal("❌ Ошибка подключения к RabbitMQ:", err)
	}
	defer rabbitMQ.Close()
	fmt.Println("✅ Подключение к RabbitMQ установлено")

	// Инициализируем репозитории
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	taskAuditRepo := repository.NewTaskAuditRepository(db)
	avatarRepo := repository.NewAvatarRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Инициализируем auth компоненты
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	// Инициализируем сервисы
	taskService := usecase.NewTaskService(taskRepo, userRepo, taskAuditRepo, rabbitMQ)
	userService := usecase.NewUserService(userRepo, avatarRepo, passwordManager, jwtManager, refreshTokenRepo)
	authService := usecase.NewAuthService(userRepo, refreshTokenRepo, passwordManager, jwtManager)

	// Запускаем pprof профилировщик на порту 6060
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("🔍 Запуск pprof сервера профилирования на порту 6060...")
		listener, err := net.Listen("tcp", ":6060")
		if err != nil {
			log.Printf("❌ Ошибка при создании listener для pprof: %v", err)
			return
		}
		if err := http.Serve(listener, nil); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ pprof сервер ошибка: %v", err)
		}
	}()

	// Запускаем воркер для обработки аудит-сообщений
	auditWorker := worker.NewAuditWorker(rabbitMQ, taskAuditRepo)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Запуск Audit Worker...")
		auditWorker.Start(workerCtx)
	}()

	// Запускаем непрерывную генерацию задач
	taskGenCtx, taskGenCancel := context.WithCancel(context.Background())
	defer taskGenCancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		continuousTaskGeneration(taskGenCtx, taskService, userRepo)
	}()

	// Запускаем gRPC сервер со всеми сервисами (Task, User, Auth)
	grpcServer := grpcapi.NewGRPCServer(taskService, userService, authService)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Запуск gRPC сервера на порту 9090...")
		fmt.Println("📋 TaskService, UserService и AuthService готовы к работе!")
		if err := grpcServer.Start("9090"); err != nil {
			log.Printf("❌ gRPC server error: %v", err)
		}
	}()

	// Запускаем gRPC Gateway (HTTP->gRPC трансляция)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Запуск gRPC Gateway на порту 8080...")
		if err := grpcServer.StartGateway(context.Background(), "9090", "8080"); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ gRPC Gateway error: %v", err)
		}
	}()

	// Запускаем непрерывную генерацию пользователей с аватарками
	userGenCtx, userGenCancel := context.WithCancel(context.Background())
	defer userGenCancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Даем серверу время на инициализацию
		time.Sleep(2 * time.Second)
		fmt.Println("\n👥 Начинаем генерацию пользователей с аватарками...")
		if err := usecase.ContinuousUserGenerationWithAvatars(userGenCtx, userService); err != nil {
			log.Printf("⚠️  Ошибка при генерации пользователей: %v", err)
		}
	}()

	fmt.Println("\n Доступные сервисы:")
	fmt.Println("  gRPC сервер: localhost:9090")
	fmt.Println("  REST Gateway: http://localhost:8080")
	fmt.Println("  🔍 pprof Профилирование: http://localhost:6060/debug/pprof/")
	fmt.Println("  RabbitMQ Management: http://localhost:15672")
	fmt.Println("\n Процессы:")
	fmt.Println("  Audit Worker запущен и ожидает сообщения...")
	fmt.Println("  Непрерывная генерация задач запущена...")
	fmt.Println("  Генерация пользователей с аватарками запущена...")
	fmt.Println("\nДля остановки нажмите Ctrl+C")

	// Ждем сигнал завершения
	waitForShutdown(workerCancel, taskGenCancel, userGenCancel)
}

// Непрерывная генерация задач для всех пользователей
func continuousTaskGeneration(ctx context.Context, taskService *usecase.TaskService, userRepo repository.IUserRepository) {
	taskCounter := 0
	statuses := []entity.TaskStatus{
		entity.StatusPending,
		entity.StatusInProgress,
		entity.StatusCompleted,
		entity.StatusCancelled,
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Генерация задач остановлена")
			return
		case <-time.After(5 * time.Second): // Генерируем задачу каждые 5 секунд
			// Получаем всех активных пользователей
			users, err := userRepo.List(ctx)
			if err != nil {
				log.Printf("❌ Ошибка получения пользователей: %v", err)
				continue
			}

			if len(users) == 0 {
				log.Println("⏳ Пользователей еще нет, ожидаем создания...")
				continue
			}

			taskCounter++

			// Распределяем задачи между пользователями
			userIdx := (taskCounter - 1) % len(users)
			user := users[userIdx]

			// Случайный статус
			status := statuses[taskCounter%len(statuses)]

			// Создаем задачу
			taskReq := &entity.CreateTaskRequest{
				Title:       fmt.Sprintf("Авто-задача #%d", taskCounter),
				Description: fmt.Sprintf("Сгенерирована автоматически в %s", time.Now().Format("15:04:05")),
				Status:      status,
				OwnerId:     user.ID,
			}

			task, err := taskService.CreateTask(ctx, taskReq, user.ID)
			if err != nil {
				log.Printf("❌ Ошибка создания авто-задачи для user %d: %v", user.ID, err)
				continue
			}

			fmt.Printf("✅ Создана авто-задача: ID=%d, Title=%s, User=%d, Status=%s\n",
				task.ID, task.Title, user.ID, task.Status)

			// Случайно обновляем или удаляем каждую 3-ю задачу
			if taskCounter%3 == 0 {
				// Обновляем задачу
				updateReq := entity.UpdateTaskRequest{
					Title:  fmt.Sprintf("обновленная задача #%d", taskCounter),
					Status: entity.StatusCompleted,
				}

				updatedTask, err := taskService.UpdateTask(ctx, task.ID, user.ID, &updateReq)
				if err != nil {
					log.Printf("❌ Ошибка обновления авто-задачи: %v", err)
				} else {
					fmt.Printf("📝 Обновлена авто-задача: %s (%s)\n", updatedTask.Title, updatedTask.Status)
				}
			}

			// Удаляем каждую 5-ю задачу
			if taskCounter%5 == 0 {
				err = taskService.DeleteTask(ctx, task.ID, user.ID)
				if err != nil {
					log.Printf("❌ Ошибка удаления авто-задачи: %v", err)
				} else {
					fmt.Printf("🗑️  Удалена авто-задача: ID=%d\n", task.ID)
				}
			}

			// Показываем статистику каждые 10 задач
			if taskCounter%10 == 0 {
				totalTasks := 0
				for _, u := range users {
					tasks, err := taskService.ListTasks(ctx, u.ID, "")
					if err == nil {
						totalTasks += len(tasks)
					}
				}
				fmt.Printf("📊 Статистика: создано %d задач, активных пользователей: %d, всего задач в БД: %d\n",
					taskCounter, len(users), totalTasks)
			}
		}
	}
}

func waitForShutdown(workerCancel, taskGenCancel, userGenCancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Ожидаем сигнал завершения (Ctrl+C)...")
	<-sigChan

	fmt.Println("Завершение работы...")

	// Останавливаем воркер и генераторы
	workerCancel()
	taskGenCancel()
	userGenCancel()

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
