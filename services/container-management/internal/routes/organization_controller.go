package routes

import (
	"encoding/json"
	"net/http"

	"backend/services/container-management/internal/services"
)

type OrganizationController struct {
	orgService *services.OrganizationService
}

func NewOrganizationController(orgService *services.OrganizationService) *OrganizationController {
	return &OrganizationController{
		orgService: orgService,
	}
}

// CreateOrganization handles POST /api/v1/organizations
func (c *OrganizationController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string  `json:"name"`
		Domain     string  `json:"domain"`
		AdminEmail string  `json:"admin_email"`
		AdminPhone *string `json:"admin_phone,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	org, err := c.orgService.CreateOrganization(req.Name, req.Domain, req.AdminEmail, req.AdminPhone)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, org)
}

// GetOrganization handles GET /api/v1/organizations/:id
func (c *OrganizationController) GetOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization id")
		return
	}

	org, err := c.orgService.GetOrganization(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, org)
}

// ListOrganizations handles GET /api/v1/organizations
func (c *OrganizationController) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	orgs, err := c.orgService.ListOrganizations(status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, orgs)
}

// UpdateOrganization handles PUT /api/v1/organizations/:id
func (c *OrganizationController) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization id")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.orgService.UpdateOrganization(id, updates); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "organization updated successfully"})
}

