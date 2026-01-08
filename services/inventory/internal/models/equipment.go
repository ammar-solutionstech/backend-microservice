package models

import "time"

// Equipment captures hardware inventory data.
// Note: Foreign keys to User, Location, Contact, HelpDesk are removed - services communicate via gRPC
type Equipment struct {
	ID                 int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name               string    `gorm:"column:name;type:char;not null" json:"name"`
	Description        string    `gorm:"column:description;type:text;not null" json:"description"`
	State              *bool     `gorm:"column:state;not null" json:"state"`
	SerialNumber       string    `gorm:"column:serial_number;type:char;not null" json:"serial_number"`
	EquipmentTypeID    int       `gorm:"column:equipment_type_id;not null" json:"equipment_type_id"`
	ModelID            int       `gorm:"column:model_id;not null" json:"model_id"`
	ProductionDate     time.Time `gorm:"column:production_date;type:date;not null" json:"production_date"`
	IPv4               string    `gorm:"column:ip_v4;type:char;not null" json:"ip_v4"`
	IPv6               string    `gorm:"column:ip_v6;type:char;not null" json:"ip_v6"`
	MACAddress         string    `gorm:"column:mac_address;type:char;not null" json:"mac_address"`
	OperatingSystemID  int       `gorm:"column:operating_system_id;not null" json:"operating_system_id"`
	LocationID         int       `gorm:"column:location_id;not null" json:"location_id"` // Reference only, no FK
	Processors        string    `gorm:"column:proccessors;type:char;not null" json:"processors"`
	CountryOfRegion    int       `gorm:"column:country_of_region;not null" json:"country_of_region"` // Reference only, no FK
	UserID             int       `gorm:"column:user_id;not null" json:"user_id"`                      // Reference only, no FK
	WarrantyStartDate  time.Time `gorm:"column:warrantly_start_date;type:date;not null" json:"warranty_start_date"`
	WarrantyEndDate    time.Time `gorm:"column:warrantly_end_date;type:date;not null" json:"warranty_end_date"`
	SupplierID         *int      `gorm:"column:supplier_id" json:"supplier_id,omitempty"` // Reference only, no FK
	Type               EquipmentType   `gorm:"foreignKey:EquipmentTypeID;references:ID" json:"type,omitempty"`
	Model              Model           `gorm:"foreignKey:ModelID;references:ID" json:"model,omitempty"`
	OperatingSystem    OperatingSystem `gorm:"foreignKey:OperatingSystemID;references:ID" json:"operating_system,omitempty"`
	Documents          []Document      `gorm:"foreignKey:EquipmentID" json:"documents,omitempty"`
	Softwares          []Software      `gorm:"many2many:Equipment_Software;foreignKey:ID;joinForeignKey:EquipmentID;References:ID;joinReferences:SoftwareID" json:"softwares,omitempty"`
}

func (Equipment) TableName() string {
	return `"Equipment"`
}

