package workers

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"backend/services/notification/internal/config"
	"backend/services/notification/internal/services"
)

func StartNotificationWorker(cfg *config.Config, notificationService *services.NotificationService) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("Warning: failed to connect to RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Warning: failed to open RabbitMQ channel: %v", err)
		return
	}
	defer ch.Close()

	// Declare exchanges and queues
	err = ch.ExchangeDeclare(
		"notifications",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error declaring exchange: %v", err)
		return
	}

	q, err := ch.QueueDeclare(
		"notification_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error declaring queue: %v", err)
		return
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		q.Name,
		"notification.*",
		"notifications",
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error binding queue: %v", err)
		return
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error consuming messages: %v", err)
		return
	}

	log.Println("Notification worker started, waiting for messages...")

	for msg := range msgs {
		go processNotification(msg, notificationService, ch)
	}
}

func processNotification(msg amqp.Delivery, notificationService *services.NotificationService, ch *amqp.Channel) {
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		msg.Nack(false, false)
		return
	}

	eventType, _ := event["event"].(string)
	variables := make(map[string]string)
	if vars, ok := event["variables"].(map[string]interface{}); ok {
		for k, v := range vars {
			if str, ok := v.(string); ok {
				variables[k] = str
			}
		}
	}

	var err error
	switch eventType {
	case "notification.send.email":
		to, _ := event["to"].(string)
		subject, _ := event["subject"].(string)
		body, _ := event["body"].(string)
		templateName, _ := event["template_name"].(string)
		err = notificationService.SendEmail(to, subject, body, templateName, variables)
	case "notification.send.sms":
		to, _ := event["to"].(string)
		content, _ := event["content"].(string)
		templateName, _ := event["template_name"].(string)
		err = notificationService.SendSMS(to, content, templateName, variables)
	default:
		log.Printf("Unknown event type: %s", eventType)
		msg.Nack(false, false)
		return
	}

	if err != nil {
		log.Printf("Error processing notification: %v", err)
		msg.Nack(false, true) // Requeue on error
		return
	}

	msg.Ack(false)
	log.Printf("Processed notification: %s", eventType)
}
