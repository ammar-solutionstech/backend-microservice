package services

import (
	"fmt"
	"time"

	"backend/services/container-management/internal/config"
	"backend/services/container-management/internal/models"

	"gorm.io/gorm"
)

// OrganizationService manages organizations
type OrganizationService struct {
	db     *gorm.DB
	config *config.Config
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(cfg *config.Config, db *gorm.DB) *OrganizationService {
	return &OrganizationService{
		db:     db,
		config: cfg,
	}
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(name, domain, adminEmail string, adminPhone *string) (*models.Organization, error) {
	// Check if domain already exists
	var existing models.Organization
	if err := s.db.Where("domain = ?", domain).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("organization with domain '%s' already exists", domain)
	}

	org := &models.Organization{
		Name:       name,
		Domain:     domain,
		AdminEmail: adminEmail,
		AdminPhone: adminPhone,
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.db.Create(org).Error; err != nil {
		return nil, fmt.Errorf("failed to create organization: %v", err)
	}

	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationService) GetOrganization(id int) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.First(&org, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %v", err)
	}

	return &org, nil
}

// GetOrganizationByDomain retrieves an organization by domain
func (s *OrganizationService) GetOrganizationByDomain(domain string) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.Where("domain = ?", domain).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %v", err)
	}

	return &org, nil
}

// UpdateOrganization updates an organization
func (s *OrganizationService) UpdateOrganization(id int, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&models.Organization{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update organization: %v", err)
	}

	return nil
}

// ListOrganizations lists all organizations
func (s *OrganizationService) ListOrganizations(status string) ([]models.Organization, error) {
	var orgs []models.Organization
	query := s.db.Model(&models.Organization{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("failed to list organizations: %v", err)
	}

	return orgs, nil
}

