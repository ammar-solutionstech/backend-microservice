package services

import (
	"context"
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
	orchestrator    *ContainerOrchestrator
	dockerClient    *DockerClient
}

// NewClientContainerService creates a new client container service
func NewClientContainerService(cfg *config.Config, db *gorm.DB, certClient *CertificateClient, orgSvc *OrganizationService, orchestrator *ContainerOrchestrator, dockerClient *DockerClient) *ClientContainerService {
	return &ClientContainerService{
		db:              db,
		certClient:      certClient,
		config:          cfg,
		organizationSvc: orgSvc,
		orchestrator:    orchestrator,
		dockerClient:    dockerClient,
	}
}

// RegisterClientContainer registers a new client container for an organization
func (s *ClientContainerService) RegisterClientContainer(orgID int, containerID, name, endpointURL, adminEmail string, adminPhone *string, csrPEM string) (*models.ClientContainer, error) {
	// Verify organization exists
	_, err := s.organizationSvc.GetOrganization(orgID)
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
		OrganizationID:       orgID,
		ContainerID:          containerID,
		Name:                 name,
		Status:               "active",
		CertificateSerial:    &cert.SerialNumber,
		ContainerEndpointURL: endpointURL,
		AdminEmail:           adminEmail,
		AdminPhone:           adminPhone,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := s.db.Create(clientContainer).Error; err != nil {
		return nil, fmt.Errorf("failed to create client container: %v", err)
	}

	// Create Docker container
	ctx := context.Background()
	dockerContainerID, err := s.orchestrator.CreateClientContainer(ctx, orgID, containerID, clientContainer, cert.SerialNumber)
	if err != nil {
		// Rollback database record
		s.db.Delete(clientContainer)
		return nil, fmt.Errorf("failed to create Docker container: %v", err)
	}

	// Get container name
	containerName := s.orchestrator.GenerateContainerName(containerID)
	dockerStatus := "running"

	// Update client container with Docker information
	clientContainer.DockerContainerID = &dockerContainerID
	clientContainer.DockerContainerName = &containerName
	clientContainer.DockerStatus = &dockerStatus
	clientContainer.UpdatedAt = time.Now()

	if err := s.db.Save(clientContainer).Error; err != nil {
		// Try to remove Docker container
		s.dockerClient.RemoveContainer(ctx, dockerContainerID, true)
		return nil, fmt.Errorf("failed to update client container with Docker info: %v", err)
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

// StartContainer starts a Docker container
func (s *ClientContainerService) StartContainer(containerID string) error {
	var clientContainer models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&clientContainer).Error; err != nil {
		return fmt.Errorf("container not found: %v", err)
	}

	if clientContainer.DockerContainerID == nil {
		return fmt.Errorf("container has no Docker container ID")
	}

	ctx := context.Background()
	if err := s.dockerClient.StartContainer(ctx, *clientContainer.DockerContainerID); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	// Update database status
	status := "running"
	clientContainer.DockerStatus = &status
	clientContainer.UpdatedAt = time.Now()
	s.db.Save(&clientContainer)

	return nil
}

// StopContainer stops a Docker container
func (s *ClientContainerService) StopContainer(containerID string) error {
	var clientContainer models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&clientContainer).Error; err != nil {
		return fmt.Errorf("container not found: %v", err)
	}

	if clientContainer.DockerContainerID == nil {
		return fmt.Errorf("container has no Docker container ID")
	}

	ctx := context.Background()
	if err := s.dockerClient.StopContainer(ctx, *clientContainer.DockerContainerID, nil); err != nil {
		return fmt.Errorf("failed to stop container: %v", err)
	}

	// Update database status
	status := "stopped"
	clientContainer.DockerStatus = &status
	clientContainer.UpdatedAt = time.Now()
	s.db.Save(&clientContainer)

	return nil
}

// RestartContainer restarts a Docker container
func (s *ClientContainerService) RestartContainer(containerID string) error {
	var clientContainer models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&clientContainer).Error; err != nil {
		return fmt.Errorf("container not found: %v", err)
	}

	if clientContainer.DockerContainerID == nil {
		return fmt.Errorf("container has no Docker container ID")
	}

	ctx := context.Background()
	if err := s.dockerClient.RestartContainer(ctx, *clientContainer.DockerContainerID, nil); err != nil {
		return fmt.Errorf("failed to restart container: %v", err)
	}

	// Update database status
	status := "running"
	clientContainer.DockerStatus = &status
	clientContainer.UpdatedAt = time.Now()
	s.db.Save(&clientContainer)

	return nil
}

// RemoveContainer removes a Docker container
func (s *ClientContainerService) RemoveContainer(containerID string) error {
	var clientContainer models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&clientContainer).Error; err != nil {
		return fmt.Errorf("container not found: %v", err)
	}

	if clientContainer.DockerContainerID == nil {
		return fmt.Errorf("container has no Docker container ID")
	}

	ctx := context.Background()
	if err := s.dockerClient.RemoveContainer(ctx, *clientContainer.DockerContainerID, true); err != nil {
		return fmt.Errorf("failed to remove container: %v", err)
	}

	// Update database status
	status := "removed"
	clientContainer.DockerStatus = &status
	clientContainer.Status = "inactive"
	clientContainer.UpdatedAt = time.Now()
	s.db.Save(&clientContainer)

	return nil
}

// GetContainerStatus gets the Docker container status
func (s *ClientContainerService) GetContainerStatus(containerID string) (string, error) {
	var clientContainer models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&clientContainer).Error; err != nil {
		return "", fmt.Errorf("container not found: %v", err)
	}

	if clientContainer.DockerContainerID == nil {
		return "", fmt.Errorf("container has no Docker container ID")
	}

	ctx := context.Background()
	status, err := s.dockerClient.GetContainerStatus(ctx, *clientContainer.DockerContainerID)
	if err != nil {
		return "", fmt.Errorf("failed to get container status: %v", err)
	}

	// Update database with current status
	clientContainer.DockerStatus = &status
	clientContainer.UpdatedAt = time.Now()
	s.db.Save(&clientContainer)

	return status, nil
}

// UpdateContainerConfiguration updates the Docker container configuration
func (s *ClientContainerService) UpdateContainerConfiguration(containerID string, updates map[string]interface{}) error {
	var clientContainer models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&clientContainer).Error; err != nil {
		return fmt.Errorf("container not found: %v", err)
	}

	if clientContainer.DockerContainerID == nil {
		return fmt.Errorf("container has no Docker container ID")
	}

	// TODO: Implement container update logic using Docker API
	// This would involve updating container resources, restart policy, etc.

	// Update database record
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&clientContainer).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update container configuration: %v", err)
	}

	return nil
}
