package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend/agent/internal/utils"
)

// LifecycleManager manages plugin lifecycle and health monitoring
type LifecycleManager struct {
	registry *Registry
	logger   *utils.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(registry *Registry, logger *utils.Logger) *LifecycleManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &LifecycleManager{
		registry: registry,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the lifecycle manager
func (lm *LifecycleManager) Start() {
	lm.logger.Info("Starting plugin lifecycle manager")

	// Start health monitoring
	lm.wg.Add(1)
	go lm.monitorHealth()
}

// Stop stops the lifecycle manager
func (lm *LifecycleManager) Stop() {
	lm.logger.Info("Stopping plugin lifecycle manager")
	lm.cancel()
	lm.wg.Wait()
}

// monitorHealth periodically checks plugin health
func (lm *LifecycleManager) monitorHealth() {
	defer lm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-lm.ctx.Done():
			return
		case <-ticker.C:
			lm.checkPluginHealth()
		}
	}
}

// checkPluginHealth checks health of all active plugins
func (lm *LifecycleManager) checkPluginHealth() {
	plugins := lm.registry.ListPlugins()

	for _, instance := range plugins {
		if instance.State != "active" {
			continue
		}

		healthy, err := instance.Plugin.HealthCheck()
		if err != nil {
			lm.logger.Error("Plugin health check failed", err, map[string]interface{}{
				"plugin": instance.Name,
			})
			instance.State = "error"
			continue
		}

		if !healthy {
			lm.logger.Warn("Plugin health check failed", map[string]interface{}{
				"plugin": instance.Name,
			})
			instance.State = "error"
		}
	}
}

// RestartPlugin restarts a plugin
func (lm *LifecycleManager) RestartPlugin(name string) error {
	instance, err := lm.registry.GetPlugin(name)
	if err != nil {
		return err
	}

	// Deactivate
	if instance.State == "active" {
		if err := lm.registry.Deactivate(name); err != nil {
			return fmt.Errorf("failed to deactivate plugin: %v", err)
		}
	}

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Activate
	if err := lm.registry.Activate(name, instance.Config); err != nil {
		return fmt.Errorf("failed to activate plugin: %v", err)
	}

	lm.logger.Info("Plugin restarted", map[string]interface{}{
		"plugin": name,
	})

	return nil
}

