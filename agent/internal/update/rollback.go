package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"backend/agent/config"
	"backend/agent/internal/utils"
)

// RollbackManager handles rollback operations
type RollbackManager struct {
	config *config.Config
	logger *utils.Logger
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager(cfg *config.Config, logger *utils.Logger) (*RollbackManager, error) {
	return &RollbackManager{
		config: cfg,
		logger: logger,
	}, nil
}

// Rollback rolls back to the previous version
func (r *RollbackManager) Rollback() error {
	r.logger.Info("Initiating rollback")

	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %v", err)
	}

	backupDir := filepath.Join(r.config.UpdateDir, "backups")
	
	// List available backups
	backups, err := r.listBackups(backupDir)
	if err != nil {
		return fmt.Errorf("failed to list backups: %v", err)
	}

	if len(backups) == 0 {
		return fmt.Errorf("no backups available for rollback")
	}

	// Get the most recent backup
	latestBackup := backups[len(backups)-1]
	backupExe := filepath.Join(latestBackup.path, filepath.Base(currentExe))

	r.logger.Info("Rolling back to version", map[string]interface{}{
		"version":    latestBackup.version,
		"backup_path": backupExe,
	})

	// Verify backup exists
	if _, err := os.Stat(backupExe); os.IsNotExist(err) {
		return fmt.Errorf("backup executable not found: %s", backupExe)
	}

	// Create staging path
	stagingExe := currentExe + ".rollback"
	
	// Copy backup to staging
	if err := copyFile(backupExe, stagingExe); err != nil {
		return fmt.Errorf("failed to copy backup: %v", err)
	}

	// Make executable
	if err := os.Chmod(stagingExe, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %v", err)
	}

	// Atomic swap
	if err := atomicSwap(stagingExe, currentExe); err != nil {
		return fmt.Errorf("failed to swap executable: %v", err)
	}

	r.logger.Info("Rollback completed successfully", map[string]interface{}{
		"version": latestBackup.version,
	})

	return nil
}

// CleanupOldVersions removes old backup versions, keeping only the specified number
func (r *RollbackManager) CleanupOldVersions(retention int) error {
	backupDir := filepath.Join(r.config.UpdateDir, "backups")
	
	backups, err := r.listBackups(backupDir)
	if err != nil {
		return fmt.Errorf("failed to list backups: %v", err)
	}

	if len(backups) <= retention {
		return nil // No cleanup needed
	}

	// Remove oldest backups
	toRemove := backups[:len(backups)-retention]
	for _, backup := range toRemove {
		r.logger.Info("Removing old backup", map[string]interface{}{
			"version": backup.version,
			"path":    backup.path,
		})

		if err := os.RemoveAll(backup.path); err != nil {
			r.logger.Error("Failed to remove backup", err, map[string]interface{}{
				"version": backup.version,
			})
		}
	}

	return nil
}

// backupInfo represents backup version information
type backupInfo struct {
	version string
	path    string
	time    time.Time
}

// listBackups lists all available backups sorted by time
func (r *RollbackManager) listBackups(backupDir string) ([]backupInfo, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []backupInfo{}, nil
		}
		return nil, err
	}

	var backups []backupInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		version := entry.Name()
		backupPath := filepath.Join(backupDir, version)
		
		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, backupInfo{
			version: version,
			path:    backupPath,
			time:    info.ModTime(),
		})
	}

	// Sort by time
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].time.Before(backups[j].time)
	})

	return backups, nil
}

// copyFile copies a file (shared with installer)
func copyFile(src, dst string) error {
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
}

// atomicSwap atomically swaps files (shared with installer)
func atomicSwap(newPath, oldPath string) error {
	if runtime.GOOS == "windows" {
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove old file: %v", err)
		}
	}

	return os.Rename(newPath, oldPath)
}


