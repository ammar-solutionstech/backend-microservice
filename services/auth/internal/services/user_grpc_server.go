package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "backend/services/auth/proto"
)

// UserGRPCServer implements the UserService gRPC interface
type UserGRPCServer struct {
	authpb.UnimplementedUserServiceServer
	authService *AuthService
}

func NewUserGRPCServer(authService *AuthService) *UserGRPCServer {
	return &UserGRPCServer{
		authService: authService,
	}
}

// GetUser retrieves a user by email
func (s *UserGRPCServer) GetUser(ctx context.Context, req *authpb.GetUserRequest) (*authpb.UserResponse, error) {
	user, err := s.authService.GetUserByEmail(req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return toUserResponse(user), nil
}

// GetUserByID retrieves a user by ID
func (s *UserGRPCServer) GetUserByID(ctx context.Context, req *authpb.GetUserByIDRequest) (*authpb.UserResponse, error) {
	user, err := s.authService.GetUserByID(int(req.UserId))
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return toUserResponse(user), nil
}

// ValidateUser validates if a user exists and is active
func (s *UserGRPCServer) ValidateUser(ctx context.Context, req *authpb.ValidateUserRequest) (*authpb.ValidateUserResponse, error) {
	user, err := s.authService.GetUserByID(int(req.UserId))
	if err != nil {
		return &authpb.ValidateUserResponse{
			Valid: false,
		}, nil
	}

	// Check if user is active
	valid := user.Active != nil && *user.Active == 1

	return &authpb.ValidateUserResponse{
		Valid: valid,
		User:  toUserResponse(user),
	}, nil
}

// GetUserRoles retrieves user roles
func (s *UserGRPCServer) GetUserRoles(ctx context.Context, req *authpb.GetUserRolesRequest) (*authpb.GetUserRolesResponse, error) {
	roles, err := s.authService.GetUserRoles(int(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user roles")
	}

	return &authpb.GetUserRolesResponse{
		Roles: roles,
	}, nil
}

// GetUserPermissions retrieves user permissions
func (s *UserGRPCServer) GetUserPermissions(ctx context.Context, req *authpb.GetUserPermissionsRequest) (*authpb.GetUserPermissionsResponse, error) {
	permissions, err := s.authService.GetUserPermissions(int(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user permissions")
	}

	return &authpb.GetUserPermissionsResponse{
		Permissions: intSliceToInt32(permissions),
	}, nil
}

// Helper function
func intSliceToInt32(slice []int) []int32 {
	result := make([]int32, len(slice))
	for i, v := range slice {
		result[i] = int32(v)
	}
	return result
}
