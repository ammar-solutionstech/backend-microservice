// +build linux

package linux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// InstallService installs the agent as a systemd service
func InstallService(serviceName, exePath string) error {
	// Create systemd unit file
	unitFile := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	
	unitContent := fmt.Sprintf(`[Unit]
Description=ITaaS Agent %s Service
After=network.target

[Service]
Type=simple
ExecStart=%s
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
`, serviceName, exePath)

	if err := os.WriteFile(unitFile, []byte(unitContent), 0644); err != nil {
		return fmt.Errorf("failed to write systemd unit file: %v", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}

	// Enable service
	if err := exec.Command("systemctl", "enable", serviceName).Run(); err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}

	return nil
}

// UninstallService uninstalls the systemd service
func UninstallService(serviceName string) error {
	// Stop and disable service
	exec.Command("systemctl", "stop", serviceName).Run()
	exec.Command("systemctl", "disable", serviceName).Run()

	// Remove unit file
	unitFile := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	if err := os.Remove(unitFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unit file: %v", err)
	}

	// Reload systemd
	exec.Command("systemctl", "daemon-reload").Run()

	return nil
}

// GetServicePath returns the path where the service executable should be installed
func GetServicePath(serviceName string) string {
	return filepath.Join("/usr/local/bin", serviceName)
}

