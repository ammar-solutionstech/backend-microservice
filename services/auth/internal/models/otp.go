package models

import "time"

// OTPVerification represents an OTP verification record
type OTPVerification struct {
	ID        int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    int        `gorm:"column:user_id;not null;index" json:"user_id"`
	Type      string     `gorm:"column:type;type:varchar(20);not null;check:type IN ('email', 'phone')" json:"type"` // 'email' or 'phone'
	Code      string     `gorm:"column:code;type:varchar(6);not null;index" json:"code"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null;index" json:"expires_at"`
	VerifiedAt *time.Time `gorm:"column:verified_at" json:"verified_at,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (OTPVerification) TableName() string {
	return "otp_verifications"
}

// PasswordReset represents a password reset token
type PasswordReset struct {
	ID        int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    int        `gorm:"column:user_id;not null;index" json:"user_id"`
	Token     string     `gorm:"column:token;type:varchar(255);not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null;index" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (PasswordReset) TableName() string {
	return "password_resets"
}

