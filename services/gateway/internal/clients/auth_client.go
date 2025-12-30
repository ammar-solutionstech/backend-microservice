package clients

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "backend/services/auth/proto"
	"backend/services/gateway/internal/config"
	"backend/services/gateway/internal/utils"
)

type AuthClient struct {
	conn       authpb.AuthServiceClient
	userClient authpb.UserServiceClient
}

func NewAuthClient(cfg *config.Config) *AuthClient {
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
		return &AuthClient{conn: nil, userClient: nil}
	}

	return &AuthClient{
		conn:       authpb.NewAuthServiceClient(conn),
		userClient: authpb.NewUserServiceClient(conn),
	}
}

func (c *AuthClient) ValidateToken(token string) (*authpb.ValidateTokenResponse, error) {
	if c.conn == nil {
		return nil, grpc.ErrClientConnClosing
	}

	ctx := context.Background()
	return c.conn.ValidateToken(ctx, &authpb.ValidateTokenRequest{Token: token})
}

func (c *AuthClient) GetUser(userID int32) (*authpb.UserResponse, error) {
	if c.userClient == nil {
		return nil, grpc.ErrClientConnClosing
	}

	ctx := context.Background()
	return c.userClient.GetUserByID(ctx, &authpb.GetUserByIDRequest{UserId: userID})
}
