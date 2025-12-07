package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"backend/services/container-management/internal/config"
	"backend/services/container-management/internal/models"

	"gorm.io/gorm"
)

// BootstrapService manages bootstrap tokens for container registration
type BootstrapService struct {
	db     *gorm.DB
	config *config.Config
}

// NewBootstrapService creates a new bootstrap service
func NewBootstrapService(cfg *config.Config, db *gorm.DB) *BootstrapService {
	return &BootstrapService{
		db:     db,
		config: cfg,
	}
}

// GenerateBootstrapToken generates a new bootstrap token
func (s *BootstrapService) GenerateBootstrapToken(containerID string, expiresIn time.Duration) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	token := hex.EncodeToString(tokenBytes)

	// Hash the token for storage
	hash := sha256.Sum256([]byte(token + s.config.BootstrapTokenSecret))
	tokenHash := hex.EncodeToString(hash[:])

	// Create bootstrap token record
	bootstrapToken := &models.BootstrapToken{
		TokenHash:   tokenHash,
		ContainerID: &containerID,
		ExpiresAt:   time.Now().Add(expiresIn),
		CreatedAt:   time.Now(),
	}

	if err := s.db.Create(bootstrapToken).Error; err != nil {
		return "", fmt.Errorf("failed to create bootstrap token: %v", err)
	}

	return token, nil
}

// ValidateBootstrapToken validates a bootstrap token and marks it as used
func (s *BootstrapService) ValidateBootstrapToken(token string) (*models.BootstrapToken, error) {
	// Hash the token
	hash := sha256.Sum256([]byte(token + s.config.BootstrapTokenSecret))
	tokenHash := hex.EncodeToString(hash[:])

	// Find the token
	var bootstrapToken models.BootstrapToken
	if err := s.db.Where("token_hash = ?", tokenHash).First(&bootstrapToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid bootstrap token")
		}
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}

	// Check if token is expired
	if time.Now().After(bootstrapToken.ExpiresAt) {
		return nil, fmt.Errorf("bootstrap token has expired")
	}

	// Check if token is already used
	if bootstrapToken.UsedAt != nil {
		return nil, fmt.Errorf("bootstrap token has already been used")
	}

	// Mark token as used
	now := time.Now()
	bootstrapToken.UsedAt = &now
	if err := s.db.Save(&bootstrapToken).Error; err != nil {
		return nil, fmt.Errorf("failed to mark token as used: %v", err)
	}

	return &bootstrapToken, nil
}

// RevokeBootstrapToken revokes an unused bootstrap token
func (s *BootstrapService) RevokeBootstrapToken(tokenID int) error {
	var bootstrapToken models.BootstrapToken
	if err := s.db.First(&bootstrapToken, tokenID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("bootstrap token not found")
		}
		return fmt.Errorf("failed to get token: %v", err)
	}

	if bootstrapToken.UsedAt != nil {
		return fmt.Errorf("token has already been used")
	}

	// Mark as expired by setting expires_at to past
	bootstrapToken.ExpiresAt = time.Now().Add(-1 * time.Hour)
	if err := s.db.Save(&bootstrapToken).Error; err != nil {
		return fmt.Errorf("failed to revoke token: %v", err)
	}

	return nil
}

