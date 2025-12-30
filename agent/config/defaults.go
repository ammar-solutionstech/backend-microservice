package config

// Default values for agent configuration
const (
	DefaultUpdateCheckInterval = 3600  // 1 hour
	DefaultUpdateRetention     = 3     // Keep 3 previous versions
	DefaultHealthReportInterval = 300  // 5 minutes
	DefaultPluginTimeout       = 30    // 30 seconds
	DefaultLogLevel            = "info"
	DefaultLogFormat           = "json"
)

