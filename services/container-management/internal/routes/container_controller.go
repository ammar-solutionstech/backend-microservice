package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"backend/services/container-management/internal/services"
)

type ContainerController struct {
	containerService *services.ContainerService
}

func NewContainerController(containerService *services.ContainerService) *ContainerController {
	return &ContainerController{
		containerService: containerService,
	}
}

// RequestContainerCertificate handles POST /api/v1/containers/:container_id/certificates/request
func (c *ContainerController) RequestContainerCertificate(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	var req struct {
		CSRPEM string `json:"csr_pem"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	certReq, err := c.containerService.RequestCertificate(containerID, req.CSRPEM)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, certReq)
}

// RequestApplicationCertificate handles POST /api/v1/containers/:container_id/applications/:app_name/certificates/request
func (c *ContainerController) RequestApplicationCertificate(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	appName := chi.URLParam(r, "app_name")

	if containerID == "" || appName == "" {
		writeError(w, http.StatusBadRequest, "container_id and app_name are required")
		return
	}

	var req struct {
		CSRPEM string `json:"csr_pem"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	certReq, err := c.containerService.RequestApplicationCertificate(containerID, appName, req.CSRPEM)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, certReq)
}

// GetContainerCertificates handles GET /api/v1/containers/:container_id/certificates
func (c *ContainerController) GetContainerCertificates(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	certs, err := c.containerService.GetContainerCertificates(containerID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, certs)
}

// RevokeContainerCertificate handles POST /api/v1/containers/:container_id/certificates/:serial/revoke
func (c *ContainerController) RevokeContainerCertificate(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "container_id")
	serial := chi.URLParam(r, "serial")

	if containerID == "" || serial == "" {
		writeError(w, http.StatusBadRequest, "container_id and serial are required")
		return
	}

	if err := c.containerService.RevokeContainerCertificate(containerID, serial); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "certificate revoked successfully"})
}

