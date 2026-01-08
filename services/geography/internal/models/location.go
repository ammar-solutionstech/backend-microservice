package models

// Location represents a physical location for equipment.
type Location struct {
	ID             int      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name           string   `gorm:"column:name;type:char;not null" json:"name"`
	ZipCode        *string  `gorm:"column:zip_code;type:char" json:"zip_code,omitempty"`
	State          *string  `gorm:"column:state;type:char" json:"state,omitempty"`
	BuildingNumber *int     `gorm:"column:building_number" json:"building_number,omitempty"`
	RoomNumber     *int     `gorm:"column:room_number" json:"room_number,omitempty"`
	Latitude       *float64 `gorm:"column:latitude;type:numeric" json:"latitude,omitempty"`
	Longitude      *float64 `gorm:"column:longitude;type:numeric" json:"longitude,omitempty"`
	CountryID      *int     `gorm:"column:country_id" json:"country_id,omitempty"`
	CityID         *int     `gorm:"column:city_id" json:"city_id,omitempty"`
	LocationMap    *string  `gorm:"column:location_map;type:text" json:"location_map,omitempty"`
	Country        *Country `gorm:"foreignKey:CountryID;references:ID" json:"country,omitempty"`
	City           *City    `gorm:"foreignKey:CityID;references:ID" json:"city,omitempty"`
}

func (Location) TableName() string {
	return `"Location"`
}

