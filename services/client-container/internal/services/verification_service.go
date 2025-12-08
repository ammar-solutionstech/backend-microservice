package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"backend/services/client-container/internal/config"
	"backend/services/client-container/internal/models"

	"gorm.io/gorm"
)

// VerificationService manages verification codes for agent certificate issuance
type VerificationService struct {
	db     *gorm.DB
	config *config.Config
}

// NewVerificationService creates a new verification service
func NewVerificationService(cfg *config.Config, db *gorm.DB) *VerificationService {
	return &VerificationService{
		db:     db,
		config: cfg,
	}
}

// GenerateVerificationCode generates a new verification code for a device
func (s *VerificationService) GenerateVerificationCode(deviceID int) (string, error) {
	// Generate random code
	code := s.generateRandomCode(s.config.VerificationCodeLength)

	// Hash the code for storage
	hash := sha256.Sum256([]byte(code))
	codeHash := hex.EncodeToString(hash[:])

	// Create verification code record
	verificationCode := &models.VerificationCode{
		DeviceID:  deviceID,
		CodeHash:  codeHash,
		ExpiresAt: time.Now().Add(s.config.VerificationCodeExpiry),
		CreatedAt: time.Now(),
	}

	// Invalidate any existing unused codes for this device
	s.db.Model(&models.VerificationCode{}).
		Where("device_id = ? AND used_at IS NULL", deviceID).
		Update("used_at", time.Now())

	// Create new code
	if err := s.db.Create(verificationCode).Error; err != nil {
		return "", fmt.Errorf("failed to create verification code: %v", err)
	}

	return code, nil
}

// ValidateVerificationCode validates a verification code and marks it as used
func (s *VerificationService) ValidateVerificationCode(deviceID int, code string) (bool, error) {
	// Hash the provided code
	hash := sha256.Sum256([]byte(code))
	codeHash := hex.EncodeToString(hash[:])

	// Find the verification code
	var verificationCode models.VerificationCode
	if err := s.db.Where("device_id = ? AND code_hash = ?", deviceID, codeHash).First(&verificationCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to validate code: %v", err)
	}

	// Check if code is expired
	if time.Now().After(verificationCode.ExpiresAt) {
		return false, fmt.Errorf("verification code has expired")
	}

	// Check if code is already used
	if verificationCode.UsedAt != nil {
		return false, fmt.Errorf("verification code has already been used")
	}

	// Mark code as used
	now := time.Now()
	verificationCode.UsedAt = &now
	if err := s.db.Save(&verificationCode).Error; err != nil {
		return false, fmt.Errorf("failed to mark code as used: %v", err)
	}

	return true, nil
}

// GetVerificationCode retrieves the active verification code for a device
func (s *VerificationService) GetVerificationCode(deviceID int) (*models.VerificationCode, error) {
	var verificationCode models.VerificationCode
	if err := s.db.Where("device_id = ? AND used_at IS NULL AND expires_at > ?", deviceID, time.Now()).
		Order("created_at DESC").
		First(&verificationCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active verification code found")
		}
		return nil, fmt.Errorf("failed to get verification code: %v", err)
	}

	return &verificationCode, nil
}

// generateRandomCode generates a random numeric code of specified length
func (s *VerificationService) generateRandomCode(length int) string {
	max := big.NewInt(0)
	max.Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	max.Sub(max, big.NewInt(1))

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// Fallback to simple random if crypto/rand fails
		code := ""
		for i := 0; i < length; i++ {
			digit, _ := rand.Int(rand.Reader, big.NewInt(10))
			code += digit.String()
		}
		return code
	}

	code := n.String()
	// Pad with zeros if needed
	for len(code) < length {
		code = "0" + code
	}

	return code
}

