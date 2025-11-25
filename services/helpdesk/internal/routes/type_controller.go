package routes

import (
	"encoding/json"
	"net/http"

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/services"
)

type HelpDeskTypeController struct {
	cfg         *config.Config
	typeService *services.HelpDeskTypeService
}

func NewHelpDeskTypeController(cfg *config.Config, typeService *services.HelpDeskTypeService) *HelpDeskTypeController {
	return &HelpDeskTypeController{cfg: cfg, typeService: typeService}
}

func (c *HelpDeskTypeController) CreateType(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		TeamID      *int   `json:"team_id"`
		IsActive    *bool  `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	helpDeskType, err := c.typeService.CreateType(req.Name, req.Description, req.TeamID, req.IsActive)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, helpDeskType)
}

func (c *HelpDeskTypeController) GetType(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid type id")
		return
	}

	helpDeskType, err := c.typeService.GetType(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "type not found")
		return
	}

	writeJSON(w, http.StatusOK, helpDeskType)
}

func (c *HelpDeskTypeController) ListTypes(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"
	types, err := c.typeService.ListTypes(activeOnly)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list types")
		return
	}

	writeJSON(w, http.StatusOK, types)
}

func (c *HelpDeskTypeController) UpdateType(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid type id")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		IsActive    *bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	helpDeskType, err := c.typeService.UpdateType(id, req.Name, req.Description, req.IsActive)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, helpDeskType)
}

func (c *HelpDeskTypeController) DeleteType(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid type id")
		return
	}

	if err := c.typeService.DeleteType(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete type")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "type deleted successfully"})
}
