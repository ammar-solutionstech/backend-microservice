package models

// EquipmentSoftware links software to equipment with versioning.
type EquipmentSoftware struct {
	SoftwareID      int       `gorm:"column:software_id;primaryKey" json:"software_id"`
	EquipmentID     int       `gorm:"column:equipment_id;primaryKey" json:"equipment_id"`
	SoftwareVersion string    `gorm:"column:software_version;type:char;not null" json:"software_version"`
	SoftwareSize    int       `gorm:"column:software_size;not null" json:"software_size"`
	License         string    `gorm:"column:license;type:char;not null" json:"license"`
	Software        Software  `gorm:"foreignKey:SoftwareID;references:ID" json:"software,omitempty"`
	Equipment       Equipment `gorm:"foreignKey:EquipmentID;references:ID" json:"equipment,omitempty"`
}

func (EquipmentSoftware) TableName() string {
	return `"equipment_softwares"`
}

