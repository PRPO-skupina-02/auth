package api

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/PRPO-skupina-02/auth/auth"
	"github.com/PRPO-skupina-02/auth/models"
	"github.com/PRPO-skupina-02/common/messaging"
	"github.com/PRPO-skupina-02/common/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"omitempty,min=1"`
	LastName  string `json:"last_name" binding:"omitempty,min=1"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

type UserResponse struct {
	ID        uuid.UUID       `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Email     string          `json:"email"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	Role      models.UserRole `json:"role"`
	Active    bool            `json:"active"`
}

func newUserResponse(user models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Active:    user.Active,
	}
}

// Register
//
//	@Id				Register
//	@Summary		Register a new user
//	@Description	Register a new user account
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterRequest	true	"Registration details"
//	@Success		201		{object}	UserResponse
//	@Failure		400		{object}	middleware.HttpError
//	@Failure		409		{object}	middleware.HttpError
//	@Failure		500		{object}	middleware.HttpError
//	@Router			/register [post]
func RegisterUser(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	// Check if user already exists
	exists, err := models.UserExists(tx, req.Email)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Create user with customer role (only customers can self-register)
	user := models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      models.RoleCustomer,
		Active:    true,
	}

	if err := user.SetPassword(req.Password); err != nil {
		_ = c.Error(err)
		return
	}

	if err := user.Create(tx); err != nil {
		_ = c.Error(err)
		return
	}

	// Send welcome email asynchronously
	go func() {
		rabbitmqURL := os.Getenv("RABBITMQ_URL")
		if rabbitmqURL == "" {
			rabbitmqURL = "amqp://guest:guest@localhost:5672/"
		}

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}

		emailMsg := messaging.NewEmailMessage(
			user.Email,
			"welcome",
			map[string]interface{}{
				"Subject":  "Welcome to CineCore!",
				"UserName": user.FirstName,
				"AppLink":  frontendURL,
			},
		)

		if err := messaging.PublishEmailSimple(rabbitmqURL, emailMsg.To, emailMsg.Template, emailMsg.TemplateData); err != nil {
			slog.Error("Failed to publish welcome email", "email", user.Email, "error", err)
		}
	}()

	c.JSON(http.StatusCreated, newUserResponse(user))
}

// Login
//
//	@Id				Login
//	@Summary		Login user
//	@Description	Authenticate user and return JWT tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"Login credentials"
//	@Success		200		{object}	TokenResponse
//	@Failure		400		{object}	middleware.HttpError
//	@Failure		401		{object}	middleware.HttpError
//	@Failure		500		{object}	middleware.HttpError
//	@Router			/login [post]
func Login(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	// Validate credentials
	user, err := models.ValidateCredentials(tx, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Generate tokens
	accessToken, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		_ = c.Error(err)
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400, // 24 hours in seconds
	})
}

// VerifyToken
//
//	@Id				VerifyToken
//	@Summary		Verify JWT token
//	@Description	Verify a JWT token and return user information
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			token	body		object{token=string}	true	"Token to verify"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	middleware.HttpError
//	@Failure		401		{object}	middleware.HttpError
//	@Failure		500		{object}	middleware.HttpError
//	@Router			/verify [post]
func VerifyToken(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)

	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	claims, err := auth.ValidateToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Get user from database
	user, err := models.GetUser(tx, claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is inactive"})
		return
	}

	c.JSON(http.StatusOK, newUserResponse(user))
}

// RefreshToken
//
//	@Id				RefreshToken
//	@Summary		Refresh access token
//	@Description	Use refresh token to get a new access token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			refresh_token	body		object{refresh_token=string}	true	"Refresh token"
//	@Success		200				{object}	TokenResponse
//	@Failure		400				{object}	middleware.HttpError
//	@Failure		401				{object}	middleware.HttpError
//	@Failure		500				{object}	middleware.HttpError
//	@Router			/refresh [post]
func RefreshToken(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	claims, err := auth.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Verify user still exists and is active
	user, err := models.GetUser(tx, claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is inactive"})
		return
	}

	// Generate new tokens
	accessToken, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		_ = c.Error(err)
		return
	}

	newRefreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    86400,
	})
}

// GetCurrentUser
//
//	@Id				GetCurrentUser
//	@Summary		Get current user
//	@Description	Get information about the currently authenticated user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	UserResponse
//	@Failure		401	{object}	middleware.HttpError
//	@Failure		500	{object}	middleware.HttpError
//	@Router			/me [get]
func GetCurrentUser(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)
	userID := c.MustGet("user_id").(uuid.UUID)

	user, err := models.GetUser(tx, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, newUserResponse(user))
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name" binding:"omitempty,min=1"`
	LastName  string `json:"last_name" binding:"omitempty,min=1"`
}

// UpdateCurrentUser
//
//	@Id				UpdateCurrentUser
//	@Summary		Update current user
//	@Description	Update information about the currently authenticated user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		UpdateUserRequest	true	"User update details"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	middleware.HttpError
//	@Failure		401		{object}	middleware.HttpError
//	@Failure		500		{object}	middleware.HttpError
//	@Router			/me [put]
func UpdateCurrentUser(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)
	userID := c.MustGet("user_id").(uuid.UUID)

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	user, err := models.GetUser(tx, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}

	if err := user.Save(tx); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, newUserResponse(user))
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword
//
//	@Id				ChangePassword
//	@Summary		Change password
//	@Description	Change password for the currently authenticated user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		ChangePasswordRequest	true	"Password change details"
//	@Success		200		{object}	object{message=string}
//	@Failure		400		{object}	middleware.HttpError
//	@Failure		401		{object}	middleware.HttpError
//	@Failure		500		{object}	middleware.HttpError
//	@Router			/me/password [put]
func ChangePassword(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)
	userID := c.MustGet("user_id").(uuid.UUID)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	user, err := models.GetUser(tx, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Verify old password
	if err := user.CheckPassword(req.OldPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
		return
	}

	// Set new password
	if err := user.SetPassword(req.NewPassword); err != nil {
		_ = c.Error(err)
		return
	}

	if err := user.Save(tx); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
