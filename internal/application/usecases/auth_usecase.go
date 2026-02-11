package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/Sistal/ms-authentication/internal/domain/entities"
	"github.com/Sistal/ms-authentication/internal/domain/repositories"
	"github.com/Sistal/ms-authentication/internal/domain/services"
	"github.com/Sistal/ms-authentication/internal/utils"
	"github.com/google/uuid"
)

// AuthUseCase maneja los casos de uso de autenticación
type AuthUseCase struct {
	usuarioRepo repositories.UsuarioRepository
	authService services.AuthService
}

// NewAuthUseCase crea una nueva instancia del caso de uso
func NewAuthUseCase(
	usuarioRepo repositories.UsuarioRepository,
	authService services.AuthService,
) *AuthUseCase {
	return &AuthUseCase{
		usuarioRepo: usuarioRepo,
		authService: authService,
	}
}

// CreateUsuario crea un nuevo usuario (puede ser público o admin)
func (uc *AuthUseCase) CreateUsuario(ctx context.Context, req entities.CreateUsuarioDTO, requireAdmin bool) (*entities.UsuarioResponseDTO, error) {
	// Validar que el username no exista
	existingUser, _ := uc.usuarioRepo.GetUsuarioByUsername(ctx, req.NombreUsuario)
	if existingUser != nil {
		return nil, fmt.Errorf("el nombre de usuario ya está registrado")
	}

	// Validar que el RUT no exista
	existingRut, _ := uc.usuarioRepo.GetUsuarioByRut(ctx, req.RUT)
	if existingRut != nil {
		return nil, fmt.Errorf("el RUT ya está registrado")
	}

	// Validar RUT
	if !utils.ValidateRUT(req.RUT) {
		return nil, fmt.Errorf("el formato del RUT no es válido")
	}

	// Hashear password
	hashedPassword, err := uc.authService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Establecer estado por defecto si no se proporciona
	idEstado := 1 // Activo por defecto
	if req.IDEstadoUsuario != nil {
		idEstado = *req.IDEstadoUsuario
	}

	// Crear usuario
	now := time.Now()
	usuario := &entities.Usuario{
		NombreUsuario:     req.NombreUsuario,
		NombreCompleto:    req.NombreCompleto,
		Rut:               utils.FormatRUT(req.RUT),
		IDRol:             req.IDRol,
		IDEstadoUsuario:   idEstado,
		FechaCreacion:     now,
		FechaModificacion: now,
	}

	if err := uc.usuarioRepo.CreateUsuario(ctx, usuario); err != nil {
		return nil, fmt.Errorf("failed to create usuario: %w", err)
	}

	// Crear credencial
	credencial := &entities.Credencial{
		IDUsuario:     usuario.IDUsuario,
		PasswordHash:  hashedPassword,
		FechaCreacion: now,
	}

	if err := uc.usuarioRepo.CreateCredencial(ctx, credencial); err != nil {
		return nil, fmt.Errorf("failed to create credencial: %w", err)
	}

	// Obtener rol y estado
	rol, _ := uc.usuarioRepo.GetRolByID(ctx, usuario.IDRol)
	estado, _ := uc.usuarioRepo.GetEstadoByID(ctx, usuario.IDEstadoUsuario)

	return buildUsuarioResponse(usuario, rol, estado), nil
}

// Login autentica un usuario y genera un token JWT según el contrato
func (uc *AuthUseCase) Login(ctx context.Context, req entities.LoginDTO) (*entities.LoginResponseDTO, error) {
	// Obtener usuario por username o RUT
	usuario, err := uc.usuarioRepo.GetUsuarioByUsername(ctx, req.NombreUsuario)
	if err != nil {
		// Intentar por RUT
		usuario, err = uc.usuarioRepo.GetUsuarioByRut(ctx, req.NombreUsuario)
		if err != nil {
			return nil, fmt.Errorf("invalid credentials")
		}
	}

	// Verificar estado del usuario (debe retornar 403 si no está activo)
	if usuario.IDEstadoUsuario != 1 {
		return nil, fmt.Errorf("user is not active")
	}

	// Obtener credenciales
	credencial, err := uc.usuarioRepo.GetCredencialByUsuarioID(ctx, usuario.IDUsuario)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Comparar password
	if err := uc.authService.ComparePassword(credencial.PasswordHash, req.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Obtener rol para incluir nombre en el token
	rol, err := uc.usuarioRepo.GetRolByID(ctx, usuario.IDRol)
	if err != nil {
		return nil, fmt.Errorf("failed to get rol: %w", err)
	}

	estado, _ := uc.usuarioRepo.GetEstadoByID(ctx, usuario.IDEstadoUsuario)

	// Generar token con todos los claims requeridos
	token, err := uc.authService.GenerateToken(
		usuario.IDUsuario,
		usuario.NombreUsuario,
		usuario.NombreCompleto,
		usuario.Rut,
		usuario.IDRol,
		rol.NombreRol,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generar refresh token
	refreshToken := uuid.New().String()
	refreshTokenEntity := &entities.RefreshToken{
		IDUsuario: usuario.IDUsuario,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 días
		CreatedAt: time.Now(),
	}
	_ = uc.usuarioRepo.CreateRefreshToken(ctx, refreshTokenEntity)

	// Construir respuesta según contrato
	return &entities.LoginResponseDTO{
		Usuario:   *buildUsuarioResponse(usuario, rol, estado),
		Token:     token,
		ExpiresIn: 86400, // 24 horas en segundos
		TokenType: "Bearer",
	}, nil
}

// ValidateTokenComplete valida un token y retorna información completa del usuario
func (uc *AuthUseCase) ValidateTokenComplete(ctx context.Context, tokenString string) (*entities.ValidateTokenResponseDTO, error) {
	claims, err := uc.authService.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Retornar claims completos según contrato
	return &entities.ValidateTokenResponseDTO{
		IDUsuario:       claims.UserID,
		NombreUsuario:   claims.Username,
		NombreCompleto:  claims.NombreCompleto,
		RUT:             claims.RUT,
		IDRol:           claims.Role,
		NombreRol:       claims.NombreRol,
		IDEstadoUsuario: 1, // TODO: obtener de BD si es necesario
		NombreEstado:    "Activo",
		Exp:             claims.ExpiresAt,
		Iat:             claims.IssuedAt,
	}, nil
}

// GetUsuarioCompleto obtiene información completa del usuario desde la BD
func (uc *AuthUseCase) GetUsuarioCompleto(ctx context.Context, userID int) (*entities.UsuarioResponseDTO, error) {
	usuario, err := uc.usuarioRepo.GetUsuarioByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usuario no encontrado")
	}

	rol, _ := uc.usuarioRepo.GetRolByID(ctx, usuario.IDRol)
	estado, _ := uc.usuarioRepo.GetEstadoByID(ctx, usuario.IDEstadoUsuario)

	return buildUsuarioResponse(usuario, rol, estado), nil
}

// ChangePassword cambia la contraseña de un usuario
func (uc *AuthUseCase) ChangePassword(ctx context.Context, userID int, req entities.ChangePasswordDTO) error {
	// Obtener credencial actual
	credencial, err := uc.usuarioRepo.GetCredencialByUsuarioID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get credencial: %w", err)
	}

	// Verificar contraseña actual
	if err := uc.authService.ComparePassword(credencial.PasswordHash, req.PasswordActual); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Validar nueva contraseña
	if len(req.PasswordNueva) < utils.MinPasswordLength {
		return fmt.Errorf("la contraseña debe tener al menos %d caracteres", utils.MinPasswordLength)
	}

	// Hashear nueva contraseña
	hashedPassword, err := uc.authService.HashPassword(req.PasswordNueva)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Actualizar credencial
	if err := uc.usuarioRepo.UpdateCredencial(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update credencial: %w", err)
	}

	return nil
}

// RefreshToken renueva el access token usando un refresh token
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshTokenString string) (*entities.TokenResponseDTO, error) {
	// Buscar refresh token en BD
	refreshToken, err := uc.usuarioRepo.GetRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Verificar expiración
	if refreshToken.ExpiresAt.Before(time.Now()) {
		_ = uc.usuarioRepo.DeleteRefreshToken(ctx, refreshTokenString)
		return nil, fmt.Errorf("refresh token expired")
	}

	// Obtener usuario
	usuario, err := uc.usuarioRepo.GetUsuarioByID(ctx, refreshToken.IDUsuario)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Verificar que el usuario esté activo
	if usuario.IDEstadoUsuario != 1 {
		return nil, fmt.Errorf("user is not active")
	}

	// Obtener rol
	rol, err := uc.usuarioRepo.GetRolByID(ctx, usuario.IDRol)
	if err != nil {
		return nil, fmt.Errorf("failed to get rol: %w", err)
	}

	// Generar nuevo access token
	token, err := uc.authService.GenerateToken(
		usuario.IDUsuario,
		usuario.NombreUsuario,
		usuario.NombreCompleto,
		usuario.Rut,
		usuario.IDRol,
		rol.NombreRol,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &entities.TokenResponseDTO{
		Token:     token,
		ExpiresIn: 86400,
		TokenType: "Bearer",
	}, nil
}

// GetRoles obtiene la lista de roles
func (uc *AuthUseCase) GetRoles(ctx context.Context, activosOnly bool) ([]*entities.RolDTO, error) {
	roles, err := uc.usuarioRepo.ListRoles(ctx, activosOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	result := make([]*entities.RolDTO, len(roles))
	for i, rol := range roles {
		result[i] = &entities.RolDTO{
			IDRol:       rol.IDRol,
			NombreRol:   rol.NombreRol,
			Descripcion: rol.Descripcion,
		}
	}

	return result, nil
}

// UpdateUsuario actualiza un usuario existente
func (uc *AuthUseCase) UpdateUsuario(ctx context.Context, id int, req entities.UpdateUsuarioDTO) (*entities.UsuarioResponseDTO, error) {
	// Obtener usuario actual
	usuario, err := uc.usuarioRepo.GetUsuarioByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usuario no encontrado")
	}

	// Actualizar campos si se proporcionan
	if req.NombreCompleto != nil {
		usuario.NombreCompleto = *req.NombreCompleto
	}
	if req.IDRol != nil {
		usuario.IDRol = *req.IDRol
	}
	if req.IDEstadoUsuario != nil {
		usuario.IDEstadoUsuario = *req.IDEstadoUsuario
	}

	usuario.FechaModificacion = time.Now()

	// Actualizar en BD
	if err := uc.usuarioRepo.UpdateUsuario(ctx, id, usuario); err != nil {
		return nil, fmt.Errorf("failed to update usuario: %w", err)
	}

	// Obtener información completa
	return uc.GetUsuarioCompleto(ctx, id)
}

// ListUsuarios lista usuarios con filtros y paginación
func (uc *AuthUseCase) ListUsuarios(ctx context.Context, filters entities.ListUsuariosFilters) ([]*entities.UsuarioResponseDTO, int, error) {
	usuarios, total, err := uc.usuarioRepo.ListUsuarios(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list usuarios: %w", err)
	}

	result := make([]*entities.UsuarioResponseDTO, len(usuarios))
	for i, usuario := range usuarios {
		rol, _ := uc.usuarioRepo.GetRolByID(ctx, usuario.IDRol)
		estado, _ := uc.usuarioRepo.GetEstadoByID(ctx, usuario.IDEstadoUsuario)
		result[i] = buildUsuarioResponse(usuario, rol, estado)
	}

	return result, total, nil
}

// buildUsuarioResponse construye un UsuarioResponseDTO desde entidades
func buildUsuarioResponse(usuario *entities.Usuario, rol *entities.Rol, estado *entities.Estado) *entities.UsuarioResponseDTO {
	response := &entities.UsuarioResponseDTO{
		IDUsuario:      usuario.IDUsuario,
		NombreUsuario:  usuario.NombreUsuario,
		NombreCompleto: usuario.NombreCompleto,
		RUT:            usuario.Rut,
		FechaCreacion:  usuario.FechaCreacion.Format("2006-01-02"),
	}

	if rol != nil {
		response.Rol = entities.RolDTO{
			IDRol:       rol.IDRol,
			NombreRol:   rol.NombreRol,
			Descripcion: rol.Descripcion,
		}
	}

	if estado != nil {
		response.Estado = entities.EstadoDTO{
			IDEstado:     estado.IDEstado,
			NombreEstado: estado.NombreEstado,
			TablaEstado:  estado.TablaEstado,
		}
	}

	if !usuario.FechaModificacion.IsZero() && usuario.FechaModificacion != usuario.FechaCreacion {
		fechaModif := usuario.FechaModificacion.Format("2006-01-02")
		response.FechaModificacion = &fechaModif
	}

	return response
}
