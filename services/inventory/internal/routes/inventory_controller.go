package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"gorm.io/gorm"

	"backend/services/inventory/internal/models"
	"backend/services/inventory/internal/services"
)

type InventoryController struct {
	cfg                interface{} // Config placeholder
	db                 *gorm.DB
	equipmentService   *services.EquipmentService
	brandService       *services.GenericService[models.Brand]
	modelService        *services.GenericService[models.Model]
	equipmentTypeService *services.GenericService[models.EquipmentType]
	osService           *services.GenericService[models.OperatingSystem]
	softwareCatService  *services.GenericService[models.SoftwareCategory]
	softwareService     *services.GenericService[models.Software]
	equipmentService2   *services.GenericService[models.Equipment]
	documentService     *services.GenericService[models.Document]
	maintenanceService  *services.GenericService[models.Maintenance]
}

func NewInventoryController(db *gorm.DB) *InventoryController {
	return &InventoryController{
		db:                 db,
		equipmentService:   services.NewEquipmentService(db),
		brandService:       services.NewGenericService[models.Brand](db),
		modelService:       services.NewGenericService[models.Model](db),
		equipmentTypeService: services.NewGenericService[models.EquipmentType](db),
		osService:          services.NewGenericService[models.OperatingSystem](db),
		softwareCatService: services.NewGenericService[models.SoftwareCategory](db),
		softwareService:    services.NewGenericService[models.Software](db),
		equipmentService2:   services.NewGenericService[models.Equipment](db),
		documentService:     services.NewGenericService[models.Document](db),
		maintenanceService:  services.NewGenericService[models.Maintenance](db),
	}
}

// Generic CRUD handlers
func (c *InventoryController) ListBrands(w http.ResponseWriter, r *http.Request) {
	var records []models.Brand
	if err := c.brandService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list brands")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *InventoryController) GetBrand(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.Brand
	if err := c.brandService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "brand not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve brand")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) CreateBrand(w http.ResponseWriter, r *http.Request) {
	var payload models.Brand
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.brandService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create brand")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateBrand(w http.ResponseWriter, r *http.Request) {
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
	record, err := c.brandService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "brand not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update brand")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) DeleteBrand(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.brandService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "brand not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete brand")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Equipment Software handlers
func (c *InventoryController) GetEquipmentSoftware(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	if err := c.equipmentService.ValidateEquipmentExists(equipmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "equipment validation failed")
		return
	}
	links, err := c.equipmentService.ListSoftware(equipmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list software")
		return
	}
	writeJSON(w, http.StatusOK, links)
}

func (c *InventoryController) CreateEquipmentSoftware(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	var payload models.EquipmentSoftware
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	payload.EquipmentID = equipmentID
	if err := c.equipmentService.ValidateEquipmentExists(payload.EquipmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "equipment validation failed")
		return
	}
	if err := c.equipmentService.ValidateSoftwareExists(payload.SoftwareID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusBadRequest, "software not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "software validation failed")
		return
	}
	if err := c.equipmentService.CreateSoftwareLink(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to link software to equipment")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateEquipmentSoftware(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	softwareID, err := parseIDParam(r, "softwareId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid software id")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	delete(updates, "equipment_id")
	delete(updates, "software_id")
	record, err := c.equipmentService.UpdateSoftwareLink(equipmentID, softwareID, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "software assignment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update software assignment")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) DeleteEquipmentSoftware(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	softwareID, err := parseIDParam(r, "softwareId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid software id")
		return
	}
	if err := c.equipmentService.DeleteSoftwareLink(equipmentID, softwareID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "software assignment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete software assignment")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Equipment Help Desk handlers
func (c *InventoryController) GetEquipmentHelpDeskLinks(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	if err := c.equipmentService.ValidateEquipmentExists(equipmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "equipment validation failed")
		return
	}
	links, err := c.equipmentService.ListHelpDeskLinks(equipmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list equipment help desk links")
		return
	}
	writeJSON(w, http.StatusOK, links)
}

func (c *InventoryController) CreateEquipmentHelpDeskLink(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	var payload struct {
		HelpDeskID int `json:"help_desk_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.HelpDeskID <= 0 {
		writeError(w, http.StatusBadRequest, "help_desk_id is required")
		return
	}
	if err := c.equipmentService.ValidateEquipmentExists(equipmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "equipment validation failed")
		return
	}
	// Note: HelpDeskID validation should be done via gRPC call to Help Desk service
	if err := c.equipmentService.AddHelpDeskLink(equipmentID, payload.HelpDeskID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to link equipment to help desk")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{
		"equipment_id": equipmentID,
		"help_desk_id":    payload.HelpDeskID,
	})
}

func (c *InventoryController) DeleteEquipmentHelpDeskLink(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	helpDeskID, err := parseIDParam(r, "helpDeskId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid help desk id")
		return
	}
	if err := c.equipmentService.RemoveHelpDeskLink(equipmentID, helpDeskID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to unlink equipment from help desk")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Equipment User History handlers
func (c *InventoryController) GetEquipmentUserHistory(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	if err := c.equipmentService.ValidateEquipmentExists(equipmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "equipment validation failed")
		return
	}
	history, err := c.equipmentService.ListUserHistory(equipmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list equipment user history")
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func (c *InventoryController) CreateEquipmentUserHistory(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	var payload models.EquipmentUserHistory
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	payload.EquipmentID = equipmentID
	if err := c.equipmentService.ValidateEquipmentExists(payload.EquipmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "equipment validation failed")
		return
	}
	// Note: UserID validation should be done via gRPC call to Auth service
	if payload.EndDate.Before(payload.StartDate) {
		writeError(w, http.StatusBadRequest, "end_date must not precede start_date")
		return
	}
	if err := c.equipmentService.AddUserHistory(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create history entry")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateEquipmentUserHistory(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	userID, err := parseIDParam(r, "userId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	startDateRaw := chi.URLParam(r, "startDate")
	startDate, err := time.Parse("2006-01-02", startDateRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start date format, expected YYYY-MM-DD")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	delete(updates, "equipment_id")
	delete(updates, "user_id")
	delete(updates, "start_date")
	if endRaw, ok := updates["end_date"].(string); ok {
		if parsed, err := time.Parse("2006-01-02", endRaw); err == nil {
			if parsed.Before(startDate) {
				writeError(w, http.StatusBadRequest, "end_date must not precede start_date")
				return
			}
		}
	}
	entry, err := c.equipmentService.UpdateUserHistory(equipmentID, userID, startDate, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "history entry not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update history entry")
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (c *InventoryController) DeleteEquipmentUserHistory(w http.ResponseWriter, r *http.Request) {
	equipmentID, err := parseIDParam(r, "equipmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid equipment id")
		return
	}
	userID, err := parseIDParam(r, "userId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	startDateRaw := chi.URLParam(r, "startDate")
	startDate, err := time.Parse("2006-01-02", startDateRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start date format, expected YYYY-MM-DD")
		return
	}
	if err := c.equipmentService.DeleteUserHistory(equipmentID, userID, startDate); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "history entry not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete history entry")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Generic CRUD handlers for other resources (following Brand pattern)
func (c *InventoryController) ListModels(w http.ResponseWriter, r *http.Request) {
	var records []models.Model
	if err := c.modelService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list models")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *InventoryController) GetModel(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.Model
	if err := c.modelService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "model not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve model")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) CreateModel(w http.ResponseWriter, r *http.Request) {
	var payload models.Model
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.modelService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create model")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateModel(w http.ResponseWriter, r *http.Request) {
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
	record, err := c.modelService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "model not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update model")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) DeleteModel(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.modelService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "model not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete model")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Equipment Type handlers
func (c *InventoryController) ListEquipmentTypes(w http.ResponseWriter, r *http.Request) {
	var records []models.EquipmentType
	if err := c.equipmentTypeService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list equipment types")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *InventoryController) GetEquipmentType(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.EquipmentType
	if err := c.equipmentTypeService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment type not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve equipment type")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) CreateEquipmentType(w http.ResponseWriter, r *http.Request) {
	var payload models.EquipmentType
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.equipmentTypeService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create equipment type")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateEquipmentType(w http.ResponseWriter, r *http.Request) {
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
	record, err := c.equipmentTypeService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment type not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update equipment type")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) DeleteEquipmentType(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.equipmentTypeService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment type not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete equipment type")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Operating System handlers (similar pattern)
func (c *InventoryController) ListOperatingSystems(w http.ResponseWriter, r *http.Request) {
	var records []models.OperatingSystem
	if err := c.osService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list operating systems")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *InventoryController) GetOperatingSystem(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.OperatingSystem
	if err := c.osService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "operating system not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve operating system")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) CreateOperatingSystem(w http.ResponseWriter, r *http.Request) {
	var payload models.OperatingSystem
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.osService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create operating system")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateOperatingSystem(w http.ResponseWriter, r *http.Request) {
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
	record, err := c.osService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "operating system not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update operating system")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) DeleteOperatingSystem(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.osService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "operating system not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete operating system")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Software Category handlers
func (c *InventoryController) ListSoftwareCategories(w http.ResponseWriter, r *http.Request) {
	var records []models.SoftwareCategory
	if err := c.softwareCatService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list software categories")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *InventoryController) GetSoftwareCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.SoftwareCategory
	if err := c.softwareCatService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "software category not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve software category")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) CreateSoftwareCategory(w http.ResponseWriter, r *http.Request) {
	var payload models.SoftwareCategory
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.softwareCatService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create software category")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateSoftwareCategory(w http.ResponseWriter, r *http.Request) {
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
	record, err := c.softwareCatService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "software category not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update software category")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) DeleteSoftwareCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.softwareCatService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "software category not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete software category")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Software handlers
func (c *InventoryController) ListSoftware(w http.ResponseWriter, r *http.Request) {
	var records []models.Software
	if err := c.softwareService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list software")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *InventoryController) GetSoftware(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.Software
	if err := c.softwareService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "software not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve software")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) CreateSoftware(w http.ResponseWriter, r *http.Request) {
	var payload models.Software
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.softwareService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create software")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateSoftware(w http.ResponseWriter, r *http.Request) {
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
	record, err := c.softwareService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "software not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update software")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) DeleteSoftware(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.softwareService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "software not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete software")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Equipment handlers
func (c *InventoryController) ListEquipment(w http.ResponseWriter, r *http.Request) {
	var records []models.Equipment
	if err := c.equipmentService2.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list equipment")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *InventoryController) GetEquipment(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.Equipment
	if err := c.equipmentService2.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve equipment")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	var payload models.Equipment
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.equipmentService2.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create equipment")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateEquipment(w http.ResponseWriter, r *http.Request) {
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
	record, err := c.equipmentService2.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update equipment")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) DeleteEquipment(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.equipmentService2.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "equipment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete equipment")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Document handlers
func (c *InventoryController) ListDocuments(w http.ResponseWriter, r *http.Request) {
	var records []models.Document
	if err := c.documentService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list documents")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *InventoryController) GetDocument(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.Document
	if err := c.documentService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve document")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var payload models.Document
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.documentService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create document")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateDocument(w http.ResponseWriter, r *http.Request) {
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
	record, err := c.documentService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update document")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.documentService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete document")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Maintenance handlers
func (c *InventoryController) ListMaintenance(w http.ResponseWriter, r *http.Request) {
	var records []models.Maintenance
	if err := c.maintenanceService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list maintenance records")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *InventoryController) GetMaintenance(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.Maintenance
	if err := c.maintenanceService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "maintenance record not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve maintenance record")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	var payload models.Maintenance
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.maintenanceService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create maintenance record")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *InventoryController) UpdateMaintenance(w http.ResponseWriter, r *http.Request) {
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
	record, err := c.maintenanceService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "maintenance record not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update maintenance record")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *InventoryController) DeleteMaintenance(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.maintenanceService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "maintenance record not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete maintenance record")
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

