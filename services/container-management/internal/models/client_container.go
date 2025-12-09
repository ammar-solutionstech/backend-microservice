package models

import (
	"backend/services/container-management/internal/utils"
	"time"
)

// ClientContainer represents a client container instance for an organization
type ClientContainer struct {
	// ID is the primary key, auto-incrementing integer
	ID int `gorm:"column:id;primaryKey;autoIncrement" json:"id"`

	// OrganizationID is the foreign key to organizations table, required, indexed
	OrganizationID int `gorm:"column:organization_id;not null;index" json:"organization_id"`

	// ContainerID is the unique container identifier (e.g., "container-123"), required, unique index, max 255 characters
	ContainerID string `gorm:"column:container_id;type:varchar(255);uniqueIndex;not null" json:"container_id"`

	// Name is the human-readable container name, optional, max 255 characters
	Name string `gorm:"column:name;type:varchar(255)" json:"name"`

	// Status is the container status. Options: "active", "inactive". Default: "active"
	Status string `gorm:"column:status;type:varchar(50);default:'active'" json:"status"`

	// CertificateSerial is the serial number of container's certificate, optional, max 255 characters
	CertificateSerial *string `gorm:"column:certificate_serial;type:varchar(255)" json:"certificate_serial,omitempty"`

	// ContainerEndpointURL is the full URL endpoint for the container (e.g., "https://container.example.com:8006"), optional, max 500 characters
	ContainerEndpointURL string `gorm:"column:container_endpoint_url;type:varchar(500)" json:"container_endpoint_url"`

	// AdminEmail is the container administrator email address, required, max 255 characters
	AdminEmail string `gorm:"column:admin_email;type:varchar(255);not null" json:"admin_email"`

	// AdminPhone is the container administrator phone number, optional, max 255 characters
	AdminPhone *string `gorm:"column:admin_phone;type:varchar(255)" json:"admin_phone,omitempty"`

	// DockerContainerID is the Docker container ID from Docker daemon (64-character hex string), optional, max 255 characters
	DockerContainerID *string `gorm:"column:docker_container_id;type:varchar(255)" json:"docker_container_id,omitempty"`

	// DockerContainerName is the Docker container name (format: "client-container-{container_id}"), optional, max 255 characters
	DockerContainerName *string `gorm:"column:docker_container_name;type:varchar(255)" json:"docker_container_name,omitempty"`

	// DockerStatus is the Docker container status. Options: "running", "stopped", "restarting", "paused", "exited", "dead". Optional, max 50 characters
	DockerStatus *string `gorm:"column:docker_status;type:varchar(50)" json:"docker_status,omitempty"`

	// Services is the active services/plugins configuration (map structure: {"service_name": {"type": "internal|external", "config": {...}, "enabled": true/false}}), optional
	Services utils.JSONB `gorm:"column:services;type:jsonb" json:"services,omitempty"`

	// ContainerConfig is the container configuration snapshot including environment variables, volumes, network settings, optional
	ContainerConfig utils.JSONB `gorm:"column:container_config;type:jsonb" json:"container_config,omitempty"`

	// CreatedAt is the timestamp when container was created
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`

	// UpdatedAt is the timestamp when container was last updated
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (ClientContainer) TableName() string {
	return "client_containers"
}
