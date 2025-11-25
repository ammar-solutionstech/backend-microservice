package models

import "time"

// Permission represents a permission in the auth service
type Permission struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Module    string    `gorm:"column:module;type:varchar(255);not null" json:"module"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Permission) TableName() string {
	return `"Permission"`
}

