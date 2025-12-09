package models

import (
	"backend/services/client-container/internal/utils"
	"time"
)

// Telemetry represents telemetry data from devices
type Telemetry struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DeviceID    int       `gorm:"column:device_id;not null;index" json:"device_id"`
	MetricName  string    `gorm:"column:metric_name;type:varchar(255);not null;index" json:"metric_name"`
	MetricValue utils.JSONB     `gorm:"column:metric_value;type:jsonb" json:"metric_value"`
	Timestamp   time.Time `gorm:"column:timestamp;not null;index" json:"timestamp"`
	Tags        utils.JSONB     `gorm:"column:tags;type:jsonb" json:"tags,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (Telemetry) TableName() string {
	return "telemetry"
}
