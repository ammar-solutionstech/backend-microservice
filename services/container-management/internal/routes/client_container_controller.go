package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"backend/services/container-management/internal/services"
)

type ClientContainerController struct {
	containerService *services.ClientContainerService
}

func NewClientContainerController(containerService *services.ClientContainerService) *ClientContainerController {
	return &ClientContainerController{
		containerService: containerService,
	}
}

// RegisterClientContainer handles POST /api/v1/organizations/:id/containers
func (c *ClientContainerController) RegisterClientContainer(w http.ResponseWriter, r *http.Request) {
	orgID, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization id")
		return
	}

	var req struct {
		ContainerID string  `json:"container_id"`
		Name        string  `json:"name"`
		EndpointURL string  `json:"endpoint_url"`
		AdminEmail  string  `json:"admin_email"`
		AdminPhone  *string `json:"admin_phone,omitempty"`
		CSRPEM      string  `json:"csr_pem"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	container, err := c.containerService.RegisterClientContainer(
		orgID,
		req.ContainerID,
		req.Name,
		req.EndpointURL,
		req.AdminEmail,
		req.AdminPhone,
		req.CSRPEM,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, container)
}

// GetClientContainer handles GET /api/v1/containers/:container_id
func (c *ClientContainerController) GetClientContainer(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	container, err := c.containerService.GetClientContainer(containerID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, container)
}

// GetClientContainerByOrg handles GET /api/v1/organizations/:id/containers
func (c *ClientContainerController) GetClientContainerByOrg(w http.ResponseWriter, r *http.Request) {
	orgID, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization id")
		return
	}

	container, err := c.containerService.GetClientContainerByOrg(orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, container)
}

// StartContainer handles POST /api/v1/containers/:container_id/start
func (c *ClientContainerController) StartContainer(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	if err := c.containerService.StartContainer(containerID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "container started"})
}

// StopContainer handles POST /api/v1/containers/:container_id/stop
func (c *ClientContainerController) StopContainer(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	if err := c.containerService.StopContainer(containerID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "container stopped"})
}

// RestartContainer handles POST /api/v1/containers/:container_id/restart
func (c *ClientContainerController) RestartContainer(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	if err := c.containerService.RestartContainer(containerID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "container restarted"})
}

// UpdateContainer handles PUT /api/v1/containers/:container_id/update
func (c *ClientContainerController) UpdateContainer(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.containerService.UpdateClientContainer(containerID, updates); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "container updated successfully"})
}

// RemoveContainer handles DELETE /api/v1/containers/:container_id
func (c *ClientContainerController) RemoveContainer(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	if err := c.containerService.RemoveContainer(containerID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "container removed"})
}

// GetContainerStatus handles GET /api/v1/containers/:container_id/status
func (c *ClientContainerController) GetContainerStatus(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	status, err := c.containerService.GetContainerStatus(containerID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": status})
}
