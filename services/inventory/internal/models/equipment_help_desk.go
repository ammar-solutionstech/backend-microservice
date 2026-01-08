package models

// EquipmentHelpDesk links equipment to help desk tickets.
// Note: HelpDeskID is a reference only, no foreign key (Help Desk is in another service)
type EquipmentHelpDesk struct {
	EquipmentID int `gorm:"column:equipment_id" json:"equipment_id"`
	HelpDeskID  int `gorm:"column:help_desk_id" json:"help_desk_id"`
}

func (EquipmentHelpDesk) TableName() string {
	return `"equipment_help_desk"`
}

