package models

import "time"

// Role represents a role in the auth service
type Role struct {
	ID          int          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"column:name;type:varchar(255);not null;uniqueIndex" json:"name"`
	Description *string      `gorm:"column:description;type:text" json:"description,omitempty"`
	CreatedAt   time.Time    `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Permissions []Permission `gorm:"many2many:role_permission;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:PermissionID" json:"permissions,omitempty"`
	Users       []User       `gorm:"many2many:user_role;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:UserID" json:"users,omitempty"`
}

func (Role) TableName() string {
	return `"Role"`
}

// RolePermission represents the role_permission join table
type RolePermission struct {
	RoleID       int `gorm:"column:role_id;primaryKey" json:"role_id"`
	PermissionID int `gorm:"column:permission_id;primaryKey" json:"permission_id"`
}

func (RolePermission) TableName() string {
	return "role_permission"
}

