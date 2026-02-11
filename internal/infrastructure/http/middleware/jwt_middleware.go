package middleware

import (
	"net/http"
	"strings"

	"github.com/Sistal/ms-authentication/internal/domain/services"
	"github.com/gin-gonic/gin"
)

// JWTMiddleware crea un middleware de autenticación JWT
func JWTMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extraer token del header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Validar formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validar token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Agregar claims al contexto para uso posterior
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole crea un middleware que verifica el rol del usuario
func RequireRole(allowedRoles ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user role not found in context",
			})
			c.Abort()
			return
		}

		userRole, ok := role.(int)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid role type",
			})
			c.Abort()
			return
		}

		// Verificar si el rol del usuario está en la lista de roles permitidos
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "insufficient permissions",
		})
		c.Abort()
	}
}
