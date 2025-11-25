package services

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"backend/services/auth/internal/config"
)

type rabbitMQPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func newRabbitMQPublisher(cfg *config.Config) *rabbitMQPublisher {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("Warning: failed to connect to RabbitMQ: %v", err)
		return &rabbitMQPublisher{conn: nil, ch: nil}
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Warning: failed to open RabbitMQ channel: %v", err)
		return &rabbitMQPublisher{conn: conn, ch: nil}
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		"auth_events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error declaring exchange: %v", err)
	}

	return &rabbitMQPublisher{conn: conn, ch: ch}
}

func (p *rabbitMQPublisher) PublishUserRegistered(userID int, email string) {
	if p.ch == nil {
		return
	}

	event := map[string]interface{}{
		"event":   "user.registered",
		"user_id": userID,
		"email":   email,
	}

	p.publishEvent("user.registered", event)
}

func (p *rabbitMQPublisher) PublishPasswordResetRequested(userID int, email string) {
	if p.ch == nil {
		return
	}

	event := map[string]interface{}{
		"event":   "password.reset.requested",
		"user_id": userID,
		"email":   email,
	}

	p.publishEvent("password.reset.requested", event)
}

func (p *rabbitMQPublisher) publishEvent(routingKey string, event map[string]interface{}) {
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	err = p.ch.Publish(
		"auth_events",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("Error publishing event: %v", err)
	}
}
