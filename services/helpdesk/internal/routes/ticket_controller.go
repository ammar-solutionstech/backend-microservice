package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/services"
)

type TicketController struct {
	cfg           *config.Config
	ticketService *services.TicketService
}

func NewTicketController(cfg *config.Config, ticketService *services.TicketService) *TicketController {
	return &TicketController{cfg: cfg, ticketService: ticketService}
}

func (c *TicketController) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		HelpDeskTypeID int    `json:"help_desk_type_id"`
		PortalUserID   int    `json:"portal_user_id"`
		State          string `json:"state"`
		ParentID       *int   `json:"parent_id"`
		ProjectID      *int   `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	ticket, err := c.ticketService.CreateTicket(req.Name, req.Description, req.HelpDeskTypeID, req.PortalUserID, req.State, req.ParentID, req.ProjectID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ticket)
}

func (c *TicketController) GetTicket(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	ticket, err := c.ticketService.GetTicket(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

func (c *TicketController) ListTickets(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}

	var userID *int
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		if id, err := strconv.Atoi(uid); err == nil {
			userID = &id
		}
	}

	var state *string
	if s := r.URL.Query().Get("state"); s != "" {
		state = &s
	}

	tickets, total, err := c.ticketService.ListTickets(page, pageSize, userID, state)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tickets")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tickets":   tickets,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (c *TicketController) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	var req struct {
		Name           *string `json:"name"`
		Description    *string `json:"description"`
		State          *string `json:"state"`
		HelpDeskTypeID *int    `json:"help_desk_type_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	ticket, err := c.ticketService.UpdateTicket(id, req.Name, req.Description, req.State, req.HelpDeskTypeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

func (c *TicketController) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	if err := c.ticketService.DeleteTicket(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete ticket")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ticket deleted successfully"})
}

func (c *TicketController) AddComment(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	var req struct {
		UserID  int    `json:"user_id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	comment, err := c.ticketService.AddComment(id, req.UserID, req.Content)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (c *TicketController) GetComments(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	comments, err := c.ticketService.GetComments(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get comments")
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

func (c *TicketController) AddAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	// Simplified - in production, handle multipart form data
	var req struct {
		Name         string `json:"name"`
		Content      []byte `json:"content"`
		DocumentType string `json:"document_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	document, err := c.ticketService.AddAttachment(id, req.Name, req.Content, req.DocumentType)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, document)
}

func (c *TicketController) GetAttachments(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	attachments, err := c.ticketService.GetAttachments(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get attachments")
		return
	}

	writeJSON(w, http.StatusOK, attachments)
}

func (c *TicketController) AssignUser(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	var req struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.ticketService.AssignUser(id, req.UserID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user assigned successfully"})
}

func (c *TicketController) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.ticketService.UpdateStatus(id, req.Status); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "status updated successfully"})
}

func getIDParam(r *http.Request, param string) (int, error) {
	idStr := chi.URLParam(r, param)
	return strconv.Atoi(idStr)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
