package routes

import (
	"net/http"

	"github.com/go-chi/chi"

	"backend/services/certificate/internal/services"
)

type VAController struct {
	vaService *services.VAService
}

func NewVAController(vaService *services.VAService) *VAController {
	return &VAController{
		vaService: vaService,
	}
}

// GetOCSPStatus handles GET /api/ocsp/:serial
func (c *VAController) GetOCSPStatus(w http.ResponseWriter, r *http.Request) {
	serial := chi.URLParam(r, "serial")
	if serial == "" {
		writeError(w, http.StatusBadRequest, "serial number is required")
		return
	}

	status, err := c.vaService.GetOCSPStatus(serial)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, status)
}

// GetCRL handles GET /api/crl
func (c *VAController) GetCRL(w http.ResponseWriter, r *http.Request) {
	crl, err := c.vaService.GetCRL()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/pkix-crl")
	w.WriteHeader(http.StatusOK)
	w.Write(crl)
}

// ValidateCertificate handles GET /api/validate/:serial
func (c *VAController) ValidateCertificate(w http.ResponseWriter, r *http.Request) {
	serial := chi.URLParam(r, "serial")
	if serial == "" {
		writeError(w, http.StatusBadRequest, "serial number is required")
		return
	}

	valid, err := c.vaService.ValidateCertificate(serial)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"serial": serial,
		"valid":  valid,
	})
}
