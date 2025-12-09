package models

import (
	"backend/services/client-container/internal/utils"
	"time"
)

// Plugin represents a plugin that can be deployed to devices
type Plugin struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ContainerID int       `gorm:"column:container_id;not null;index" json:"container_id"`
	Name        string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Version     string    `gorm:"column:version;type:varchar(50);not null" json:"version"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	PluginType  string    `gorm:"column:plugin_type;type:varchar(50)" json:"plugin_type"`
	DownloadURL string    `gorm:"column:download_url;type:varchar(500)" json:"download_url"`
	Checksum    string    `gorm:"column:checksum;type:varchar(255)" json:"checksum"`
	ConfigSchema utils.JSONB    `gorm:"column:config_schema;type:jsonb" json:"config_schema,omitempty"`
	Status      string    `gorm:"column:status;type:varchar(50);default:'active'" json:"status"` // active, deprecated
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Plugin) TableName() string {
	return "plugins"
}
