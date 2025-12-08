package models

import "time"

// Organization represents an organization/tenant
type Organization struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Domain      string    `gorm:"column:domain;type:varchar(255);uniqueIndex;not null" json:"domain"`
	AdminEmail  string    `gorm:"column:admin_email;type:varchar(255);not null" json:"admin_email"`
	AdminPhone  *string   `gorm:"column:admin_phone;type:varchar(255)" json:"admin_phone,omitempty"`
	Status      string    `gorm:"column:status;type:varchar(50);default:'active'" json:"status"` // active, suspended, revoked
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Organization) TableName() string {
	return "organizations"
}

