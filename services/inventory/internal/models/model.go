package models

// Model represents a specific hardware model tied to a brand.
type Model struct {
	ID      int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name    string `gorm:"column:name;type:char;not null" json:"name"`
	BrandID int    `gorm:"column:brand_id;not null" json:"brand_id"`
	Brand   Brand  `gorm:"foreignKey:BrandID;references:ID" json:"brand,omitempty"`
}

func (Model) TableName() string {
	return `"Model"`
}

