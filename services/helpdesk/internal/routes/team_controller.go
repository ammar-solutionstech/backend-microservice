package routes

import (
	"encoding/json"
	"net/http"

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/services"
)

type TeamController struct {
	cfg         *config.Config
	teamService *services.TeamService
}

func NewTeamController(cfg *config.Config, teamService *services.TeamService) *TeamController {
	return &TeamController{cfg: cfg, teamService: teamService}
}

func (c *TeamController) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	team, err := c.teamService.CreateTeam(req.Name, req.Description)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, team)
}

func (c *TeamController) GetTeam(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	team, err := c.teamService.GetTeam(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}

	writeJSON(w, http.StatusOK, team)
}

func (c *TeamController) ListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := c.teamService.ListTeams()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list teams")
		return
	}

	writeJSON(w, http.StatusOK, teams)
}

func (c *TeamController) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	team, err := c.teamService.UpdateTeam(id, req.Name, req.Description)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, team)
}

func (c *TeamController) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	if err := c.teamService.DeleteTeam(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete team")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "team deleted successfully"})
}

func (c *TeamController) GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	members, err := c.teamService.GetTeamMembers(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get team members")
		return
	}

	writeJSON(w, http.StatusOK, members)
}

func (c *TeamController) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	var req struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.teamService.AddTeamMember(id, req.UserID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "member added successfully"})
}

func (c *TeamController) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	userID, err := getIDParam(r, "userId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := c.teamService.RemoveTeamMember(teamID, userID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "member removed successfully"})
}
