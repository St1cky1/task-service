package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/St1cky1/task-service/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewRabbitMQClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Объявляем очередь для аудита
	queue, err := channel.QueueDeclare(
		"task_audit_logs", // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:    conn,
		channel: channel,
		queue:   queue,
	}, nil
}

// GetChannel возвращает AMQP channel для использования в consumer'ах
func (c *Client) GetChannel() *amqp.Channel {
	return c.channel
}

// GetQueueName возвращает имя очереди
func (c *Client) GetQueueName() string {
	return c.queue.Name
}

func (c *Client) PublishAuditMessage(ctx context.Context, message *models.AuditMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = c.channel.PublishWithContext(
		ctx,
		"",           // exchange
		c.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Сообщения сохраняются на диск
		},
	)

	if err != nil {
		return err
	}

	log.Printf("Отправлено сообщение в RabbitMQ: %s для задачи ID=%d", message.Action, message.EntityID)
	return nil
}

func (c *Client) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
