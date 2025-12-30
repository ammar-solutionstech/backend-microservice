package services

import (
	"fmt"
	"os"
	"path/filepath"

	"backend/services/client-container/internal/config"
	"backend/services/client-container/internal/models"
	"gorm.io/gorm"
)

// PluginService manages plugin registry and distribution
type PluginService struct {
	db     *gorm.DB
	config *config.Config
}

// NewPluginService creates a new plugin service
func NewPluginService(cfg *config.Config, db *gorm.DB) *PluginService {
	return &PluginService{
		db:     db,
		config: cfg,
	}
}

// ListPlugins lists all available plugins, optionally filtered by status
func (s *PluginService) ListPlugins(status string) ([]models.PluginRegistry, error) {
	var plugins []models.PluginRegistry
	
	query := s.db.Model(&models.PluginRegistry{})
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	err := query.Order("name ASC, version DESC").Find(&plugins).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %v", err)
	}

	return plugins, nil
}

// GetPluginByName gets a plugin by name and optional version (latest if version is empty)
func (s *PluginService) GetPluginByName(name, version string) (*models.PluginRegistry, error) {
	var plugin models.PluginRegistry
	
	query := s.db.Where("name = ?", name)
	
	if version != "" {
		query = query.Where("version = ?", version)
	} else {
		// Get latest version
		query = query.Order("version DESC")
	}
	
	err := query.First(&plugin).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("plugin not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get plugin: %v", err)
	}

	return &plugin, nil
}

// GetPluginManifest returns the manifest for a specific plugin
func (s *PluginService) GetPluginManifest(name, version string) (map[string]interface{}, error) {
	plugin, err := s.GetPluginByName(name, version)
	if err != nil {
		return nil, err
	}

	if plugin.Manifest == nil {
		return nil, fmt.Errorf("manifest not found for plugin %s version %s", name, version)
	}

	return map[string]interface{}(plugin.Manifest), nil
}

// GetPluginPackagePath returns the file system path to the plugin package
func (s *PluginService) GetPluginPackagePath(name, version string) (string, error) {
	plugin, err := s.GetPluginByName(name, version)
	if err != nil {
		return "", err
	}

	// If download_url is a relative path, use storage path
	if plugin.DownloadURL != "" {
		if filepath.IsAbs(plugin.DownloadURL) {
			// Absolute path
			return plugin.DownloadURL, nil
		}
		// Relative path - use storage directory
		return filepath.Join(s.config.PluginStoragePath, plugin.DownloadURL), nil
	}

	// Default path structure
	filename := fmt.Sprintf("%s-%s.so", name, version)
	return filepath.Join(s.config.PluginStoragePath, name, version, filename), nil
}

// GetPluginPackage returns the file path and metadata for a plugin package
func (s *PluginService) GetPluginPackage(name, version string) (string, *models.PluginRegistry, error) {
	plugin, err := s.GetPluginByName(name, version)
	if err != nil {
		return "", nil, err
	}

	packagePath, err := s.GetPluginPackagePath(name, version)
	if err != nil {
		return "", nil, err
	}

	// Verify file exists
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("plugin package not found: %s", packagePath)
	}

	return packagePath, plugin, nil
}

// CreatePlugin creates a new plugin registry entry
func (s *PluginService) CreatePlugin(plugin *models.PluginRegistry) error {
	if err := s.db.Create(plugin).Error; err != nil {
		return fmt.Errorf("failed to create plugin: %v", err)
	}
	return nil
}


