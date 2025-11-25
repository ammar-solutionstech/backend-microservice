package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"

	"backend/services/auth/internal/models"
)

type OTPService struct {
	db *gorm.DB
}

func NewOTPService(db *gorm.DB) *OTPService {
	return &OTPService{db: db}
}

// GenerateOTP generates a 6-digit OTP code
func (s *OTPService) GenerateOTP() (string, error) {
	max := big.NewInt(1000000) // 6 digits
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// CreateOTPVerification creates an OTP verification record
func (s *OTPService) CreateOTPVerification(userID int, otpType string) (string, error) {
	// Generate OTP code
	code, err := s.GenerateOTP()
	if err != nil {
		return "", err
	}

	// Invalidate any existing unverified OTPs of the same type
	s.db.Model(&models.OTPVerification{}).
		Where("user_id = ? AND type = ? AND verified_at IS NULL", userID, otpType).
		Update("verified_at", time.Now().UTC())

	// Create new OTP verification
	otp := models.OTPVerification{
		UserID:    userID,
		Type:      otpType, // "email" or "phone"
		Code:      code,
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute), // OTP expires in 10 minutes
	}

	if err := s.db.Create(&otp).Error; err != nil {
		return "", err
	}

	return code, nil
}

// VerifyOTP verifies an OTP code
func (s *OTPService) VerifyOTP(userID int, otpType string, code string) error {
	var otp models.OTPVerification
	if err := s.db.Where("user_id = ? AND type = ? AND code = ? AND expires_at > ? AND verified_at IS NULL",
		userID, otpType, code, time.Now().UTC()).First(&otp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired OTP code")
		}
		return err
	}

	// Mark as verified
	now := time.Now().UTC()
	if err := s.db.Model(&otp).Update("verified_at", now).Error; err != nil {
		return err
	}

	return nil
}

// ResendOTP creates a new OTP for the user
func (s *OTPService) ResendOTP(userID int, otpType string) (string, error) {
	return s.CreateOTPVerification(userID, otpType)
}

// CleanupExpiredOTPs removes expired OTPs from database
func (s *OTPService) CleanupExpiredOTPs() {
	now := time.Now().UTC()
	s.db.Where("expires_at < ?", now).Delete(&models.OTPVerification{})
}

