package models

import "time"

// HelpDesk represents a help desk ticket (no foreign keys to User)
type HelpDesk struct {
	ID             int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name           string    `gorm:"column:name;type:char;not null" json:"name"`
	CreateDate     time.Time `gorm:"column:create_date;type:date;not null" json:"create_date"`
	HelpDeskTypeID int       `gorm:"column:help_desk_type_id;not null;index" json:"help_desk_type_id"`
	PortalUserID   int       `gorm:"column:portal_user_id;not null;index" json:"portal_user_id"` // Reference to Auth Service user
	Description    string    `gorm:"column:description;type:text;not null" json:"description"`
	State          string    `gorm:"column:state;type:char;not null;index" json:"state"`
	ResolveDate    time.Time `gorm:"column:resolve_date;type:date;not null" json:"resolve_date"`
	ParentID       *int      `gorm:"column:help_desk_id" json:"parent_id,omitempty"`
	ProjectID      *int      `gorm:"column:project_id" json:"project_id,omitempty"`
	CreatedAt      time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships (no foreign keys)
	Type         HelpDeskType         `gorm:"foreignKey:HelpDeskTypeID;references:ID" json:"type,omitempty"`
	Parent       *HelpDesk            `gorm:"foreignKey:ParentID;references:ID" json:"parent,omitempty"`
	Documents    []Document           `gorm:"foreignKey:HelpDeskID;references:ID" json:"documents,omitempty"`
	Comments     []TicketComment      `gorm:"foreignKey:TicketID;references:ID" json:"comments,omitempty"`
	Attachments  []TicketAttachment   `gorm:"foreignKey:TicketID;references:ID" json:"attachments,omitempty"`
	Transactions []HelpDeskTransaction `gorm:"foreignKey:HelpDeskID;references:ID" json:"transactions,omitempty"`
	Ratings      []HelpDeskRating     `gorm:"foreignKey:HelpDeskID;references:ID" json:"ratings,omitempty"`
}

func (HelpDesk) TableName() string {
	return `"Help_desk"`
}

// TicketComment represents a comment on a ticket
type TicketComment struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TicketID  int       `gorm:"column:ticket_id;not null;index" json:"ticket_id"`
	UserID    int       `gorm:"column:user_id;not null;index" json:"user_id"` // Reference to Auth Service user
	Content   string    `gorm:"column:content;type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (TicketComment) TableName() string {
	return "ticket_comments"
}

// TicketAttachment represents an attachment on a ticket
type TicketAttachment struct {
	ID         int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TicketID   int       `gorm:"column:ticket_id;not null;index" json:"ticket_id"`
	DocumentID int       `gorm:"column:document_id;not null" json:"document_id"`
	CreatedAt  time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (TicketAttachment) TableName() string {
	return "ticket_attachments"
}

// Document represents a document/attachment
type Document struct {
	ID           int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string `gorm:"column:name;type:char;not null" json:"name"`
	DocumentSize string `gorm:"column:document_size;type:char;not null" json:"document_size"`
	Picture      []byte `gorm:"column:picture;type:bytea;not null" json:"picture"`
	DocumentType string `gorm:"column:document_type;type:char;not null" json:"document_type"`
	EquipmentID  int    `gorm:"column:equipment_id;default:0" json:"equipment_id"` // Reference only
	SupplierID   int    `gorm:"column:supplier_id;default:0" json:"supplier_id"`   // Reference only
	HelpDeskID   int    `gorm:"column:help_desk_id;not null;index" json:"help_desk_id"`
}

func (Document) TableName() string {
	return `"Documents"`
}

