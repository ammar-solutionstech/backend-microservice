package models

// Software represents software assets.
type Software struct {
	ID             int              `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SoftwareName   string           `gorm:"column:software_name;type:char;not null" json:"software_name"`
	CategoryID     int              `gorm:"column:category_id;not null" json:"category_id"`
	LicenseExpDate int64            `gorm:"column:license_exp_date;not null" json:"license_exp_date"`
	Category       SoftwareCategory `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"`
}

func (Software) TableName() string {
	return `"Software"`
}

