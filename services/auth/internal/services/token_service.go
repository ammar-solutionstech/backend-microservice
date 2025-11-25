package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"backend/services/auth/internal/config"
	"backend/services/auth/internal/models"
)

type TokenService struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewTokenService(cfg *config.Config, db *gorm.DB) *TokenService {
	return &TokenService{cfg: cfg, db: db}
}

type CustomClaims struct {
	UserID      int      `json:"sub"`
	Roles       []string `json:"roles"`
	Permissions []int    `json:"permissions"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a short-lived JWT access token
func (s *TokenService) GenerateAccessToken(userID int, roles []string, permissions []int) (string, error) {
	now := time.Now().UTC()
	claims := CustomClaims{
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.JWTIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTAccessExpiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// GenerateRefreshToken generates a long-lived refresh token and stores it in DB
func (s *TokenService) GenerateRefreshToken(userID int) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Store in database
	refreshToken := models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(s.cfg.JWTRefreshExpiry),
	}

	if err := s.db.Create(&refreshToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

// ValidateAccessToken validates an access token and returns claims
func (s *TokenService) ValidateAccessToken(tokenStr string) (*CustomClaims, error) {
	// Check if token is blacklisted
	var blacklist models.TokenBlacklist
	if err := s.db.Where("token_id = ? AND expires_at > ?", tokenStr, time.Now().UTC()).First(&blacklist).Error; err == nil {
		return nil, errors.New("token is blacklisted")
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now().UTC()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (s *TokenService) ValidateRefreshToken(tokenStr string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	if err := s.db.Where("token = ? AND expires_at > ? AND revoked_at IS NULL", tokenStr, time.Now().UTC()).First(&refreshToken).Error; err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}
	return &refreshToken, nil
}

// RevokeRefreshToken revokes a refresh token
func (s *TokenService) RevokeRefreshToken(tokenStr string) error {
	now := time.Now().UTC()
	result := s.db.Model(&models.RefreshToken{}).
		Where("token = ?", tokenStr).
		Update("revoked_at", now)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("refresh token not found")
	}
	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *TokenService) RevokeAllUserTokens(userID int) error {
	now := time.Now().UTC()
	return s.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

// BlacklistToken adds a token to the blacklist
func (s *TokenService) BlacklistToken(tokenID string, expiresAt time.Time) error {
	blacklist := models.TokenBlacklist{
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
	}
	return s.db.Create(&blacklist).Error
}

// RotateTokens generates new access and refresh tokens, revoking the old refresh token
func (s *TokenService) RotateTokens(oldRefreshToken string, userID int, roles []string, permissions []int) (string, string, error) {
	// Validate old refresh token
	_, err := s.ValidateRefreshToken(oldRefreshToken)
	if err != nil {
		return "", "", err
	}

	// Revoke old token
	if err := s.RevokeRefreshToken(oldRefreshToken); err != nil {
		return "", "", err
	}

	// Generate new tokens
	accessToken, err := s.GenerateAccessToken(userID, roles, permissions)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	// Clean up expired tokens
	go s.cleanupExpiredTokens()

	return accessToken, newRefreshToken, nil
}

// cleanupExpiredTokens removes expired tokens from database
func (s *TokenService) cleanupExpiredTokens() {
	now := time.Now().UTC()

	// Delete expired refresh tokens
	s.db.Where("expires_at < ?", now).Delete(&models.RefreshToken{})

	// Delete expired blacklist entries
	s.db.Where("expires_at < ?", now).Delete(&models.TokenBlacklist{})
}
