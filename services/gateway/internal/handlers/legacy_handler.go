package handlers

import (
	"net/http"

	"gorm.io/gorm"

	"backend/services/gateway/internal/config"
)

type LegacyHandler struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewLegacyHandler(cfg *config.Config) *LegacyHandler {
	return &LegacyHandler{
		cfg: cfg,
		db:  cfg.LegacyDB,
	}
}

// HandleInventory handles inventory-related requests (temporary)
func (h *LegacyHandler) HandleInventory(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		writeError(w, http.StatusServiceUnavailable, "legacy database not available")
		return
	}

	// Placeholder - implement actual inventory logic
	writeJSON(w, http.StatusOK, map[string]string{"message": "Inventory endpoint - to be implemented"})
}

// HandleGeography handles geography-related requests (temporary)
func (h *LegacyHandler) HandleGeography(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		writeError(w, http.StatusServiceUnavailable, "legacy database not available")
		return
	}

	// Placeholder - implement actual geography logic
	writeJSON(w, http.StatusOK, map[string]string{"message": "Geography endpoint - to be implemented"})
}

// HandleNavigation handles navigation-related requests (temporary)
func (h *LegacyHandler) HandleNavigation(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		writeError(w, http.StatusServiceUnavailable, "legacy database not available")
		return
	}

	// Placeholder - implement actual navigation logic
	writeJSON(w, http.StatusOK, map[string]string{"message": "Navigation endpoint - to be implemented"})
}

// HandleRoles handles roles-related requests (temporary)
func (h *LegacyHandler) HandleRoles(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		writeError(w, http.StatusServiceUnavailable, "legacy database not available")
		return
	}

	// Placeholder - implement actual roles logic
	writeJSON(w, http.StatusOK, map[string]string{"message": "Roles endpoint - to be implemented"})
}

// HandlePermissions handles permissions-related requests (temporary)
func (h *LegacyHandler) HandlePermissions(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		writeError(w, http.StatusServiceUnavailable, "legacy database not available")
		return
	}

	// Placeholder - implement actual permissions logic
	writeJSON(w, http.StatusOK, map[string]string{"message": "Permissions endpoint - to be implemented"})
}

/* func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
*/
