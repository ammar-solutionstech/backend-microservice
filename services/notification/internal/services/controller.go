package services

import (
	"encoding/json"
	"net/http"
)

type NotificationController struct {
	notificationService *NotificationService
	templateService     *TemplateService
}

func NewNotificationController(notificationService *NotificationService, templateService *TemplateService) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
		templateService:     templateService,
	}
}

func (c *NotificationController) SendEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		To           string            `json:"to"`
		Subject      string            `json:"subject"`
		Body         string            `json:"body"`
		TemplateName string            `json:"template_name"`
		Variables    map[string]string `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.notificationService.SendEmail(req.To, req.Subject, req.Body, req.TemplateName, req.Variables); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "email sent successfully"})
}

func (c *NotificationController) SendSMS(w http.ResponseWriter, r *http.Request) {
	var req struct {
		To           string            `json:"to"`
		Content      string            `json:"content"`
		TemplateName string            `json:"template_name"`
		Variables    map[string]string `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.notificationService.SendSMS(req.To, req.Content, req.TemplateName, req.Variables); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "sms sent successfully"})
}

func (c *NotificationController) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateName := r.URL.Query().Get("name")
	templateType := r.URL.Query().Get("type")

	if templateType == "email" {
		template, err := c.templateService.GetEmailTemplate(templateName)
		if err != nil {
			writeError(w, http.StatusNotFound, "template not found")
			return
		}
		writeJSON(w, http.StatusOK, template)
	} else if templateType == "sms" {
		template, err := c.templateService.GetSMSTemplate(templateName)
		if err != nil {
			writeError(w, http.StatusNotFound, "template not found")
			return
		}
		writeJSON(w, http.StatusOK, template)
	} else {
		writeError(w, http.StatusBadRequest, "invalid template type")
	}
}

func (c *NotificationController) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string   `json:"name"`
		Type      string   `json:"type"`
		Subject   string   `json:"subject"`
		Content   string   `json:"content"`
		Variables []string `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Type == "email" {
		template, err := c.templateService.CreateEmailTemplate(req.Name, req.Subject, req.Content, req.Variables)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, template)
	} else if req.Type == "sms" {
		template, err := c.templateService.CreateSMSTemplate(req.Name, req.Content, req.Variables)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, template)
	} else {
		writeError(w, http.StatusBadRequest, "invalid template type")
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
