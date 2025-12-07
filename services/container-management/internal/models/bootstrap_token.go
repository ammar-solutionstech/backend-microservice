package models

import "time"

// BootstrapToken represents a bootstrap token for container registration
type BootstrapToken struct {
	ID          int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TokenHash   string     `gorm:"column:token_hash;type:varchar(255);uniqueIndex;not null" json:"-"`
	ContainerID *string    `gorm:"column:container_id;type:varchar(255)" json:"container_id,omitempty"`
	ExpiresAt   time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	UsedAt      *time.Time `gorm:"column:used_at" json:"used_at,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (BootstrapToken) TableName() string {
	return "bootstrap_tokens"
}

