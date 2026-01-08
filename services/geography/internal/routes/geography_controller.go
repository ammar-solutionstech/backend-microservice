package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gorm.io/gorm"

	"backend/services/geography/internal/models"
	"backend/services/geography/internal/services"
)

type GeographyController struct {
	db              *gorm.DB
	countryService  *services.GenericService[models.Country]
	cityService     *services.GenericService[models.City]
	locationService *services.GenericService[models.Location]
	contactService  *services.GenericService[models.Contact]
}

func NewGeographyController(db *gorm.DB) *GeographyController {
	return &GeographyController{
		db:              db,
		countryService:  services.NewGenericService[models.Country](db),
		cityService:     services.NewGenericService[models.City](db),
		locationService: services.NewGenericService[models.Location](db),
		contactService:  services.NewGenericService[models.Contact](db),
	}
}

// Country handlers
func (c *GeographyController) ListCountries(w http.ResponseWriter, r *http.Request) {
	var records []models.Country
	if err := c.countryService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list countries")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *GeographyController) GetCountry(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.Country
	if err := c.countryService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "country not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve country")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *GeographyController) CreateCountry(w http.ResponseWriter, r *http.Request) {
	var payload models.Country
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.countryService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create country")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *GeographyController) UpdateCountry(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	record, err := c.countryService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "country not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update country")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *GeographyController) DeleteCountry(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.countryService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "country not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete country")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// City handlers (similar pattern)
func (c *GeographyController) ListCities(w http.ResponseWriter, r *http.Request) {
	var records []models.City
	if err := c.cityService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list cities")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *GeographyController) GetCity(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.City
	if err := c.cityService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "city not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve city")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *GeographyController) CreateCity(w http.ResponseWriter, r *http.Request) {
	var payload models.City
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.cityService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create city")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *GeographyController) UpdateCity(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	record, err := c.cityService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "city not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update city")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *GeographyController) DeleteCity(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.cityService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "city not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete city")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Location handlers
func (c *GeographyController) ListLocations(w http.ResponseWriter, r *http.Request) {
	var records []models.Location
	if err := c.locationService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list locations")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *GeographyController) GetLocation(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.Location
	if err := c.locationService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "location not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve location")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *GeographyController) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var payload models.Location
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.locationService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create location")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *GeographyController) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	record, err := c.locationService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "location not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update location")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *GeographyController) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.locationService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "location not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete location")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Contact handlers
func (c *GeographyController) ListContacts(w http.ResponseWriter, r *http.Request) {
	var records []models.Contact
	if err := c.contactService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list contacts")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *GeographyController) GetContact(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.Contact
	if err := c.contactService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "contact not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve contact")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *GeographyController) CreateContact(w http.ResponseWriter, r *http.Request) {
	var payload models.Contact
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.contactService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create contact")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *GeographyController) UpdateContact(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	record, err := c.contactService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "contact not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update contact")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *GeographyController) DeleteContact(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.contactService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "contact not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete contact")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Helper functions
func parseIDParam(r *http.Request, param string) (int, error) {
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

