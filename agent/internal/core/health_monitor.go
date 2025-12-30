package core

import (
	"context"
	"time"

	"backend/agent/config"
	"backend/agent/internal/communication"
	"backend/agent/internal/plugin"
	"backend/agent/internal/utils"
	agentpb "backend/agent/proto"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// HealthMonitor monitors agent and plugin health
type HealthMonitor struct {
	config          *config.Config
	logger          *utils.Logger
	grpcClient      *communication.Client
	pluginRegistry  *plugin.Registry
	deviceID        int32
	startTime       time.Time
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(
	cfg *config.Config,
	logger *utils.Logger,
	grpcClient *communication.Client,
	pluginRegistry *plugin.Registry,
	deviceID int32,
) *HealthMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &HealthMonitor{
		config:         cfg,
		logger:         logger,
		grpcClient:     grpcClient,
		pluginRegistry: pluginRegistry,
		deviceID:       deviceID,
		startTime:      time.Now(),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start starts the health monitoring
func (hm *HealthMonitor) Start() {
	hm.logger.Info("Starting health monitor", map[string]interface{}{
		"report_interval": hm.config.HealthReportInterval,
	})

	// Start periodic health reporting
	go hm.healthReportLoop()
}

// Stop stops the health monitoring
func (hm *HealthMonitor) Stop() {
	hm.logger.Info("Stopping health monitor")
	hm.cancel()
}

// healthReportLoop periodically reports health status
func (hm *HealthMonitor) healthReportLoop() {
	ticker := time.NewTicker(time.Duration(hm.config.HealthReportInterval) * time.Second)
	defer ticker.Stop()

	// Report immediately on start
	hm.reportHealth()

	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-ticker.C:
			hm.reportHealth()
		}
	}
}

// reportHealth reports current health status to backend
func (hm *HealthMonitor) reportHealth() {
	healthReport := hm.collectHealthData()

	client := hm.grpcClient.GetClient()
	if client == nil {
		hm.logger.Warn("gRPC client not connected, skipping health report")
		return
	}

	req := &agentpb.HealthReport{
		DeviceId:    hm.deviceID,
		AgentVersion: utils.GetVersion(),
		Status:      hm.getAgentStatus(),
		PluginHealth: hm.getPluginHealth(),
		Metrics:     healthReport,
		Timestamp:   time.Now().Unix(),
	}

	ctx, cancel := context.WithTimeout(hm.ctx, 10*time.Second)
	defer cancel()

	resp, err := client.ReportHealth(ctx, req)
	if err != nil {
		hm.logger.Error("Failed to report health", err)
		return
	}

	if !resp.Success {
		hm.logger.Warn("Health report failed", map[string]interface{}{
			"message": resp.Message,
		})
		return
	}

	hm.logger.Debug("Health report sent successfully", map[string]interface{}{
		"next_interval": resp.NextReportIntervalSeconds,
	})
}

// collectHealthData collects system metrics
func (hm *HealthMonitor) collectHealthData() *agentpb.SystemMetrics {
	metrics := &agentpb.SystemMetrics{}

	// CPU usage
	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		metrics.CpuUsagePercent = cpuPercent[0]
	}

	// Memory usage
	if memInfo, err := mem.VirtualMemory(); err == nil {
		metrics.MemoryUsedBytes = int64(memInfo.Used)
		metrics.MemoryTotalBytes = int64(memInfo.Total)
	}

	// Disk usage (simplified - would need disk package)
	metrics.DiskUsedBytes = 0
	metrics.DiskTotalBytes = 0

	return metrics
}

// getAgentStatus returns agent status
func (hm *HealthMonitor) getAgentStatus() *agentpb.AgentStatus {
	uptime := time.Since(hm.startTime).Seconds()

	status := &agentpb.AgentStatus{
		State:        "active",
		UptimeSeconds: int64(uptime),
	}

	if !hm.grpcClient.IsConnected() {
		status.State = "error"
		status.LastError = "gRPC client not connected"
	}

	return status
}

// getPluginHealth returns health status of all plugins
func (hm *HealthMonitor) getPluginHealth() []*agentpb.PluginHealth {
	plugins := hm.pluginRegistry.ListPlugins()
	pluginHealth := make([]*agentpb.PluginHealth, 0, len(plugins))

	for _, instance := range plugins {
		healthy := false
		var lastError string

		if instance.State == "active" {
			healthy, err := instance.Plugin.HealthCheck()
			if err != nil {
				lastError = err.Error()
			} else {
				healthy = healthy
			}
		}

		pluginHealth = append(pluginHealth, &agentpb.PluginHealth{
			Name:      instance.Name,
			Version:   instance.Version,
			State:     instance.State,
			Healthy:   healthy,
			LastError: lastError,
		})
	}

	return pluginHealth
}

