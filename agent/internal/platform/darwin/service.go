// +build darwin

package darwin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// InstallService installs the agent as a LaunchAgent
func InstallService(serviceName, exePath string) error {
	home := os.Getenv("HOME")
	if home == "" {
		return fmt.Errorf("HOME environment variable not set")
	}

	// Create LaunchAgents directory
	launchAgentsDir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %v", err)
	}

	// Create plist file
	plistFile := filepath.Join(launchAgentsDir, fmt.Sprintf("com.agent.%s.plist", serviceName))
	
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.agent.%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>
`, serviceName, exePath)

	if err := os.WriteFile(plistFile, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %v", err)
	}

	// Load the service
	if err := exec.Command("launchctl", "load", plistFile).Run(); err != nil {
		return fmt.Errorf("failed to load service: %v", err)
	}

	return nil
}

// UninstallService uninstalls the LaunchAgent
func UninstallService(serviceName string) error {
	home := os.Getenv("HOME")
	if home == "" {
		return fmt.Errorf("HOME environment variable not set")
	}

	plistFile := filepath.Join(home, "Library", "LaunchAgents", fmt.Sprintf("com.agent.%s.plist", serviceName))

	// Unload the service
	exec.Command("launchctl", "unload", plistFile).Run()

	// Remove plist file
	if err := os.Remove(plistFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file: %v", err)
	}

	return nil
}

// GetServicePath returns the path where the service executable should be installed
func GetServicePath(serviceName string) string {
	return filepath.Join("/usr/local/bin", serviceName)
}

