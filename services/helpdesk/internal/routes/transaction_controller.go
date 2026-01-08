package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gorm.io/gorm"

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/models"
	"backend/services/helpdesk/internal/services"
)

type TransactionController struct {
	cfg                *config.Config
	transactionService *services.TransactionService
	ticketService      *services.TicketService
}

func NewTransactionController(cfg *config.Config, transactionService *services.TransactionService, ticketService *services.TicketService) *TransactionController {
	return &TransactionController{
		cfg:                cfg,
		transactionService: transactionService,
		ticketService:      ticketService,
	}
}

func (c *TransactionController) ListTransactions(w http.ResponseWriter, r *http.Request) {
	helpDeskIDStr := r.URL.Query().Get("help_desk_id")
	if helpDeskIDStr == "" {
		writeError(w, http.StatusBadRequest, "help_desk_id query parameter is required")
		return
	}
	helpDeskID, err := strconv.Atoi(helpDeskIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid help_desk_id")
		return
	}
	transactions, err := c.transactionService.ListTransactions(helpDeskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}
	writeJSON(w, http.StatusOK, transactions)
}

func (c *TransactionController) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction models.HelpDeskTransaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.transactionService.CreateTransaction(&transaction); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create transaction")
		return
	}
	writeJSON(w, http.StatusCreated, transaction)
}

func (c *TransactionController) GetTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	transaction, err := c.transactionService.GetTransaction(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "transaction not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve transaction")
		return
	}
	writeJSON(w, http.StatusOK, transaction)
}

func (c *TransactionController) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	transaction, err := c.transactionService.UpdateTransaction(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "transaction not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update transaction")
		return
	}
	writeJSON(w, http.StatusOK, transaction)
}

func (c *TransactionController) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	if err := c.transactionService.DeleteTransaction(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "transaction not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete transaction")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *TransactionController) GetTransactionUsers(w http.ResponseWriter, r *http.Request) {
	transactionID, err := strconv.Atoi(chi.URLParam(r, "transactionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	users, err := c.transactionService.ListTransactionUsers(transactionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list transaction users")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (c *TransactionController) AddTransactionUser(w http.ResponseWriter, r *http.Request) {
	transactionID, err := strconv.Atoi(chi.URLParam(r, "transactionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
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
	if err := c.transactionService.AddTransactionUser(transactionID, payload.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add transaction user")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{
		"transaction_id": transactionID,
		"user_id":        payload.UserID,
	})
}

func (c *TransactionController) RemoveTransactionUser(w http.ResponseWriter, r *http.Request) {
	transactionID, err := strconv.Atoi(chi.URLParam(r, "transactionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	userID, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := c.transactionService.RemoveTransactionUser(transactionID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "transaction user link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to remove transaction user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

