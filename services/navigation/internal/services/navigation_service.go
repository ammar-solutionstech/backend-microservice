package services

import (
	"gorm.io/gorm"

	"backend/services/navigation/internal/models"
)

// NavigationService manages menu-role relationships.
type NavigationService struct {
	db *gorm.DB
}

// NewNavigationService constructs a NavigationService.
func NewNavigationService(db *gorm.DB) *NavigationService {
	return &NavigationService{db: db}
}

func (s *NavigationService) ListMenuRoles(menuID int) ([]models.MenuRole, error) {
	var items []models.MenuRole
	if err := s.db.Where("menu_id = ?", menuID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *NavigationService) AddMenuRole(menuID, roleID int) error {
	link := models.MenuRole{
		MenuID: menuID,
		RoleID: roleID,
	}
	return s.db.Where(link).FirstOrCreate(&link).Error
}

func (s *NavigationService) RemoveMenuRole(menuID, roleID int) error {
	result := s.db.Where("menu_id = ? AND role_id = ?", menuID, roleID).Delete(&models.MenuRole{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *NavigationService) ValidateMenu(menuID int) error {
	var exists bool
	if err := s.db.Model(&models.Menu{}).Select("count(1) > 0").Where("id = ?", menuID).Scan(&exists).Error; err != nil {
		return err
	}
	if !exists {
		return gorm.ErrRecordNotFound
	}
	return nil
}

