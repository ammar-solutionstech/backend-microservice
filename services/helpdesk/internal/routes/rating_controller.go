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

type RatingController struct {
	cfg           *config.Config
	ratingService *services.RatingService
}

func NewRatingController(cfg *config.Config, ratingService *services.RatingService) *RatingController {
	return &RatingController{
		cfg:           cfg,
		ratingService: ratingService,
	}
}

func (c *RatingController) ListRatings(w http.ResponseWriter, r *http.Request) {
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
	ratings, err := c.ratingService.ListRatings(helpDeskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list ratings")
		return
	}
	writeJSON(w, http.StatusOK, ratings)
}

func (c *RatingController) CreateRating(w http.ResponseWriter, r *http.Request) {
	var rating models.HelpDeskRating
	if err := json.NewDecoder(r.Body).Decode(&rating); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.ratingService.CreateRating(&rating); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create rating")
		return
	}
	writeJSON(w, http.StatusCreated, rating)
}

func (c *RatingController) GetRating(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rating id")
		return
	}
	rating, err := c.ratingService.GetRating(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "rating not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve rating")
		return
	}
	writeJSON(w, http.StatusOK, rating)
}

func (c *RatingController) UpdateRating(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rating id")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rating, err := c.ratingService.UpdateRating(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "rating not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update rating")
		return
	}
	writeJSON(w, http.StatusOK, rating)
}

func (c *RatingController) DeleteRating(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rating id")
		return
	}
	if err := c.ratingService.DeleteRating(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "rating not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete rating")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

