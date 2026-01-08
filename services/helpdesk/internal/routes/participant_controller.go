package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gorm.io/gorm"

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/services"
)

type ParticipantController struct {
	cfg           *config.Config
	ticketService *services.TicketService
}

func NewParticipantController(cfg *config.Config, ticketService *services.TicketService) *ParticipantController {
	return &ParticipantController{
		cfg:           cfg,
		ticketService: ticketService,
	}
}

func (c *ParticipantController) GetParticipants(w http.ResponseWriter, r *http.Request) {
	helpDeskID, err := strconv.Atoi(chi.URLParam(r, "helpDeskId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid help desk id")
		return
	}
	participants, err := c.ticketService.ListParticipants(helpDeskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list participants")
		return
	}
	writeJSON(w, http.StatusOK, participants)
}

func (c *ParticipantController) AddParticipant(w http.ResponseWriter, r *http.Request) {
	helpDeskID, err := strconv.Atoi(chi.URLParam(r, "helpDeskId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid help desk id")
		return
	}
	var payload struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.UserID <= 0 {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	// Note: UserID validation should be done via gRPC call to Auth service
	if err := c.ticketService.AddParticipant(helpDeskID, payload.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add participant")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{
		"help_desk_id": helpDeskID,
		"user_id":      payload.UserID,
	})
}

func (c *ParticipantController) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	helpDeskID, err := strconv.Atoi(chi.URLParam(r, "helpDeskId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid help desk id")
		return
	}
	userID, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := c.ticketService.RemoveParticipant(helpDeskID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "participant link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to remove participant")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

