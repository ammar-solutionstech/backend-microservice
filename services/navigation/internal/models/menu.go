package models

// Menu represents a navigation menu item.
type Menu struct {
	ID    int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name  string `gorm:"column:name;type:char;not null" json:"name"`
	Title string `gorm:"column:title;type:char;not null" json:"title"`
	URI   string `gorm:"column:uri;type:char;not null" json:"uri"`
}

func (Menu) TableName() string {
	return `"Menu"`
}

