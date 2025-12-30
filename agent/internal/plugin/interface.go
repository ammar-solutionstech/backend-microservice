package plugin

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Initialize initializes the plugin with configuration
	Initialize(config map[string]interface{}) error

	// Start starts the plugin
	Start() error

	// Stop stops the plugin
	Stop() error

	// HealthCheck performs a health check and returns status
	HealthCheck() (bool, error)

	// GetStatus returns the current status of the plugin
	GetStatus() map[string]interface{}
}

