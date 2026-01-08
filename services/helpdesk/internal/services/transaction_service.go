package services

import (
	"gorm.io/gorm"

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/models"
)

type TransactionService struct {
	db *gorm.DB
}

func NewTransactionService(cfg *config.Config, db *gorm.DB) *TransactionService {
	return &TransactionService{db: db}
}

func (s *TransactionService) ListTransactions(helpDeskID int) ([]models.HelpDeskTransaction, error) {
	var transactions []models.HelpDeskTransaction
	if err := s.db.Where("help_desk_id = ?", helpDeskID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *TransactionService) CreateTransaction(transaction *models.HelpDeskTransaction) error {
	return s.db.Create(transaction).Error
}

func (s *TransactionService) GetTransaction(id int) (*models.HelpDeskTransaction, error) {
	var transaction models.HelpDeskTransaction
	if err := s.db.First(&transaction, id).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (s *TransactionService) UpdateTransaction(id int, updates map[string]interface{}) (*models.HelpDeskTransaction, error) {
	var transaction models.HelpDeskTransaction
	if err := s.db.First(&transaction, id).Error; err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return &transaction, nil
	}
	if err := s.db.Model(&transaction).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (s *TransactionService) DeleteTransaction(id int) error {
	result := s.db.Delete(&models.HelpDeskTransaction{}, id)
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *TransactionService) ListTransactionUsers(transactionID int) ([]models.UserHelpDeskTransaction, error) {
	var links []models.UserHelpDeskTransaction
	if err := s.db.Where("transaction_id = ?", transactionID).Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

func (s *TransactionService) AddTransactionUser(transactionID, userID int) error {
	link := models.UserHelpDeskTransaction{
		TransactionID: transactionID,
		UserID:        userID,
	}
	return s.db.Where(link).FirstOrCreate(&link).Error
}

func (s *TransactionService) RemoveTransactionUser(transactionID, userID int) error {
	result := s.db.Where("transaction_id = ? AND user_id = ?", transactionID, userID).
		Delete(&models.UserHelpDeskTransaction{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

