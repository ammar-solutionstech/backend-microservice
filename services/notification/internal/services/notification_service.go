package services

import (
	"time"

	"gorm.io/gorm"

	"backend/services/notification/internal/config"
	"backend/services/notification/internal/models"
	"backend/services/notification/internal/providers"
)

type NotificationService struct {
	cfg             *config.Config
	db              *gorm.DB
	smtpProvider    *providers.SMTPProvider
	smsProvider     providers.SMSProvider
	templateService *TemplateService
}

func NewNotificationService(cfg *config.Config, db *gorm.DB, smtpProvider *providers.SMTPProvider, smsProvider providers.SMSProvider, templateService *TemplateService) *NotificationService {
	return &NotificationService{
		cfg:             cfg,
		db:              db,
		smtpProvider:    smtpProvider,
		smsProvider:     smsProvider,
		templateService: templateService,
	}
}

// SendEmail sends an email notification
func (s *NotificationService) SendEmail(to, subject, body string, templateName string, variables map[string]string) error {
	// If template name provided, use template
	if templateName != "" {
		template, err := s.templateService.GetEmailTemplate(templateName)
		if err == nil {
			subject, body, _ = s.templateService.RenderEmailTemplate(template, variables)
		}
	}

	// Create notification record
	notification := &models.Notification{
		Type:         "email",
		Recipient:    to,
		TemplateName: stringPtr(templateName),
		Subject:      stringPtr(subject),
		Content:      body,
		Status:       "pending",
	}

	if err := s.db.Create(notification).Error; err != nil {
		return err
	}

	// Send email
	if err := s.smtpProvider.SendEmail(to, subject, body); err != nil {
		s.db.Model(notification).Updates(map[string]interface{}{
			"status": "failed",
			"error":  err.Error(),
		})
		return err
	}

	// Update status
	s.db.Model(notification).Updates(map[string]interface{}{
		"status":  "sent",
		"sent_at": time.Now(),
	})

	return nil
}

// SendSMS sends an SMS notification
func (s *NotificationService) SendSMS(to, content string, templateName string, variables map[string]string) error {
	// If template name provided, use template
	if templateName != "" {
		template, err := s.templateService.GetSMSTemplate(templateName)
		if err == nil {
			content, _ = s.templateService.RenderSMSTemplate(template, variables)
		}
	}

	// Create notification record
	notification := &models.Notification{
		Type:         "sms",
		Recipient:    to,
		TemplateName: stringPtr(templateName),
		Content:      content,
		Status:       "pending",
	}

	if err := s.db.Create(notification).Error; err != nil {
		return err
	}

	// Send SMS
	if err := s.smsProvider.SendSMS(to, content); err != nil {
		s.db.Model(notification).Updates(map[string]interface{}{
			"status": "failed",
			"error":  err.Error(),
		})
		return err
	}

	// Update status
	s.db.Model(notification).Updates(map[string]interface{}{
		"status":  "sent",
		"sent_at": time.Now(),
	})

	return nil
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
