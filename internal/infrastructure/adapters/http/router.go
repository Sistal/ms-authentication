package http

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter configures all HTTP routes
func SetupRouter(userHandler *UserHandler, authService interface {
	ValidateToken(token string) (string, error)
}) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// Protected routes
		user := v1.Group("/user")
		user.Use(AuthMiddleware(authService))
		{
			user.GET("/profile", userHandler.GetProfile)
		}
	}

	return router
}
