package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"backend/agent/config"
	"backend/agent/internal/utils"
)

// Installer handles atomic update installation
type Installer struct {
	config *config.Config
	logger *utils.Logger
}

// NewInstaller creates a new installer
func NewInstaller(cfg *config.Config, logger *utils.Logger) (*Installer, error) {
	return &Installer{
		config: cfg,
		logger: logger,
	}, nil
}

// InstallUpdate installs the update atomically
func (i *Installer) InstallUpdate(updatePath, version string) error {
	i.logger.Info("Installing update", map[string]interface{}{
		"version": version,
		"path":    updatePath,
	})

	extractDir := filepath.Join(i.config.UpdateDir, "extracted")

	// Determine current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %v", err)
	}

	// Create backup of current version
	backupDir := filepath.Join(i.config.UpdateDir, "backups", version)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Copy current executable to backup
	backupExe := filepath.Join(backupDir, filepath.Base(currentExe))
	if err := copyFile(currentExe, backupExe); err != nil {
		return fmt.Errorf("failed to backup current executable: %v", err)
	}

	i.logger.Info("Backed up current version", map[string]interface{}{
		"backup_path": backupExe,
	})

	// Determine which binaries to update based on service type
	// This is a simplified version - in reality, we need to know which service is running
	binaries := []string{"update-service", "agent-core"}

	for _, binaryName := range binaries {
		// Find binary in extracted directory
		sourceBinary := filepath.Join(extractDir, binaryName)
		if runtime.GOOS == "windows" {
			sourceBinary += ".exe"
		}

		// Check if binary exists in update
		if _, err := os.Stat(sourceBinary); os.IsNotExist(err) {
			i.logger.Debug("Binary not in update, skipping", map[string]interface{}{
				"binary": binaryName,
			})
			continue
		}

		// Determine target path
		// For now, assume binaries are in the same directory as current executable
		targetDir := filepath.Dir(currentExe)
		targetBinary := filepath.Join(targetDir, binaryName)
		if runtime.GOOS == "windows" {
			targetBinary += ".exe"
		}

		// Create staging path
		stagingBinary := targetBinary + ".new"

		// Copy new binary to staging
		if err := copyFile(sourceBinary, stagingBinary); err != nil {
			return fmt.Errorf("failed to copy new binary: %v", err)
		}

		// Make staging binary executable
		if err := os.Chmod(stagingBinary, 0755); err != nil {
			return fmt.Errorf("failed to set permissions: %v", err)
		}

		// Atomic swap: rename staging to target
		// On Windows, we may need to use MoveFileEx with MOVEFILE_DELAY_UNTIL_REBOOT
		if err := atomicSwap(stagingBinary, targetBinary); err != nil {
			return fmt.Errorf("failed to swap binary: %v", err)
		}

		i.logger.Info("Updated binary", map[string]interface{}{
			"binary":  binaryName,
			"version": version,
		})
	}

	i.logger.Info("Update installation completed", map[string]interface{}{
		"version": version,
	})

	return nil
}

// copyFile copies a file from source to destination
/* func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
} */

// atomicSwap atomically swaps the new binary with the old one
/* func atomicSwap(newPath, oldPath string) error {
	// On Unix systems, rename is atomic
	// On Windows, we need special handling
	if runtime.GOOS == "windows" {
		// Remove old file first (if it exists)
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove old file: %v", err)
		}
	}

	// Atomic rename
	return os.Rename(newPath, oldPath)
} */
