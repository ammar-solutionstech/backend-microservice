package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"backend/services/client-container/internal/services"
)

// ServiceController handles internal service management API endpoints
// These endpoints are used by container-management service to manage services/plugins
type ServiceController struct {
	serviceRegistry *services.ServiceRegistry
}

// NewServiceController creates a new service controller
func NewServiceController(serviceRegistry *services.ServiceRegistry) *ServiceController {
	return &ServiceController{
		serviceRegistry: serviceRegistry,
	}
}

// AddService handles POST /api/v1/internal/services
// Adds a new service/plugin to the client-container
func (c *ServiceController) AddService(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string                 `json:"name"`
		Type    string                 `json:"type"` // "internal" or "external"
		Config  map[string]interface{} `json:"config"`
		Enabled bool                   `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate service type
	if req.Type != "internal" && req.Type != "external" {
		writeError(w, http.StatusBadRequest, "invalid service type: must be 'internal' or 'external'")
		return
	}

	// For now, we'll create a placeholder service
	// In a full implementation, this would load the actual service/plugin
	// For external services, this might involve loading a plugin file
	// For internal services, this would instantiate the built-in service

	// TODO: Implement actual service loading based on type
	// For now, return an error indicating this needs to be implemented
	writeError(w, http.StatusNotImplemented, "service loading not yet implemented")
}

// ListServices handles GET /api/v1/internal/services
// Lists all registered services
func (c *ServiceController) ListServices(w http.ResponseWriter, r *http.Request) {
	serviceNames := c.serviceRegistry.ListServices()
	allServices := c.serviceRegistry.GetAllServices()

	services := make([]map[string]interface{}, 0, len(serviceNames))
	for _, name := range serviceNames {
		service := allServices[name]
		services = append(services, map[string]interface{}{
			"name":   name,
			"status": service.Status(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"services": services,
	})
}

// GetService handles GET /api/v1/internal/services/:name
// Gets details about a specific service
func (c *ServiceController) GetService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	_, err := c.serviceRegistry.GetService(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	status, err := c.serviceRegistry.GetServiceStatus(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"name":   name,
		"status": status,
	})
}

// UpdateService handles PUT /api/v1/internal/services/:name
// Updates the configuration of a service
func (c *ServiceController) UpdateService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var req struct {
		Config map[string]interface{} `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := c.serviceRegistry.ConfigureService(name, req.Config); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "service configuration updated successfully",
	})
}

// StartService handles POST /api/v1/internal/services/:name/start
// Starts a service
func (c *ServiceController) StartService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if err := c.serviceRegistry.StartService(name); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "service started successfully",
	})
}

// StopService handles POST /api/v1/internal/services/:name/stop
// Stops a service
func (c *ServiceController) StopService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if err := c.serviceRegistry.StopService(name); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "service stopped successfully",
	})
}

// RemoveService handles DELETE /api/v1/internal/services/:name
// Removes a service from the registry
func (c *ServiceController) RemoveService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if err := c.serviceRegistry.UnregisterService(name); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "service removed successfully",
	})
}
