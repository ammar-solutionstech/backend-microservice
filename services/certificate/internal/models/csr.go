package models

import "time"

// CSRRequest represents a Certificate Signing Request
type CSRRequest struct {
	ID              int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CSRPEM          string    `gorm:"column:csr_pem;type:text;not null" json:"csr_pem"`
	Status          string    `gorm:"column:status;type:varchar(50);not null;default:'pending'" json:"status"` // pending, approved, rejected
	RequesterUserID *int      `gorm:"column:requester_user_id" json:"requester_user_id,omitempty"`
	RequesterEmail  string    `gorm:"column:requester_email;type:varchar(255)" json:"requester_email"`
	ApprovedBy      *int      `gorm:"column:approved_by" json:"approved_by,omitempty"`
	RejectionReason *string   `gorm:"column:rejection_reason;type:text" json:"rejection_reason,omitempty"`
	CreatedAt       time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (CSRRequest) TableName() string {
	return "csr_requests"
}
