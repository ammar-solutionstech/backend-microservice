package services

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"backend/services/certificate/internal/config"
	"backend/services/certificate/internal/models"

	"gorm.io/gorm"
)

// RAService handles Registration Authority operations
type RAService struct {
	db     *gorm.DB
	caSvc  *CAService
	config *config.Config
}

// NewRAService creates a new RA service
func NewRAService(cfg *config.Config, db *gorm.DB, caSvc *CAService) *RAService {
	return &RAService{
		db:     db,
		caSvc:  caSvc,
		config: cfg,
	}
}

// SubmitCSR submits a new CSR request
func (s *RAService) SubmitCSR(csrPEM string, requesterUserID *int, requesterEmail string) (*models.CSRRequest, error) {
	// Validate CSR format
	if err := s.validateCSR(csrPEM); err != nil {
		return nil, fmt.Errorf("invalid CSR format: %v", err)
	}

	csr := &models.CSRRequest{
		CSRPEM:          csrPEM,
		Status:          "pending",
		RequesterUserID: requesterUserID,
		RequesterEmail:  requesterEmail,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.db.Create(csr).Error; err != nil {
		return nil, fmt.Errorf("failed to create CSR request: %v", err)
	}

	return csr, nil
}

// GetCSR retrieves a CSR by ID
func (s *RAService) GetCSR(id int) (*models.CSRRequest, error) {
	var csr models.CSRRequest
	if err := s.db.First(&csr, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("CSR not found")
		}
		return nil, fmt.Errorf("failed to get CSR: %v", err)
	}
	return &csr, nil
}

// ApproveCSR approves a CSR and optionally issues the certificate
func (s *RAService) ApproveCSR(csrID int, approverID int, autoIssue bool) error {
	csr, err := s.GetCSR(csrID)
	if err != nil {
		return err
	}

	if csr.Status != "pending" {
		return fmt.Errorf("CSR is not in pending status")
	}

	// Update CSR status
	csr.Status = "approved"
	csr.ApprovedBy = &approverID
	csr.UpdatedAt = time.Now()

	if err := s.db.Save(csr).Error; err != nil {
		return fmt.Errorf("failed to update CSR: %v", err)
	}

	// If auto-issue is enabled, issue the certificate immediately
	if autoIssue && s.caSvc != nil {
		_, err := s.caSvc.IssueCertificate(csrID)
		if err != nil {
			// Log error but don't fail the approval
			fmt.Printf("Warning: failed to auto-issue certificate: %v\n", err)
		}
	}

	return nil
}

// RejectCSR rejects a CSR
func (s *RAService) RejectCSR(csrID int, reason string) error {
	csr, err := s.GetCSR(csrID)
	if err != nil {
		return err
	}

	if csr.Status != "pending" {
		return fmt.Errorf("CSR is not in pending status")
	}

	csr.Status = "rejected"
	csr.RejectionReason = &reason
	csr.UpdatedAt = time.Now()

	if err := s.db.Save(csr).Error; err != nil {
		return fmt.Errorf("failed to update CSR: %v", err)
	}

	return nil
}

// ListCSRs lists CSR requests with optional filters
func (s *RAService) ListCSRs(status string, requesterUserID *int, limit, offset int) ([]models.CSRRequest, int64, error) {
	var csrs []models.CSRRequest
	var total int64

	query := s.db.Model(&models.CSRRequest{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if requesterUserID != nil {
		query = query.Where("requester_user_id = ?", *requesterUserID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count CSRs: %v", err)
	}

	// Get paginated results
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&csrs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list CSRs: %v", err)
	}

	return csrs, total, nil
}

// validateCSR validates the CSR PEM format
func (s *RAService) validateCSR(csrPEM string) error {
	block, _ := pem.Decode([]byte(csrPEM))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "CERTIFICATE REQUEST" && block.Type != "NEW CERTIFICATE REQUEST" {
		return fmt.Errorf("invalid PEM type: %s", block.Type)
	}

	_, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate request: %v", err)
	}

	return nil
}
