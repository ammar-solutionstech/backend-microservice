package models

import "time"

// Certificate represents an issued certificate
type Certificate struct {
	ID             int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SerialNumber   string    `gorm:"column:serial_number;type:varchar(255);not null;uniqueIndex" json:"serial_number"`
	CSRID          *int      `gorm:"column:csr_id" json:"csr_id,omitempty"`
	CertificatePEM string    `gorm:"column:certificate_pem;type:text;not null" json:"certificate_pem"`
	IssuedAt       time.Time `gorm:"column:issued_at;not null" json:"issued_at"`
	ExpiresAt      time.Time `gorm:"column:expires_at;not null" json:"expires_at"`
	Status         string    `gorm:"column:status;type:varchar(50);not null;default:'active'" json:"status"` // active, revoked, expired
	StepCACertID   *string   `gorm:"column:step_ca_cert_id;type:varchar(255)" json:"step_ca_cert_id,omitempty"`
	CreatedAt      time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Certificate) TableName() string {
	return "certificates"
}
