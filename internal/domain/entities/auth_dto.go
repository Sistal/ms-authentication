package entities

import "time"

// LoginDTO estructura para solicitud de login
type LoginDTO struct {
	NombreUsuario string `json:"nombre_usuario" binding:"required"`
	Password      string `json:"password" binding:"required"`
}

// TokenResponseDTO estructura para respuesta con token
type TokenResponseDTO struct {
	Token        string `json:"token"`
	ExpiresIn    int    `json:"expires_in"` // segundos
	TokenType    string `json:"token_type"` // "Bearer"
	RefreshToken string `json:"refresh_token,omitempty"`
}

// RolDTO estructura para información de rol
type RolDTO struct {
	IDRol       int    `json:"id_rol"`
	NombreRol   string `json:"nombre_rol"`
	Descripcion string `json:"descripcion,omitempty"`
}

// EstadoDTO estructura para información de estado
type EstadoDTO struct {
	IDEstado     int    `json:"id_estado"`
	NombreEstado string `json:"nombre_estado"`
	TablaEstado  string `json:"tabla_estado,omitempty"`
}

// UsuarioResponseDTO estructura para respuesta con información del usuario
type UsuarioResponseDTO struct {
	IDUsuario         int       `json:"id_usuario"`
	NombreUsuario     string    `json:"nombre_usuario"`
	NombreCompleto    string    `json:"nombre_completo"`
	RUT               string    `json:"rut"`
	Rol               RolDTO    `json:"rol"`
	Estado            EstadoDTO `json:"estado"`
	FechaCreacion     string    `json:"fecha_creacion,omitempty"`
	FechaModificacion *string   `json:"fecha_modificacion,omitempty"`
}

// LoginResponseDTO estructura completa para respuesta de login
type LoginResponseDTO struct {
	Usuario   UsuarioResponseDTO `json:"usuario"`
	Token     string             `json:"token"`
	ExpiresIn int                `json:"expires_in"`
	TokenType string             `json:"token_type"`
}

// ValidateTokenResponseDTO estructura para respuesta de validación de token
type ValidateTokenResponseDTO struct {
	IDUsuario       int    `json:"id_usuario"`
	NombreUsuario   string `json:"nombre_usuario"`
	NombreCompleto  string `json:"nombre_completo"`
	RUT             string `json:"rut"`
	IDRol           int    `json:"id_rol"`
	NombreRol       string `json:"nombre_rol"`
	IDEstadoUsuario int    `json:"id_estado_usuario"`
	NombreEstado    string `json:"nombre_estado"`
	Exp             int64  `json:"exp"` // timestamp
	Iat             int64  `json:"iat"` // timestamp
}

// ChangePasswordDTO estructura para cambio de contraseña
type ChangePasswordDTO struct {
	PasswordActual       string `json:"password_actual" binding:"required"`
	PasswordNueva        string `json:"password_nueva" binding:"required,min=8"`
	PasswordConfirmacion string `json:"password_confirmacion" binding:"required"`
}

// RefreshTokenDTO estructura para solicitud de refresh token
type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// CreateUsuarioDTO estructura para crear un usuario (Admin)
type CreateUsuarioDTO struct {
	NombreUsuario   string `json:"nombre_usuario" binding:"required"`
	NombreCompleto  string `json:"nombre_completo" binding:"required"`
	RUT             string `json:"rut" binding:"required"`
	Password        string `json:"password" binding:"required,min=8"`
	IDRol           int    `json:"id_rol" binding:"required"`
	IDEstadoUsuario *int   `json:"id_estado_usuario,omitempty"` // opcional, default: 1 (Activo)
}

// UpdateUsuarioDTO estructura para actualizar un usuario (Admin)
type UpdateUsuarioDTO struct {
	NombreCompleto  *string `json:"nombre_completo,omitempty"`
	IDRol           *int    `json:"id_rol,omitempty"`
	IDEstadoUsuario *int    `json:"id_estado_usuario,omitempty"`
}

// ListUsuariosFilters estructura para filtros de listado de usuarios
type ListUsuariosFilters struct {
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	IDRol    *int   `json:"id_rol,omitempty"`
	IDEstado *int   `json:"id_estado,omitempty"`
	Search   string `json:"search,omitempty"`
	SortBy   string `json:"sort_by"`
	Order    string `json:"order"`
}

// PaginationMeta estructura para metadatos de paginación
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// RefreshToken entidad para almacenar refresh tokens
type RefreshToken struct {
	ID        int       `json:"id"`
	IDUsuario int       `json:"id_usuario"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
