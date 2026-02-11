package routes

import (
	"github.com/Sistal/ms-authentication/internal/domain/services"
	"github.com/Sistal/ms-authentication/internal/infrastructure/http/handlers"
	"github.com/Sistal/ms-authentication/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configura todas las rutas de la aplicación
func SetupRouter(authHandler *handlers.AuthHandler, authService services.AuthService, allowedOrigins string) *gin.Engine {
	router := gin.Default()

	// CORS Middleware
	router.Use(middleware.CORSMiddleware(allowedOrigins))

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	// @Summary Health Check
	// @Description Verifica el estado del servicio
	// @Tags Health
	// @Produce json
	// @Success 200 {object} map[string]string
	// @Router /health [get]
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "ms-authentication",
		})
	})

	// Rutas de autenticación (públicas) con prefijo /api/v1
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/validate", authHandler.Validate)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Rutas protegidas con JWT
	authProtected := router.Group("/api/v1/auth")
	authProtected.Use(middleware.JWTMiddleware(authService))
	{
		authProtected.GET("/me", authHandler.GetMe)
		authProtected.POST("/logout", authHandler.Logout)
		authProtected.PUT("/change-password", authHandler.ChangePassword)
		authProtected.GET("/roles", authHandler.GetRoles)
	}

	// Rutas de administración de usuarios (solo Admin y Super Admin)
	adminUsers := router.Group("/api/v1/auth/users")
	adminUsers.Use(middleware.JWTMiddleware(authService))
	adminUsers.Use(middleware.RequireRole(2, 3)) // Roles 2=Admin, 3=Super Admin
	{
		adminUsers.POST("", authHandler.CreateUser)
		adminUsers.GET("", authHandler.ListUsers)
		adminUsers.GET("/:id_usuario", authHandler.GetUser)
		adminUsers.PUT("/:id_usuario", authHandler.UpdateUser)
	}

	return router
}
