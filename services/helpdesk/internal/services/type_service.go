package services

import (
	"gorm.io/gorm"

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/models"
)

type HelpDeskTypeService struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewHelpDeskTypeService(cfg *config.Config, db *gorm.DB) *HelpDeskTypeService {
	return &HelpDeskTypeService{cfg: cfg, db: db}
}

func (s *HelpDeskTypeService) CreateType(name, description string, teamID *int, isActive *bool) (*models.HelpDeskType, error) {
	helpDeskType := &models.HelpDeskType{
		Name:        name,
		Description: description,
		TeamID:      teamID,
		IsActive:    isActive,
	}
	if err := s.db.Create(helpDeskType).Error; err != nil {
		return nil, err
	}
	return helpDeskType, nil
}

func (s *HelpDeskTypeService) GetType(typeID int) (*models.HelpDeskType, error) {
	var helpDeskType models.HelpDeskType
	if err := s.db.Where("id = ?", typeID).First(&helpDeskType).Error; err != nil {
		return nil, err
	}
	return &helpDeskType, nil
}

func (s *HelpDeskTypeService) ListTypes(activeOnly bool) ([]models.HelpDeskType, error) {
	var types []models.HelpDeskType
	query := s.db
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

func (s *HelpDeskTypeService) UpdateType(typeID int, name, description *string, isActive *bool) (*models.HelpDeskType, error) {
	var helpDeskType models.HelpDeskType
	if err := s.db.Where("id = ?", typeID).First(&helpDeskType).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if description != nil {
		updates["description"] = *description
	}
	if isActive != nil {
		updates["is_active"] = *isActive
	}

	if err := s.db.Model(&helpDeskType).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &helpDeskType, nil
}

func (s *HelpDeskTypeService) DeleteType(typeID int) error {
	return s.db.Delete(&models.HelpDeskType{}, typeID).Error
}
