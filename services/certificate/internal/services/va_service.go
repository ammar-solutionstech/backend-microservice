package services

import (
	"fmt"
	"time"

	"backend/services/certificate/internal/config"
	"backend/services/certificate/internal/models"

	"gorm.io/gorm"
)

// VAService handles Validation Authority operations
type VAService struct {
	db           *gorm.DB
	stepCAClient *StepCAClient
	config       *config.Config
}

// NewVAService creates a new VA service
func NewVAService(cfg *config.Config, db *gorm.DB, stepCAClient *StepCAClient) *VAService {
	return &VAService{
		db:           db,
		stepCAClient: stepCAClient,
		config:       cfg,
	}
}

// OCSPStatus represents OCSP validation status
type OCSPStatus struct {
	Status       string  `json:"status"` // good, revoked, unknown
	Serial       string  `json:"serial"`
	RevokedAt    *string `json:"revoked_at,omitempty"`
	RevokedBy    *int    `json:"revoked_by,omitempty"`
	RevokeReason *int    `json:"revoke_reason,omitempty"`
}

// GetOCSPStatus gets OCSP status for a certificate
func (s *VAService) GetOCSPStatus(serial string) (*OCSPStatus, error) {
	// First check database cache
	var cert models.Certificate
	if err := s.db.Where("serial_number = ?", serial).First(&cert).Error; err == nil {
		status := "good"
		if cert.Status == "revoked" {
			status = "revoked"
			// Get revocation details
			var revocation models.Revocation
			if err := s.db.Where("serial_number = ?", serial).First(&revocation).Error; err == nil {
				revokedAt := revocation.RevokedAt.Format("2006-01-02T15:04:05Z")
				return &OCSPStatus{
					Status:       status,
					Serial:       serial,
					RevokedAt:    &revokedAt,
					RevokedBy:    revocation.RevokedBy,
					RevokeReason: &revocation.Reason,
				}, nil
			}
		}
		return &OCSPStatus{
			Status: status,
			Serial: serial,
		}, nil
	}

	// If not in database, query step-ca
	ocspResp, err := s.stepCAClient.GetOCSPStatus(serial)
	if err != nil {
		return &OCSPStatus{
			Status: "unknown",
			Serial: serial,
		}, nil
	}

	return &OCSPStatus{
		Status: ocspResp.Status,
		Serial: serial,
	}, nil
}

// GetCRL retrieves the Certificate Revocation List
func (s *VAService) GetCRL() ([]byte, error) {
	crl, err := s.stepCAClient.GetCRL()
	if err != nil {
		return nil, fmt.Errorf("failed to get CRL from step-ca: %v", err)
	}
	return crl, nil
}

// ValidateCertificate checks if a certificate is valid (not revoked, not expired)
func (s *VAService) ValidateCertificate(serial string) (bool, error) {
	var cert models.Certificate
	if err := s.db.Where("serial_number = ?", serial).First(&cert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Certificate not in our database, query step-ca
			ocspResp, err := s.stepCAClient.GetOCSPStatus(serial)
			if err != nil {
				return false, fmt.Errorf("certificate not found: %v", err)
			}
			return ocspResp.Status == "good", nil
		}
		return false, fmt.Errorf("failed to get certificate: %v", err)
	}

	// Check if revoked
	if cert.Status == "revoked" {
		return false, nil
	}

	// Check if expired
	if time.Now().After(cert.ExpiresAt) {
		return false, nil
	}

	return true, nil
}
