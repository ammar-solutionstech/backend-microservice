package models

// SoftwareCategory represents a software category.
type SoftwareCategory struct {
	ID   int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"column:name;type:char;not null" json:"name"`
}

func (SoftwareCategory) TableName() string {
	return `"Software_category"`
}

