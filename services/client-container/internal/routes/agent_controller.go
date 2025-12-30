package routes

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
	"gorm.io/gorm"

	"backend/services/client-container/internal/services"
)

// AgentController handles agent update endpoints
type AgentController struct {
	versionService *services.AgentVersionService
	db             *gorm.DB
	publicKey      string
}

// NewAgentController creates a new agent controller
func NewAgentController(versionService *services.AgentVersionService, db *gorm.DB, publicKey string) *AgentController {
	return &AgentController{
		versionService: versionService,
		db:             db,
		publicKey:      publicKey,
	}
}

// GetLatestVersion handles GET /api/v1/agents/updates/latest
func (c *AgentController) GetLatestVersion(w http.ResponseWriter, r *http.Request) {
	osType := r.URL.Query().Get("os")
	arch := r.URL.Query().Get("arch")

	if osType == "" || arch == "" {
		writeError(w, http.StatusBadRequest, "os and arch query parameters are required")
		return
	}

	version, err := c.versionService.GetLatestVersion(osType, arch)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "no version found for specified platform")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"version":     version.Version,
		"os_type":     version.OSType,
		"architecture": version.Architecture,
		"manifest":    version.Manifest,
		"checksum":    version.Checksum,
		"release_date": version.ReleaseDate,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetManifest handles GET /api/v1/agents/updates/manifest/{version}
func (c *AgentController) GetManifest(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	osType := r.URL.Query().Get("os")
	arch := r.URL.Query().Get("arch")

	if osType == "" || arch == "" {
		writeError(w, http.StatusBadRequest, "os and arch query parameters are required")
		return
	}

	manifest, err := c.versionService.GetManifest(version, osType, arch)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "manifest not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, manifest)
}

// DownloadUpdate handles GET /api/v1/agents/updates/download/{version}
func (c *AgentController) DownloadUpdate(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	osType := r.URL.Query().Get("os")
	arch := r.URL.Query().Get("arch")

	if osType == "" || arch == "" {
		writeError(w, http.StatusBadRequest, "os and arch query parameters are required")
		return
	}

	// Get update package path
	packagePath, agentVersion, err := c.versionService.GetUpdatePackage(version, osType, arch)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "update package not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Verify file exists
	fileInfo, err := os.Stat(packagePath)
	if err != nil {
		writeError(w, http.StatusNotFound, "update package file not found")
		return
	}

	// Set headers for file download
	filename := filepath.Base(packagePath)
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	
	// Include checksum in headers
	if agentVersion.Checksum != "" {
		w.Header().Set("X-Update-Checksum", agentVersion.Checksum)
	}

	// Serve file
	http.ServeFile(w, r, packagePath)
}

// GetPublicKey handles GET /api/v1/agents/public-key
func (c *AgentController) GetPublicKey(w http.ResponseWriter, r *http.Request) {
	if c.publicKey == "" {
		writeError(w, http.StatusNotFound, "public key not configured")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"public_key": c.publicKey,
		"algorithm":  "ed25519",
	})
}

