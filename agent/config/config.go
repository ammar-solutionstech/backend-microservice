package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

// Config holds agent configuration
type Config struct {
	// Agent identity
	DeviceID     string
	ContainerID  string
	AgentVersion string

	// Backend connection
	BackendURL      string
	BackendGRPCURL  string
	BackendCACert   string
	ClientCertPath  string
	ClientKeyPath   string

	// Data directories (platform-specific)
	DataDir        string
	ConfigDir      string
	LogDir         string
	PluginDir      string
	UpdateDir      string
	CertificateDir string

	// Update settings
	UpdateCheckInterval int // seconds
	UpdateRetention     int // number of versions to keep

	// Health reporting
	HealthReportInterval int // seconds

	// Plugin settings
	PluginTimeout int // seconds for plugin operations

	// Logging
	LogLevel  string
	LogFile   string
	LogFormat string
}

// Load loads configuration from environment and files
func Load() (*Config, error) {
	// Try to load .env file
	_ = godotenv.Load()

	cfg := &Config{
		DeviceID:            getEnv("AGENT_DEVICE_ID", ""),
		ContainerID:         getEnv("AGENT_CONTAINER_ID", ""),
		AgentVersion:        getEnv("AGENT_VERSION", "1.0.0"),
		BackendURL:          getEnv("AGENT_BACKEND_URL", "https://localhost:8006"),
		BackendGRPCURL:      getEnv("AGENT_BACKEND_GRPC_URL", "localhost:9006"),
		BackendCACert:       getEnv("AGENT_BACKEND_CA_CERT", ""),
		ClientCertPath:      getEnv("AGENT_CLIENT_CERT", ""),
		ClientKeyPath:       getEnv("AGENT_CLIENT_KEY", ""),
		UpdateCheckInterval: getIntEnv("AGENT_UPDATE_CHECK_INTERVAL", 3600),
		UpdateRetention:     getIntEnv("AGENT_UPDATE_RETENTION", 3),
		HealthReportInterval: getIntEnv("AGENT_HEALTH_REPORT_INTERVAL", 300),
		PluginTimeout:        getIntEnv("AGENT_PLUGIN_TIMEOUT", 30),
		LogLevel:             getEnv("AGENT_LOG_LEVEL", "info"),
		LogFile:              getEnv("AGENT_LOG_FILE", ""),
		LogFormat:            getEnv("AGENT_LOG_FORMAT", "json"),
	}

	// Set platform-specific directories
	cfg.setPlatformPaths()

	return cfg, nil
}

// setPlatformPaths sets platform-specific directory paths
func (c *Config) setPlatformPaths() {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		c.DataDir = filepath.Join(appData, "agent")
		c.ConfigDir = filepath.Join(c.DataDir, "config")
		c.LogDir = filepath.Join(c.DataDir, "logs")
		c.PluginDir = filepath.Join(c.DataDir, "plugins")
		c.UpdateDir = filepath.Join(c.DataDir, "updates")
		c.CertificateDir = filepath.Join(c.DataDir, "certs")

	case "darwin": // macOS
		home := os.Getenv("HOME")
		c.DataDir = filepath.Join(home, "Library", "Application Support", "agent")
		c.ConfigDir = filepath.Join(c.DataDir, "config")
		c.LogDir = filepath.Join(c.DataDir, "logs")
		c.PluginDir = filepath.Join(c.DataDir, "plugins")
		c.UpdateDir = filepath.Join(c.DataDir, "updates")
		c.CertificateDir = filepath.Join(c.DataDir, "certs")

	case "linux":
		home := os.Getenv("HOME")
		if home == "" {
			home = "/root"
		}
		c.DataDir = filepath.Join(home, ".config", "agent")
		c.ConfigDir = filepath.Join(c.DataDir, "config")
		c.LogDir = filepath.Join(c.DataDir, "logs")
		c.PluginDir = filepath.Join(c.DataDir, "plugins")
		c.UpdateDir = filepath.Join(c.DataDir, "updates")
		c.CertificateDir = filepath.Join(c.DataDir, "certs")

	default:
		// Fallback to current directory
		c.DataDir = "./agent-data"
		c.ConfigDir = filepath.Join(c.DataDir, "config")
		c.LogDir = filepath.Join(c.DataDir, "logs")
		c.PluginDir = filepath.Join(c.DataDir, "plugins")
		c.UpdateDir = filepath.Join(c.DataDir, "updates")
		c.CertificateDir = filepath.Join(c.DataDir, "certs")
	}

	// Ensure directories exist
	dirs := []string{c.DataDir, c.ConfigDir, c.LogDir, c.PluginDir, c.UpdateDir, c.CertificateDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create directory %s: %v\n", dir, err)
		}
	}
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.BackendURL == "" {
		return fmt.Errorf("backend URL is required")
	}
	if c.BackendGRPCURL == "" {
		return fmt.Errorf("backend gRPC URL is required")
	}
	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var intValue int
	if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
		return defaultValue
	}
	return intValue
}

