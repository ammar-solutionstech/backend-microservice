package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"backend/services/certificate/internal/services"
)

type CAController struct {
	caService *services.CAService
}

func NewCAController(caService *services.CAService) *CAController {
	return &CAController{
		caService: caService,
	}
}

// IssueCertificate handles POST /api/certificates
func (c *CAController) IssueCertificate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CSRID int `json:"csr_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	cert, err := c.caService.IssueCertificate(req.CSRID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cert)
}

// GetCertificate handles GET /api/certificates/:serial
func (c *CAController) GetCertificate(w http.ResponseWriter, r *http.Request) {
	serial := chi.URLParam(r, "serial")
	if serial == "" {
		writeError(w, http.StatusBadRequest, "serial number is required")
		return
	}

	cert, err := c.caService.GetCertificate(serial)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cert)
}

// RevokeCertificate handles POST /api/certificates/:serial/revoke
func (c *CAController) RevokeCertificate(w http.ResponseWriter, r *http.Request) {
	serial := chi.URLParam(r, "serial")
	if serial == "" {
		writeError(w, http.StatusBadRequest, "serial number is required")
		return
	}

	var req struct {
		Reason    int  `json:"reason"`
		RevokedBy *int `json:"revoked_by,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.caService.RevokeCertificate(serial, req.Reason, req.RevokedBy); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "certificate revoked successfully"})
}

// RenewCertificate handles POST /api/certificates/renew
func (c *CAController) RenewCertificate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Serial    string `json:"serial"`
		NewCSRPEM string `json:"new_csr_pem"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	cert, err := c.caService.RenewCertificate(req.Serial, req.NewCSRPEM)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cert)
}

// ListCertificates handles GET /api/certificates
func (c *CAController) ListCertificates(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	certs, total, err := c.caService.ListCertificates(status, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"certificates": certs,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}
