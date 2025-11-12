package main

import (
	"context"
	"fmt"
	"log"

	"github.com/St1cky1/task-service/internal/models"
	"github.com/St1cky1/task-service/internal/rabbitmq"
	"github.com/St1cky1/task-service/internal/repo"
	"github.com/St1cky1/task-service/internal/service"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgresql://user:pass@localhost:54321/tasks?sslmode=disable"
	rabbitMQURL := "amqp://guest:guest@localhost:15672/"

	runMigrations(dbURL)

	// Подключаемся к БД
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	// Подключаемся к RabbitMQ
	rabbitMQ, err := rabbitmq.NewRabbitMQClient(rabbitMQURL)
	if err != nil {
		log.Fatal("Ошибка подключения к RabbitMQ:", err)
	}
	defer rabbitMQ.Close()

	// Инициализируем репозитории и сервисы
	userRepo := repo.NewUserRepository(db)
	taskRepo := repo.NewTaskRepository(db)
	taskAuditRepo := repo.NewTaskAuditRepository(db)
	taskService := service.NewTaskService(taskRepo, userRepo, taskAuditRepo, rabbitMQ)

	// Тестируем полный цикл сервиса
	testFullServiceCycle(taskService, userRepo)

	fmt.Println("🎉 Сервисный слой с RabbitMQ работает!")
}
func testFullServiceCycle(taskService *service.TaskService, userRepo *repo.UserRepository) {
	ctx := context.Background()

	// Создаем пользователя
	userReq := &models.CreateUserRequest{Name: "Service User"}
	user, err := userRepo.Create(ctx, userReq)
	if err != nil {
		log.Printf("❌ Ошибка создания пользователя: %v", err)
		return
	}
	fmt.Printf("✅ Создан пользователь: ID=%d\n", user.ID)

	// 1. Создаем задачу через сервис
	taskReq := &models.CreateTaskRequest{
		Title:       "Первая задача через сервис",
		Description: "Тестируем полный цикл",
		Status:      models.StatusPending,
		OwnerId:     user.ID, // будет перезаписано сервисом для безопасности
	}

	task, err := taskService.CreateTask(ctx, taskReq, user.ID)
	if err != nil {
		log.Printf("❌ Ошибка создания задачи: %v", err)
		return
	}
	fmt.Printf("✅ Создана задача: ID=%d\n", task.ID)

	// 2. Получаем задачу
	foundTask, err := taskService.GetTask(ctx, task.ID, user.ID)
	if err != nil {
		log.Printf("❌ Ошибка получения задачи: %v", err)
		return
	}
	fmt.Printf("✅ Получена задача: %s\n", foundTask.Title)

	// 3. Обновляем задачу
	updateReq := &models.UpdateTaskRequest{
		Title:  "Обновленное название",
		Status: models.StatusInProgres,
	}

	updatedTask, err := taskService.UpdateTask(ctx, task.ID, user.ID, updateReq)
	if err != nil {
		log.Printf("❌ Ошибка обновления задачи: %v", err)
		return
	}
	fmt.Printf("✅ Обновлена задача: %s (%s)\n", updatedTask.Title, updatedTask.Status)

	// 4. Получаем список задач
	tasks, err := taskService.ListTasks(ctx, user.ID, "")
	if err != nil {
		log.Printf("❌ Ошибка получения списка: %v", err)
		return
	}
	fmt.Printf("✅ Найдено задач пользователя: %d\n", len(tasks))

	// 5. Удаляем задачу
	err = taskService.DeleteTask(ctx, task.ID, user.ID)
	if err != nil {
		log.Printf("❌ Ошибка удаления задачи: %v", err)
		return
	}
	fmt.Printf("✅ Задача удалена: ID=%d\n", task.ID)

	fmt.Println("📨 Все аудит-сообщения отправлены в RabbitMQ!")
	fmt.Println("👀 Проверь RabbitMQ Management: http://localhost:15672")
}

func runMigrations(dbURL string) {
	m, err := migrate.New("file:/Users/v.petrov/task-service/migrations", dbURL)
	if err != nil {
		log.Fatal("Ошибка создания мигратора:", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("Ошибка выполнения миграций:", err)
	}
	fmt.Println("Миграции выполнены успешно")
}
