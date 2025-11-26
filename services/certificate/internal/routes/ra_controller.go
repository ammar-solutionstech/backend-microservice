package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"backend/services/certificate/internal/services"
)

type RAController struct {
	raService *services.RAService
}

func NewRAController(raService *services.RAService) *RAController {
	return &RAController{
		raService: raService,
	}
}

// SubmitCSR handles POST /api/csr
func (c *RAController) SubmitCSR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CSRPEM          string `json:"csr_pem"`
		RequesterUserID *int   `json:"requester_user_id,omitempty"`
		RequesterEmail  string `json:"requester_email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	csr, err := c.raService.SubmitCSR(req.CSRPEM, req.RequesterUserID, req.RequesterEmail)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, csr)
}

// GetCSR handles GET /api/csr/:id
func (c *RAController) GetCSR(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid CSR id")
		return
	}

	csr, err := c.raService.GetCSR(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, csr)
}

// ApproveCSR handles POST /api/csr/:id/approve
func (c *RAController) ApproveCSR(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid CSR id")
		return
	}

	var req struct {
		ApproverID int  `json:"approver_id"`
		AutoIssue  bool `json:"auto_issue,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.raService.ApproveCSR(id, req.ApproverID, req.AutoIssue); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "CSR approved successfully"})
}

// RejectCSR handles POST /api/csr/:id/reject
func (c *RAController) RejectCSR(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid CSR id")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.raService.RejectCSR(id, req.Reason); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "CSR rejected successfully"})
}

// ListCSRs handles GET /api/csr
func (c *RAController) ListCSRs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	var requesterUserID *int
	if uid := r.URL.Query().Get("requester_user_id"); uid != "" {
		if id, err := strconv.Atoi(uid); err == nil {
			requesterUserID = &id
		}
	}

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

	csrs, total, err := c.raService.ListCSRs(status, requesterUserID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"csrs":   csrs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func getIDParam(r *http.Request, param string) (int, error) {
	idStr := chi.URLParam(r, param)
	return strconv.Atoi(idStr)
}
