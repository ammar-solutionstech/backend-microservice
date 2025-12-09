package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"backend/services/container-management/internal/models"
	"backend/services/container-management/internal/services"
)

type ServiceController struct {
	serviceManager *services.ServiceManager
}

func NewServiceController(serviceManager *services.ServiceManager) *ServiceController {
	return &ServiceController{
		serviceManager: serviceManager,
	}
}

// AddService handles POST /api/v1/containers/:container_id/services
func (c *ServiceController) AddService(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	var req models.ContainerService
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.serviceManager.AddServiceToContainer(containerID, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, req)
}

// ListServices handles GET /api/v1/containers/:container_id/services
func (c *ServiceController) ListServices(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	services, err := c.serviceManager.ListContainerServices(containerID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, services)
}

// GetService handles GET /api/v1/containers/:container_id/services/:service_name
func (c *ServiceController) GetService(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	serviceName := chi.URLParam(r, "service_name")
	if containerID == "" || serviceName == "" {
		writeError(w, http.StatusBadRequest, "container_id and service_name are required")
		return
	}

	services, err := c.serviceManager.ListContainerServices(containerID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// Find the specific service
	for _, svc := range services {
		if svc.ServiceName == serviceName {
			writeJSON(w, http.StatusOK, svc)
			return
		}
	}

	writeError(w, http.StatusNotFound, "service not found")
}

// UpdateService handles PUT /api/v1/containers/:container_id/services/:service_name
func (c *ServiceController) UpdateService(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	serviceName := chi.URLParam(r, "service_name")
	if containerID == "" || serviceName == "" {
		writeError(w, http.StatusBadRequest, "container_id and service_name are required")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.serviceManager.UpdateServiceInContainer(containerID, serviceName, updates); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "service updated successfully"})
}

// EnableService handles POST /api/v1/containers/:container_id/services/:service_name/enable
func (c *ServiceController) EnableService(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	serviceName := chi.URLParam(r, "service_name")
	if containerID == "" || serviceName == "" {
		writeError(w, http.StatusBadRequest, "container_id and service_name are required")
		return
	}

	if err := c.serviceManager.EnableService(containerID, serviceName); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "service enabled"})
}

// DisableService handles POST /api/v1/containers/:container_id/services/:service_name/disable
func (c *ServiceController) DisableService(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	serviceName := chi.URLParam(r, "service_name")
	if containerID == "" || serviceName == "" {
		writeError(w, http.StatusBadRequest, "container_id and service_name are required")
		return
	}

	if err := c.serviceManager.DisableService(containerID, serviceName); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "service disabled"})
}

// RemoveService handles DELETE /api/v1/containers/:container_id/services/:service_name
func (c *ServiceController) RemoveService(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	serviceName := chi.URLParam(r, "service_name")
	if containerID == "" || serviceName == "" {
		writeError(w, http.StatusBadRequest, "container_id and service_name are required")
		return
	}

	if err := c.serviceManager.RemoveServiceFromContainer(containerID, serviceName); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "service removed"})
}
