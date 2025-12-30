package update

import (
	"context"
	"fmt"
	"time"

	"backend/agent/config"
	"backend/agent/internal/utils"
)

// Service manages agent updates
type Service struct {
	config     *config.Config
	logger     *utils.Logger
	downloader *Downloader
	verifier   *Verifier
	installer  *Installer
	rollback   *RollbackManager
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewService creates a new update service
func NewService(cfg *config.Config, logger *utils.Logger) (*Service, error) {
	downloader, err := NewDownloader(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create downloader: %v", err)
	}

	// Verifier will be created - public key will be loaded on first verification
	verifier, err := NewVerifier(cfg, logger)
	if err != nil {
		// Non-fatal error - verifier can still be created, key loaded later
		logger.Warn("Verifier created without initial public key", map[string]interface{}{
			"error": err.Error(),
		})
	}

	installer, err := NewInstaller(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create installer: %v", err)
	}

	rollback, err := NewRollbackManager(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create rollback manager: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		config:     cfg,
		logger:     logger,
		downloader: downloader,
		verifier:   verifier,
		installer:  installer,
		rollback:   rollback,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start starts the update service
func (s *Service) Start() error {
	s.logger.Info("Starting update service", map[string]interface{}{
		"check_interval": s.config.UpdateCheckInterval,
	})

	// Start periodic update checking
	go s.updateLoop()

	return nil
}

// Stop stops the update service
func (s *Service) Stop() error {
	s.logger.Info("Stopping update service")
	s.cancel()
	return nil
}

// updateLoop periodically checks for updates
func (s *Service) updateLoop() {
	ticker := time.NewTicker(time.Duration(s.config.UpdateCheckInterval) * time.Second)
	defer ticker.Stop()

	// Check immediately on start
	s.checkForUpdates()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkForUpdates()
		}
	}
}

// checkForUpdates checks if updates are available and installs them
func (s *Service) checkForUpdates() {
	s.logger.Info("Checking for updates")

	currentVersion := utils.GetVersion()
	s.logger.Debug("Current agent version", map[string]interface{}{
		"version": currentVersion,
	})

	// Check for latest version from backend
	latestVersion, manifest, err := s.downloader.CheckForUpdates(currentVersion)
	if err != nil {
		s.logger.Error("Failed to check for updates", err)
		return
	}

	if latestVersion == "" || latestVersion == currentVersion {
		s.logger.Debug("No updates available", map[string]interface{}{
			"current_version": currentVersion,
			"latest_version":  latestVersion,
		})
		return
	}

	s.logger.Info("Update available", map[string]interface{}{
		"current_version": currentVersion,
		"latest_version":  latestVersion,
	})

	// Download update
	updatePath, err := s.downloader.DownloadUpdate(latestVersion)
	if err != nil {
		s.logger.Error("Failed to download update", err)
		return
	}

	// Verify update
	if err := s.verifier.VerifyUpdate(updatePath, manifest); err != nil {
		s.logger.Error("Update verification failed", err)
		return
	}

	s.logger.Info("Update verified successfully")

	// Install update
	if err := s.installer.InstallUpdate(updatePath, latestVersion); err != nil {
		s.logger.Error("Update installation failed", err)
		// Attempt rollback
		if rollbackErr := s.rollback.Rollback(); rollbackErr != nil {
			s.logger.Error("Rollback failed", rollbackErr)
		} else {
			s.logger.Info("Rolled back to previous version")
		}
		return
	}

	s.logger.Info("Update installed successfully", map[string]interface{}{
		"version": latestVersion,
	})

	// Clean up old versions
	if err := s.rollback.CleanupOldVersions(s.config.UpdateRetention); err != nil {
		s.logger.Error("Failed to cleanup old versions", err)
	}
}

// CheckForUpdatesNow manually triggers an update check
func (s *Service) CheckForUpdatesNow() error {
	s.checkForUpdates()
	return nil
}

