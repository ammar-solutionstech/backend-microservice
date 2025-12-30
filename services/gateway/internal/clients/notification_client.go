package clients

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"backend/services/gateway/internal/config"
	"backend/services/gateway/internal/utils"
	notificationpb "backend/services/notification/proto"
)

type NotificationClient struct {
	conn notificationpb.NotificationServiceClient
}

func NewNotificationClient(cfg *config.Config) *NotificationClient {
	var opts []grpc.DialOption

	// Use mTLS if certificates are configured
	if cfg.NotificationServiceGRPCMTLSCA != "" && cfg.NotificationServiceGRPCMTLSClientCert != "" && cfg.NotificationServiceGRPCMTLSClientKey != "" {
		// Extract server name from address (e.g., "notification-service:9003" -> "notification-service")
		serverName := strings.Split(cfg.NotificationServiceGRPC, ":")[0]
		if serverName == "" || serverName == "localhost" {
			serverName = "notification-service"
		}

		creds, err := utils.LoadGRPCClientCredentials(
			cfg.NotificationServiceGRPCMTLSClientCert,
			cfg.NotificationServiceGRPCMTLSClientKey,
			cfg.NotificationServiceGRPCMTLSCA,
			serverName,
		)
		if err != nil {
			log.Printf("Warning: failed to load Notification Service gRPC TLS credentials: %v", err)
			log.Println("Falling back to insecure connection")
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(creds))
			log.Println("Notification Service gRPC client configured with mTLS")
		}
	} else {
		log.Println("Warning: Notification Service gRPC mTLS not configured, using insecure connection")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.NotificationServiceGRPC, opts...)
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
