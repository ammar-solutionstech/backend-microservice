package models

import "time"

// Maintenance captures maintenance events.
// Note: ContactID is a reference only, no foreign key (Contact is in Geography service)
type Maintenance struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;type:char;not null" json:"name"`
	StartDate time.Time `gorm:"column:start_date;type:date;not null" json:"start_date"`
	EndDate   time.Time `gorm:"column:end_date;type:date;not null" json:"end_date"`
	ContactID int       `gorm:"column:contact_id;not null" json:"contact_id"` // Reference only, no FK
}

func (Maintenance) TableName() string {
	return `"Maintenance"`
}

