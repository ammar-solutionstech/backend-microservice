package models

// Brand represents a hardware brand/manufacturer record.
type Brand struct {
	ID   int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"column:name;type:char;not null" json:"name"`
}

func (Brand) TableName() string {
	return `"Brand"`
}

