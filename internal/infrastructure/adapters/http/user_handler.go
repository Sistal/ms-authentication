package http

import (
	"net/http"
	"strings"

	"github.com/Sistal/ms-authentication/internal/application/usecases"
	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userUseCase *usecases.UserUseCase
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userUseCase *usecases.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUseCase}
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Register handles user registration
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userUseCase.RegisterUser(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if err == usecases.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	token, user, err := h.userUseCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == usecases.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// GetProfile returns the current user's profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "invalid user ID format"})
		return
	}

	user, err := h.userUseCase.GetUserByID(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// AuthMiddleware validates JWT token
func AuthMiddleware(authService interface {
	ValidateToken(token string) (string, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		userID, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
