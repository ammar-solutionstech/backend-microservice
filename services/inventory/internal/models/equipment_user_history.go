package models

import "time"

// EquipmentUserHistory tracks assignment history.
// Note: UserID is a reference only, no foreign key (User is in Auth service)
type EquipmentUserHistory struct {
	EquipmentID int       `gorm:"column:equipment_id;primaryKey" json:"equipment_id"`
	UserID      int       `gorm:"column:user_id;primaryKey" json:"user_id"`
	StartDate   time.Time `gorm:"column:start_date;type:date;primaryKey" json:"start_date"`
	EndDate     time.Time `gorm:"column:end_date;type:date;not null" json:"end_date"`
	Comment     string    `gorm:"column:comment;type:char;not null" json:"comment"`
	Equipment   Equipment `gorm:"foreignKey:EquipmentID;references:ID" json:"equipment,omitempty"`
}

func (EquipmentUserHistory) TableName() string {
	return `"Equipment_User_History"`
}

