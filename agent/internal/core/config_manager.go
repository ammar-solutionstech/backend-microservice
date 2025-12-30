package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"backend/agent/config"
	"backend/agent/internal/communication"
	"backend/agent/internal/utils"
	agentpb "backend/agent/proto"
)

// ConfigManager manages configuration synchronization
type ConfigManager struct {
	config     *config.Config
	logger     *utils.Logger
	grpcClient *communication.Client
	deviceID   int32
	configPath string
}

// NewConfigManager creates a new config manager
func NewConfigManager(
	cfg *config.Config,
	logger *utils.Logger,
	grpcClient *communication.Client,
	deviceID int32,
) *ConfigManager {
	configPath := filepath.Join(cfg.ConfigDir, "agent_config.json")

	return &ConfigManager{
		config:     cfg,
		logger:     logger,
		grpcClient: grpcClient,
		deviceID:   deviceID,
		configPath: configPath,
	}
}

// SyncConfiguration syncs configuration from backend
func (cm *ConfigManager) SyncConfiguration() error {
	cm.logger.Info("Syncing configuration from backend")

	client := cm.grpcClient.GetClient()
	if client == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	req := &agentpb.ConfigRequest{
		DeviceId: cm.deviceID,
		Keys:      []string{}, // Empty means all config
	}

	ctx := context.Background()
	resp, err := client.GetConfiguration(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get configuration: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("configuration sync failed: %s", resp.Message)
	}

	// Merge with local configuration
	if err := cm.mergeConfiguration(resp.Configuration); err != nil {
		return fmt.Errorf("failed to merge configuration: %v", err)
	}

	// Save merged configuration
	if err := cm.saveConfiguration(resp.Configuration); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}

	cm.logger.Info("Configuration synced successfully", map[string]interface{}{
		"version": resp.Version,
	})

	return nil
}

// mergeConfiguration merges remote configuration with local
func (cm *ConfigManager) mergeConfiguration(remoteConfig map[string]string) error {
	// For now, just update config values
	// In production, implement proper merging logic
	for key, value := range remoteConfig {
		switch key {
		case "update_check_interval":
			// Update config if needed
			_ = value
		case "health_report_interval":
			// Update config if needed
			_ = value
		}
	}

	return nil
}

// saveConfiguration saves configuration to file
func (cm *ConfigManager) saveConfiguration(config map[string]string) error {
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(cm.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	if err := os.WriteFile(cm.configPath, configData, 0600); err != nil {
		return fmt.Errorf("failed to write configuration: %v", err)
	}

	return nil
}

// LoadConfiguration loads configuration from file
func (cm *ConfigManager) LoadConfiguration() (map[string]string, error) {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		return make(map[string]string), nil
	}

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %v", err)
	}

	var config map[string]string
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %v", err)
	}

	return config, nil
}

