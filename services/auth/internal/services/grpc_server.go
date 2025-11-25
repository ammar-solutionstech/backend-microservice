package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"backend/services/auth/internal/config"
	"backend/services/auth/internal/models"
	authpb "backend/services/auth/proto"
)

// AuthGRPCServer implements the AuthService gRPC interface
type AuthGRPCServer struct {
	authpb.UnimplementedAuthServiceServer
	authService  *AuthService
	tokenService *TokenService
	cfg          *config.Config
}

func NewAuthGRPCServer(cfg *config.Config, authService *AuthService, tokenService *TokenService) *AuthGRPCServer {
	return &AuthGRPCServer{
		authService:  authService,
		tokenService: tokenService,
		cfg:          cfg,
	}
}

// Login handles user login
func (s *AuthGRPCServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	user, err := s.authService.Authenticate(req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Get user roles and permissions
	roles, err := s.authService.GetUserRoles(user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user roles")
	}

	permissions, err := s.authService.GetUserPermissions(user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user permissions")
	}

	// Generate tokens
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, roles, permissions)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}

	return &authpb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.JWTAccessExpiry.Seconds()),
		User:         toUserResponse(user),
	}, nil
}

// Register handles user registration
func (s *AuthGRPCServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	registerReq := &RegisterRequest{
		FirstName:     req.FirstName,
		LatestName:    req.LatestName,
		FatherName:    req.FatherName,
		WorkEmail:     req.WorkEmail,
		Password:      req.Password,
		WorkMobile:    stringPtr(req.WorkMobile),
		NationalityID: intPtr(int(req.NationalityId)),
	}

	user, err := s.authService.Register(registerReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &authpb.RegisterResponse{
		User:    toUserResponse(user),
		Message: "User registered successfully. Please verify your email and phone.",
	}, nil
}

// RefreshToken handles token refresh
func (s *AuthGRPCServer) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	// Validate refresh token
	refreshToken, err := s.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	// Get user
	user, err := s.authService.GetUserByID(refreshToken.UserID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Get roles and permissions
	roles, err := s.authService.GetUserRoles(user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user roles")
	}

	permissions, err := s.authService.GetUserPermissions(user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user permissions")
	}

	// Rotate tokens
	accessToken, newRefreshToken, err := s.tokenService.RotateTokens(req.RefreshToken, user.ID, roles, permissions)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to rotate tokens")
	}

	return &authpb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.cfg.JWTAccessExpiry.Seconds()),
	}, nil
}

// ValidateToken validates an access token
func (s *AuthGRPCServer) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	claims, err := s.tokenService.ValidateAccessToken(req.Token)
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:       true,
		UserId:      int32(claims.UserID),
		Roles:       claims.Roles,
		Permissions: int32SliceToInt32(claims.Permissions),
	}, nil
}

// Logout handles user logout
func (s *AuthGRPCServer) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	if err := s.tokenService.RevokeRefreshToken(req.RefreshToken); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &authpb.LogoutResponse{
		Success: true,
	}, nil
}

// Helper functions
func toUserResponse(user *models.User) *authpb.UserResponse {
	return &authpb.UserResponse{
		Id:            int32(user.ID),
		FirstName:     user.FirstName,
		LatestName:    user.LatestName,
		FatherName:    user.FatherName,
		WorkEmail:     user.WorkEmail,
		WorkMobile:    stringValue(user.WorkMobile),
		EmailVerified: user.EmailVerified,
		PhoneVerified: user.PhoneVerified,
		Active:        int32(intValue(user.Active)),
	}
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func int32SliceToInt32(slice []int) []int32 {
	result := make([]int32, len(slice))
	for i, v := range slice {
		result[i] = int32(v)
	}
	return result
}
