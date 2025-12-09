package services

import (
	"fmt"
	"sync"
)

// ServiceRegistry manages the registration and lifecycle of services/plugins
type ServiceRegistry struct {
	services map[string]Service
	mu       sync.RWMutex
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]Service),
	}
}

// RegisterService registers a service with the registry
func (r *ServiceRegistry) RegisterService(name string, service Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service '%s' is already registered", name)
	}

	// Validate service name matches
	if service.Name() != name {
		return fmt.Errorf("service name mismatch: expected '%s', got '%s'", name, service.Name())
	}

	r.services[name] = service
	return nil
}

// UnregisterService unregisters a service from the registry
func (r *ServiceRegistry) UnregisterService(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service, exists := r.services[name]
	if !exists {
		return fmt.Errorf("service '%s' is not registered", name)
	}

	// Stop the service before unregistering
	if service.Status() == "active" || service.Status() == "starting" {
		if err := service.Stop(); err != nil {
			return fmt.Errorf("failed to stop service '%s' before unregistering: %v", name, err)
		}
	}

	delete(r.services, name)
	return nil
}

// GetService retrieves a service by name
func (r *ServiceRegistry) GetService(name string) (Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("service '%s' is not registered", name)
	}

	return service, nil
}

// ListServices returns a list of all registered service names
func (r *ServiceRegistry) ListServices() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}

	return names
}

// StartService starts a service by name
func (r *ServiceRegistry) StartService(name string) error {
	service, err := r.GetService(name)
	if err != nil {
		return err
	}

	if service.Status() == "active" {
		return fmt.Errorf("service '%s' is already active", name)
	}

	return service.Start()
}

// StopService stops a service by name
func (r *ServiceRegistry) StopService(name string) error {
	service, err := r.GetService(name)
	if err != nil {
		return err
	}

	if service.Status() == "inactive" || service.Status() == "stopping" {
		return fmt.Errorf("service '%s' is not active", name)
	}

	return service.Stop()
}

// GetServiceStatus returns the status of a service
func (r *ServiceRegistry) GetServiceStatus(name string) (string, error) {
	service, err := r.GetService(name)
	if err != nil {
		return "", err
	}

	return service.Status(), nil
}

// ConfigureService configures a service with the provided configuration
func (r *ServiceRegistry) ConfigureService(name string, config map[string]interface{}) error {
	service, err := r.GetService(name)
	if err != nil {
		return err
	}

	return service.Configure(config)
}

// GetAllServices returns all registered services
func (r *ServiceRegistry) GetAllServices() map[string]Service {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]Service)
	for name, service := range r.services {
		result[name] = service
	}

	return result
}
