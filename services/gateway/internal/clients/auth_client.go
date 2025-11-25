package clients

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "backend/services/auth/proto"
)

type AuthClient struct {
	conn       authpb.AuthServiceClient
	userClient authpb.UserServiceClient
}

func NewAuthClient(authServiceAddr string) *AuthClient {
	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
