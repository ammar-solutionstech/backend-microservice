package models

import "time"

// Revocation represents a certificate revocation record
type Revocation struct {
	ID            int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CertificateID int       `gorm:"column:certificate_id;not null" json:"certificate_id"`
	SerialNumber  string    `gorm:"column:serial_number;type:varchar(255);not null" json:"serial_number"`
	RevokedAt     time.Time `gorm:"column:revoked_at;not null" json:"revoked_at"`
	Reason        int       `gorm:"column:reason;not null;default:0" json:"reason"` // RFC 5280 revocation reason codes
	RevokedBy     *int      `gorm:"column:revoked_by" json:"revoked_by,omitempty"`
	CreatedAt     time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (Revocation) TableName() string {
	return "revocations"
}
