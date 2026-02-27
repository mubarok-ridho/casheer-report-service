package messaging

import (
	"encoding/json"
	"log"
	"os"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

type OrderCompletedEvent struct {
	OrderID     uint    `json:"order_id"`
	TenantID    uint    `json:"tenant_id"`
	TotalAmount float64 `json:"total_amount"`
	Date        string  `json:"date"`
}

type ExpenseAddedEvent struct {
	ExpenseID uint    `json:"expense_id"`
	TenantID  uint    `json:"tenant_id"`
	Amount    float64 `json:"amount"`
	Category  string  `json:"category"`
	Date      string  `json:"date"`
}

func NewRabbitMQ() (*RabbitMQ, error) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		Conn:    conn,
		Channel: ch,
	}, nil
}

func (r *RabbitMQ) ConsumeOrderCompleted(handler func(OrderCompletedEvent)) error {
	queueName := os.Getenv("RABBITMQ_QUEUE_ORDER_COMPLETED")
	if queueName == "" {
		queueName = "order_completed"
	}

	q, err := r.Channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := r.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			var event OrderCompletedEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("Error decoding order completed event: %v", err)
				continue
			}
			handler(event)
		}
	}()

	log.Println("📡 Listening for order completed events...")
	return nil
}

func (r *RabbitMQ) PublishExpenseAdded(event ExpenseAddedEvent) error {
	queueName := os.Getenv("RABBITMQ_QUEUE_EXPENSE_ADDED")
	if queueName == "" {
		queueName = "expense_added"
	}

	_, err := r.Channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = r.Channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		return err
	}

	log.Printf("📤 Published expense added event: ExpenseID=%d", event.ExpenseID)
	return nil
}

func (r *RabbitMQ) Close() {
	r.Channel.Close()
	r.Conn.Close()
}
