package utils

// Version information
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// GetVersion returns the agent version
func GetVersion() string {
	return Version
}

// GetBuildInfo returns build information
func GetBuildInfo() map[string]string {
	return map[string]string{
		"version":   Version,
		"build_time": BuildTime,
		"git_commit": GitCommit,
	}
}

