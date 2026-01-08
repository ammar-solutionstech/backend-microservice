package services

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"backend/services/inventory/internal/models"
)

// EquipmentService encapsulates relationship operations for equipment.
type EquipmentService struct {
	db *gorm.DB
}

// NewEquipmentService constructs an EquipmentService.
func NewEquipmentService(db *gorm.DB) *EquipmentService {
	return &EquipmentService{db: db}
}

// ListSoftware returns software assignments for the given equipment.
func (s *EquipmentService) ListSoftware(equipmentID int) ([]models.EquipmentSoftware, error) {
	var links []models.EquipmentSoftware
	if err := s.db.Where("equipment_id = ?", equipmentID).Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

// CreateSoftwareLink adds software assignment to equipment.
func (s *EquipmentService) CreateSoftwareLink(link *models.EquipmentSoftware) error {
	return s.db.Create(link).Error
}

// UpdateSoftwareLink updates metadata for a software assignment.
func (s *EquipmentService) UpdateSoftwareLink(equipmentID, softwareID int, updates map[string]interface{}) (*models.EquipmentSoftware, error) {
	var link models.EquipmentSoftware
	if err := s.db.Where("equipment_id = ? AND software_id = ?", equipmentID, softwareID).First(&link).Error; err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return &link, nil
	}
	if err := s.db.Model(&link).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

// DeleteSoftwareLink removes a software assignment.
func (s *EquipmentService) DeleteSoftwareLink(equipmentID, softwareID int) error {
	result := s.db.Where("equipment_id = ? AND software_id = ?", equipmentID, softwareID).Delete(&models.EquipmentSoftware{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListHelpDeskLinks returns help desk tickets linked to equipment.
func (s *EquipmentService) ListHelpDeskLinks(equipmentID int) ([]models.EquipmentHelpDesk, error) {
	var links []models.EquipmentHelpDesk
	if err := s.db.Where("equipment_id = ?", equipmentID).Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

// AddHelpDeskLink links equipment to a help desk ticket.
// Note: HelpDeskID validation should be done via gRPC call to Help Desk service
func (s *EquipmentService) AddHelpDeskLink(equipmentID, helpDeskID int) error {
	link := models.EquipmentHelpDesk{
		EquipmentID: equipmentID,
		HelpDeskID:  helpDeskID,
	}
	return s.db.Where(link).FirstOrCreate(&link).Error
}

// RemoveHelpDeskLink unlinks equipment from a help desk ticket.
func (s *EquipmentService) RemoveHelpDeskLink(equipmentID, helpDeskID int) error {
	result := s.db.Where("equipment_id = ? AND help_desk_id = ?", equipmentID, helpDeskID).
		Delete(&models.EquipmentHelpDesk{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListUserHistory retrieves equipment assignment history.
func (s *EquipmentService) ListUserHistory(equipmentID int) ([]models.EquipmentUserHistory, error) {
	var history []models.EquipmentUserHistory
	if err := s.db.Where("equipment_id = ?", equipmentID).Order("start_date DESC").Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

// AddUserHistory records a new history entry.
// Note: UserID validation should be done via gRPC call to Auth service
func (s *EquipmentService) AddUserHistory(entry *models.EquipmentUserHistory) error {
	// Validate chronological consistency
	if entry.EndDate.Before(entry.StartDate) {
		return errors.New("end_date must not precede start_date")
	}
	return s.db.Create(entry).Error
}

// UpdateUserHistory updates fields on an existing history entry.
func (s *EquipmentService) UpdateUserHistory(equipmentID, userID int, startDate time.Time, updates map[string]interface{}) (*models.EquipmentUserHistory, error) {
	var entry models.EquipmentUserHistory
	err := s.db.
		Where("equipment_id = ? AND user_id = ? AND start_date = ?", equipmentID, userID, startDate).
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return &entry, nil
	}
	// Validate end_date if being updated
	if endDate, ok := updates["end_date"].(time.Time); ok {
		if endDate.Before(entry.StartDate) {
			return nil, errors.New("end_date must not precede start_date")
		}
	}
	if err := s.db.Model(&entry).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// DeleteUserHistory removes a history entry.
func (s *EquipmentService) DeleteUserHistory(equipmentID, userID int, startDate time.Time) error {
	result := s.db.
		Where("equipment_id = ? AND user_id = ? AND start_date = ?", equipmentID, userID, startDate).
		Delete(&models.EquipmentUserHistory{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ValidateEquipmentExists ensures the equipment record exists.
func (s *EquipmentService) ValidateEquipmentExists(equipmentID int) error {
	var exists bool
	err := s.db.Model(&models.Equipment{}).Select("count(1) > 0").Where("id = ?", equipmentID).Scan(&exists).Error
	if err != nil {
		return err
	}
	if !exists {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ValidateSoftwareExists ensures software exists.
func (s *EquipmentService) ValidateSoftwareExists(softwareID int) error {
	var exists bool
	if err := s.db.Model(&models.Software{}).Select("count(1) > 0").Where("id = ?", softwareID).Scan(&exists).Error; err != nil {
		return err
	}
	if !exists {
		return gorm.ErrRecordNotFound
	}
	return nil
}

