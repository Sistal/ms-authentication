package middleware

import (
	"log/slog"
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
			slog.Warn("[JWTMiddleware] Authorization header faltante",
				slog.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Validar formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			slog.Warn("[JWTMiddleware] Formato de Authorization header inválido",
				slog.String("client_ip", c.ClientIP()),
				slog.String("header", authHeader),
			)
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
			slog.Warn("[JWTMiddleware] Token inválido o expirado",
				slog.String("client_ip", c.ClientIP()),
				slog.String("error", err.Error()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		slog.Info("[JWTMiddleware] Usuario autenticado",
			slog.Int("user_id", claims.UserID),
			slog.Int("role", claims.Role),
			slog.String("endpoint", c.Request.URL.Path),
		)

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
		userID, _ := c.Get("user_id")

		if !exists {
			slog.Error("[RequireRole] Rol no encontrado en contexto",
				slog.Any("user_id", userID),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user role not found in context",
			})
			c.Abort()
			return
		}

		userRole, ok := role.(int)
		if !ok {
			slog.Error("[RequireRole] Tipo de rol inválido en contexto",
				slog.Any("role", role),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid role type",
			})
			c.Abort()
			return
		}

		// Verificar si el rol del usuario está en la lista de roles permitidos
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				slog.Debug("[RequireRole] Acceso autorizado por rol",
					slog.Int("user_role", userRole),
					slog.Any("allowed_roles", allowedRoles),
					slog.String("endpoint", c.Request.URL.Path),
				)
				c.Next()
				return
			}
		}

		slog.Warn("[RequireRole] Acceso denegado: rol insuficiente",
			slog.Int("user_role", userRole),
			slog.Any("allowed_roles", allowedRoles),
			slog.Any("user_id", userID),
			slog.String("endpoint", c.Request.URL.Path),
		)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "insufficient permissions",
		})
		c.Abort()
	}
}
