package services

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "backend/services/auth/proto"
	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/utils"
)

type authClient struct {
	conn authpb.UserServiceClient
}

func newAuthClient(cfg *config.Config) *authClient {
	var opts []grpc.DialOption

	// Use mTLS if certificates are configured
	if cfg.AuthServiceGRPCMTLSCA != "" && cfg.AuthServiceGRPCMTLSClientCert != "" && cfg.AuthServiceGRPCMTLSClientKey != "" {
		// Extract server name from address (e.g., "auth-service:9001" -> "auth-service")
		serverName := strings.Split(cfg.AuthServiceGRPC, ":")[0]
		if serverName == "" || serverName == "localhost" {
			serverName = "auth-service"
		}

		creds, err := utils.LoadGRPCClientCredentials(
			cfg.AuthServiceGRPCMTLSClientCert,
			cfg.AuthServiceGRPCMTLSClientKey,
			cfg.AuthServiceGRPCMTLSCA,
			serverName,
		)
		if err != nil {
			log.Printf("Warning: failed to load Auth Service gRPC TLS credentials: %v", err)
			log.Println("Falling back to insecure connection")
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(creds))
			log.Println("Auth Service gRPC client configured with mTLS")
		}
	} else {
		log.Println("Warning: Auth Service gRPC mTLS not configured, using insecure connection")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.AuthServiceGRPC, opts...)
	if err != nil {
		log.Printf("Warning: failed to connect to Auth Service: %v", err)
		return &authClient{conn: nil}
	}

	client := authpb.NewUserServiceClient(conn)
	return &authClient{conn: client}
}

func (c *authClient) ValidateUser(userID int) bool {
	if c.conn == nil {
		// If connection failed, allow for now (graceful degradation)
		return true
	}

	ctx := context.Background()
	resp, err := c.conn.ValidateUser(ctx, &authpb.ValidateUserRequest{UserId: int32(userID)})
	if err != nil {
		log.Printf("Error validating user %d: %v", userID, err)
		return false
	}

	return resp.Valid
}
