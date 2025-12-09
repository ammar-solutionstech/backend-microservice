package models

import (
	"backend/services/container-management/internal/utils"
	"time"
)

// ContainerService represents a service (plugin) within a client-container
type ContainerService struct {
	// ID is the primary key, auto-incrementing integer
	ID int `gorm:"column:id;primaryKey;autoIncrement" json:"id"`

	// ContainerID is the reference to client_container.container_id, required, indexed, max 255 characters
	ContainerID string `gorm:"column:container_id;type:varchar(255);not null;index" json:"container_id"`

	// ServiceName is the unique service name within container (e.g., "telemetry", "device-management"), required, max 255 characters
	ServiceName string `gorm:"column:service_name;type:varchar(255);not null" json:"service_name"`

	// ServiceType is the service type. Options: "internal" (built-in service), "external" (custom plugin). Required, max 50 characters
	ServiceType string `gorm:"column:service_type;type:varchar(50);not null" json:"service_type"`

	// ServiceConfig is the service-specific configuration (structure varies by service type), optional
	ServiceConfig utils.JSONB `gorm:"column:service_config;type:jsonb" json:"service_config,omitempty"`

	// Status is the service status. Options: "active", "inactive", "starting", "stopping", "error". Default: "inactive", max 50 characters
	Status string `gorm:"column:status;type:varchar(50);default:'inactive'" json:"status"`

	// Enabled is whether service is enabled (can be disabled without removing), default: true
	Enabled bool `gorm:"column:enabled;default:true" json:"enabled"`

	// CreatedAt is the timestamp when service was added to container
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`

	// UpdatedAt is the timestamp when service was last updated
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (ContainerService) TableName() string {
	return "container_services"
}
