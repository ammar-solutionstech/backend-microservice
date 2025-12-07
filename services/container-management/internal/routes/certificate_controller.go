package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"backend/services/container-management/internal/services"
)

type CertificateController struct {
	certClient   *services.CertificateClient
	csrValidator *services.CSRValidator
}

func NewCertificateController(certClient *services.CertificateClient, csrValidator *services.CSRValidator) *CertificateController {
	return &CertificateController{
		certClient:   certClient,
		csrValidator: csrValidator,
	}
}

// RequestCertificate handles POST /api/v1/certificates/request
func (c *CertificateController) RequestCertificate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CSRPEM string `json:"csr_pem"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Validate CSR
	if err := c.csrValidator.Validate(req.CSRPEM); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Submit to certificate service
	csrResp, err := c.certClient.SubmitCSR(req.CSRPEM, "api-request")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, csrResp)
}

// GetCertificate handles GET /api/v1/certificates/:serial
func (c *CertificateController) GetCertificate(w http.ResponseWriter, r *http.Request) {
	serial := chi.URLParam(r, "serial")
	if serial == "" {
		writeError(w, http.StatusBadRequest, "serial number is required")
		return
	}

	cert, err := c.certClient.GetCertificate(serial)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cert)
}

// RevokeCertificate handles POST /api/v1/certificates/:serial/revoke
func (c *CertificateController) RevokeCertificate(w http.ResponseWriter, r *http.Request) {
	serial := chi.URLParam(r, "serial")
	if serial == "" {
		writeError(w, http.StatusBadRequest, "serial number is required")
		return
	}

	var req struct {
		Reason int `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.certClient.RevokeCertificate(serial, req.Reason); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "certificate revoked successfully"})
}

// ListCertificates handles GET /api/v1/certificates
func (c *CertificateController) ListCertificates(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	certs, err := c.certClient.ListCertificates(status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, certs)
}

