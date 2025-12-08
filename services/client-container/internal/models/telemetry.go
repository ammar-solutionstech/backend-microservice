package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Telemetry represents telemetry data from devices
type Telemetry struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DeviceID    int       `gorm:"column:device_id;not null;index" json:"device_id"`
	MetricName  string    `gorm:"column:metric_name;type:varchar(255);not null;index" json:"metric_name"`
	MetricValue JSONB     `gorm:"column:metric_value;type:jsonb" json:"metric_value"`
	Timestamp   time.Time `gorm:"column:timestamp;not null;index" json:"timestamp"`
	Tags        JSONB     `gorm:"column:tags;type:jsonb" json:"tags,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (Telemetry) TableName() string {
	return "telemetry"
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

