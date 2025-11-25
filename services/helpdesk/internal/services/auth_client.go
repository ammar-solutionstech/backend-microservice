package services

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "backend/services/auth/proto"
)

type authClient struct {
	conn authpb.UserServiceClient
}

func newAuthClient(authServiceAddr string) *authClient {
	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
