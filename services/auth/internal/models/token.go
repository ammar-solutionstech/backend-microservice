package models

import "time"

// RefreshToken represents a refresh token for a user
type RefreshToken struct {
	ID        int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    int        `gorm:"column:user_id;not null;index" json:"user_id"`
	Token     string     `gorm:"column:token;type:text;not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null;index" json:"expires_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at" json:"revoked_at,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// TokenBlacklist represents a blacklisted token
type TokenBlacklist struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TokenID   string    `gorm:"column:token_id;type:varchar(255);not null;uniqueIndex" json:"token_id"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

