package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Container represents a registered container
type Container struct {
	ID                int            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ContainerID       string         `gorm:"column:container_id;type:varchar(255);uniqueIndex;not null" json:"container_id"`
	Name              string         `gorm:"column:name;type:varchar(255)" json:"name"`
	RegisteredAt      time.Time      `gorm:"column:registered_at;default:CURRENT_TIMESTAMP" json:"registered_at"`
	CertificateSerial *string        `gorm:"column:certificate_serial;type:varchar(255)" json:"certificate_serial,omitempty"`
	Status            string         `gorm:"column:status;type:varchar(50);default:'active'" json:"status"` // active, suspended, revoked
	Metadata          JSONB          `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt         time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Container) TableName() string {
	return "containers"
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

