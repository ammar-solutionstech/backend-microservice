package models

import "time"

// HelpDeskType represents a type/category for help desk tickets
type HelpDeskType struct {
	ID          int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description string     `gorm:"column:description;type:text;not null" json:"description"`
	TeamID      *int       `gorm:"column:team_id" json:"team_id,omitempty"`
	IsActive    *bool      `gorm:"column:is_active" json:"is_active,omitempty"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`
	DeletedAt   *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`
}

func (HelpDeskType) TableName() string {
	return `"Help_desk_type"`
}

// Team represents a help desk team
type Team struct {
	ID          int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description *string    `gorm:"column:description;type:text" json:"description,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Team) TableName() string {
	return `"Team"`
}

