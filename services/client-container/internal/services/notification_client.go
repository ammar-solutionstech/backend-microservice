package services

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	notificationpb "backend/services/notification/proto"
)

// NotificationClient is a gRPC client for the notification service
type NotificationClient struct {
	conn   *grpc.ClientConn
	client notificationpb.NotificationServiceClient
}

// NewNotificationClient creates a new notification service client
func NewNotificationClient(grpcAddr string) (*NotificationClient, error) {
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification service: %v", err)
	}

	return &NotificationClient{
		conn:   conn,
		client: notificationpb.NewNotificationServiceClient(conn),
	}, nil
}

// SendVerificationCode sends a verification code to organization admin via email and/or SMS
func (c *NotificationClient) SendVerificationCode(email, phone, code string) error {
	ctx := context.Background()

	// Send email if provided
	if email != "" {
		emailReq := &notificationpb.SendEmailRequest{
			To:      email,
			Subject: "Agent Certificate Verification Code",
			Body:    fmt.Sprintf("Your verification code is: %s\n\nThis code will expire in 15 minutes.", code),
		}

		_, err := c.client.SendEmail(ctx, emailReq)
		if err != nil {
			return fmt.Errorf("failed to send verification email: %v", err)
		}
	}

	// Send SMS if provided
	if phone != "" {
		smsReq := &notificationpb.SendSMSRequest{
			To:      phone,
			Content: fmt.Sprintf("Your verification code is: %s. Expires in 15 minutes.", code),
		}

		_, err := c.client.SendSMS(ctx, smsReq)
		if err != nil {
			return fmt.Errorf("failed to send verification SMS: %v", err)
		}
	}

	return nil
}

// Close closes the gRPC connection
func (c *NotificationClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
