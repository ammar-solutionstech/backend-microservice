package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

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
		ContainerID        string  `json:"container_id"`
		Name               string  `json:"name"`
		EndpointURL        string  `json:"endpoint_url"`
		AdminEmail         string  `json:"admin_email"`
		AdminPhone         *string `json:"admin_phone,omitempty"`
		CSRPEM             string  `json:"csr_pem"`
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

