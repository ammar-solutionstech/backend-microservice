package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"backend/services/gateway/internal/clients"
	"backend/services/gateway/internal/config"
	helpdeskpb "backend/services/helpdesk/proto"
)

type HelpDeskHandler struct {
	cfg            *config.Config
	helpDeskClient *clients.HelpDeskClient
}

func NewHelpDeskHandler(cfg *config.Config) *HelpDeskHandler {
	return &HelpDeskHandler{
		cfg:            cfg,
		helpDeskClient: clients.NewHelpDeskClient(cfg.HelpDeskServiceGRPC),
	}
}

func (h *HelpDeskHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req helpdeskpb.CreateTicketRequest
	json.Unmarshal(body, &req)

	resp, err := h.helpDeskClient.CreateTicket(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *HelpDeskHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	resp, err := h.helpDeskClient.GetTicket(r.Context(), &helpdeskpb.GetTicketRequest{TicketId: int32(id)})
	if err != nil {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *HelpDeskHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}

	req := &helpdeskpb.ListTicketsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if userID := r.URL.Query().Get("user_id"); userID != "" {
		if id, err := strconv.Atoi(userID); err == nil {
			req.UserId = int32(id)
		}
	}

	if state := r.URL.Query().Get("state"); state != "" {
		req.State = state
	}

	resp, err := h.helpDeskClient.ListTickets(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tickets")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
