package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gorm.io/gorm"

	"backend/services/navigation/internal/models"
	"backend/services/navigation/internal/services"
)

type NavigationController struct {
	db                 *gorm.DB
	menuService        *services.GenericService[models.Menu]
	navigationService  *services.NavigationService
}

func NewNavigationController(db *gorm.DB) *NavigationController {
	return &NavigationController{
		db:                db,
		menuService:       services.NewGenericService[models.Menu](db),
		navigationService: services.NewNavigationService(db),
	}
}

// Menu handlers
func (c *NavigationController) ListMenus(w http.ResponseWriter, r *http.Request) {
	var records []models.Menu
	if err := c.menuService.List(&records); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list menus")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (c *NavigationController) GetMenu(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	var record models.Menu
	if err := c.menuService.Get(id, &record); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "menu not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve menu")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *NavigationController) CreateMenu(w http.ResponseWriter, r *http.Request) {
	var payload models.Menu
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.menuService.Create(&payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create menu")
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (c *NavigationController) UpdateMenu(w http.ResponseWriter, r *http.Request) {
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
	record, err := c.menuService.Update(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "menu not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update menu")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (c *NavigationController) DeleteMenu(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}
	if err := c.menuService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "menu not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete menu")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Menu Role handlers
func (c *NavigationController) GetMenuRoles(w http.ResponseWriter, r *http.Request) {
	menuID, err := parseIDParam(r, "menuId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid menu id")
		return
	}
	if err := c.navigationService.ValidateMenu(menuID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "menu not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "menu validation failed")
		return
	}
	items, err := c.navigationService.ListMenuRoles(menuID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list menu roles")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (c *NavigationController) AddMenuRole(w http.ResponseWriter, r *http.Request) {
	menuID, err := parseIDParam(r, "menuId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid menu id")
		return
	}
	var payload struct {
		RoleID int `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.RoleID <= 0 {
		writeError(w, http.StatusBadRequest, "role_id is required")
		return
	}
	if err := c.navigationService.ValidateMenu(menuID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "menu not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "menu validation failed")
		return
	}
	// Note: RoleID validation should be done via gRPC call to Auth service
	if err := c.navigationService.AddMenuRole(menuID, payload.RoleID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add menu role")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{
		"menu_id": menuID,
		"role_id": payload.RoleID,
	})
}

func (c *NavigationController) RemoveMenuRole(w http.ResponseWriter, r *http.Request) {
	menuID, err := parseIDParam(r, "menuId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid menu id")
		return
	}
	roleID, err := parseIDParam(r, "roleId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}
	if err := c.navigationService.RemoveMenuRole(menuID, roleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "menu role link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to remove menu role")
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

