package models

// Document stores binary documents linked to equipment, suppliers, or help desk tickets.
// Note: Foreign keys to Contact, HelpDesk are removed - services communicate via gRPC
type Document struct {
	ID           int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string `gorm:"column:name;type:char;not null" json:"name"`
	DocumentSize string `gorm:"column:document_size;type:char;not null" json:"document_size"`
	Picture      []byte `gorm:"column:picture;type:bytea;not null" json:"picture"`
	DocumentType string `gorm:"column:document_type;type:char;not null" json:"document_type"`
	EquipmentID  int    `gorm:"column:equipment_id;not null" json:"equipment_id"`
	SupplierID   int    `gorm:"column:supplier_id;not null" json:"supplier_id"`   // Reference only, no FK
	HelpDeskID   int    `gorm:"column:help_desk_id;not null" json:"help_desk_id"` // Reference only, no FK
	Equipment    Equipment `gorm:"foreignKey:EquipmentID;references:ID" json:"equipment,omitempty"`
}

func (Document) TableName() string {
	return `"Documents"`
}

