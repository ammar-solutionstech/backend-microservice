package models

import "time"

// ClientContainer represents a client container instance for an organization
type ClientContainer struct {
	ID                  int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrganizationID      int       `gorm:"column:organization_id;not null;index" json:"organization_id"`
	ContainerID         string    `gorm:"column:container_id;type:varchar(255);uniqueIndex;not null" json:"container_id"`
	Name                string    `gorm:"column:name;type:varchar(255)" json:"name"`
	Status              string    `gorm:"column:status;type:varchar(50);default:'active'" json:"status"` // active, inactive
	CertificateSerial   *string   `gorm:"column:certificate_serial;type:varchar(255)" json:"certificate_serial,omitempty"`
	ContainerEndpointURL string   `gorm:"column:container_endpoint_url;type:varchar(500)" json:"container_endpoint_url"`
	AdminEmail          string    `gorm:"column:admin_email;type:varchar(255);not null" json:"admin_email"`
	AdminPhone          *string   `gorm:"column:admin_phone;type:varchar(255)" json:"admin_phone,omitempty"`
	CreatedAt           time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (ClientContainer) TableName() string {
	return "client_containers"
}

