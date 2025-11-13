package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/St1cky1/task-service/internal/models"
	"github.com/St1cky1/task-service/internal/rabbitmq"
	"github.com/St1cky1/task-service/internal/repo"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AuditWorker struct {
	rabbitMQ  *rabbitmq.Client
	auditRepo *repo.TaskAuditRepository
}

func NewAuditWorker(rabbitMQ *rabbitmq.Client, auditRepo *repo.TaskAuditRepository) *AuditWorker {
	return &AuditWorker{
		rabbitMQ:  rabbitMQ,
		auditRepo: auditRepo,
	}
}

func (w *AuditWorker) Start(ctx context.Context) {
	// Создаем отдельное соединение и канал для consumer'а
	rabbitMQURL := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Printf("❌ Ошибка подключения к RabbitMQ для воркера: %v", err)
		return
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Printf("❌ Ошибка создания канала для воркера: %v", err)
		return
	}
	defer channel.Close()

	// Убеждаемся, что очередь существует
	queueName := "task_audit_logs"
	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Printf("❌ Ошибка объявления очереди: %v", err)
		return
	}

	// Создаем consumer для очереди
	msgs, err := channel.Consume(
		queueName,      // queue
		"audit_worker", // consumer tag (уникальный идентификатор)
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Printf("❌ Ошибка создания consumer: %v", err)
		return
	}

	fmt.Println("✅ Audit Worker запущен. Ожидаем сообщения...")

	// Обрабатываем сообщения
	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Audit Worker остановлен")
			return
		case msg, ok := <-msgs:
			if !ok {
				fmt.Println("📨 Канал сообщений закрыт")
				return
			}
			w.processMessage(msg, channel)
		}
	}
}

func (w *AuditWorker) processMessage(msg amqp.Delivery, channel *amqp.Channel) {
	ctx := context.Background()

	log.Printf("📥 Получено сообщение: %s", msg.Body)

	// 1. Парсим сообщение
	var auditMsg models.AuditMessage
	if err := json.Unmarshal(msg.Body, &auditMsg); err != nil {
		log.Printf("❌ Ошибка парсинга сообщения: %v", err)
		msg.Nack(false, false) // Не возвращаем в очередь
		return
	}

	// 2. Конвертируем в TaskAudit
	taskAudit, err := w.convertToTaskAudit(&auditMsg)
	if err != nil {
		log.Printf("❌ Ошибка конвертации: %v", err)
		msg.Nack(false, true) // Возвращаем в очередь для повторной обработки
		return
	}

	// 3. Сохраняем в БД
	if err := w.auditRepo.Create(ctx, taskAudit); err != nil {
		log.Printf("❌ Ошибка сохранения аудита: %v", err)
		msg.Nack(false, true) // Возвращаем в очередь для повторной обработки
		return
	}

	// 4. Подтверждаем обработку
	msg.Ack(false)
	log.Printf("✅ Аудит сохранен: %s задача ID=%d", taskAudit.Action, taskAudit.EntityID)
}

func (w *AuditWorker) convertToTaskAudit(msg *models.AuditMessage) (*models.TaskAudit, error) {
	// Конвертируем map[string]any в JSON строки
	var oldValuesJSON, newValuesJSON, changesJSON *string

	if msg.OldValues != nil {
		oldJSON, err := json.Marshal(msg.OldValues)
		if err != nil {
			return nil, err
		}
		oldStr := string(oldJSON)
		oldValuesJSON = &oldStr
	}

	if msg.NewValues != nil {
		newJSON, err := json.Marshal(msg.NewValues)
		if err != nil {
			return nil, err
		}
		newStr := string(newJSON)
		newValuesJSON = &newStr
	}

	if msg.Changes != nil {
		changesJSONBytes, err := json.Marshal(msg.Changes)
		if err != nil {
			return nil, err
		}
		changesStr := string(changesJSONBytes)
		changesJSON = &changesStr
	}

	return &models.TaskAudit{
		UserID:     msg.UserID,
		Action:     msg.Action,
		EntityType: "task",
		EntityID:   msg.EntityID,
		OldValues:  oldValuesJSON,
		NewValues:  newValuesJSON,
		Changes:    changesJSON,
		ChangesAt:  msg.Timestamp,
	}, nil
}
