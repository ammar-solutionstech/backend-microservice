package models

import "time"

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(255);not null;uniqueIndex" json:"name"`
	Subject   string    `gorm:"column:subject;type:varchar(500);not null" json:"subject"`
	Body      string    `gorm:"column:body;type:text;not null" json:"body"`
	Variables []string  `gorm:"column:variables;type:text[]" json:"variables"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (EmailTemplate) TableName() string {
	return "email_templates"
}

// SMSTemplate represents an SMS template
type SMSTemplate struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(255);not null;uniqueIndex" json:"name"`
	Content   string    `gorm:"column:content;type:text;not null" json:"content"`
	Variables []string  `gorm:"column:variables;type:text[]" json:"variables"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SMSTemplate) TableName() string {
	return "sms_templates"
}

