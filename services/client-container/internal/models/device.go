package models

import "time"

// Device represents a registered device
type Device struct {
	ID              int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ContainerID     int       `gorm:"column:container_id;not null;index" json:"container_id"`
	DeviceID        string    `gorm:"column:device_id;type:varchar(255);not null;index" json:"device_id"` // Unique per container
	Hostname        string    `gorm:"column:hostname;type:varchar(255)" json:"hostname"`
	OSType          string    `gorm:"column:os_type;type:varchar(50)" json:"os_type"` // windows, linux, darwin
	OSVersion       string    `gorm:"column:os_version;type:varchar(255)" json:"os_version"`
	Architecture    string    `gorm:"column:architecture;type:varchar(50)" json:"architecture"`
	IPAddress       string    `gorm:"column:ip_address;type:varchar(45)" json:"ip_address"`
	DeviceSerial    string    `gorm:"column:device_serial;type:varchar(255);not null" json:"device_serial"` // Required for certificate validation
	LastSeen        *time.Time `gorm:"column:last_seen" json:"last_seen,omitempty"`
	Status          string    `gorm:"column:status;type:varchar(50);default:'active'" json:"status"` // active, inactive, revoked
	CertificateSerial *string `gorm:"column:certificate_serial;type:varchar(255)" json:"certificate_serial,omitempty"`
	RegisteredAt    time.Time `gorm:"column:registered_at;default:CURRENT_TIMESTAMP" json:"registered_at"`
	RegisteredBy    *int      `gorm:"column:registered_by" json:"registered_by,omitempty"` // User ID
	CreatedAt           time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Device) TableName() string {
	return "devices"
}

