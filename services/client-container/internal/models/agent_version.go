package models

import (
	"backend/services/client-container/internal/utils"
	"time"
)

// AgentVersion represents an agent version available for download
type AgentVersion struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Version     string    `gorm:"column:version;type:varchar(50);not null;index" json:"version"`
	OSType      string    `gorm:"column:os_type;type:varchar(50);not null;index" json:"os_type"`
	Architecture string   `gorm:"column:architecture;type:varchar(50);not null;index" json:"architecture"`
	Manifest    utils.JSONB `gorm:"column:manifest;type:jsonb" json:"manifest,omitempty"`
	DownloadURL string    `gorm:"column:download_url;type:varchar(500)" json:"download_url"`
	Checksum    string    `gorm:"column:checksum;type:varchar(255)" json:"checksum"`
	Signature   string    `gorm:"column:signature;type:text" json:"signature,omitempty"`
	ReleaseDate *time.Time `gorm:"column:release_date" json:"release_date,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (AgentVersion) TableName() string {
	return "agent_versions"
}




