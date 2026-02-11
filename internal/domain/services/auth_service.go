package services

// AuthService define las operaciones de autenticación
type AuthService interface {
	// GenerateToken genera un JWT para un usuario
	GenerateToken(userID int, username string, nombreCompleto string, rut string, role int, nombreRol string) (string, error)

	// ValidateToken valida un JWT y retorna los claims
	ValidateToken(token string) (*TokenClaims, error)

	// HashPassword hashea una contraseña usando bcrypt
	HashPassword(password string) (string, error)

	// ComparePassword compara una contraseña con su hash
	ComparePassword(hashedPassword, password string) error
}

// TokenClaims representa los claims del JWT
type TokenClaims struct {
	UserID         int    `json:"sub"`
	Username       string `json:"nombre_usuario"`
	NombreCompleto string `json:"nombre_completo"`
	RUT            string `json:"rut"`
	Role           int    `json:"id_rol"`
	NombreRol      string `json:"nombre_rol"`
	IssuedAt       int64  `json:"iat"`
	ExpiresAt      int64  `json:"exp"`
}
