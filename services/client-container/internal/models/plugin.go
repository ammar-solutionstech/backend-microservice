package models

import (
	"database/sql/driver"
	"encoding/json"
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
	ConfigSchema JSONB    `gorm:"column:config_schema;type:jsonb" json:"config_schema,omitempty"`
	Status      string    `gorm:"column:status;type:varchar(50);default:'active'" json:"status"` // active, deprecated
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Plugin) TableName() string {
	return "plugins"
}

// JSONB is a custom type for PostgreSQL JSONB
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

