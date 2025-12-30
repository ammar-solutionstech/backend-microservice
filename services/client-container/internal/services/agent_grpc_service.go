package services

import (
	"context"
	"fmt"
	"strconv"

	"backend/services/client-container/internal/config"
	"backend/services/client-container/internal/models"
	agentpb "backend/services/client-container/proto"

	"gorm.io/gorm"
)

// AgentGRPCService implements the AgentService gRPC interface
type AgentGRPCService struct {
	agentpb.UnimplementedAgentServiceServer
	deviceService *DeviceService
	config        *config.Config
	db            *gorm.DB
}

// NewAgentGRPCService creates a new agent gRPC service
func NewAgentGRPCService(cfg *config.Config, db *gorm.DB, deviceService *DeviceService) *AgentGRPCService {
	return &AgentGRPCService{
		deviceService: deviceService,
		config:        cfg,
		db:            db,
	}
}

// RegisterDevice handles device registration
func (s *AgentGRPCService) RegisterDevice(ctx context.Context, req *agentpb.RegisterDeviceRequest) (*agentpb.RegisterDeviceResponse, error) {
	// Extract container ID from context (set by mTLS middleware)
	containerIDStr := ctx.Value("container_id").(string)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		return &agentpb.RegisterDeviceResponse{
			Success: false,
			Message: fmt.Sprintf("invalid container ID: %v", err),
		}, nil
	}

	// Prepare device info map
	deviceInfo := map[string]interface{}{
		"device_id":      req.DeviceId,
		"hostname":       req.Hostname,
		"os_type":        req.OsType,
		"os_version":     req.OsVersion,
		"architecture":   req.Architecture,
		"ip_address":     req.IpAddress,
		"device_serial":  req.DeviceSerial,
		"agent_version": req.AgentVersion,
	}

	// Add hardware info
	for k, v := range req.HardwareInfo {
		deviceInfo[k] = v
	}

	// Register device
	device, code, err := s.deviceService.RegisterDevice(containerID, deviceInfo, req.CsrPem)
	if err != nil {
		return &agentpb.RegisterDeviceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &agentpb.RegisterDeviceResponse{
		Success:          true,
		DeviceId:         int32(device.ID),
		VerificationCode: code,
		Message:          "Device registered successfully",
	}, nil
}

// VerifyDevice handles device verification
func (s *AgentGRPCService) VerifyDevice(ctx context.Context, req *agentpb.VerifyDeviceRequest) (*agentpb.VerifyDeviceResponse, error) {
	// Verify and issue certificate
	agentCert, err := s.deviceService.VerifyAndIssueCertificateWithCSR(int(req.DeviceId), req.VerificationCode, req.CsrPem)
	if err != nil {
		return &agentpb.VerifyDeviceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Get device to retrieve container ID
	var device models.Device
	if err := s.db.First(&device, req.DeviceId).Error; err != nil {
		return &agentpb.VerifyDeviceResponse{
			Success: false,
			Message: fmt.Sprintf("device not found: %v", err),
		}, nil
	}

	var container models.ClientContainer
	if err := s.db.First(&container, device.ContainerID).Error; err != nil {
		return &agentpb.VerifyDeviceResponse{
			Success: false,
			Message: fmt.Sprintf("container not found: %v", err),
		}, nil
	}

	return &agentpb.VerifyDeviceResponse{
		Success:          true,
		CertificatePem:   agentCert.CertificatePEM,
		CertificateSerial: agentCert.CertificateSerial,
		ContainerId:      container.ContainerID,
		Message:          "Device verified and certificate issued",
	}, nil
}

// ReportHealth handles health reporting
func (s *AgentGRPCService) ReportHealth(ctx context.Context, req *agentpb.HealthReport) (*agentpb.HealthResponse, error) {
	// Update device last seen
	if err := s.deviceService.UpdateDeviceLastSeen(int(req.DeviceId)); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: Failed to update device last seen: %v\n", err)
	}

	// TODO: Store health metrics in telemetry table

	return &agentpb.HealthResponse{
		Success:                 true,
		Message:                 "Health report received",
		NextReportIntervalSeconds: 300, // 5 minutes default
	}, nil
}

// GetConfiguration handles configuration retrieval
func (s *AgentGRPCService) GetConfiguration(ctx context.Context, req *agentpb.ConfigRequest) (*agentpb.ConfigResponse, error) {
	// TODO: Implement configuration retrieval from database
	config := make(map[string]string)
	config["update_check_interval"] = "3600"
	config["health_report_interval"] = "300"

	return &agentpb.ConfigResponse{
		Success:      true,
		Configuration: config,
		Version:      1,
		Message:      "Configuration retrieved",
	}, nil
}

// UpdateConfiguration handles configuration updates
func (s *AgentGRPCService) UpdateConfiguration(ctx context.Context, req *agentpb.UpdateConfigRequest) (*agentpb.ConfigResponse, error) {
	// TODO: Implement configuration storage
	return &agentpb.ConfigResponse{
		Success:      true,
		Configuration: req.Configuration,
		Version:      1,
		Message:      "Configuration updated",
	}, nil
}

// ListPlugins lists available plugins
func (s *AgentGRPCService) ListPlugins(ctx context.Context, req *agentpb.PluginListRequest) (*agentpb.PluginListResponse, error) {
	// TODO: Implement plugin listing from database
	return &agentpb.PluginListResponse{
		Success: true,
		Plugins: []*agentpb.PluginInfo{},
		Message: "Plugins listed",
	}, nil
}

// InstallPlugin handles plugin installation
func (s *AgentGRPCService) InstallPlugin(ctx context.Context, req *agentpb.InstallPluginRequest) (*agentpb.InstallPluginResponse, error) {
	// TODO: Implement plugin installation
	return &agentpb.InstallPluginResponse{
		Success:     true,
		PluginName:  req.PluginName,
		Version:     req.Version,
		Message:     "Plugin installation initiated",
	}, nil
}

// UninstallPlugin handles plugin uninstallation
func (s *AgentGRPCService) UninstallPlugin(ctx context.Context, req *agentpb.UninstallPluginRequest) (*agentpb.UninstallPluginResponse, error) {
	// TODO: Implement plugin uninstallation
	return &agentpb.UninstallPluginResponse{
		Success: true,
		Message: "Plugin uninstallation initiated",
	}, nil
}

// ActivatePlugin handles plugin activation
func (s *AgentGRPCService) ActivatePlugin(ctx context.Context, req *agentpb.ActivatePluginRequest) (*agentpb.ActivatePluginResponse, error) {
	// TODO: Implement plugin activation
	return &agentpb.ActivatePluginResponse{
		Success: true,
		Message: "Plugin activation initiated",
	}, nil
}

// DeactivatePlugin handles plugin deactivation
func (s *AgentGRPCService) DeactivatePlugin(ctx context.Context, req *agentpb.DeactivatePluginRequest) (*agentpb.DeactivatePluginResponse, error) {
	// TODO: Implement plugin deactivation
	return &agentpb.DeactivatePluginResponse{
		Success: true,
		Message: "Plugin deactivation initiated",
	}, nil
}

// GetPluginStatus gets plugin status
func (s *AgentGRPCService) GetPluginStatus(ctx context.Context, req *agentpb.PluginStatusRequest) (*agentpb.PluginStatusResponse, error) {
	// TODO: Implement plugin status retrieval
	return &agentpb.PluginStatusResponse{
		Success: true,
		Plugin:  &agentpb.PluginInfo{},
		Message: "Plugin status retrieved",
	}, nil
}

