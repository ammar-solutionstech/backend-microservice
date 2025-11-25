package services

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"backend/services/auth/internal/config"
	"backend/services/auth/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
)

type AuthService struct {
	db           *gorm.DB
	tokenService *TokenService
	otpService   *OTPService
	cfg          *config.Config
	rabbitMQ     *rabbitMQPublisher
}

func NewAuthService(cfg *config.Config, db *gorm.DB) *AuthService {
	return &AuthService{
		db:           db,
		cfg:          cfg,
		tokenService: NewTokenService(cfg, db),
		otpService:   NewOTPService(db),
		rabbitMQ:     newRabbitMQPublisher(cfg),
	}
}

// Authenticate authenticates a user with email and password
func (s *AuthService) Authenticate(email, password string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles").Preload("Permissions").Where("work_email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if user.Active != nil && *user.Active != 1 {
		return nil, ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

// Register creates a new user
func (s *AuthService) Register(req *RegisterRequest) (*models.User, error) {
	// Check if email already exists
	var existingUser models.User
	if err := s.db.Where("work_email = ?", req.WorkEmail).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := models.User{
		FirstName:     req.FirstName,
		LatestName:    req.LatestName,
		FatherName:    req.FatherName,
		WorkEmail:     req.WorkEmail,
		WorkMobile:    req.WorkMobile,
		NationalityID: req.NationalityID,
		Password:      string(hashedPassword),
		Active:        intPtr(1),
		EmailVerified: false,
		PhoneVerified: false,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// Generate OTPs for email and phone verification
	if user.WorkEmail != "" {
		_, _ = s.otpService.CreateOTPVerification(user.ID, "email")
	}
	if user.WorkMobile != nil && *user.WorkMobile != "" {
		_, _ = s.otpService.CreateOTPVerification(user.ID, "phone")
	}

	// Publish event
	s.rabbitMQ.PublishUserRegistered(user.ID, user.WorkEmail)

	return &user, nil
}

// GetUserRoles returns user roles as strings
func (s *AuthService) GetUserRoles(userID int) ([]string, error) {
	var roles []string
	if err := s.db.Model(&models.User{}).
		Select(`"Role".name`).
		Joins(`JOIN user_role ON user_role.user_id = "User".id`).
		Joins(`JOIN "Role" ON "Role".id = user_role.role_id`).
		Where(`"User".id = ?`, userID).
		Scan(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GetUserPermissions returns user permission IDs
func (s *AuthService) GetUserPermissions(userID int) ([]int, error) {
	var permissions []int
	query := `
		SELECT permission_id
		FROM user_special_permission
		WHERE user_id = ?
		UNION
		SELECT rp.permission_id
		FROM role_permission rp
		JOIN user_role ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = ?`
	if err := s.db.Raw(query, userID, userID).Scan(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// VerifyEmail verifies user email with OTP
func (s *AuthService) VerifyEmail(userID int, code string) error {
	if err := s.otpService.VerifyOTP(userID, "email", code); err != nil {
		return err
	}

	// Update user email_verified flag
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("email_verified", true).Error
}

// VerifyPhone verifies user phone with OTP
func (s *AuthService) VerifyPhone(userID int, code string) error {
	if err := s.otpService.VerifyOTP(userID, "phone", code); err != nil {
		return err
	}

	// Update user phone_verified flag
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("phone_verified", true).Error
}

// ForgotPassword initiates password reset
func (s *AuthService) ForgotPassword(email string) (string, error) {
	var user models.User
	if err := s.db.Where("work_email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	// Generate OTP for password reset
	code, err := s.otpService.CreateOTPVerification(user.ID, "email")
	if err != nil {
		return "", err
	}

	// Publish event
	s.rabbitMQ.PublishPasswordResetRequested(user.ID, email)

	return code, nil
}

// ResetPassword resets user password with OTP
func (s *AuthService) ResetPassword(email string, code string, newPassword string) error {
	var user models.User
	if err := s.db.Where("work_email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Verify OTP
	if err := s.otpService.VerifyOTP(user.ID, "email", code); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	return s.db.Model(&user).Update("password", string(hashedPassword)).Error
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID int) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles").Preload("Permissions").Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *AuthService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles").Preload("Permissions").Where("work_email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetDB returns the database connection (for controllers)
func (s *AuthService) GetDB() *gorm.DB {
	return s.db
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	FirstName     string
	LatestName    string
	FatherName    string
	WorkEmail     string
	Password      string
	WorkMobile    *string
	NationalityID *int
}
