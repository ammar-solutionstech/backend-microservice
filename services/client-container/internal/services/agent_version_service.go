package services

import (
	"fmt"
	"os"
	"path/filepath"

	"backend/services/client-container/internal/config"
	"backend/services/client-container/internal/models"
	"gorm.io/gorm"
)

// AgentVersionService manages agent versions and update distribution
type AgentVersionService struct {
	db     *gorm.DB
	config *config.Config
}

// NewAgentVersionService creates a new agent version service
func NewAgentVersionService(cfg *config.Config, db *gorm.DB) *AgentVersionService {
	return &AgentVersionService{
		db:     db,
		config: cfg,
	}
}

// GetLatestVersion returns the latest agent version for the specified OS and architecture
func (s *AgentVersionService) GetLatestVersion(osType, arch string) (*models.AgentVersion, error) {
	var version models.AgentVersion
	
	err := s.db.Where("os_type = ? AND architecture = ?", osType, arch).
		Order("release_date DESC NULLS LAST, created_at DESC").
		First(&version).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no agent version found for %s/%s", osType, arch)
		}
		return nil, fmt.Errorf("failed to get latest version: %v", err)
	}

	return &version, nil
}

// GetVersionByVersion returns a specific agent version
func (s *AgentVersionService) GetVersionByVersion(version, osType, arch string) (*models.AgentVersion, error) {
	var agentVersion models.AgentVersion
	
	err := s.db.Where("version = ? AND os_type = ? AND architecture = ?", version, osType, arch).
		First(&agentVersion).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("agent version %s for %s/%s not found", version, osType, arch)
		}
		return nil, fmt.Errorf("failed to get version: %v", err)
	}

	return &agentVersion, nil
}

// GetManifest returns the manifest for a specific version
func (s *AgentVersionService) GetManifest(version, osType, arch string) (map[string]interface{}, error) {
	agentVersion, err := s.GetVersionByVersion(version, osType, arch)
	if err != nil {
		return nil, err
	}

	if agentVersion.Manifest == nil {
		return nil, fmt.Errorf("manifest not found for version %s", version)
	}

	return map[string]interface{}(agentVersion.Manifest), nil
}

// GetUpdatePackagePath returns the file system path to the update package
func (s *AgentVersionService) GetUpdatePackagePath(version, osType, arch string) (string, error) {
	agentVersion, err := s.GetVersionByVersion(version, osType, arch)
	if err != nil {
		return "", err
	}

	// If download_url is a relative path, use storage path
	if agentVersion.DownloadURL != "" {
		if filepath.IsAbs(agentVersion.DownloadURL) {
			// Absolute path
			return agentVersion.DownloadURL, nil
		}
		// Relative path - use storage directory
		return filepath.Join(s.config.UpdateStoragePath, agentVersion.DownloadURL), nil
	}

	// Default path structure
	filename := fmt.Sprintf("update-%s-%s-%s.tar.gz", version, osType, arch)
	return filepath.Join(s.config.UpdateStoragePath, version, filename), nil
}

// GetUpdatePackage returns the file path and metadata for an update package
func (s *AgentVersionService) GetUpdatePackage(version, osType, arch string) (string, *models.AgentVersion, error) {
	agentVersion, err := s.GetVersionByVersion(version, osType, arch)
	if err != nil {
		return "", nil, err
	}

	packagePath, err := s.GetUpdatePackagePath(version, osType, arch)
	if err != nil {
		return "", nil, err
	}

	// Verify file exists
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("update package not found: %s", packagePath)
	}

	return packagePath, agentVersion, nil
}

// CreateVersion creates a new agent version record
func (s *AgentVersionService) CreateVersion(version *models.AgentVersion) error {
	if err := s.db.Create(version).Error; err != nil {
		return fmt.Errorf("failed to create agent version: %v", err)
	}
	return nil
}

// ListVersions lists all agent versions, optionally filtered by OS and architecture
func (s *AgentVersionService) ListVersions(osType, arch string) ([]models.AgentVersion, error) {
	var versions []models.AgentVersion
	
	query := s.db.Model(&models.AgentVersion{})
	
	if osType != "" {
		query = query.Where("os_type = ?", osType)
	}
	if arch != "" {
		query = query.Where("architecture = ?", arch)
	}
	
	err := query.Order("release_date DESC NULLS LAST, created_at DESC").Find(&versions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %v", err)
	}

	return versions, nil
}

