package api

import (
	"net/http"

	_ "github.com/PRPO-skupina-02/auth/api/docs"
	"github.com/PRPO-skupina-02/auth/auth"
	"github.com/PRPO-skupina-02/auth/models"
	"github.com/PRPO-skupina-02/common/middleware"
	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

//	@title			Auth API
//	@version		1.0
//	@description	Authentication and authorization service for the PRPO project

//	@host		localhost:8080
//	@BasePath	/api/v1/auth

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func Register(router *gin.Engine, db *gorm.DB, trans ut.Translator) {
	// Healthcheck
	router.GET("/healthcheck", healthcheck)

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// REST API
	v1 := router.Group("/api/v1/auth")
	v1.Use(middleware.TransactionMiddleware(db))
	v1.Use(middleware.TranslationMiddleware(trans))
	v1.Use(middleware.ErrorMiddleware)

	// Public routes
	v1.POST("/register", RegisterUser)
	v1.POST("/login", Login)
	v1.POST("/refresh", RefreshToken)
	v1.POST("/verify", VerifyToken)

	// Protected routes
	protected := v1.Group("")
	protected.Use(AuthMiddleware())

	protected.GET("/me", GetCurrentUser)
	protected.PUT("/me", UpdateCurrentUser)
	protected.PUT("/me/password", ChangePassword)

	// Admin routes (for managing users)
	admin := v1.Group("/users")
	admin.Use(AuthMiddleware())
	admin.Use(RequireAdmin())

	admin.GET("", UsersList)
	admin.GET("/:userID", UsersShow)
	admin.POST("", AdminCreateUser)
	admin.PUT("/:userID", UsersUpdate)
	admin.DELETE("/:userID", UsersDelete)
}

func healthcheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// AuthMiddleware validates JWT token and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Extract token from "Bearer <token>"
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenString := authHeader[len(bearerPrefix):]
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Get user from database to verify and get role
		tx := middleware.GetContextTransaction(c)
		user, err := models.GetUser(tx, claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		if !user.Active {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User account is inactive"})
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", user.Role)

		c.Next()
	}
}
