package entities

// Rol entidad para la tabla Roles
type Rol struct {
	IDRol       int    `json:"id_rol" db:"id_rol"`
	NombreRol   string `json:"nombre_rol" db:"nombre_rol"`
	Descripcion string `json:"descripcion" db:"descripcion"`
	Activo      bool   `json:"activo" db:"activo"`
}

// Estado entidad para la tabla Estado
type Estado struct {
	IDEstado     int    `json:"id_estado" db:"id_estado"`
	NombreEstado string `json:"nombre_estado" db:"nombre_estado"`
	TablaEstado  string `json:"tabla_estado" db:"tabla_estado"`
}
