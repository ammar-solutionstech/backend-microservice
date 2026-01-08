package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gorm.io/gorm"

	"backend/services/auth/internal/config"
	"backend/services/auth/internal/models"
	"backend/services/auth/internal/services"
)

type RolePermissionController struct {
	cfg          *config.Config
	authService  *services.AuthService
}

func NewRolePermissionController(cfg *config.Config, authService *services.AuthService) *RolePermissionController {
	return &RolePermissionController{
		cfg:         cfg,
		authService: authService,
	}
}

// Role handlers
func (c *RolePermissionController) ListRoles(w http.ResponseWriter, r *http.Request) {
	var roles []models.Role
	if err := c.authService.GetDB().Preload("Permissions").Find(&roles).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list roles")
		return
	}
	writeJSON(w, http.StatusOK, roles)
}

func (c *RolePermissionController) GetRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}
	var role models.Role
	if err := c.authService.GetDB().Preload("Permissions").First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "role not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve role")
		return
	}
	writeJSON(w, http.StatusOK, role)
}

func (c *RolePermissionController) CreateRole(w http.ResponseWriter, r *http.Request) {
	var role models.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.authService.GetDB().Create(&role).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create role")
		return
	}
	writeJSON(w, http.StatusCreated, role)
}

func (c *RolePermissionController) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.authService.GetDB().Model(&models.Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "role not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update role")
		return
	}
	var role models.Role
	c.authService.GetDB().Preload("Permissions").First(&role, id)
	writeJSON(w, http.StatusOK, role)
}

func (c *RolePermissionController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}
	if err := c.authService.GetDB().Delete(&models.Role{}, id).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete role")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *RolePermissionController) UpdateRolePermissions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}
	var permIDs []int
	if err := json.NewDecoder(r.Body).Decode(&permIDs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// Delete existing permissions
	c.authService.GetDB().Where("role_id = ?", id).Delete(&models.RolePermission{})
	// Add new permissions
	for _, permID := range permIDs {
		c.authService.GetDB().Create(&models.RolePermission{RoleID: id, PermissionID: permID})
	}
	var role models.Role
	c.authService.GetDB().Preload("Permissions").First(&role, id)
	writeJSON(w, http.StatusOK, role)
}

func (c *RolePermissionController) AddPermissionToRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}
	var perm struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&perm); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c.authService.GetDB().FirstOrCreate(&models.RolePermission{RoleID: id, PermissionID: perm.ID})
	var role models.Role
	c.authService.GetDB().Preload("Permissions").First(&role, id)
	writeJSON(w, http.StatusOK, role)
}

func (c *RolePermissionController) DeletePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}
	var perm struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&perm); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c.authService.GetDB().Where("role_id = ? AND permission_id = ?", id, perm.ID).Delete(&models.RolePermission{})
	var role models.Role
	c.authService.GetDB().Preload("Permissions").First(&role, id)
	writeJSON(w, http.StatusOK, role)
}

// Permission handlers
func (c *RolePermissionController) ListPermissions(w http.ResponseWriter, r *http.Request) {
	var permissions []models.Permission
	if err := c.authService.GetDB().Find(&permissions).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list permissions")
		return
	}
	writeJSON(w, http.StatusOK, permissions)
}

func (c *RolePermissionController) GetPermission(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid permission id")
		return
	}
	var permission models.Permission
	if err := c.authService.GetDB().First(&permission, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "permission not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve permission")
		return
	}
	writeJSON(w, http.StatusOK, permission)
}

func (c *RolePermissionController) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var permission models.Permission
	if err := json.NewDecoder(r.Body).Decode(&permission); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.authService.GetDB().Create(&permission).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create permission")
		return
	}
	writeJSON(w, http.StatusCreated, permission)
}

func (c *RolePermissionController) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid permission id")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.authService.GetDB().Model(&models.Permission{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "permission not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update permission")
		return
	}
	var permission models.Permission
	c.authService.GetDB().First(&permission, id)
	writeJSON(w, http.StatusOK, permission)
}

func (c *RolePermissionController) DeletePermission(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid permission id")
		return
	}
	if err := c.authService.GetDB().Delete(&models.Permission{}, id).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete permission")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// User Role/Permission handlers
func (c *RolePermissionController) UpdateUserRoles(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var roleIDs []int
	if err := json.NewDecoder(r.Body).Decode(&roleIDs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c.authService.GetDB().Where("user_id = ?", id).Delete(&models.UserRole{})
	for _, roleID := range roleIDs {
		c.authService.GetDB().Create(&models.UserRole{UserID: id, RoleID: roleID})
	}
	var user models.User
	c.authService.GetDB().Preload("Roles").First(&user, id)
	writeJSON(w, http.StatusOK, user)
}

func (c *RolePermissionController) AddRoleToUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var role struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c.authService.GetDB().FirstOrCreate(&models.UserRole{UserID: id, RoleID: role.ID})
	var user models.User
	c.authService.GetDB().Preload("Roles").First(&user, id)
	writeJSON(w, http.StatusOK, user)
}

func (c *RolePermissionController) DeleteRoleFromUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var role struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c.authService.GetDB().Where("user_id = ? AND role_id = ?", id, role.ID).Delete(&models.UserRole{})
	var user models.User
	c.authService.GetDB().Preload("Roles").First(&user, id)
	writeJSON(w, http.StatusOK, user)
}

func (c *RolePermissionController) UpdateUserPermissions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var permIDs []int
	if err := json.NewDecoder(r.Body).Decode(&permIDs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c.authService.GetDB().Where("user_id = ?", id).Delete(&models.UserPermission{})
	for _, permID := range permIDs {
		c.authService.GetDB().Create(&models.UserPermission{UserID: id, PermissionID: permID})
	}
	var user models.User
	c.authService.GetDB().Preload("Permissions").First(&user, id)
	writeJSON(w, http.StatusOK, user)
}

func (c *RolePermissionController) AddPermissionToUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var perm struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&perm); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c.authService.GetDB().FirstOrCreate(&models.UserPermission{UserID: id, PermissionID: perm.ID})
	var user models.User
	c.authService.GetDB().Preload("Permissions").First(&user, id)
	writeJSON(w, http.StatusOK, user)
}

func (c *RolePermissionController) DeletePermissionFromUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var perm struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&perm); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c.authService.GetDB().Where("user_id = ? AND permission_id = ?", id, perm.ID).Delete(&models.UserPermission{})
	var user models.User
	c.authService.GetDB().Preload("Permissions").First(&user, id)
	writeJSON(w, http.StatusOK, user)
}

