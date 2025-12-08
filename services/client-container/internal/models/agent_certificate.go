package models

import "time"

// AgentCertificate represents an agent certificate
type AgentCertificate struct {
	ID              int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DeviceID        int        `gorm:"column:device_id;not null;index" json:"device_id"`
	CertificateSerial string   `gorm:"column:certificate_serial;type:varchar(255);not null" json:"certificate_serial"`
	CertificatePEM  string    `gorm:"column:certificate_pem;type:text;not null" json:"certificate_pem"`
	IssuedAt        time.Time `gorm:"column:issued_at;not null" json:"issued_at"`
	ExpiresAt       time.Time `gorm:"column:expires_at;not null" json:"expires_at"`
	RevokedAt       *time.Time `gorm:"column:revoked_at" json:"revoked_at,omitempty"`
	Status          string    `gorm:"column:status;type:varchar(50);default:'active'" json:"status"` // active, revoked, expired
	CreatedAt       time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (AgentCertificate) TableName() string {
	return "agent_certificates"
}

