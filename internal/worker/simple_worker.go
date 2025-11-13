package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/St1cky1/task-service/internal/models"
	"github.com/St1cky1/task-service/internal/repo"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleAuditWorker struct {
	auditRepo *repo.TaskAuditRepository
}

func NewSimpleAuditWorker(auditRepo *repo.TaskAuditRepository) *SimpleAuditWorker {
	return &SimpleAuditWorker{
		auditRepo: auditRepo,
	}
}

func (w *SimpleAuditWorker) Start(ctx context.Context) {
	fmt.Println("🔄 Simple Worker: Начинаем подключение к RabbitMQ...")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Simple Worker остановлен")
			return
		default:
			err := w.runWorker(ctx)
			if err != nil {
				log.Printf("❌ Simple Worker ошибка: %v, переподключение через 5 секунд...", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (w *SimpleAuditWorker) runWorker(ctx context.Context) error {
	// Создаем отдельное соединение
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("ошибка подключения: %w", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("ошибка создания канала: %w", err)
	}
	defer channel.Close()

	// Убеждаемся, что очередь существует
	_, err = channel.QueueDeclarePassive(
		"task_audit_logs", // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("очередь не найдена: %w", err)
	}

	fmt.Println("✅ Simple Worker: Очередь найдена, начинаем потребление...")

	// Создаем consumer
	msgs, err := channel.Consume(
		"task_audit_logs", // queue
		"simple_worker",   // consumer tag
		false,             // auto-ack (false - подтверждаем вручную)
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return fmt.Errorf("ошибка создания consumer: %w", err)
	}

	fmt.Println("🎯 Simple Worker: Успешно запущен, ожидаем сообщения...")

	// Обрабатываем сообщения
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("канал сообщений закрыт")
			}

			fmt.Printf("📥 ПОЛУЧЕНО СООБЩЕНИЕ! Длина: %d байт\n", len(msg.Body))

			// Простая обработка - логируем и подтверждаем
			var auditMsg models.AuditMessage
			if err := json.Unmarshal(msg.Body, &auditMsg); err != nil {
				log.Printf("❌ Ошибка парсинга JSON: %v", err)
				log.Printf("📄 Сырое сообщение: %s", string(msg.Body))
				msg.Nack(false, false) // Не возвращаем в очередь
			} else {
				fmt.Printf("✅ Сообщение распаршено: %s задача %d от пользователя %d\n",
					auditMsg.Action, auditMsg.EntityID, auditMsg.UserID)

				// Сохраняем в БД
				taskAudit, err := w.convertToTaskAudit(&auditMsg)
				if err != nil {
					log.Printf("❌ Ошибка конвертации: %v", err)
					msg.Nack(false, true) // Возвращаем в очередь
				} else {
					if err := w.auditRepo.Create(context.Background(), taskAudit); err != nil {
						log.Printf("❌ Ошибка сохранения в БД: %v", err)
						msg.Nack(false, true) // Возвращаем в очередь
					} else {
						msg.Ack(false) // Подтверждаем обработку
						fmt.Printf("💾 Аудит сохранен в БД: %s задача ID=%d\n",
							taskAudit.Action, taskAudit.EntityID)
					}
				}
			}
		}
	}
}

func (w *SimpleAuditWorker) convertToTaskAudit(msg *models.AuditMessage) (*models.TaskAudit, error) {
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
