package routes

import (
	"encoding/json"
	"net/http"

	"backend/services/auth/internal/config"
	"backend/services/auth/internal/middleware"
	"backend/services/auth/internal/models"
	"backend/services/auth/internal/services"
)

type AuthController struct {
	cfg          *config.Config
	authService  *services.AuthService
	tokenService *services.TokenService
}

func NewAuthController(cfg *config.Config, authService *services.AuthService, tokenService *services.TokenService) *AuthController {
	return &AuthController{
		cfg:          cfg,
		authService:  authService,
		tokenService: tokenService,
	}
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	user, err := c.authService.Authenticate(req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	roles, _ := c.authService.GetUserRoles(user.ID)
	permissions, _ := c.authService.GetUserPermissions(user.ID)

	accessToken, err := c.tokenService.GenerateAccessToken(user.ID, roles, permissions)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := c.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    int(c.cfg.JWTAccessExpiry.Seconds()),
		"user":          user,
	})
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FirstName     string  `json:"first_name"`
		LatestName    string  `json:"latest_name"`
		FatherName    string  `json:"father_name"`
		WorkEmail     string  `json:"work_email"`
		Password      string  `json:"password"`
		WorkMobile    *string `json:"work_mobile"`
		NationalityID *int    `json:"nationality_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	registerReq := &services.RegisterRequest{
		FirstName:     req.FirstName,
		LatestName:    req.LatestName,
		FatherName:    req.FatherName,
		WorkEmail:     req.WorkEmail,
		Password:      req.Password,
		WorkMobile:    req.WorkMobile,
		NationalityID: req.NationalityID,
	}

	user, err := c.authService.Register(registerReq)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user":    user,
		"message": "User registered successfully. Please verify your email and phone.",
	})
}

func (c *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	refreshToken, err := c.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	user, err := c.authService.GetUserByID(refreshToken.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	roles, _ := c.authService.GetUserRoles(user.ID)
	permissions, _ := c.authService.GetUserPermissions(user.ID)

	accessToken, newRefreshToken, err := c.tokenService.RotateTokens(req.RefreshToken, user.ID, roles, permissions)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to rotate tokens")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    int(c.cfg.JWTAccessExpiry.Seconds()),
	})
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.tokenService.RevokeRefreshToken(req.RefreshToken); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (c *AuthController) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.authService.VerifyEmail(userID, req.Code); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (c *AuthController) VerifyPhone(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.authService.VerifyPhone(userID, req.Code); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (c *AuthController) ResendOTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Type string `json:"type"` // "email" or "phone"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	otpService := services.NewOTPService(c.authService.GetDB())
	_, err := otpService.ResendOTP(userID, req.Type)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "OTP sent successfully"})
}

func (c *AuthController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	code, err := c.authService.ForgotPassword(req.Email)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// In production, send OTP via notification service
	// For now, return it (remove in production)
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset OTP sent to email",
		"code":    code, // Remove this in production
	})
}

func (c *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.authService.ResetPassword(req.Email, req.Code, req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Password reset successfully"})
}

// User CRUD handlers
func (c *AuthController) GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := c.authService.GetDB().Preload("Roles").Preload("Permissions").Find(&users).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve users")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (c *AuthController) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.GetIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := c.authService.GetUserByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (c *AuthController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.authService.GetDB().Create(&user).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (c *AuthController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.GetIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := c.authService.GetDB().Model(&models.User{}).Where("id = ?", id).Updates(user).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (c *AuthController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.GetIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := c.authService.GetDB().Delete(&models.User{}, id).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user deleted successfully"})
}

// Helper functions
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
