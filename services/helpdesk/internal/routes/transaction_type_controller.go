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

type TransactionTypeController struct {
	cfg     *config.Config
	service *services.GenericService[models.TransactionType]
}

func NewTransactionTypeController(cfg *config.Config, service *services.GenericService[models.TransactionType]) *TransactionTypeController {
	return &TransactionTypeController{
		cfg:     cfg,
		service: service,
	}
}

func (c *TransactionTypeController) ListTransactionTypes(w http.ResponseWriter, r *http.Request) {
	var types []models.TransactionType
	if err := c.service.List(&types); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list transaction types")
		return
	}
	writeJSON(w, http.StatusOK, types)
}

func (c *TransactionTypeController) CreateTransactionType(w http.ResponseWriter, r *http.Request) {
	var transactionType models.TransactionType
	if err := json.NewDecoder(r.Body).Decode(&transactionType); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.service.Create(&transactionType); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create transaction type")
		return
	}
	writeJSON(w, http.StatusCreated, transactionType)
}

func (c *TransactionTypeController) GetTransactionType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction type id")
		return
	}
	var transactionType models.TransactionType
	if err := c.service.Get(id, &transactionType); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "transaction type not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve transaction type")
		return
	}
	writeJSON(w, http.StatusOK, transactionType)
}

func (c *TransactionTypeController) UpdateTransactionType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction type id")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	transactionType, err := c.service.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "transaction type not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update transaction type")
		return
	}
	writeJSON(w, http.StatusOK, transactionType)
}

func (c *TransactionTypeController) DeleteTransactionType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction type id")
		return
	}
	if err := c.service.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "transaction type not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete transaction type")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

