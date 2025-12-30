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

// PluginController handles plugin distribution endpoints
type PluginController struct {
	pluginService *services.PluginService
}

// NewPluginController creates a new plugin controller
func NewPluginController(pluginService *services.PluginService) *PluginController {
	return &PluginController{
		pluginService: pluginService,
	}
}

// ListPlugins handles GET /api/v1/plugins
func (c *PluginController) ListPlugins(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active" // Default to active plugins
	}

	plugins, err := c.pluginService.ListPlugins(status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to response format
	pluginList := make([]map[string]interface{}, 0, len(plugins))
	for _, plugin := range plugins {
		pluginList = append(pluginList, map[string]interface{}{
			"name":        plugin.Name,
			"version":     plugin.Version,
			"description": plugin.Description,
			"status":      plugin.Status,
			"checksum":    plugin.Checksum,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plugins": pluginList,
	})
}

// GetPluginManifest handles GET /api/v1/plugins/{name}/manifest
func (c *PluginController) GetPluginManifest(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := r.URL.Query().Get("version") // Optional version parameter

	manifest, err := c.pluginService.GetPluginManifest(name, version)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "plugin manifest not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, manifest)
}

// DownloadPlugin handles GET /api/v1/plugins/{name}/download
func (c *PluginController) DownloadPlugin(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := r.URL.Query().Get("version") // Optional version parameter

	// Get plugin package path
	packagePath, plugin, err := c.pluginService.GetPluginPackage(name, version)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "plugin package not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Verify file exists
	fileInfo, err := os.Stat(packagePath)
	if err != nil {
		writeError(w, http.StatusNotFound, "plugin package file not found")
		return
	}

	// Set headers for file download
	filename := filepath.Base(packagePath)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	
	// Include checksum in headers
	if plugin.Checksum != "" {
		w.Header().Set("X-Plugin-Checksum", plugin.Checksum)
	}

	// Serve file
	http.ServeFile(w, r, packagePath)
}

