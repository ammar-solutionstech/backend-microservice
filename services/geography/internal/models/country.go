package models

// Country represents entries in the Country table.
type Country struct {
	ID   int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Code string `gorm:"column:code;type:varchar(50);not null" json:"code"`
}

func (Country) TableName() string {
	return `"Country"`
}

// Nationality is kept as an alias for backwards compatibility
type Nationality = Country

