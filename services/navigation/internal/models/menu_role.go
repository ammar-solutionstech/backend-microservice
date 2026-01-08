package models

// MenuRole links menus to roles.
// Note: RoleID is a reference only, no foreign key (Role is in Auth service)
type MenuRole struct {
	MenuID int `gorm:"column:menu_id;primaryKey" json:"menu_id"`
	RoleID int `gorm:"column:role_id;primaryKey" json:"role_id"` // Reference only, no FK
	Menu   Menu `gorm:"foreignKey:MenuID;references:ID" json:"menu,omitempty"`
}

func (MenuRole) TableName() string {
	return `"menu_role"`
}

