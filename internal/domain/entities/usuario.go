package entities

import "time"

// Usuario representa la entidad Usuario del sistema
type Usuario struct {
	IDUsuario         int       `json:"id_usuario"`
	NombreUsuario     string    `json:"nombre_usuario"` // Almacena el email del usuario
	NombreCompleto    string    `json:"nombre_completo"`
	Rut               string    `json:"rut"`
	IDRol             int       `json:"id_rol"`
	IDEstadoUsuario   int       `json:"id_estado_usuario"`
	FechaCreacion     time.Time `json:"fecha_creacion"`
	FechaModificacion time.Time `json:"fecha_modificacion"`
}

// Credencial representa las credenciales de acceso de un usuario
type Credencial struct {
	IDCredencial  int       `json:"id_credencial"`
	IDUsuario     int       `json:"id_usuario"`
	PasswordHash  string    `json:"-"` // No exponer en JSON
	FechaCreacion time.Time `json:"fecha_creacion"`
}
