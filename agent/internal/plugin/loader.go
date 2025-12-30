package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"backend/agent/config"
	"backend/agent/internal/utils"
)

// Loader handles loading plugins
type Loader struct {
	config *config.Config
	logger *utils.Logger
}

// NewLoader creates a new plugin loader
func NewLoader(cfg *config.Config, logger *utils.Logger) *Loader {
	return &Loader{
		config: cfg,
		logger: logger,
	}
}

// LoadPlugin loads a plugin from a file
func (l *Loader) LoadPlugin(pluginPath string) (Plugin, error) {
	l.logger.Info("Loading plugin", map[string]interface{}{
		"path": pluginPath,
	})

	// Check if plugin file exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", pluginPath)
	}

	// Load Go plugin (requires buildmode=plugin)
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %v", err)
	}

	// Look for exported symbol "Plugin"
	symbol, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("plugin symbol not found: %v", err)
	}

	// Type assert to Plugin interface
	pluginInstance, ok := symbol.(Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin does not implement Plugin interface")
	}

	l.logger.Info("Plugin loaded successfully", map[string]interface{}{
		"name":    pluginInstance.Name(),
		"version": pluginInstance.Version(),
	})

	return pluginInstance, nil
}

// FindPlugins finds all plugins in the plugin directory
func (l *Loader) FindPlugins() ([]string, error) {
	var pluginPaths []string

	err := filepath.Walk(l.config.PluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for .so files (Linux), .dylib (macOS), or .dll (Windows)
		ext := filepath.Ext(path)
		if ext == ".so" || ext == ".dylib" || ext == ".dll" {
			pluginPaths = append(pluginPaths, path)
		}

		return nil
	})

	return pluginPaths, err
}

