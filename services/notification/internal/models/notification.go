package models

import "time"

// Notification represents a notification record
type Notification struct {
	ID           int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Type         string     `gorm:"column:type;type:varchar(20);not null;check:type IN ('email', 'sms')" json:"type"`
	Recipient    string     `gorm:"column:recipient;type:varchar(255);not null;index" json:"recipient"`
	TemplateID   *int       `gorm:"column:template_id" json:"template_id,omitempty"`
	TemplateName *string    `gorm:"column:template_name;type:varchar(255)" json:"template_name,omitempty"`
	Subject      *string    `gorm:"column:subject;type:varchar(500)" json:"subject,omitempty"`
	Content      string     `gorm:"column:content;type:text;not null" json:"content"`
	Status       string     `gorm:"column:status;type:varchar(20);default:'pending';check:status IN ('pending', 'sent', 'failed');index" json:"status"`
	SentAt       *time.Time `gorm:"column:sent_at" json:"sent_at,omitempty"`
	Error        *string    `gorm:"column:error;type:text" json:"error,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP;index" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

// NotificationQueue represents a notification queue entry
type NotificationQueue struct {
	ID             int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	NotificationID int        `gorm:"column:notification_id;not null;index" json:"notification_id"`
	QueueName      string     `gorm:"column:queue_name;type:varchar(255);not null" json:"queue_name"`
	Status         string     `gorm:"column:status;type:varchar(20);default:'pending';check:status IN ('pending', 'processing', 'completed', 'failed');index" json:"status"`
	RetryCount     int        `gorm:"column:retry_count;default:0" json:"retry_count"`
	Error          *string    `gorm:"column:error;type:text" json:"error,omitempty"`
	CreatedAt      time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (NotificationQueue) TableName() string {
	return "notification_queue"
}

