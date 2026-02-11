package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Sistal/ms-authentication/internal/application/usecases"
	"github.com/Sistal/ms-authentication/internal/domain/entities"
	"github.com/Sistal/ms-authentication/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthHandler maneja las peticiones HTTP de autenticación
type AuthHandler struct {
	authUseCase *usecases.AuthUseCase
}

// NewAuthHandler crea una nueva instancia del handler
func NewAuthHandler(authUseCase *usecases.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Register maneja el endpoint POST /api/v1/auth/register
// @Summary Registrar nuevo usuario (público)
// @Description Crea un nuevo usuario en el sistema
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body entities.CreateUsuarioDTO true "Datos de registro"
// @Success 201 {object} entities.APIResponse
// @Failure 400 {object} entities.APIResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req entities.CreateUsuarioDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Datos de entrada inválidos",
			[]entities.ValidationError{{Field: "request", Message: err.Error()}},
		))
		return
	}

	// Validar RUT
	if !utils.ValidateRUT(req.RUT) {
		c.JSON(http.StatusBadRequest, entities.ErrorResponse(
			"Error en la validación",
			entities.ErrorDetail{Code: "INVALID_RUT_FORMAT", Details: "El formato del RUT no es válido"},
		))
		return
	}

	// Validar longitud de contraseña
	if len(req.Password) < utils.MinPasswordLength {
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Error en la validación",
			[]entities.ValidationError{{
				Field:   "password",
				Message: "La contraseña debe tener al menos 8 caracteres",
			}},
		))
		return
	}

	response, err := h.authUseCase.CreateUsuario(c.Request.Context(), req, false) // false = no requiere ser admin
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "ya está registrado") {
			c.JSON(http.StatusBadRequest, entities.ErrorResponse(
				"Error en la validación",
				entities.ErrorDetail{Code: "DUPLICATE_USERNAME", Details: err.Error()},
			))
			return
		}
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error interno del servidor"))
		return
	}

	c.JSON(http.StatusCreated, entities.SuccessResponse("Usuario creado exitosamente", response))
}

// Login maneja el endpoint POST /api/v1/auth/login
// @Summary Iniciar sesión
// @Description Autentica un usuario y retorna un token JWT
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body entities.LoginDTO true "Credenciales de login"
// @Success 200 {object} entities.APIResponse
// @Failure 400 {object} entities.APIResponse
// @Failure 401 {object} entities.APIResponse
// @Failure 403 {object} entities.APIResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req entities.LoginDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Datos de entrada inválidos",
			[]entities.ValidationError{{Field: "nombre_usuario", Message: "El nombre de usuario es requerido"}},
		))
		return
	}

	response, err := h.authUseCase.Login(c.Request.Context(), req)
	if err != nil {
		errorMsg := err.Error()

		// Usuario inactivo = 403
		if strings.Contains(errorMsg, "not active") || strings.Contains(errorMsg, "inactivo") {
			c.JSON(http.StatusForbidden, entities.ErrorResponse(
				"Acceso denegado",
				entities.ErrorDetail{Code: "USER_INACTIVE", Details: "La cuenta está inactiva o suspendida"},
			))
			return
		}

		// Credenciales incorrectas = 401
		if strings.Contains(errorMsg, "invalid credentials") || strings.Contains(errorMsg, "credenciales incorrectas") {
			c.JSON(http.StatusUnauthorized, entities.ErrorResponse(
				"Credenciales incorrectas",
				entities.ErrorDetail{Code: "INVALID_CREDENTIALS", Details: "Nombre de usuario o contraseña incorrectos"},
			))
			return
		}

		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error interno del servidor"))
		return
	}

	c.JSON(http.StatusOK, entities.SuccessResponse("Autenticación exitosa", response))
}

// Validate maneja el endpoint GET /api/v1/auth/validate
// @Summary Validar token JWT
// @Description Valida un token JWT y retorna información del usuario
// @Tags Authentication
// @Security BearerAuth
// @Success 200 {object} entities.APIResponse
// @Failure 401 {object} entities.APIResponse
// @Router /api/v1/auth/validate [get]
func (h *AuthHandler) Validate(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse(
			"Token inválido o expirado",
			entities.ErrorDetail{Code: "INVALID_TOKEN", Details: "El token proporcionado no es válido"},
		))
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse(
			"Token inválido o expirado",
			entities.ErrorDetail{Code: "INVALID_TOKEN", Details: "El token proporcionado no es válido"},
		))
		return
	}

	token := parts[1]
	response, err := h.authUseCase.ValidateTokenComplete(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse(
			"Token inválido o expirado",
			entities.ErrorDetail{Code: "INVALID_TOKEN", Details: "El token proporcionado no es válido"},
		))
		return
	}

	c.JSON(http.StatusOK, entities.SuccessResponse("Token válido", response))
}

// GetMe maneja el endpoint GET /api/v1/auth/me
// @Summary Obtener perfil de usuario actual
// @Description Retorna información completa del usuario autenticado
// @Tags Authentication
// @Security BearerAuth
// @Success 200 {object} entities.APIResponse
// @Failure 401 {object} entities.APIResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponseSimple("Usuario no autenticado"))
		return
	}

	// Obtener usuario completo de la BD
	usuario, err := h.authUseCase.GetUsuarioCompleto(c.Request.Context(), userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error al obtener información del usuario"))
		return
	}

	c.JSON(http.StatusOK, entities.SuccessResponse("", usuario))
}

// ChangePassword maneja el endpoint PUT /api/v1/auth/change-password
// @Summary Cambiar contraseña
// @Description Permite al usuario cambiar su contraseña
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body entities.ChangePasswordDTO true "Datos de cambio de contraseña"
// @Success 200 {object} entities.APIResponse
// @Failure 400 {object} entities.APIResponse
// @Failure 401 {object} entities.APIResponse
// @Router /api/v1/auth/change-password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponseSimple("Usuario no autenticado"))
		return
	}

	var req entities.ChangePasswordDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Error en la validación",
			[]entities.ValidationError{{Field: "request", Message: err.Error()}},
		))
		return
	}

	// Validar que las contraseñas coincidan
	if req.PasswordNueva != req.PasswordConfirmacion {
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Error en la validación",
			[]entities.ValidationError{{
				Field:   "password_confirmacion",
				Message: "Las contraseñas no coinciden",
			}},
		))
		return
	}

	err := h.authUseCase.ChangePassword(c.Request.Context(), userID.(int), req)
	if err != nil {
		if strings.Contains(err.Error(), "incorrect") || strings.Contains(err.Error(), "incorrecta") {
			c.JSON(http.StatusUnauthorized, entities.ErrorResponseSimple("La contraseña actual es incorrecta"))
			return
		}
		c.JSON(http.StatusBadRequest, entities.ErrorResponseSimple(err.Error()))
		return
	}

	c.JSON(http.StatusOK, entities.SuccessResponse("Contraseña actualizada exitosamente", nil))
}

// Logout maneja el endpoint POST /api/v1/auth/logout
// @Summary Cerrar sesión
// @Description Invalida el token actual (implementación básica)
// @Tags Authentication
// @Security BearerAuth
// @Success 200 {object} entities.APIResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Por ahora, solo retorna éxito
	// TODO: Implementar blacklist de tokens cuando sea necesario
	c.JSON(http.StatusOK, entities.SuccessResponse("Sesión cerrada exitosamente", nil))
}

// RefreshToken maneja el endpoint POST /api/v1/auth/refresh
// @Summary Renovar token
// @Description Genera un nuevo token JWT usando un refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body entities.RefreshTokenDTO true "Refresh token"
// @Success 200 {object} entities.APIResponse
// @Failure 401 {object} entities.APIResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req entities.RefreshTokenDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponseSimple("Datos inválidos"))
		return
	}

	response, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponseSimple("Refresh token inválido o expirado"))
		return
	}

	c.JSON(http.StatusOK, entities.SuccessResponse("Token renovado exitosamente", response))
}

// GetRoles maneja el endpoint GET /api/v1/auth/roles
// @Summary Listar roles disponibles
// @Description Retorna la lista de roles disponibles en el sistema
// @Tags Authentication
// @Security BearerAuth
// @Param activos_solo query boolean false "Solo roles activos" default(true)
// @Success 200 {object} entities.APIResponse
// @Router /api/v1/auth/roles [get]
func (h *AuthHandler) GetRoles(c *gin.Context) {
	activosOnly := c.DefaultQuery("activos_solo", "true") == "true"

	roles, err := h.authUseCase.GetRoles(c.Request.Context(), activosOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error al obtener roles"))
		return
	}

	c.JSON(http.StatusOK, entities.SuccessResponseWithMeta(
		"",
		roles,
		gin.H{"total": len(roles)},
	))
}

// CreateUser maneja el endpoint POST /api/v1/auth/users (Admin)
// @Summary Crear usuario (Admin)
// @Description Crea un nuevo usuario en el sistema (requiere permisos de Admin)
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body entities.CreateUsuarioDTO true "Datos del usuario"
// @Success 201 {object} entities.APIResponse
// @Failure 400 {object} entities.APIResponse
// @Failure 403 {object} entities.APIResponse
// @Router /api/v1/auth/users [post]
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req entities.CreateUsuarioDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Error en la validación",
			[]entities.ValidationError{{Field: "request", Message: err.Error()}},
		))
		return
	}

	// Validar RUT
	if !utils.ValidateRUT(req.RUT) {
		c.JSON(http.StatusBadRequest, entities.ErrorResponse(
			"Error en la validación",
			entities.ErrorDetail{Code: "INVALID_RUT_FORMAT", Details: "El formato del RUT no es válido"},
		))
		return
	}

	response, err := h.authUseCase.CreateUsuario(c.Request.Context(), req, true) // true = requiere admin
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "ya está registrado") {
			c.JSON(http.StatusBadRequest, entities.ErrorResponse(
				"Error en la validación",
				entities.ErrorDetail{Code: "DUPLICATE_USERNAME", Details: err.Error()},
			))
			return
		}
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error interno del servidor"))
		return
	}

	c.JSON(http.StatusCreated, entities.SuccessResponse("Usuario creado exitosamente", response))
}

// GetUser maneja el endpoint GET /api/v1/auth/users/:id_usuario (Admin)
// @Summary Obtener usuario por ID (Admin)
// @Description Obtiene información detallada de un usuario específico
// @Tags Admin
// @Security BearerAuth
// @Param id_usuario path int true "ID del usuario"
// @Success 200 {object} entities.APIResponse
// @Failure 404 {object} entities.APIResponse
// @Router /api/v1/auth/users/{id_usuario} [get]
func (h *AuthHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id_usuario")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponseSimple("ID de usuario inválido"))
		return
	}

	usuario, err := h.authUseCase.GetUsuarioCompleto(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, entities.ErrorResponseSimple("Usuario no encontrado"))
		return
	}

	c.JSON(http.StatusOK, entities.SuccessResponse("", usuario))
}

// UpdateUser maneja el endpoint PUT /api/v1/auth/users/:id_usuario (Admin)
// @Summary Actualizar usuario (Admin)
// @Description Actualiza información de un usuario existente
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id_usuario path int true "ID del usuario"
// @Param request body entities.UpdateUsuarioDTO true "Datos a actualizar"
// @Success 200 {object} entities.APIResponse
// @Failure 400 {object} entities.APIResponse
// @Failure 404 {object} entities.APIResponse
// @Router /api/v1/auth/users/{id_usuario} [put]
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id_usuario")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponseSimple("ID de usuario inválido"))
		return
	}

	var req entities.UpdateUsuarioDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Error en la validación",
			[]entities.ValidationError{{Field: "request", Message: err.Error()}},
		))
		return
	}

	usuario, err := h.authUseCase.UpdateUsuario(c.Request.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no encontrado") {
			c.JSON(http.StatusNotFound, entities.ErrorResponseSimple("Usuario no encontrado"))
			return
		}
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error al actualizar usuario"))
		return
	}

	c.JSON(http.StatusOK, entities.SuccessResponse("Usuario actualizado exitosamente", usuario))
}

// ListUsers maneja el endpoint GET /api/v1/auth/users (Admin)
// @Summary Listar usuarios (Admin)
// @Description Lista todos los usuarios del sistema con filtros y paginación
// @Tags Admin
// @Security BearerAuth
// @Param page query int false "Página" default(1)
// @Param limit query int false "Límite por página" default(20)
// @Param id_rol query int false "Filtrar por rol"
// @Param id_estado query int false "Filtrar por estado"
// @Param search query string false "Buscar en nombre_usuario, nombre_completo, rut"
// @Param sort_by query string false "Ordenar por campo" default(fecha_creacion)
// @Param order query string false "Orden (asc|desc)" default(desc)
// @Success 200 {object} entities.APIResponse
// @Router /api/v1/auth/users [get]
func (h *AuthHandler) ListUsers(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}

	var idRol *int
	if idRolStr := c.Query("id_rol"); idRolStr != "" {
		val, _ := strconv.Atoi(idRolStr)
		idRol = &val
	}

	var idEstado *int
	if idEstadoStr := c.Query("id_estado"); idEstadoStr != "" {
		val, _ := strconv.Atoi(idEstadoStr)
		idEstado = &val
	}

	filters := entities.ListUsuariosFilters{
		Page:     page,
		Limit:    limit,
		IDRol:    idRol,
		IDEstado: idEstado,
		Search:   c.Query("search"),
		SortBy:   c.DefaultQuery("sort_by", "fecha_creacion"),
		Order:    c.DefaultQuery("order", "desc"),
	}

	usuarios, total, err := h.authUseCase.ListUsuarios(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error al listar usuarios"))
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := entities.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, entities.SuccessResponseWithMeta("", usuarios, meta))
}
