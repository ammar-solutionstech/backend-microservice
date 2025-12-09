package services

// Service defines the interface that all plugins/services must implement
type Service interface {
	// Name returns the unique name of the service
	Name() string

	// Start starts the service
	Start() error

	// Stop stops the service
	Stop() error

	// Status returns the current status of the service
	// Possible values: "active", "inactive", "starting", "stopping", "error"
	Status() string

	// Configure configures the service with the provided configuration
	// The config map structure varies by service type
	Configure(config map[string]interface{}) error
}
