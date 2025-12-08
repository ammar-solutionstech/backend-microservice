package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// PluginDeployment represents a plugin deployment to a device
type PluginDeployment struct {
	ID            int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PluginID      int        `gorm:"column:plugin_id;not null;index" json:"plugin_id"`
	DeviceID      int        `gorm:"column:device_id;not null;index" json:"device_id"`
	Version       string     `gorm:"column:version;type:varchar(50);not null" json:"version"`
	Status        string     `gorm:"column:status;type:varchar(50);default:'pending'" json:"status"` // pending, installed, failed, updating, removed
	InstalledAt   *time.Time `gorm:"column:installed_at" json:"installed_at,omitempty"`
	LastHeartbeat *time.Time `gorm:"column:last_heartbeat" json:"last_heartbeat,omitempty"`
	Config        JSONB      `gorm:"column:config;type:jsonb" json:"config,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (PluginDeployment) TableName() string {
	return "plugin_deployments"
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

