package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Sistal/ms-authentication/internal/domain/entities"
	_ "github.com/lib/pq"
)

// PostgresUsuarioRepository implementa UsuarioRepository para PostgreSQL
type PostgresUsuarioRepository struct {
	db *sql.DB
}

// NewPostgresUsuarioRepository crea una nueva instancia del repositorio
func NewPostgresUsuarioRepository(db *sql.DB) *PostgresUsuarioRepository {
	return &PostgresUsuarioRepository{db: db}
}

// InitSchema inicializa el esquema de la base de datos
func (r *PostgresUsuarioRepository) InitSchema(ctx context.Context) error {
	queries := []string{
		// Tabla Usuario (nombre_usuario almacena el email del usuario)
		`CREATE TABLE IF NOT EXISTS "Usuario" (
			id_usuario SERIAL PRIMARY KEY,
			nombre_usuario VARCHAR(100) NOT NULL UNIQUE,
			nombre_completo VARCHAR(200),
			rut VARCHAR(20) UNIQUE,
			id_rol INT NOT NULL DEFAULT 1,
			id_estado_usuario INT NOT NULL DEFAULT 1,
			fecha_creacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_modificacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		// Tabla Credencial
		`CREATE TABLE IF NOT EXISTS "Credencial" (
			id_credencial SERIAL PRIMARY KEY,
			id_usuario INT NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			fecha_creacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_credencial_usuario 
				FOREIGN KEY (id_usuario) 
				REFERENCES "Usuario"(id_usuario) 
				ON DELETE CASCADE
		)`,
		// Crear índices
		`CREATE INDEX IF NOT EXISTS idx_usuario_nombre_usuario ON "Usuario"(nombre_usuario)`,
		`CREATE INDEX IF NOT EXISTS idx_usuario_rut ON "Usuario"(rut)`,
	}

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// CreateUsuario crea un nuevo usuario en la base de datos
func (r *PostgresUsuarioRepository) CreateUsuario(ctx context.Context, usuario *entities.Usuario) error {
	// Obtener el siguiente ID disponible
	var nextID int
	getNextIDQuery := `
		SELECT COALESCE(MAX(id_usuario), 0) + 1 
		FROM "Usuario"
	`
	
	err := r.db.QueryRowContext(ctx, getNextIDQuery).Scan(&nextID)
	if err != nil {
		return fmt.Errorf("failed to get next id: %w", err)
	}

	// Insertar el usuario con el ID generado
	query := `
		INSERT INTO "Usuario" (id_usuario, nombre_usuario, nombre_completo, rut, id_rol, id_estado_usuario, fecha_creacion, fecha_modificacion)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		nextID,
		usuario.NombreUsuario,
		usuario.NombreCompleto,
		usuario.Rut,
		usuario.IDRol,
		usuario.IDEstadoUsuario,
		usuario.FechaCreacion,
		usuario.FechaModificacion,
	)

	if err != nil {
		return fmt.Errorf("failed to create usuario: %w", err)
	}

	// Asignar el ID generado al objeto usuario
	usuario.IDUsuario = nextID

	return nil
}

// GetUsuarioByID obtiene un usuario por su ID
func (r *PostgresUsuarioRepository) GetUsuarioByID(ctx context.Context, id int) (*entities.Usuario, error) {
	query := `
		SELECT id_usuario, nombre_usuario, nombre_completo, rut, id_rol, id_estado_usuario, fecha_creacion, fecha_modificacion
		FROM "Usuario"
		WHERE id_usuario = $1
	`

	usuario := &entities.Usuario{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&usuario.IDUsuario,
		&usuario.NombreUsuario,
		&usuario.NombreCompleto,
		&usuario.Rut,
		&usuario.IDRol,
		&usuario.IDEstadoUsuario,
		&usuario.FechaCreacion,
		&usuario.FechaModificacion,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("usuario not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get usuario: %w", err)
	}

	return usuario, nil
}

// GetUsuarioByUsername obtiene un usuario por su nombre de usuario
func (r *PostgresUsuarioRepository) GetUsuarioByUsername(ctx context.Context, username string) (*entities.Usuario, error) {
	query := `
		SELECT id_usuario, nombre_usuario, nombre_completo, rut, id_rol, id_estado_usuario, fecha_creacion, fecha_modificacion
		FROM "Usuario"
		WHERE nombre_usuario = $1
	`

	usuario := &entities.Usuario{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&usuario.IDUsuario,
		&usuario.NombreUsuario,
		&usuario.NombreCompleto,
		&usuario.Rut,
		&usuario.IDRol,
		&usuario.IDEstadoUsuario,
		&usuario.FechaCreacion,
		&usuario.FechaModificacion,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("usuario not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get usuario: %w", err)
	}

	return usuario, nil
}

// GetUsuarioByRut obtiene un usuario por su RUT
func (r *PostgresUsuarioRepository) GetUsuarioByRut(ctx context.Context, rut string) (*entities.Usuario, error) {
	query := `
		SELECT id_usuario, nombre_usuario, nombre_completo, rut, id_rol, id_estado_usuario, fecha_creacion, fecha_modificacion
		FROM "Usuario"
		WHERE rut = $1
	`

	usuario := &entities.Usuario{}
	err := r.db.QueryRowContext(ctx, query, rut).Scan(
		&usuario.IDUsuario,
		&usuario.NombreUsuario,
		&usuario.NombreCompleto,
		&usuario.Rut,
		&usuario.IDRol,
		&usuario.IDEstadoUsuario,
		&usuario.FechaCreacion,
		&usuario.FechaModificacion,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("usuario not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get usuario: %w", err)
	}

	return usuario, nil
}

// CreateCredencial crea las credenciales de un usuario
func (r *PostgresUsuarioRepository) CreateCredencial(ctx context.Context, credencial *entities.Credencial) error {
	// Obtener el siguiente ID disponible
	var nextID int
	getNextIDQuery := `
		SELECT COALESCE(MAX(id_credencial), 0) + 1 
		FROM "Credencial"
	`
	
	err := r.db.QueryRowContext(ctx, getNextIDQuery).Scan(&nextID)
	if err != nil {
		return fmt.Errorf("failed to get next credencial id: %w", err)
	}

	// Insertar la credencial con el ID generado
	query := `
		INSERT INTO "Credencial" (id_credencial, id_usuario, password_hash, fecha_creacion)
		VALUES ($1, $2, $3, $4)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		nextID,
		credencial.IDUsuario,
		credencial.PasswordHash,
		credencial.FechaCreacion,
	)

	if err != nil {
		return fmt.Errorf("failed to create credencial: %w", err)
	}

	// Asignar el ID generado al objeto credencial
	credencial.IDCredencial = nextID

	return nil
}

// GetCredencialByUsuarioID obtiene las credenciales de un usuario
func (r *PostgresUsuarioRepository) GetCredencialByUsuarioID(ctx context.Context, idUsuario int) (*entities.Credencial, error) {
	query := `
		SELECT id_credencial, id_usuario, password_hash, fecha_creacion
		FROM "Credencial"
		WHERE id_usuario = $1
	`

	credencial := &entities.Credencial{}
	err := r.db.QueryRowContext(ctx, query, idUsuario).Scan(
		&credencial.IDCredencial,
		&credencial.IDUsuario,
		&credencial.PasswordHash,
		&credencial.FechaCreacion,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("credencial not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credencial: %w", err)
	}

	return credencial, nil
}
