package handlers

import (
	"log/slog"
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
	clientIP := c.ClientIP()
	slog.Info("[Register] Solicitud de registro recibida",
		slog.String("endpoint", "POST /api/v1/auth/register"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", c.Request.UserAgent()),
	)

	var req entities.CreateUsuarioDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("[Register] Payload inválido o malformado",
			slog.String("client_ip", clientIP),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Datos de entrada inválidos",
			[]entities.ValidationError{{Field: "request", Message: err.Error()}},
		))
		return
	}

	slog.Debug("[Register] Payload recibido",
		slog.String("nombre_usuario", req.NombreUsuario),
		slog.String("rut", req.RUT),
		slog.Int("id_rol", req.IDRol),
	)

	// Validar RUT
	if !utils.ValidateRUT(req.RUT) {
		slog.Warn("[Register] Formato de RUT inválido",
			slog.String("rut", req.RUT),
			slog.String("client_ip", clientIP),
		)
		c.JSON(http.StatusBadRequest, entities.ErrorResponse(
			"Error en la validación",
			entities.ErrorDetail{Code: "INVALID_RUT_FORMAT", Details: "El formato del RUT no es válido"},
		))
		return
	}

	// Validar longitud de contraseña
	if len(req.Password) < utils.MinPasswordLength {
		slog.Warn("[Register] Contraseña demasiado corta",
			slog.String("nombre_usuario", req.NombreUsuario),
			slog.String("client_ip", clientIP),
			slog.Int("min_length", utils.MinPasswordLength),
		)
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Error en la validación",
			[]entities.ValidationError{{
				Field:   "password",
				Message: "La contraseña debe tener al menos 8 caracteres",
			}},
		))
		return
	}

	slog.Info("[Register] Validaciones básicas superadas, creando usuario",
		slog.String("nombre_usuario", req.NombreUsuario),
		slog.String("client_ip", clientIP),
	)

	response, err := h.authUseCase.CreateUsuario(c.Request.Context(), req, false) // false = no requiere ser admin
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "ya está registrado") {
			slog.Warn("[Register] Conflicto: usuario o RUT ya registrado",
				slog.String("nombre_usuario", req.NombreUsuario),
				slog.String("rut", req.RUT),
				slog.String("error", err.Error()),
				slog.Int("http_status", http.StatusBadRequest),
			)
			c.JSON(http.StatusBadRequest, entities.ErrorResponse(
				"Error en la validación",
				entities.ErrorDetail{Code: "DUPLICATE_USERNAME", Details: err.Error()},
			))
			return
		}
		slog.Error("[Register] Error interno al crear usuario",
			slog.String("nombre_usuario", req.NombreUsuario),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusInternalServerError),
		)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error interno del servidor"))
		return
	}

	slog.Info("[Register] Usuario registrado exitosamente",
		slog.String("nombre_usuario", req.NombreUsuario),
		slog.Int("http_status", http.StatusCreated),
	)
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
	clientIP := c.ClientIP()
	slog.Info("[Login] Solicitud de autenticación recibida",
		slog.String("endpoint", "POST /api/v1/auth/login"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", c.Request.UserAgent()),
	)

	var req entities.LoginDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("[Login] Payload inválido o malformado",
			slog.String("client_ip", clientIP),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Datos de entrada inválidos",
			[]entities.ValidationError{{Field: "nombre_usuario", Message: "El nombre de usuario es requerido"}},
		))
		return
	}

	slog.Info("[Login] Credenciales recibidas, delegando a usecase",
		slog.String("nombre_usuario", req.NombreUsuario),
		slog.String("client_ip", clientIP),
	)

	response, err := h.authUseCase.Login(c.Request.Context(), req)
	if err != nil {
		errorMsg := err.Error()

		// Usuario inactivo = 403
		if strings.Contains(errorMsg, "not active") || strings.Contains(errorMsg, "inactivo") {
			slog.Warn("[Login] Intento de acceso con cuenta inactiva",
				slog.String("nombre_usuario", req.NombreUsuario),
				slog.String("client_ip", clientIP),
				slog.Int("http_status", http.StatusForbidden),
			)
			c.JSON(http.StatusForbidden, entities.ErrorResponse(
				"Acceso denegado",
				entities.ErrorDetail{Code: "USER_INACTIVE", Details: "La cuenta está inactiva o suspendida"},
			))
			return
		}

		// Credenciales incorrectas = 401
		if strings.Contains(errorMsg, "invalid credentials") || strings.Contains(errorMsg, "credenciales incorrectas") {
			slog.Warn("[Login] Credenciales incorrectas",
				slog.String("nombre_usuario", req.NombreUsuario),
				slog.String("client_ip", clientIP),
				slog.Int("http_status", http.StatusUnauthorized),
			)
			c.JSON(http.StatusUnauthorized, entities.ErrorResponse(
				"Credenciales incorrectas",
				entities.ErrorDetail{Code: "INVALID_CREDENTIALS", Details: "Nombre de usuario o contraseña incorrectos"},
			))
			return
		}

		slog.Error("[Login] Error interno al procesar el login",
			slog.String("nombre_usuario", req.NombreUsuario),
			slog.String("client_ip", clientIP),
			slog.String("error", errorMsg),
			slog.Int("http_status", http.StatusInternalServerError),
		)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error interno del servidor"))
		return
	}

	slog.Info("[Login] Autenticación exitosa",
		slog.String("nombre_usuario", req.NombreUsuario),
		slog.String("client_ip", clientIP),
		slog.Int("http_status", http.StatusOK),
	)
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
	clientIP := c.ClientIP()
	slog.Info("[Validate] Solicitud de validación de token recibida",
		slog.String("endpoint", "GET /api/v1/auth/validate"),
		slog.String("client_ip", clientIP),
	)

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		slog.Warn("[Validate] Header Authorization ausente",
			slog.String("client_ip", clientIP),
			slog.Int("http_status", http.StatusUnauthorized),
		)
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse(
			"Token inválido o expirado",
			entities.ErrorDetail{Code: "INVALID_TOKEN", Details: "El token proporcionado no es válido"},
		))
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		slog.Warn("[Validate] Formato de Authorization header inválido",
			slog.String("client_ip", clientIP),
			slog.String("auth_header_prefix", parts[0]),
			slog.Int("http_status", http.StatusUnauthorized),
		)
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse(
			"Token inválido o expirado",
			entities.ErrorDetail{Code: "INVALID_TOKEN", Details: "El token proporcionado no es válido"},
		))
		return
	}

	token := parts[1]
	slog.Debug("[Validate] Token extraído, validando con usecase",
		slog.String("client_ip", clientIP),
	)

	response, err := h.authUseCase.ValidateTokenComplete(c.Request.Context(), token)
	if err != nil {
		slog.Warn("[Validate] Token inválido o expirado",
			slog.String("client_ip", clientIP),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusUnauthorized),
		)
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse(
			"Token inválido o expirado",
			entities.ErrorDetail{Code: "INVALID_TOKEN", Details: "El token proporcionado no es válido"},
		))
		return
	}

	slog.Info("[Validate] Token válido",
		slog.String("client_ip", clientIP),
		slog.Int("id_usuario", response.IDUsuario),
		slog.String("nombre_usuario", response.NombreUsuario),
		slog.Int("http_status", http.StatusOK),
	)
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
	clientIP := c.ClientIP()
	slog.Info("[GetMe] Solicitud de perfil propio recibida",
		slog.String("endpoint", "GET /api/v1/auth/me"),
		slog.String("client_ip", clientIP),
	)

	userID, exists := c.Get("user_id")
	if !exists {
		slog.Warn("[GetMe] user_id no encontrado en contexto (middleware no ejecutado o token inválido)",
			slog.String("client_ip", clientIP),
			slog.Int("http_status", http.StatusUnauthorized),
		)
		c.JSON(http.StatusUnauthorized, entities.ErrorResponseSimple("Usuario no autenticado"))
		return
	}

	slog.Debug("[GetMe] Obteniendo información completa del usuario desde BD",
		slog.Int("id_usuario", userID.(int)),
	)

	// Obtener usuario completo de la BD
	usuario, err := h.authUseCase.GetUsuarioCompleto(c.Request.Context(), userID.(int))
	if err != nil {
		slog.Error("[GetMe] Error al obtener información del usuario desde BD",
			slog.Int("id_usuario", userID.(int)),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusInternalServerError),
		)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error al obtener información del usuario"))
		return
	}

	slog.Info("[GetMe] Perfil de usuario retornado exitosamente",
		slog.Int("id_usuario", userID.(int)),
		slog.Int("http_status", http.StatusOK),
	)
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
	clientIP := c.ClientIP()
	slog.Info("[ChangePassword] Solicitud de cambio de contraseña recibida",
		slog.String("endpoint", "PUT /api/v1/auth/change-password"),
		slog.String("client_ip", clientIP),
	)

	userID, exists := c.Get("user_id")
	if !exists {
		slog.Warn("[ChangePassword] user_id no encontrado en contexto",
			slog.String("client_ip", clientIP),
			slog.Int("http_status", http.StatusUnauthorized),
		)
		c.JSON(http.StatusUnauthorized, entities.ErrorResponseSimple("Usuario no autenticado"))
		return
	}

	var req entities.ChangePasswordDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("[ChangePassword] Payload inválido o malformado",
			slog.Int("id_usuario", userID.(int)),
			slog.String("client_ip", clientIP),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Error en la validación",
			[]entities.ValidationError{{Field: "request", Message: err.Error()}},
		))
		return
	}

	// Validar que las contraseñas coincidan
	if req.PasswordNueva != req.PasswordConfirmacion {
		slog.Warn("[ChangePassword] Las contraseñas nueva y confirmación no coinciden",
			slog.Int("id_usuario", userID.(int)),
			slog.String("client_ip", clientIP),
		)
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Error en la validación",
			[]entities.ValidationError{{
				Field:   "password_confirmacion",
				Message: "Las contraseñas no coinciden",
			}},
		))
		return
	}

	slog.Debug("[ChangePassword] Validaciones de payload superadas, procesando cambio",
		slog.Int("id_usuario", userID.(int)),
	)

	err := h.authUseCase.ChangePassword(c.Request.Context(), userID.(int), req)
	if err != nil {
		if strings.Contains(err.Error(), "incorrect") || strings.Contains(err.Error(), "incorrecta") {
			slog.Warn("[ChangePassword] La contraseña actual proporcionada es incorrecta",
				slog.Int("id_usuario", userID.(int)),
				slog.String("client_ip", clientIP),
				slog.Int("http_status", http.StatusUnauthorized),
			)
			c.JSON(http.StatusUnauthorized, entities.ErrorResponseSimple("La contraseña actual es incorrecta"))
			return
		}
		slog.Error("[ChangePassword] Error al actualizar la contraseña",
			slog.Int("id_usuario", userID.(int)),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusBadRequest),
		)
		c.JSON(http.StatusBadRequest, entities.ErrorResponseSimple(err.Error()))
		return
	}

	slog.Info("[ChangePassword] Contraseña actualizada exitosamente",
		slog.Int("id_usuario", userID.(int)),
		slog.String("client_ip", clientIP),
		slog.Int("http_status", http.StatusOK),
	)
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
	clientIP := c.ClientIP()
	userID, _ := c.Get("user_id")

	slog.Info("[Logout] Solicitud de cierre de sesión recibida",
		slog.String("endpoint", "POST /api/v1/auth/logout"),
		slog.String("client_ip", clientIP),
		slog.Any("id_usuario", userID),
	)

	// Por ahora, solo retorna éxito
	// TODO: Implementar blacklist de tokens cuando sea necesario
	slog.Info("[Logout] Sesión cerrada exitosamente",
		slog.Any("id_usuario", userID),
		slog.String("client_ip", clientIP),
		slog.Int("http_status", http.StatusOK),
	)
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
	clientIP := c.ClientIP()
	slog.Info("[RefreshToken] Solicitud de renovación de token recibida",
		slog.String("endpoint", "POST /api/v1/auth/refresh"),
		slog.String("client_ip", clientIP),
	)

	var req entities.RefreshTokenDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("[RefreshToken] Payload inválido o malformado",
			slog.String("client_ip", clientIP),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusBadRequest),
		)
		c.JSON(http.StatusBadRequest, entities.ErrorResponseSimple("Datos inválidos"))
		return
	}

	slog.Debug("[RefreshToken] Refresh token recibido, validando con usecase",
		slog.String("client_ip", clientIP),
	)

	response, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		slog.Warn("[RefreshToken] Refresh token inválido o expirado",
			slog.String("client_ip", clientIP),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusUnauthorized),
		)
		c.JSON(http.StatusUnauthorized, entities.ErrorResponseSimple("Refresh token inválido o expirado"))
		return
	}

	slog.Info("[RefreshToken] Token renovado exitosamente",
		slog.String("client_ip", clientIP),
		slog.Int("http_status", http.StatusOK),
	)
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
	clientIP := c.ClientIP()
	activosOnly := c.DefaultQuery("activos_solo", "true") == "true"

	slog.Info("[GetRoles] Solicitud de listado de roles recibida",
		slog.String("endpoint", "GET /api/v1/auth/roles"),
		slog.String("client_ip", clientIP),
		slog.Bool("activos_solo", activosOnly),
	)

	roles, err := h.authUseCase.GetRoles(c.Request.Context(), activosOnly)
	if err != nil {
		slog.Error("[GetRoles] Error al obtener roles desde BD",
			slog.String("client_ip", clientIP),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusInternalServerError),
		)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error al obtener roles"))
		return
	}

	slog.Info("[GetRoles] Roles retornados exitosamente",
		slog.String("client_ip", clientIP),
		slog.Int("total_roles", len(roles)),
		slog.Int("http_status", http.StatusOK),
	)
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
	clientIP := c.ClientIP()
	adminID, _ := c.Get("user_id")

	slog.Info("[CreateUser] Solicitud admin de creación de usuario recibida",
		slog.String("endpoint", "POST /api/v1/auth/users"),
		slog.String("client_ip", clientIP),
		slog.Any("admin_id", adminID),
	)

	var req entities.CreateUsuarioDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("[CreateUser] Payload inválido o malformado",
			slog.String("client_ip", clientIP),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Error en la validación",
			[]entities.ValidationError{{Field: "request", Message: err.Error()}},
		))
		return
	}

	slog.Debug("[CreateUser] Payload recibido",
		slog.String("nombre_usuario", req.NombreUsuario),
		slog.String("rut", req.RUT),
		slog.Int("id_rol", req.IDRol),
	)

	// Validar RUT
	if !utils.ValidateRUT(req.RUT) {
		slog.Warn("[CreateUser] Formato de RUT inválido",
			slog.String("rut", req.RUT),
			slog.String("client_ip", clientIP),
		)
		c.JSON(http.StatusBadRequest, entities.ErrorResponse(
			"Error en la validación",
			entities.ErrorDetail{Code: "INVALID_RUT_FORMAT", Details: "El formato del RUT no es válido"},
		))
		return
	}

	slog.Info("[CreateUser] Validaciones básicas superadas, creando usuario",
		slog.String("nombre_usuario", req.NombreUsuario),
		slog.Any("admin_id", adminID),
	)

	response, err := h.authUseCase.CreateUsuario(c.Request.Context(), req, true) // true = requiere admin
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "ya está registrado") {
			slog.Warn("[CreateUser] Conflicto: usuario o RUT ya registrado",
				slog.String("nombre_usuario", req.NombreUsuario),
				slog.String("rut", req.RUT),
				slog.String("error", err.Error()),
				slog.Int("http_status", http.StatusBadRequest),
			)
			c.JSON(http.StatusBadRequest, entities.ErrorResponse(
				"Error en la validación",
				entities.ErrorDetail{Code: "DUPLICATE_USERNAME", Details: err.Error()},
			))
			return
		}
		slog.Error("[CreateUser] Error interno al crear usuario",
			slog.String("nombre_usuario", req.NombreUsuario),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusInternalServerError),
		)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error interno del servidor"))
		return
	}

	slog.Info("[CreateUser] Usuario creado exitosamente por admin",
		slog.String("nombre_usuario", req.NombreUsuario),
		slog.Any("admin_id", adminID),
		slog.Int("http_status", http.StatusCreated),
	)
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
	clientIP := c.ClientIP()
	idStr := c.Param("id_usuario")

	slog.Info("[GetUser] Solicitud admin de obtención de usuario recibida",
		slog.String("endpoint", "GET /api/v1/auth/users/:id_usuario"),
		slog.String("client_ip", clientIP),
		slog.String("id_usuario_param", idStr),
	)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("[GetUser] Parámetro id_usuario no es un entero válido",
			slog.String("id_usuario_param", idStr),
			slog.String("client_ip", clientIP),
			slog.Int("http_status", http.StatusBadRequest),
		)
		c.JSON(http.StatusBadRequest, entities.ErrorResponseSimple("ID de usuario inválido"))
		return
	}

	slog.Debug("[GetUser] Buscando usuario en BD",
		slog.Int("id_usuario", id),
	)

	usuario, err := h.authUseCase.GetUsuarioCompleto(c.Request.Context(), id)
	if err != nil {
		slog.Warn("[GetUser] Usuario no encontrado",
			slog.Int("id_usuario", id),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusNotFound),
		)
		c.JSON(http.StatusNotFound, entities.ErrorResponseSimple("Usuario no encontrado"))
		return
	}

	slog.Info("[GetUser] Usuario retornado exitosamente",
		slog.Int("id_usuario", id),
		slog.String("client_ip", clientIP),
		slog.Int("http_status", http.StatusOK),
	)
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
	clientIP := c.ClientIP()
	adminID, _ := c.Get("user_id")
	idStr := c.Param("id_usuario")

	slog.Info("[UpdateUser] Solicitud admin de actualización de usuario recibida",
		slog.String("endpoint", "PUT /api/v1/auth/users/:id_usuario"),
		slog.String("client_ip", clientIP),
		slog.String("id_usuario_param", idStr),
		slog.Any("admin_id", adminID),
	)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("[UpdateUser] Parámetro id_usuario no es un entero válido",
			slog.String("id_usuario_param", idStr),
			slog.String("client_ip", clientIP),
			slog.Int("http_status", http.StatusBadRequest),
		)
		c.JSON(http.StatusBadRequest, entities.ErrorResponseSimple("ID de usuario inválido"))
		return
	}

	var req entities.UpdateUsuarioDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("[UpdateUser] Payload inválido o malformado",
			slog.Int("id_usuario", id),
			slog.String("client_ip", clientIP),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, entities.ValidationErrorResponse(
			"Error en la validación",
			[]entities.ValidationError{{Field: "request", Message: err.Error()}},
		))
		return
	}

	slog.Debug("[UpdateUser] Payload recibido, actualizando usuario",
		slog.Int("id_usuario", id),
		slog.Any("admin_id", adminID),
	)

	usuario, err := h.authUseCase.UpdateUsuario(c.Request.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no encontrado") {
			slog.Warn("[UpdateUser] Usuario no encontrado para actualizar",
				slog.Int("id_usuario", id),
				slog.String("error", err.Error()),
				slog.Int("http_status", http.StatusNotFound),
			)
			c.JSON(http.StatusNotFound, entities.ErrorResponseSimple("Usuario no encontrado"))
			return
		}
		slog.Error("[UpdateUser] Error interno al actualizar usuario",
			slog.Int("id_usuario", id),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusInternalServerError),
		)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponseSimple("Error al actualizar usuario"))
		return
	}

	slog.Info("[UpdateUser] Usuario actualizado exitosamente",
		slog.Int("id_usuario", id),
		slog.Any("admin_id", adminID),
		slog.String("client_ip", clientIP),
		slog.Int("http_status", http.StatusOK),
	)
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
	clientIP := c.ClientIP()
	adminID, _ := c.Get("user_id")

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

	slog.Info("[ListUsers] Solicitud admin de listado de usuarios recibida",
		slog.String("endpoint", "GET /api/v1/auth/users"),
		slog.String("client_ip", clientIP),
		slog.Any("admin_id", adminID),
		slog.Int("page", page),
		slog.Int("limit", limit),
		slog.String("search", filters.Search),
		slog.String("sort_by", filters.SortBy),
		slog.String("order", filters.Order),
	)

	usuarios, total, err := h.authUseCase.ListUsuarios(c.Request.Context(), filters)
	if err != nil {
		slog.Error("[ListUsers] Error al listar usuarios desde BD",
			slog.String("client_ip", clientIP),
			slog.Any("admin_id", adminID),
			slog.String("error", err.Error()),
			slog.Int("http_status", http.StatusInternalServerError),
		)
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

	slog.Info("[ListUsers] Listado de usuarios retornado exitosamente",
		slog.String("client_ip", clientIP),
		slog.Any("admin_id", adminID),
		slog.Int("total", total),
		slog.Int("total_pages", totalPages),
		slog.Int("page", page),
		slog.Int("http_status", http.StatusOK),
	)
	c.JSON(http.StatusOK, entities.SuccessResponseWithMeta("", usuarios, meta))
}
