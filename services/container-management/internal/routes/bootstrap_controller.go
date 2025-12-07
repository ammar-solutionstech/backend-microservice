package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"backend/services/container-management/internal/services"
)

type BootstrapController struct {
	bootstrapService *services.BootstrapService
	containerService *services.ContainerService
	certClient       *services.CertificateClient
	csrValidator     *services.CSRValidator
}

func NewBootstrapController(
	bootstrapService *services.BootstrapService,
	containerService *services.ContainerService,
	certClient *services.CertificateClient,
	csrValidator *services.CSRValidator,
) *BootstrapController {
	return &BootstrapController{
		bootstrapService: bootstrapService,
		containerService: containerService,
		certClient:       certClient,
		csrValidator:     csrValidator,
	}
}

// Register handles POST /api/v1/bootstrap/register
func (c *BootstrapController) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BootstrapToken string `json:"bootstrap_token"`
		ContainerID    string `json:"container_id"`
		Name           string `json:"name"`
		CSRPEM         string `json:"csr_pem"`
		Metadata       map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Validate bootstrap token
	bootstrapToken, err := c.bootstrapService.ValidateBootstrapToken(req.BootstrapToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Validate container ID matches token (if specified in token)
	if bootstrapToken.ContainerID != nil && *bootstrapToken.ContainerID != req.ContainerID {
		writeError(w, http.StatusBadRequest, "container_id does not match bootstrap token")
		return
	}

	// Validate CSR
	if err := c.csrValidator.Validate(req.CSRPEM); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Register container
	container, err := c.containerService.RegisterContainer(req.ContainerID, req.Name, req.Metadata)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Submit CSR to certificate service
	csrResp, err := c.certClient.SubmitCSR(req.CSRPEM, fmt.Sprintf("bootstrap:container:%s", req.ContainerID))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Auto-approve and issue certificate (bootstrap flow)
	approverID := 0 // System approver
	if err := c.certClient.ApproveCSR(csrResp.ID, approverID, true); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to approve CSR: "+err.Error())
		return
	}

	// Issue certificate if not auto-issued
	cert, err := c.certClient.IssueCertificate(csrResp.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue certificate: "+err.Error())
		return
	}

	// Update container with certificate serial
	if err := c.containerService.UpdateContainerCertificateSerial(req.ContainerID, cert.SerialNumber); err != nil {
		// Log error but don't fail the request
		// In production, you might want to log this
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"container":   container,
		"certificate": cert,
		"message":     "container registered and certificate issued successfully",
	})
}

// Status handles GET /api/v1/bootstrap/status
func (c *BootstrapController) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "bootstrap service is operational",
	})
}

