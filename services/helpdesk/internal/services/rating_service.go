package services

import (
	"gorm.io/gorm"

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/models"
)

type RatingService struct {
	db *gorm.DB
}

func NewRatingService(cfg *config.Config, db *gorm.DB) *RatingService {
	return &RatingService{db: db}
}

func (s *RatingService) ListRatings(helpDeskID int) ([]models.HelpDeskRating, error) {
	var ratings []models.HelpDeskRating
	if err := s.db.Where("help_desk_id = ?", helpDeskID).Find(&ratings).Error; err != nil {
		return nil, err
	}
	return ratings, nil
}

func (s *RatingService) CreateRating(rating *models.HelpDeskRating) error {
	return s.db.Create(rating).Error
}

func (s *RatingService) GetRating(id int) (*models.HelpDeskRating, error) {
	var rating models.HelpDeskRating
	if err := s.db.First(&rating, id).Error; err != nil {
		return nil, err
	}
	return &rating, nil
}

func (s *RatingService) UpdateRating(id int, updates map[string]interface{}) (*models.HelpDeskRating, error) {
	var rating models.HelpDeskRating
	if err := s.db.First(&rating, id).Error; err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return &rating, nil
	}
	if err := s.db.Model(&rating).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &rating, nil
}

func (s *RatingService) DeleteRating(id int) error {
	result := s.db.Delete(&models.HelpDeskRating{}, id)
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

