package services

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"backend/services/helpdesk/internal/config"
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

	return &rabbitMQPublisher{conn: conn, ch: ch}
}

func (p *rabbitMQPublisher) PublishTicketCreated(ticketID, userID int) {
	if p.ch == nil {
		return
	}

	event := map[string]interface{}{
		"event":     "ticket.created",
		"ticket_id": ticketID,
		"user_id":   userID,
	}

	p.publishEvent("ticket.created", event)
}

func (p *rabbitMQPublisher) PublishTicketUpdated(ticketID, userID int) {
	if p.ch == nil {
		return
	}

	event := map[string]interface{}{
		"event":     "ticket.updated",
		"ticket_id": ticketID,
		"user_id":   userID,
	}

	p.publishEvent("ticket.updated", event)
}

func (p *rabbitMQPublisher) PublishTicketAssigned(ticketID, userID int) {
	if p.ch == nil {
		return
	}

	event := map[string]interface{}{
		"event":     "ticket.assigned",
		"ticket_id": ticketID,
		"user_id":   userID,
	}

	p.publishEvent("ticket.assigned", event)
}

func (p *rabbitMQPublisher) publishEvent(routingKey string, event map[string]interface{}) {
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	err = p.ch.Publish(
		"helpdesk_events",
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
