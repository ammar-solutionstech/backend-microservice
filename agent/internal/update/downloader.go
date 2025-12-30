package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"backend/agent/config"
	"backend/agent/internal/utils"
)

// UpdateManifest represents the update package manifest
type UpdateManifest struct {
	Version     string            `json:"version"`
	OS          string            `json:"os"`
	Arch        string            `json:"arch"`
	Checksums   map[string]string `json:"checksums"`
	Signatures  map[string]string `json:"signatures"`
	Size        int64             `json:"size"`
	ReleaseDate string            `json:"release_date"`
}

// Downloader handles downloading updates
type Downloader struct {
	config *config.Config
	logger *utils.Logger
	client *http.Client
}

// NewDownloader creates a new downloader
func NewDownloader(cfg *config.Config, logger *utils.Logger) (*Downloader, error) {
	client := &http.Client{
		Timeout: 30 * time.Minute, // Allow time for large downloads
	}

	return &Downloader{
		config: cfg,
		logger: logger,
		client: client,
	}, nil
}

// CheckForUpdates checks if updates are available
func (d *Downloader) CheckForUpdates(currentVersion string) (string, *UpdateManifest, error) {
	osType := runtime.GOOS
	arch := runtime.GOARCH

	url := fmt.Sprintf("%s/api/v1/agents/updates/latest?os=%s&arch=%s", d.config.BackendURL, osType, arch)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %v", err)
	}

	// TODO: Add authentication headers if needed

	resp, err := d.client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("update check failed with status: %d", resp.StatusCode)
	}

	var response struct {
		Version  string          `json:"version"`
		Manifest UpdateManifest  `json:"manifest"`
		Message  string          `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Version == "" || response.Version == currentVersion {
		return "", nil, nil // No update available
	}

	return response.Version, &response.Manifest, nil
}

// DownloadUpdate downloads the update package
func (d *Downloader) DownloadUpdate(version string) (string, error) {
	osType := runtime.GOOS
	arch := runtime.GOARCH

	url := fmt.Sprintf("%s/api/v1/agents/updates/download/%s?os=%s&arch=%s", d.config.BackendURL, version, osType, arch)

	d.logger.Info("Downloading update", map[string]interface{}{
		"version": version,
		"url":     url,
	})

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// TODO: Add authentication headers if needed

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download update: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create download directory
	downloadDir := filepath.Join(d.config.UpdateDir, "downloads")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create download directory: %v", err)
	}

	// Create temporary file for download
	updateFile := filepath.Join(downloadDir, fmt.Sprintf("update-%s.tar.gz", version))
	file, err := os.Create(updateFile)
	if err != nil {
		return "", fmt.Errorf("failed to create update file: %v", err)
	}
	defer file.Close()

	// Download with progress tracking
	written, err := io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(updateFile)
		return "", fmt.Errorf("failed to write update file: %v", err)
	}

	d.logger.Info("Update downloaded", map[string]interface{}{
		"version": version,
		"size":    written,
		"path":    updateFile,
	})

	return updateFile, nil
}

