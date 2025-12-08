package services

import (
	"fmt"
	"time"

	"backend/services/client-container/internal/config"
	"backend/services/client-container/internal/models"

	"gorm.io/gorm"
)

// ContainerService manages client container lifecycle
type ContainerService struct {
	db           *gorm.DB
	certClient   *CertificateClient
	config       *config.Config
}

// NewContainerService creates a new container service
func NewContainerService(cfg *config.Config, db *gorm.DB, certClient *CertificateClient) *ContainerService {
	return &ContainerService{
		db:         db,
		certClient: certClient,
		config:     cfg,
	}
}

// GetContainer retrieves a container by ID
func (s *ContainerService) GetContainer(containerID string) (*models.ClientContainer, error) {
	var container models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container not found")
		}
		return nil, fmt.Errorf("failed to get container: %v", err)
	}

	return &container, nil
}

// GetContainerByOrg retrieves a container by organization ID
func (s *ContainerService) GetContainerByOrg(orgID int) (*models.ClientContainer, error) {
	var container models.ClientContainer
	if err := s.db.Where("organization_id = ?", orgID).First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container not found for organization")
		}
		return nil, fmt.Errorf("failed to get container: %v", err)
	}

	return &container, nil
}

// UpdateContainerStatus updates container status
func (s *ContainerService) UpdateContainerStatus(containerID string, status string) error {
	if err := s.db.Model(&models.ClientContainer{}).
		Where("container_id = ?", containerID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update container status: %v", err)
	}

	return nil
}

// GetContainerCertificate loads the container's certificate
func (s *ContainerService) GetContainerCertificate(containerID string) (*models.ClientContainer, error) {
	container, err := s.GetContainer(containerID)
	if err != nil {
		return nil, err
	}

	if container.CertificateSerial == nil {
		return nil, fmt.Errorf("container certificate not found")
	}

	return container, nil
}

// RenewContainerCertificate renews the container's certificate
func (s *ContainerService) RenewContainerCertificate(containerID, csrPEM string) error {
	container, err := s.GetContainer(containerID)
	if err != nil {
		return err
	}

	// Request new certificate
	certResp, err := s.certClient.RequestContainerCertificate(container.OrganizationID, csrPEM)
	if err != nil {
		return fmt.Errorf("failed to request certificate: %v", err)
	}

	// Update container with new certificate serial
	container.CertificateSerial = &certResp.SerialNumber
	container.UpdatedAt = time.Now()
	if err := s.db.Save(container).Error; err != nil {
		return fmt.Errorf("failed to update container: %v", err)
	}

	return nil
}

