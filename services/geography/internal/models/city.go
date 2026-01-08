package models

// City represents a city linked to a country.
type City struct {
	ID        int     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string  `gorm:"column:name;type:char;not null" json:"name"`
	CountryID int     `gorm:"column:country_id;not null" json:"country_id"`
	Country   Country `gorm:"foreignKey:CountryID;references:ID" json:"country,omitempty"`
}

func (City) TableName() string {
	return `"City"`
}

