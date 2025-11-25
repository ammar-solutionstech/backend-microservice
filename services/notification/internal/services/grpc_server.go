package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notificationpb "backend/services/notification/proto"
)

type NotificationGRPCServer struct {
	notificationpb.UnimplementedNotificationServiceServer
	notificationService *NotificationService
	templateService     *TemplateService
}

func NewNotificationGRPCServer(notificationService *NotificationService, templateService *TemplateService) *NotificationGRPCServer {
	return &NotificationGRPCServer{
		notificationService: notificationService,
		templateService:     templateService,
	}
}

func (s *NotificationGRPCServer) SendEmail(ctx context.Context, req *notificationpb.SendEmailRequest) (*notificationpb.SendEmailResponse, error) {
	variables := make(map[string]string)
	for k, v := range req.Variables {
		variables[k] = v
	}

	err := s.notificationService.SendEmail(req.To, req.Subject, req.Body, req.TemplateName, variables)
	if err != nil {
		return &notificationpb.SendEmailResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &notificationpb.SendEmailResponse{
		Success: true,
	}, nil
}

func (s *NotificationGRPCServer) SendSMS(ctx context.Context, req *notificationpb.SendSMSRequest) (*notificationpb.SendSMSResponse, error) {
	variables := make(map[string]string)
	for k, v := range req.Variables {
		variables[k] = v
	}

	err := s.notificationService.SendSMS(req.To, req.Content, req.TemplateName, variables)
	if err != nil {
		return &notificationpb.SendSMSResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &notificationpb.SendSMSResponse{
		Success: true,
	}, nil
}

func (s *NotificationGRPCServer) GetTemplate(ctx context.Context, req *notificationpb.GetTemplateRequest) (*notificationpb.TemplateResponse, error) {
	if req.Type == "email" {
		emailTemplate, err := s.templateService.GetEmailTemplate(req.TemplateName)
		if err != nil {
			return nil, status.Error(codes.NotFound, "template not found")
		}
		return &notificationpb.TemplateResponse{
			Id:        int32(emailTemplate.ID),
			Name:      emailTemplate.Name,
			Type:      "email",
			Subject:   emailTemplate.Subject,
			Content:   emailTemplate.Body,
			Variables: emailTemplate.Variables,
		}, nil
	} else if req.Type == "sms" {
		smsTemplate, err := s.templateService.GetSMSTemplate(req.TemplateName)
		if err != nil {
			return nil, status.Error(codes.NotFound, "template not found")
		}
		return &notificationpb.TemplateResponse{
			Id:        int32(smsTemplate.ID),
			Name:      smsTemplate.Name,
			Type:      "sms",
			Content:   smsTemplate.Content,
			Variables: smsTemplate.Variables,
		}, nil
	}

	return nil, status.Error(codes.InvalidArgument, "invalid template type")
}

func (s *NotificationGRPCServer) CreateTemplate(ctx context.Context, req *notificationpb.CreateTemplateRequest) (*notificationpb.TemplateResponse, error) {
	if req.Type == "email" {
		emailTemplate, err := s.templateService.CreateEmailTemplate(req.Name, req.Subject, req.Content, req.Variables)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return &notificationpb.TemplateResponse{
			Id:        int32(emailTemplate.ID),
			Name:      emailTemplate.Name,
			Type:      "email",
			Subject:   emailTemplate.Subject,
			Content:   emailTemplate.Body,
			Variables: emailTemplate.Variables,
		}, nil
	} else if req.Type == "sms" {
		smsTemplate, err := s.templateService.CreateSMSTemplate(req.Name, req.Content, req.Variables)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return &notificationpb.TemplateResponse{
			Id:        int32(smsTemplate.ID),
			Name:      smsTemplate.Name,
			Type:      "sms",
			Content:   smsTemplate.Content,
			Variables: smsTemplate.Variables,
		}, nil
	}

	return nil, status.Error(codes.InvalidArgument, "invalid template type")
}
