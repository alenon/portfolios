package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/lenon/portfolios/internal/services"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService          services.AuthService
	passwordResetService services.PasswordResetService
	userRepo             repository.UserRepository
	accessTokenDuration  int // in seconds
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(
	authService services.AuthService,
	passwordResetService services.PasswordResetService,
	userRepo repository.UserRepository,
	accessTokenDuration int,
) *AuthHandler {
	return &AuthHandler{
		authService:          authService,
		passwordResetService: passwordResetService,
		userRepo:             userRepo,
		accessTokenDuration:  accessTokenDuration,
	}
}

// Register handles user registration
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Call auth service to register user
	user, accessToken, refreshToken, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		// Check for duplicate email error
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: "A user with this email already exists",
				Code:  "EMAIL_ALREADY_EXISTS",
			})
			return
		}

		// Check for password validation error
		if strings.Contains(err.Error(), "invalid password") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
				Code:  "INVALID_PASSWORD",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to register user",
			Code:  "REGISTRATION_FAILED",
		})
		return
	}

	// Build response
	response := dto.AuthResponse{
		User:         userToResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    h.accessTokenDuration,
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Call auth service to login user
	user, accessToken, refreshToken, err := h.authService.Login(req.Email, req.Password, req.RememberMe)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Invalid email or password",
			Code:  "INVALID_CREDENTIALS",
		})
		return
	}

	// Build response
	response := dto.AuthResponse{
		User:         userToResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    h.accessTokenDuration,
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Call auth service to refresh token
	accessToken, err := h.authService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Invalid or expired refresh token",
			Code:  "INVALID_REFRESH_TOKEN",
		})
		return
	}

	// Build response
	response := dto.RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   h.accessTokenDuration,
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Call auth service to logout user
	if err := h.authService.Logout(req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to logout",
			Code:  "LOGOUT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Logged out successfully",
	})
}

// GetCurrentUser returns the current authenticated user's profile
// GET /api/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "NOT_AUTHENTICATED",
		})
		return
	}

	// Fetch user from repository
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "User not found",
			Code:  "USER_NOT_FOUND",
		})
		return
	}

	// Build response
	response := userToResponse(user)

	c.JSON(http.StatusOK, response)
}

// ForgotPassword initiates the password reset process
// POST /api/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Call password reset service
	// Always return success to prevent email enumeration
	_ = h.passwordResetService.InitiateReset(req.Email)

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "If an account exists with this email, a password reset link has been sent",
	})
}

// ResetPassword completes the password reset process
// POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Call password reset service
	if err := h.passwordResetService.ResetPassword(req.Token, req.NewPassword); err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "already used") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid or expired reset token",
				Code:  "INVALID_RESET_TOKEN",
			})
			return
		}

		if strings.Contains(err.Error(), "invalid password") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
				Code:  "INVALID_PASSWORD",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to reset password",
			Code:  "RESET_PASSWORD_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Password reset successfully",
	})
}

// userToResponse converts a User model to UserResponse DTO
func userToResponse(user *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		CreatedAt:   user.CreatedAt,
		LastLoginAt: user.LastLoginAt,
	}
}
