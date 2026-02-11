package repositories

import (
	"context"

	"github.com/Sistal/ms-authentication/internal/domain/entities"
)

// UsuarioRepository define las operaciones de persistencia para Usuario
type UsuarioRepository interface {
	// CreateUsuario crea un nuevo usuario en la base de datos
	CreateUsuario(ctx context.Context, usuario *entities.Usuario) error

	// GetUsuarioByID obtiene un usuario por su ID
	GetUsuarioByID(ctx context.Context, id int) (*entities.Usuario, error)

	// GetUsuarioByUsername obtiene un usuario por su nombre de usuario (email)
	GetUsuarioByUsername(ctx context.Context, username string) (*entities.Usuario, error)

	// GetUsuarioByRut obtiene un usuario por su RUT
	GetUsuarioByRut(ctx context.Context, rut string) (*entities.Usuario, error)

	// UpdateUsuario actualiza un usuario existente
	UpdateUsuario(ctx context.Context, id int, usuario *entities.Usuario) error

	// UpdateCredencial actualiza las credenciales de un usuario
	UpdateCredencial(ctx context.Context, idUsuario int, passwordHash string) error

	// ListUsuarios lista usuarios con filtros y paginación
	ListUsuarios(ctx context.Context, filters entities.ListUsuariosFilters) ([]*entities.Usuario, int, error)

	// GetRolByID obtiene un rol por su ID
	GetRolByID(ctx context.Context, idRol int) (*entities.Rol, error)

	// GetEstadoByID obtiene un estado por su ID
	GetEstadoByID(ctx context.Context, idEstado int) (*entities.Estado, error)

	// ListRoles lista todos los roles
	ListRoles(ctx context.Context, activosOnly bool) ([]*entities.Rol, error)

	// CreateCredencial crea las credenciales de un usuario
	CreateCredencial(ctx context.Context, credencial *entities.Credencial) error

	// GetCredencialByUsuarioID obtiene las credenciales de un usuario
	GetCredencialByUsuarioID(ctx context.Context, idUsuario int) (*entities.Credencial, error)

	// CreateRefreshToken crea un refresh token
	CreateRefreshToken(ctx context.Context, refreshToken *entities.RefreshToken) error

	// GetRefreshToken obtiene un refresh token por su valor
	GetRefreshToken(ctx context.Context, token string) (*entities.RefreshToken, error)

	// DeleteRefreshToken elimina un refresh token
	DeleteRefreshToken(ctx context.Context, token string) error

	// InitSchema inicializa el esquema de la base de datos
	InitSchema(ctx context.Context) error
}
