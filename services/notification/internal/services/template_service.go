package services

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"backend/services/notification/internal/models"
)

type TemplateService struct {
	db *gorm.DB
}

func NewTemplateService(db *gorm.DB) *TemplateService {
	return &TemplateService{db: db}
}

// GetEmailTemplate retrieves an email template by name
func (s *TemplateService) GetEmailTemplate(name string) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	if err := s.db.Where("name = ?", name).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

// GetSMSTemplate retrieves an SMS template by name
func (s *TemplateService) GetSMSTemplate(name string) (*models.SMSTemplate, error) {
	var template models.SMSTemplate
	if err := s.db.Where("name = ?", name).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

// RenderEmailTemplate renders an email template with variables
func (s *TemplateService) RenderEmailTemplate(template *models.EmailTemplate, variables map[string]string) (string, string, error) {
	subject := s.replaceVariables(template.Subject, variables)
	body := s.replaceVariables(template.Body, variables)
	return subject, body, nil
}

// RenderSMSTemplate renders an SMS template with variables
func (s *TemplateService) RenderSMSTemplate(template *models.SMSTemplate, variables map[string]string) (string, error) {
	content := s.replaceVariables(template.Content, variables)
	return content, nil
}

// replaceVariables replaces {{variable}} placeholders with actual values
func (s *TemplateService) replaceVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// CreateEmailTemplate creates a new email template
func (s *TemplateService) CreateEmailTemplate(name, subject, body string, variables []string) (*models.EmailTemplate, error) {
	template := &models.EmailTemplate{
		Name:      name,
		Subject:   subject,
		Body:      body,
		Variables: variables,
	}
	if err := s.db.Create(template).Error; err != nil {
		return nil, err
	}
	return template, nil
}

// CreateSMSTemplate creates a new SMS template
func (s *TemplateService) CreateSMSTemplate(name, content string, variables []string) (*models.SMSTemplate, error) {
	template := &models.SMSTemplate{
		Name:      name,
		Content:   content,
		Variables: variables,
	}
	if err := s.db.Create(template).Error; err != nil {
		return nil, err
	}
	return template, nil
}

