//go:build windows
// +build windows

package windows

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// InstallService installs the agent as a Windows service
func InstallService(serviceName, exePath string) error {
	// Use sc.exe to create the service
	cmd := exec.Command("sc.exe", "create", serviceName,
		fmt.Sprintf("binPath= \"%s\"", exePath),
		"start= auto",
		"DisplayName= ITaaS Agent Service",
		"Description= ITaaS Agent Program Service",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install service: %v, output: %s", err, string(output))
	}

	return nil
}

// UninstallService uninstalls the Windows service
func UninstallService(serviceName string) error {
	// Stop service first
	exec.Command("sc.exe", "stop", serviceName).Run()

	// Delete service
	cmd := exec.Command("sc.exe", "delete", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall service: %v, output: %s", err, string(output))
	}

	return nil
}

// GetServicePath returns the path where the service executable should be installed
func GetServicePath(serviceName string) string {
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = "C:\\Program Files"
	}
	return filepath.Join(programFiles, "Agent", serviceName+".exe")
}

