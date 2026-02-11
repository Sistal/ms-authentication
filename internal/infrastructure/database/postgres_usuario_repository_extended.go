package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Sistal/ms-authentication/internal/domain/entities"
)

// UpdateUsuario actualiza un usuario existente
func (r *PostgresUsuarioRepository) UpdateUsuario(ctx context.Context, id int, usuario *entities.Usuario) error {
	query := `
		UPDATE "Usuario"
		SET nombre_completo = $1,
		    id_rol = $2,
		    id_estado_usuario = $3,
		    fecha_modificacion = $4
		WHERE id_usuario = $5
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		usuario.NombreCompleto,
		usuario.IDRol,
		usuario.IDEstadoUsuario,
		usuario.FechaModificacion,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update usuario: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("usuario not found")
	}

	return nil
}

// UpdateCredencial actualiza la contraseña de un usuario
func (r *PostgresUsuarioRepository) UpdateCredencial(ctx context.Context, idUsuario int, passwordHash string) error {
	query := `
		UPDATE "Credencial"
		SET password_hash = $1
		WHERE id_usuario = $2
	`

	result, err := r.db.ExecContext(ctx, query, passwordHash, idUsuario)
	if err != nil {
		return fmt.Errorf("failed to update credencial: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("credencial not found")
	}

	return nil
}

// ListUsuarios lista usuarios con filtros y paginación
func (r *PostgresUsuarioRepository) ListUsuarios(ctx context.Context, filters entities.ListUsuariosFilters) ([]*entities.Usuario, int, error) {
	// Construir query base
	baseQuery := `FROM "Usuario" WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Aplicar filtros
	if filters.IDRol != nil {
		baseQuery += fmt.Sprintf(" AND id_rol = $%d", argCount)
		args = append(args, *filters.IDRol)
		argCount++
	}

	if filters.IDEstado != nil {
		baseQuery += fmt.Sprintf(" AND id_estado_usuario = $%d", argCount)
		args = append(args, *filters.IDEstado)
		argCount++
	}

	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		baseQuery += fmt.Sprintf(" AND (nombre_usuario ILIKE $%d OR nombre_completo ILIKE $%d OR rut ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, searchPattern)
		argCount++
	}

	// Contar total
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count usuarios: %w", err)
	}

	// Construir query con ordenamiento y paginación
	orderClause := " ORDER BY "
	switch filters.SortBy {
	case "nombre_usuario":
		orderClause += "nombre_usuario"
	case "nombre_completo":
		orderClause += "nombre_completo"
	case "id_rol":
		orderClause += "id_rol"
	default:
		orderClause += "fecha_creacion"
	}

	if strings.ToUpper(filters.Order) == "ASC" {
		orderClause += " ASC"
	} else {
		orderClause += " DESC"
	}

	offset := (filters.Page - 1) * filters.Limit
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filters.Limit, offset)

	selectQuery := "SELECT id_usuario, nombre_usuario, nombre_completo, rut, id_rol, id_estado_usuario, fecha_creacion, fecha_modificacion " +
		baseQuery + orderClause + paginationClause

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query usuarios: %w", err)
	}
	defer rows.Close()

	usuarios := []*entities.Usuario{}
	for rows.Next() {
		usuario := &entities.Usuario{}
		err := rows.Scan(
			&usuario.IDUsuario,
			&usuario.NombreUsuario,
			&usuario.NombreCompleto,
			&usuario.Rut,
			&usuario.IDRol,
			&usuario.IDEstadoUsuario,
			&usuario.FechaCreacion,
			&usuario.FechaModificacion,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan usuario: %w", err)
		}
		usuarios = append(usuarios, usuario)
	}

	return usuarios, total, nil
}

// GetRolByID obtiene un rol por su ID
func (r *PostgresUsuarioRepository) GetRolByID(ctx context.Context, idRol int) (*entities.Rol, error) {
	query := `
		SELECT id_rol, nombre_rol, descripcion, activo
		FROM "Roles"
		WHERE id_rol = $1
	`

	rol := &entities.Rol{}
	err := r.db.QueryRowContext(ctx, query, idRol).Scan(
		&rol.IDRol,
		&rol.NombreRol,
		&rol.Descripcion,
		&rol.Activo,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rol not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rol: %w", err)
	}

	return rol, nil
}

// GetEstadoByID obtiene un estado por su ID
func (r *PostgresUsuarioRepository) GetEstadoByID(ctx context.Context, idEstado int) (*entities.Estado, error) {
	query := `
		SELECT id_estado, nombre_estado, tabla_estado
		FROM "Estado"
		WHERE id_estado = $1
	`

	estado := &entities.Estado{}
	err := r.db.QueryRowContext(ctx, query, idEstado).Scan(
		&estado.IDEstado,
		&estado.NombreEstado,
		&estado.TablaEstado,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("estado not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get estado: %w", err)
	}

	return estado, nil
}

// ListRoles lista todos los roles
func (r *PostgresUsuarioRepository) ListRoles(ctx context.Context, activosOnly bool) ([]*entities.Rol, error) {
	query := `
		SELECT id_rol, nombre_rol, descripcion, activo
		FROM "Roles"
	`

	if activosOnly {
		query += " WHERE activo = true"
	}

	query += " ORDER BY id_rol"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	roles := []*entities.Rol{}
	for rows.Next() {
		rol := &entities.Rol{}
		err := rows.Scan(
			&rol.IDRol,
			&rol.NombreRol,
			&rol.Descripcion,
			&rol.Activo,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rol: %w", err)
		}
		roles = append(roles, rol)
	}

	return roles, nil
}

// CreateRefreshToken crea un refresh token
func (r *PostgresUsuarioRepository) CreateRefreshToken(ctx context.Context, refreshToken *entities.RefreshToken) error {
	query := `
		INSERT INTO "RefreshTokens" (id_usuario, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		refreshToken.IDUsuario,
		refreshToken.Token,
		refreshToken.ExpiresAt,
		refreshToken.CreatedAt,
	).Scan(&refreshToken.ID)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken obtiene un refresh token por su valor
func (r *PostgresUsuarioRepository) GetRefreshToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	query := `
		SELECT id, id_usuario, token, expires_at, created_at
		FROM "RefreshTokens"
		WHERE token = $1
	`

	refreshToken := &entities.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&refreshToken.ID,
		&refreshToken.IDUsuario,
		&refreshToken.Token,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return refreshToken, nil
}

// DeleteRefreshToken elimina un refresh token
func (r *PostgresUsuarioRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `
		DELETE FROM "RefreshTokens"
		WHERE token = $1
	`

	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// DeleteExpiredRefreshTokens elimina tokens expirados (función auxiliar para limpieza)
func (r *PostgresUsuarioRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	query := `
		DELETE FROM "RefreshTokens"
		WHERE expires_at < $1
	`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}
