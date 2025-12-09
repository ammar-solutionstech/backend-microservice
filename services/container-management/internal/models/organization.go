package models

import "time"

// Organization represents an organization/tenant
type Organization struct {
	// ID is the primary key, auto-incrementing integer
	ID int `gorm:"column:id;primaryKey;autoIncrement" json:"id"`

	// Name is the organization name, required, max 255 characters
	Name string `gorm:"column:name;type:varchar(255);not null" json:"name"`

	// Domain is the organization domain (e.g., "acme.example.com"), required, unique, max 255 characters
	Domain string `gorm:"column:domain;type:varchar(255);uniqueIndex;not null" json:"domain"`

	// AdminEmail is the organization administrator email address, required, max 255 characters
	AdminEmail string `gorm:"column:admin_email;type:varchar(255);not null" json:"admin_email"`

	// AdminPhone is the organization administrator phone number, optional, max 255 characters
	AdminPhone *string `gorm:"column:admin_phone;type:varchar(255)" json:"admin_phone,omitempty"`

	// Status is the organization status. Options: "active", "suspended", "revoked". Default: "active"
	Status string `gorm:"column:status;type:varchar(50);default:'active'" json:"status"`

	// CreatedAt is the timestamp when organization was created
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`

	// UpdatedAt is the timestamp when organization was last updated
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Organization) TableName() string {
	return "organizations"
}
