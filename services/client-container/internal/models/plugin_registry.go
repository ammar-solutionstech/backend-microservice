package models

import (
	"backend/services/client-container/internal/utils"
	"time"
)

// PluginRegistry represents a plugin in the registry available for deployment
type PluginRegistry struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(255);not null;index" json:"name"`
	Version     string    `gorm:"column:version;type:varchar(50);not null;index" json:"version"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Manifest    utils.JSONB `gorm:"column:manifest;type:jsonb" json:"manifest,omitempty"`
	DownloadURL string    `gorm:"column:download_url;type:varchar(500)" json:"download_url"`
	Checksum    string    `gorm:"column:checksum;type:varchar(255)" json:"checksum"`
	Signature   string    `gorm:"column:signature;type:text" json:"signature,omitempty"`
	Dependencies utils.JSONB `gorm:"column:dependencies;type:jsonb" json:"dependencies,omitempty"`
	Status      string    `gorm:"column:status;type:varchar(50);default:'active';index" json:"status"` // active, deprecated
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (PluginRegistry) TableName() string {
	return "plugin_registry"
}




