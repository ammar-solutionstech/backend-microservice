package plugin

import (
	"fmt"
	"sync"

	"backend/agent/config"
	"backend/agent/internal/utils"
)

// Registry manages plugin lifecycle
type Registry struct {
	config   *config.Config
	logger   *utils.Logger
	loader   *Loader
	plugins  map[string]*PluginInstance
	mu       sync.RWMutex
}

// PluginInstance wraps a plugin with its state
type PluginInstance struct {
	Plugin  Plugin
	Name    string
	Version string
	State   string // installed, active, inactive, error
	Config  map[string]interface{}
}

// NewRegistry creates a new plugin registry
func NewRegistry(cfg *config.Config, logger *utils.Logger) *Registry {
	return &Registry{
		config:  cfg,
		logger:  logger,
		loader:  NewLoader(cfg, logger),
		plugins: make(map[string]*PluginInstance),
	}
}

// Install installs a plugin
func (r *Registry) Install(pluginPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	plugin, err := r.loader.LoadPlugin(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %v", err)
	}

	instance := &PluginInstance{
		Plugin:  plugin,
		Name:    plugin.Name(),
		Version: plugin.Version(),
		State:   "installed",
		Config:  make(map[string]interface{}),
	}

	r.plugins[plugin.Name()] = instance

	r.logger.Info("Plugin installed", map[string]interface{}{
		"name":    instance.Name,
		"version": instance.Version,
	})

	return nil
}

// Uninstall uninstalls a plugin
func (r *Registry) Uninstall(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// Stop plugin if active
	if instance.State == "active" {
		if err := instance.Plugin.Stop(); err != nil {
			r.logger.Error("Failed to stop plugin during uninstall", err)
		}
	}

	delete(r.plugins, name)

	r.logger.Info("Plugin uninstalled", map[string]interface{}{
		"name": name,
	})

	return nil
}

// Activate activates a plugin
func (r *Registry) Activate(name string, config map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if instance.State == "active" {
		return fmt.Errorf("plugin already active: %s", name)
	}

	// Initialize plugin
	if err := instance.Plugin.Initialize(config); err != nil {
		instance.State = "error"
		return fmt.Errorf("failed to initialize plugin: %v", err)
	}

	// Start plugin
	if err := instance.Plugin.Start(); err != nil {
		instance.State = "error"
		return fmt.Errorf("failed to start plugin: %v", err)
	}

	instance.State = "active"
	instance.Config = config

	r.logger.Info("Plugin activated", map[string]interface{}{
		"name":    name,
		"version": instance.Version,
	})

	return nil
}

// Deactivate deactivates a plugin
func (r *Registry) Deactivate(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if instance.State != "active" {
		return fmt.Errorf("plugin not active: %s", name)
	}

	if err := instance.Plugin.Stop(); err != nil {
		instance.State = "error"
		return fmt.Errorf("failed to stop plugin: %v", err)
	}

	instance.State = "inactive"

	r.logger.Info("Plugin deactivated", map[string]interface{}{
		"name": name,
	})

	return nil
}

// GetPlugin returns a plugin instance
func (r *Registry) GetPlugin(name string) (*PluginInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instance, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return instance, nil
}

// ListPlugins returns all installed plugins
func (r *Registry) ListPlugins() []*PluginInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]*PluginInstance, 0, len(r.plugins))
	for _, instance := range r.plugins {
		plugins = append(plugins, instance)
	}

	return plugins
}

