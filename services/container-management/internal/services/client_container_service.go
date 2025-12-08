package services

import (
	"fmt"
	"time"

	"backend/services/container-management/internal/config"
	"backend/services/container-management/internal/models"

	"gorm.io/gorm"
)

// ClientContainerService manages client container registrations
type ClientContainerService struct {
	db              *gorm.DB
	certClient      *CertificateClient
	config          *config.Config
	organizationSvc *OrganizationService
}

// NewClientContainerService creates a new client container service
func NewClientContainerService(cfg *config.Config, db *gorm.DB, certClient *CertificateClient, orgSvc *OrganizationService) *ClientContainerService {
	return &ClientContainerService{
		db:              db,
		certClient:      certClient,
		config:          cfg,
		organizationSvc: orgSvc,
	}
}

// RegisterClientContainer registers a new client container for an organization
func (s *ClientContainerService) RegisterClientContainer(orgID int, containerID, name, endpointURL, adminEmail string, adminPhone *string, csrPEM string) (*models.ClientContainer, error) {
	// Verify organization exists
	org, err := s.organizationSvc.GetOrganization(orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %v", err)
	}

	// Check if container ID already exists
	var existing models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("client container with ID '%s' already exists", containerID)
	}

	// Request certificate from certificate service via container-management
	// Note: This goes through the certificate service, not directly
	certResp, err := s.certClient.SubmitCSR(csrPEM, fmt.Sprintf("container:%s", containerID))
	if err != nil {
		return nil, fmt.Errorf("failed to request container certificate: %v", err)
	}

	// Approve and issue certificate
	approverID := 0 // System approver
	if err := s.certClient.ApproveCSR(certResp.ID, approverID, true); err != nil {
		return nil, fmt.Errorf("failed to approve container certificate: %v", err)
	}

	// Get issued certificate
	cert, err := s.certClient.IssueCertificate(certResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to issue container certificate: %v", err)
	}

	// Create client container record
	clientContainer := &models.ClientContainer{
		OrganizationID:      orgID,
		ContainerID:         containerID,
		Name:                name,
		Status:              "active",
		CertificateSerial:   &cert.SerialNumber,
		ContainerEndpointURL: endpointURL,
		AdminEmail:          adminEmail,
		AdminPhone:          adminPhone,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := s.db.Create(clientContainer).Error; err != nil {
		return nil, fmt.Errorf("failed to create client container: %v", err)
	}

	return clientContainer, nil
}

// GetClientContainer retrieves a client container by ID
func (s *ClientContainerService) GetClientContainer(containerID string) (*models.ClientContainer, error) {
	var container models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("client container not found")
		}
		return nil, fmt.Errorf("failed to get client container: %v", err)
	}

	return &container, nil
}

// GetClientContainerByOrg retrieves client containers for an organization
func (s *ClientContainerService) GetClientContainerByOrg(orgID int) (*models.ClientContainer, error) {
	var container models.ClientContainer
	if err := s.db.Where("organization_id = ?", orgID).First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("client container not found for organization")
		}
		return nil, fmt.Errorf("failed to get client container: %v", err)
	}

	return &container, nil
}

// UpdateClientContainer updates a client container
func (s *ClientContainerService) UpdateClientContainer(containerID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&models.ClientContainer{}).Where("container_id = ?", containerID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update client container: %v", err)
	}

	return nil
}

