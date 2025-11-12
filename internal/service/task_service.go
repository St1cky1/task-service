package service

import (
	"context"
	"log"
	"time"

	"github.com/St1cky1/task-service/internal/models"
	"github.com/St1cky1/task-service/internal/rabbitmq"
	"github.com/St1cky1/task-service/internal/repo"
)

type TaskService struct {
	taskRepo  *repo.TaskRepository
	userRepo  *repo.UserRepository
	auditRepo *repo.TaskAuditRepository
	rabbitMQ  *rabbitmq.Client
}

func NewTaskService(
	taskRepo *repo.TaskRepository,
	userRepo *repo.UserRepository,
	auditRepo *repo.TaskAuditRepository,
	rabbitMQ *rabbitmq.Client,
) *TaskService {
	return &TaskService{
		taskRepo:  taskRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		rabbitMQ:  rabbitMQ,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, req *models.CreateTaskRequest, userID int) (*models.Task, error) {
	// 1. Проверяем что пользователь существует
	user, err := s.userRepo.GetById(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, models.ErrUserNotFound
	}

	// 2. Устанавливаем владельца из контекста (безопасность!)
	req.OwnerId = userID

	// 3. Создаем задачу
	task, err := s.taskRepo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// 4. Асинхронно отправляем аудит
	s.sendAuditMessage(ctx, models.ActionCreate, userID, task.ID, nil, task, nil)

	return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, taskID int, userID int) (*models.Task, error) {
	task, err := s.taskRepo.GetByTaskId(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, models.ErrTaskNotFound
	}

	// Проверяем права доступа
	if task.OwnerId != userID {
		return nil, models.ErrForbidden
	}

	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID int, userID int, req *models.UpdateTaskRequest) (*models.Task, error) {
	// 1. Получаем текущую задачу
	oldTask, err := s.taskRepo.GetByTaskId(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if oldTask == nil {
		return nil, models.ErrTaskNotFound
	}

	// 2. Проверяем права доступа
	if oldTask.OwnerId != userID {
		return nil, models.ErrForbidden
	}

	// 3. Подготавливаем обновления
	updates := make(map[string]interface{})

	if req.Title != "" {
		updates["title"] = req.Title
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		return nil, models.ErrNoFieldsToUpdate
	}

	// 4. Обновляем задачу
	updatedTask, err := s.taskRepo.Update(ctx, taskID, updates)
	if err != nil {
		return nil, err
	}

	// 5. Асинхронно отправляем аудит
	s.sendAuditMessage(ctx, models.ActionUpdate, userID, taskID, oldTask, updatedTask, updates)

	return updatedTask, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID int, userID int) error {
	// 1. Получаем задачу (для аудита и проверки прав)
	task, err := s.taskRepo.GetByTaskId(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return models.ErrTaskNotFound
	}

	// 2. Проверяем права доступа
	if task.OwnerId != userID {
		return models.ErrForbidden
	}

	// 3. Удаляем задачу
	err = s.taskRepo.Delete(ctx, taskID)
	if err != nil {
		return err
	}

	// 4. Асинхронно отправляем аудит
	s.sendAuditMessage(ctx, models.ActionDelete, userID, taskID, task, nil, nil)

	return nil
}

func (s *TaskService) ListTasks(ctx context.Context, userID int, status string) ([]models.Task, error) {
	return s.taskRepo.List(ctx, userID, status)
}

// Вспомогательный метод для отправки аудита
func (s *TaskService) sendAuditMessage(
	ctx context.Context,
	action models.ActionType,
	userID int,
	taskID int,
	oldTask *models.Task,
	newTask *models.Task,
	updates map[string]interface{},
) {
	auditMsg := &models.AuditMessage{
		Action:    action,
		UserID:    userID,
		EntityID:  taskID,
		Timestamp: time.Now(),
	}

	// Заполняем данные в зависимости от действия
	switch action {
	case models.ActionCreate:
		if newTask != nil {
			auditMsg.NewValues = map[string]interface{}{
				"title":       newTask.Title,
				"description": newTask.Description,
				"status":      newTask.Status,
				"owner_id":    newTask.OwnerId,
			}
		}

	case models.ActionUpdate:
		if oldTask != nil && newTask != nil {
			auditMsg.OldValues = map[string]interface{}{
				"title":       oldTask.Title,
				"description": oldTask.Description,
				"status":      oldTask.Status,
			}
			auditMsg.NewValues = map[string]interface{}{
				"title":       newTask.Title,
				"description": newTask.Description,
				"status":      newTask.Status,
			}
			// Вычисляем изменения
			changes := make(map[string]interface{})
			if oldTask.Title != newTask.Title {
				changes["title"] = map[string]interface{}{"old": oldTask.Title, "new": newTask.Title}
			}
			if oldTask.Description != newTask.Description {
				changes["description"] = map[string]interface{}{"old": oldTask.Description, "new": newTask.Description}
			}
			if oldTask.Status != newTask.Status {
				changes["status"] = map[string]interface{}{"old": oldTask.Status, "new": newTask.Status}
			}
			auditMsg.Changes = changes
		}

	case models.ActionDelete:
		if oldTask != nil {
			auditMsg.OldValues = map[string]interface{}{
				"title":       oldTask.Title,
				"description": oldTask.Description,
				"status":      oldTask.Status,
				"owner_id":    oldTask.OwnerId,
			}
		}
	}

	// Асинхронная отправка в RabbitMQ
	go func() {
		if err := s.rabbitMQ.PublishAuditMessage(context.Background(), auditMsg); err != nil {
			log.Printf("❌ Ошибка отправки аудита в RabbitMQ: %v", err)
		} else {
			log.Printf("📨 Аудит отправлен в RabbitMQ: %s задача ID=%d", action, taskID)
		}
	}()
}
