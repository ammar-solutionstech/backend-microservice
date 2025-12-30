package core

import (
	"context"
	"fmt"

	"backend/agent/config"
	"backend/agent/internal/communication"
	"backend/agent/internal/plugin"
	"backend/agent/internal/security"
	"backend/agent/internal/utils"
)

// Service is the main agent core service
type Service struct {
	config          *config.Config
	logger          *utils.Logger
	grpcClient      *communication.Client
	credentialStore *security.CredentialStore
	deviceRegistry  *DeviceRegistry
	pluginRegistry  *plugin.Registry
	pluginLifecycle *plugin.LifecycleManager
	healthMonitor   *HealthMonitor
	configManager   *ConfigManager
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewService creates a new agent core service
func NewService(cfg *config.Config, logger *utils.Logger) (*Service, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create credential store
	credentialStore := security.NewCredentialStore(cfg.DataDir)

	// Create gRPC client
	grpcClient, err := communication.NewClient(cfg, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create gRPC client: %v", err)
	}

	// Create device registry
	deviceRegistry, err := NewDeviceRegistry(cfg, logger, grpcClient, credentialStore)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create device registry: %v", err)
	}

	// Create plugin registry
	pluginRegistry := plugin.NewRegistry(cfg, logger)
	pluginLifecycle := plugin.NewLifecycleManager(pluginRegistry, logger)

	// Get device ID (will be set after registration)
	deviceID := deviceRegistry.GetDeviceID()

	// Create health monitor
	healthMonitor := NewHealthMonitor(cfg, logger, grpcClient, pluginRegistry, deviceID)

	// Create config manager
	configManager := NewConfigManager(cfg, logger, grpcClient, deviceID)

	return &Service{
		config:          cfg,
		logger:          logger,
		grpcClient:      grpcClient,
		credentialStore: credentialStore,
		deviceRegistry:  deviceRegistry,
		pluginRegistry:  pluginRegistry,
		pluginLifecycle: pluginLifecycle,
		healthMonitor:   healthMonitor,
		configManager:   configManager,
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Start starts the agent core service
func (s *Service) Start() error {
	s.logger.Info("Starting agent core service")

	// Check if device is registered
	deviceIDBytes, err := s.credentialStore.Retrieve("device_id")
	if err != nil {
		s.logger.Info("Device not registered, registration required")
		// Device needs to be registered first
		return fmt.Errorf("device not registered")
	}

	// Load device ID
	deviceID := string(deviceIDBytes)
	s.logger.Info("Device registered", map[string]interface{}{
		"device_id": deviceID,
	})

	// Sync configuration
	if err := s.configManager.SyncConfiguration(); err != nil {
		s.logger.Error("Failed to sync configuration", err)
	}

	// Start plugin lifecycle manager
	s.pluginLifecycle.Start()

	// Start health monitor
	s.healthMonitor.Start()

	s.logger.Info("Agent core service started")
	return nil
}

// Stop stops the agent core service
func (s *Service) Stop() error {
	s.logger.Info("Stopping agent core service")

	s.cancel()

	// Stop health monitor
	s.healthMonitor.Stop()

	// Stop plugin lifecycle
	s.pluginLifecycle.Stop()

	// Close gRPC client
	if err := s.grpcClient.Close(); err != nil {
		s.logger.Error("Failed to close gRPC client", err)
	}

	s.logger.Info("Agent core service stopped")
	return nil
}

// RegisterDevice registers the device with the backend
func (s *Service) RegisterDevice(containerID string) error {
	return s.deviceRegistry.Register(containerID)
}

// VerifyDevice verifies the device with verification code
func (s *Service) VerifyDevice(verificationCode string) error {
	return s.deviceRegistry.Verify(verificationCode)
}

// GetDeviceRegistry returns the device registry
func (s *Service) GetDeviceRegistry() *DeviceRegistry {
	return s.deviceRegistry
}

// GetPluginRegistry returns the plugin registry
func (s *Service) GetPluginRegistry() *plugin.Registry {
	return s.pluginRegistry
}
