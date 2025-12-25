package rabbitmq

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(url string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < 30; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			log.Println("Connected to RabbitMQ")
			return conn, nil
		}
		log.Printf("Waiting for RabbitMQ... attempt %d", i+1)
		time.Sleep(2 * time.Second)
	}

	return nil, err
}

func SetupExchangeAndQueues(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"payment_exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	queues := []struct {
		name       string
		routingKey string
	}{
		{"payment_requests", "payment_request"},
		{"payment_results", "payment_result"},
	}

	for _, q := range queues {
		_, err = ch.QueueDeclare(q.name, true, false, false, false, nil)
		if err != nil {
			return err
		}
		err = ch.QueueBind(q.name, q.routingKey, "payment_exchange", false, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
