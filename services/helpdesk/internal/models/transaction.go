package models

import "time"

// HelpDeskTransaction represents a transaction on a help desk ticket
type HelpDeskTransaction struct {
	ID                int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name              string    `gorm:"column:name;type:char;not null" json:"name"`
	TransactionTypeID int       `gorm:"column:transaction_type_id;not null" json:"transaction_type_id"`
	DateTime          time.Time `gorm:"column:date_time;type:date;not null" json:"date_time"`
	TimeSpent         float64   `gorm:"column:time;not null" json:"time"`
	HelpDeskID        int       `gorm:"column:help_desk_id;not null;index" json:"help_desk_id"`
	CreatedAt         time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (HelpDeskTransaction) TableName() string {
	return `"Help_desk_transaction"`
}

// TransactionType represents a type of transaction
type TransactionType struct {
	ID          int     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string  `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description *string `gorm:"column:description;type:text" json:"description,omitempty"`
}

func (TransactionType) TableName() string {
	return `"Transaction_type"`
}

// HelpDeskRating represents a rating for a help desk ticket
type HelpDeskRating struct {
	ID         int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	HelpDeskID int        `gorm:"column:help_desk_id;not null;index" json:"help_desk_id"`
	UserID     int        `gorm:"column:user_id;not null" json:"user_id"` // Reference to Auth Service user
	Rating     int        `gorm:"column:rating;not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment    *string    `gorm:"column:comment;type:text" json:"comment,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (HelpDeskRating) TableName() string {
	return `"Help_desk_rating"`
}

// UserHelpDesk represents the user_help_desk join table (no foreign key to User)
type UserHelpDesk struct {
	UserID     int `gorm:"column:user_id;primaryKey" json:"user_id"` // Reference to Auth Service user
	HelpDeskID int `gorm:"column:help_desk_id;primaryKey" json:"help_desk_id"`
}

func (UserHelpDesk) TableName() string {
	return "user_help_desk"
}

