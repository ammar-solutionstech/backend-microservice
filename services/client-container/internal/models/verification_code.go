package models

import "time"

// VerificationCode represents a verification code for agent certificate issuance
type VerificationCode struct {
	ID         int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DeviceID   int        `gorm:"column:device_id;not null;index" json:"device_id"`
	CodeHash   string     `gorm:"column:code_hash;type:varchar(255);not null;index" json:"-"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	UsedAt     *time.Time `gorm:"column:used_at" json:"used_at,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}

