package models

// Contact represents supplier or partner contact information.
type Contact struct {
	ID           int     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string  `gorm:"column:name;type:char;not null" json:"name"`
	CountryID    int     `gorm:"column:country_id;not null" json:"country_id"`
	MobileNumber string  `gorm:"column:mobile_number;type:char;not null" json:"mobile_number"`
	PhoneNumber  string  `gorm:"column:phone_number;type:char;not null" json:"phone_number"`
	Website      string  `gorm:"column:website;type:char;not null" json:"website"`
	Email        string  `gorm:"column:email;type:char;not null" json:"email"`
	Country      Country `gorm:"foreignKey:CountryID;references:ID" json:"country,omitempty"`
}

func (Contact) TableName() string {
	return `"Contact"`
}

