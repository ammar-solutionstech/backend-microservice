package services

import (
	"fmt"
	"time"

	"backend/services/container-management/internal/config"
	"backend/services/container-management/internal/models"

	"gorm.io/gorm"
)

// ContainerService manages container registrations and certificate requests
type ContainerService struct {
	db           *gorm.DB
	certClient   *CertificateClient
	csrValidator *CSRValidator
	config       *config.Config
}

// NewContainerService creates a new container service
func NewContainerService(cfg *config.Config, db *gorm.DB, certClient *CertificateClient, csrValidator *CSRValidator) *ContainerService {
	return &ContainerService{
		db:           db,
		certClient:   certClient,
		csrValidator: csrValidator,
		config:       cfg,
	}
}

// RegisterContainer registers a new container
func (s *ContainerService) RegisterContainer(containerID, name string, metadata map[string]interface{}) (*models.Container, error) {
	// Check if container already exists
	var existing models.Container
	if err := s.db.Where("container_id = ?", containerID).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("container with ID '%s' already exists", containerID)
	}

	container := &models.Container{
		ContainerID:  containerID,
		Name:         name,
		Status:       "active",
		RegisteredAt: time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if metadata != nil {
		container.Metadata = metadata
	}

	if err := s.db.Create(container).Error; err != nil {
		return nil, fmt.Errorf("failed to register container: %v", err)
	}

	return container, nil
}

// RequestCertificate requests a certificate for a container
func (s *ContainerService) RequestCertificate(containerID, csrPEM string) (*models.CertificateRequest, error) {
	// Validate container exists
	var container models.Container
	if err := s.db.Where("container_id = ?", containerID).First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container not found")
		}
		return nil, fmt.Errorf("failed to get container: %v", err)
	}

	// Validate CSR
	if err := s.csrValidator.Validate(csrPEM); err != nil {
		return nil, fmt.Errorf("CSR validation failed: %v", err)
	}

	// Submit CSR to certificate service
	csrResp, err := s.certClient.SubmitCSR(csrPEM, fmt.Sprintf("container:%s", containerID))
	if err != nil {
		return nil, fmt.Errorf("failed to submit CSR to certificate service: %v", err)
	}

	// Create certificate request record
	certReq := &models.CertificateRequest{
		ContainerID: containerID,
		CSRID:       &csrResp.ID,
		RequestType: "container",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(certReq).Error; err != nil {
		return nil, fmt.Errorf("failed to create certificate request: %v", err)
	}

	return certReq, nil
}

// RequestApplicationCertificate requests a certificate for an application under a container
func (s *ContainerService) RequestApplicationCertificate(containerID, appName, csrPEM string) (*models.CertificateRequest, error) {
	// Validate container exists
	var container models.Container
	if err := s.db.Where("container_id = ?", containerID).First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container not found")
		}
		return nil, fmt.Errorf("failed to get container: %v", err)
	}

	// Validate CSR
	if err := s.csrValidator.Validate(csrPEM); err != nil {
		return nil, fmt.Errorf("CSR validation failed: %v", err)
	}

	// Submit CSR to certificate service
	csrResp, err := s.certClient.SubmitCSR(csrPEM, fmt.Sprintf("container:%s:app:%s", containerID, appName))
	if err != nil {
		return nil, fmt.Errorf("failed to submit CSR to certificate service: %v", err)
	}

	// Create certificate request record
	certReq := &models.CertificateRequest{
		ContainerID: containerID,
		CSRID:       &csrResp.ID,
		RequestType: "application",
		AppName:     &appName,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(certReq).Error; err != nil {
		return nil, fmt.Errorf("failed to create certificate request: %v", err)
	}

	return certReq, nil
}

// GetContainerCertificates lists all certificates for a container
func (s *ContainerService) GetContainerCertificates(containerID string) ([]models.CertificateRequest, error) {
	var certReqs []models.CertificateRequest
	if err := s.db.Where("container_id = ?", containerID).Find(&certReqs).Error; err != nil {
		return nil, fmt.Errorf("failed to get container certificates: %v", err)
	}

	return certReqs, nil
}

// RevokeContainerCertificate revokes a container's certificate
func (s *ContainerService) RevokeContainerCertificate(containerID, serial string) error {
	// Verify the certificate belongs to the container
	var certReq models.CertificateRequest
	if err := s.db.Where("container_id = ? AND status = ?", containerID, "issued").First(&certReq).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("certificate not found for container")
		}
		return fmt.Errorf("failed to get certificate request: %v", err)
	}

	// Revoke via certificate service
	if err := s.certClient.RevokeCertificate(serial, 0); err != nil {
		return fmt.Errorf("failed to revoke certificate: %v", err)
	}

	// Update certificate request status
	certReq.Status = "revoked"
	certReq.UpdatedAt = time.Now()
	if err := s.db.Save(&certReq).Error; err != nil {
		return fmt.Errorf("failed to update certificate request: %v", err)
	}

	return nil
}

// GetContainer retrieves a container by ID
func (s *ContainerService) GetContainer(containerID string) (*models.Container, error) {
	var container models.Container
	if err := s.db.Where("container_id = ?", containerID).First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container not found")
		}
		return nil, fmt.Errorf("failed to get container: %v", err)
	}

	return &container, nil
}

// UpdateContainerCertificateSerial updates the certificate serial for a container
func (s *ContainerService) UpdateContainerCertificateSerial(containerID, serial string) error {
	var container models.Container
	if err := s.db.Where("container_id = ?", containerID).First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("container not found")
		}
		return fmt.Errorf("failed to get container: %v", err)
	}

	container.CertificateSerial = &serial
	container.UpdatedAt = time.Now()
	if err := s.db.Save(&container).Error; err != nil {
		return fmt.Errorf("failed to update container certificate serial: %v", err)
	}

	return nil
}
