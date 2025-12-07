package models

import "time"

// CertificateRequest represents a certificate request from a container
type CertificateRequest struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ContainerID string    `gorm:"column:container_id;type:varchar(255);not null;index" json:"container_id"`
	CSRID       *int      `gorm:"column:csr_id" json:"csr_id,omitempty"` // Reference to certificate service CSR ID
	RequestType string    `gorm:"column:request_type;type:varchar(50);not null" json:"request_type"` // container, application
	AppName     *string   `gorm:"column:app_name;type:varchar(255)" json:"app_name,omitempty"`
	Status      string    `gorm:"column:status;type:varchar(50);default:'pending'" json:"status"` // pending, approved, rejected, issued
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (CertificateRequest) TableName() string {
	return "certificate_requests"
}

