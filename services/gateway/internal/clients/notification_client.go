package clients

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	notificationpb "backend/services/notification/proto"
)

type NotificationClient struct {
	conn notificationpb.NotificationServiceClient
}

func NewNotificationClient(notificationServiceAddr string) *NotificationClient {
	conn, err := grpc.Dial(notificationServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Warning: failed to connect to Notification Service: %v", err)
		return &NotificationClient{conn: nil}
	}

	return &NotificationClient{
		conn: notificationpb.NewNotificationServiceClient(conn),
	}
}

func (c *NotificationClient) SendEmail(ctx context.Context, req *notificationpb.SendEmailRequest) (*notificationpb.SendEmailResponse, error) {
	if c.conn == nil {
		return nil, grpc.ErrClientConnClosing
	}
	return c.conn.SendEmail(ctx, req)
}

func (c *NotificationClient) SendSMS(ctx context.Context, req *notificationpb.SendSMSRequest) (*notificationpb.SendSMSResponse, error) {
	if c.conn == nil {
		return nil, grpc.ErrClientConnClosing
	}
	return c.conn.SendSMS(ctx, req)
}
