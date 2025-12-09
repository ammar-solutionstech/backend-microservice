package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"backend/services/container-management/internal/config"
	"backend/services/container-management/internal/models"

	"gorm.io/gorm"
)

// ServiceManager manages services (plugins) within client-container instances
type ServiceManager struct {
	db           *gorm.DB
	config       *config.Config
	dockerClient *DockerClient
}

// NewServiceManager creates a new service manager
func NewServiceManager(cfg *config.Config, db *gorm.DB, dockerClient *DockerClient) *ServiceManager {
	return &ServiceManager{
		db:           db,
		config:       cfg,
		dockerClient: dockerClient,
	}
}

// AddServiceToContainer adds a service/plugin to a client-container
func (s *ServiceManager) AddServiceToContainer(containerID string, serviceConfig *models.ContainerService) error {
	// Get container to verify it exists
	var clientContainer models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&clientContainer).Error; err != nil {
		return fmt.Errorf("container not found: %v", err)
	}

	// Check if service already exists
	var existing models.ContainerService
	if err := s.db.Where("container_id = ? AND service_name = ?", containerID, serviceConfig.ServiceName).First(&existing).Error; err == nil {
		return fmt.Errorf("service '%s' already exists in container", serviceConfig.ServiceName)
	}

	// Create service record
	serviceConfig.ContainerID = containerID
	serviceConfig.Status = "inactive"
	serviceConfig.Enabled = true
	serviceConfig.CreatedAt = time.Now()
	serviceConfig.UpdatedAt = time.Now()

	if err := s.db.Create(serviceConfig).Error; err != nil {
		return fmt.Errorf("failed to create service record: %v", err)
	}

	// Call client-container internal API to add service
	if err := s.callContainerServiceAPI(containerID, "POST", "/api/v1/internal/services", serviceConfig); err != nil {
		// Rollback database record
		s.db.Delete(serviceConfig)
		return fmt.Errorf("failed to add service to container: %v", err)
	}

	// Update service status
	serviceConfig.Status = "active"
	s.db.Save(serviceConfig)

	return nil
}

// RemoveServiceFromContainer removes a service/plugin from a client-container
func (s *ServiceManager) RemoveServiceFromContainer(containerID, serviceName string) error {
	// Get service record
	var service models.ContainerService
	if err := s.db.Where("container_id = ? AND service_name = ?", containerID, serviceName).First(&service).Error; err != nil {
		return fmt.Errorf("service not found: %v", err)
	}

	// Call client-container internal API to remove service
	if err := s.callContainerServiceAPI(containerID, "DELETE", fmt.Sprintf("/api/v1/internal/services/%s", serviceName), nil); err != nil {
		return fmt.Errorf("failed to remove service from container: %v", err)
	}

	// Delete service record
	if err := s.db.Delete(&service).Error; err != nil {
		return fmt.Errorf("failed to delete service record: %v", err)
	}

	return nil
}

// ListContainerServices lists all services in a container
func (s *ServiceManager) ListContainerServices(containerID string) ([]models.ContainerService, error) {
	var services []models.ContainerService
	if err := s.db.Where("container_id = ?", containerID).Find(&services).Error; err != nil {
		return nil, fmt.Errorf("failed to list services: %v", err)
	}

	return services, nil
}

// UpdateServiceInContainer updates a service configuration
func (s *ServiceManager) UpdateServiceInContainer(containerID, serviceName string, updates map[string]interface{}) error {
	// Get service record
	var service models.ContainerService
	if err := s.db.Where("container_id = ? AND service_name = ?", containerID, serviceName).First(&service).Error; err != nil {
		return fmt.Errorf("service not found: %v", err)
	}

	// Update database record
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&service).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update service record: %v", err)
	}

	// Call client-container internal API to update service
	if err := s.callContainerServiceAPI(containerID, "PUT", fmt.Sprintf("/api/v1/internal/services/%s", serviceName), updates); err != nil {
		return fmt.Errorf("failed to update service in container: %v", err)
	}

	return nil
}

// EnableService enables a service in a container
func (s *ServiceManager) EnableService(containerID, serviceName string) error {
	// Get service record
	var service models.ContainerService
	if err := s.db.Where("container_id = ? AND service_name = ?", containerID, serviceName).First(&service).Error; err != nil {
		return fmt.Errorf("service not found: %v", err)
	}

	service.Enabled = true
	service.UpdatedAt = time.Now()
	if err := s.db.Save(&service).Error; err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}

	// Call client-container internal API to start service
	if err := s.callContainerServiceAPI(containerID, "POST", fmt.Sprintf("/api/v1/internal/services/%s/start", serviceName), nil); err != nil {
		return fmt.Errorf("failed to start service in container: %v", err)
	}

	service.Status = "active"
	s.db.Save(&service)

	return nil
}

// DisableService disables a service in a container
func (s *ServiceManager) DisableService(containerID, serviceName string) error {
	// Get service record
	var service models.ContainerService
	if err := s.db.Where("container_id = ? AND service_name = ?", containerID, serviceName).First(&service).Error; err != nil {
		return fmt.Errorf("service not found: %v", err)
	}

	// Call client-container internal API to stop service
	if err := s.callContainerServiceAPI(containerID, "POST", fmt.Sprintf("/api/v1/internal/services/%s/stop", serviceName), nil); err != nil {
		return fmt.Errorf("failed to stop service in container: %v", err)
	}

	service.Enabled = false
	service.Status = "inactive"
	service.UpdatedAt = time.Now()
	if err := s.db.Save(&service).Error; err != nil {
		return fmt.Errorf("failed to disable service: %v", err)
	}

	return nil
}

// callContainerServiceAPI calls the client-container internal API
func (s *ServiceManager) callContainerServiceAPI(containerID, method, path string, body interface{}) error {
	// Get container to get endpoint URL
	var clientContainer models.ClientContainer
	if err := s.db.Where("container_id = ?", containerID).First(&clientContainer).Error; err != nil {
		return fmt.Errorf("container not found: %v", err)
	}

	// Build URL
	url := fmt.Sprintf("%s%s", clientContainer.ContainerEndpointURL, path)

	// Create request
	var req *http.Request
	var err error

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %v", err)
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}
	}

	// TODO: Add mTLS client configuration for secure communication
	// For now, using basic HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call container API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("container API returned error: %d", resp.StatusCode)
	}

	return nil
}
