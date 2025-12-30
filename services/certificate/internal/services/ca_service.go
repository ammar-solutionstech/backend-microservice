package services

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"backend/services/certificate/internal/config"
	"backend/services/certificate/internal/models"

	"gorm.io/gorm"
)

// CAService handles Certificate Authority operations
type CAService struct {
	db           *gorm.DB
	stepCAClient *StepCAClient
	config       *config.Config
}

// NewCAService creates a new CA service
func NewCAService(cfg *config.Config, db *gorm.DB, stepCAClient *StepCAClient) *CAService {
	return &CAService{
		db:           db,
		stepCAClient: stepCAClient,
		config:       cfg,
	}
}

// IssueCertificate issues a certificate from an approved CSR
func (s *CAService) IssueCertificate(csrID int) (*models.Certificate, error) {
	// Get the CSR
	var csr models.CSRRequest
	if err := s.db.First(&csr, csrID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("CSR not found")
		}
		return nil, fmt.Errorf("failed to get CSR: %v", err)
	}

	if csr.Status != "approved" {
		return nil, fmt.Errorf("CSR is not approved")
	}

	// Sign the CSR via step-ca
	signResp, err := s.stepCAClient.SignCSR(csr.CSRPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to sign CSR via step-ca: %v", err)
	}

	// Log the response for debugging
	log.Printf("SignCSR response - CertPEM length: %d, CertPEM preview: %s",
		len(signResp.CertPEM),
		/* func() string {
			if len(signResp.CertPEM) > 100 {
				return signResp.CertPEM[:100] + "..."
			}
			return signResp.CertPEM
		}() */signResp.CertPEM)

	// Parse the certificate to extract metadata
	if signResp.CertPEM == "" {
		return nil, fmt.Errorf("received empty certificate PEM from step-ca")
	}

	block, _ := pem.Decode([]byte(signResp.CertPEM))
	if block == nil {
		log.Printf("Failed to decode PEM. CertPEM content: %s", signResp.CertPEM)
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Log certificate subject for debugging
	log.Printf("Issued Certificate Subject - CN: %s, O: %v, OU: %v, C: %v",
		cert.Subject.CommonName,
		cert.Subject.Organization,
		cert.Subject.OrganizationalUnit,
		cert.Subject.Country,
	)

	// Store certificate in database
	certificate := &models.Certificate{
		SerialNumber:   cert.SerialNumber.Text(16),
		CSRID:          &csrID,
		CertificatePEM: signResp.CertPEM,
		IssuedAt:       cert.NotBefore,
		ExpiresAt:      cert.NotAfter,
		Status:         "active",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(certificate).Error; err != nil {
		return nil, fmt.Errorf("failed to store certificate: %v", err)
	}

	return certificate, nil
}

// GetCertificate retrieves a certificate by serial number
func (s *CAService) GetCertificate(serial string) (*models.Certificate, error) {
	var cert models.Certificate
	if err := s.db.Where("serial_number = ?", serial).First(&cert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Try to fetch from step-ca
			return s.fetchFromStepCA(serial)
		}
		return nil, fmt.Errorf("failed to get certificate: %v", err)
	}

	// Check if certificate is expired
	now := time.Now()
	if now.After(cert.ExpiresAt) && cert.Status == "active" {
		cert.Status = "expired"
		s.db.Save(&cert)
	}

	return &cert, nil
}

// fetchFromStepCA fetches a certificate from step-ca and stores it
func (s *CAService) fetchFromStepCA(serial string) (*models.Certificate, error) {
	certResp, err := s.stepCAClient.GetCertificate(serial)
	if err != nil {
		return nil, fmt.Errorf("certificate not found in step-ca: %v", err)
	}

	// Validate certificate format
	block, _ := pem.Decode([]byte(certResp.Certificate))
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Store in database
	certificate := &models.Certificate{
		SerialNumber:   certResp.Serial,
		CertificatePEM: certResp.Certificate,
		IssuedAt:       certResp.IssuedAt,
		ExpiresAt:      certResp.ExpiresAt,
		Status:         certResp.Status,
		StepCACertID:   &certResp.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(certificate).Error; err != nil {
		return nil, fmt.Errorf("failed to store certificate: %v", err)
	}

	return certificate, nil
}

// RevokeCertificate revokes a certificate
func (s *CAService) RevokeCertificate(serial string, reason int, revokedBy *int) error {
	// Get certificate from database
	var cert models.Certificate
	if err := s.db.Where("serial_number = ?", serial).First(&cert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("certificate not found")
		}
		return fmt.Errorf("failed to get certificate: %v", err)
	}

	if cert.Status == "revoked" {
		return fmt.Errorf("certificate is already revoked")
	}

	// Revoke via step-ca
	if err := s.stepCAClient.RevokeCertificate(serial, reason); err != nil {
		return fmt.Errorf("failed to revoke certificate via step-ca: %v", err)
	}

	// Update certificate status
	cert.Status = "revoked"
	cert.UpdatedAt = time.Now()
	if err := s.db.Save(&cert).Error; err != nil {
		return fmt.Errorf("failed to update certificate: %v", err)
	}

	// Create revocation record
	revocation := &models.Revocation{
		CertificateID: cert.ID,
		SerialNumber:  serial,
		RevokedAt:     time.Now(),
		Reason:        reason,
		RevokedBy:     revokedBy,
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(revocation).Error; err != nil {
		return fmt.Errorf("failed to create revocation record: %v", err)
	}

	return nil
}

// RenewCertificate renews a certificate by signing a new CSR
func (s *CAService) RenewCertificate(serial string, newCSRPEM string) (*models.Certificate, error) {
	// Get existing certificate
	oldCert, err := s.GetCertificate(serial)
	if err != nil {
		return nil, err
	}

	// Sign new CSR via step-ca
	signResp, err := s.stepCAClient.RenewCertificate(newCSRPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to renew certificate via step-ca: %v", err)
	}

	// Parse new certificate
	block, _ := pem.Decode([]byte(signResp.CertPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Create new certificate record
	newCert := &models.Certificate{
		SerialNumber:   cert.SerialNumber.Text(16),
		CSRID:          oldCert.CSRID,
		CertificatePEM: signResp.CertPEM,
		IssuedAt:       cert.NotBefore,
		ExpiresAt:      cert.NotAfter,
		Status:         "active",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(newCert).Error; err != nil {
		return nil, fmt.Errorf("failed to store renewed certificate: %v", err)
	}

	return newCert, nil
}

// ListCertificates lists certificates with optional filters
func (s *CAService) ListCertificates(status string, limit, offset int) ([]models.Certificate, int64, error) {
	var certs []models.Certificate
	var total int64

	query := s.db.Model(&models.Certificate{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count certificates: %v", err)
	}

	// Get paginated results
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&certs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list certificates: %v", err)
	}

	return certs, total, nil
}
