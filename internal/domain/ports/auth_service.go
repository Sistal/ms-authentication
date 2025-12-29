package ports

// AuthService defines the interface for authentication operations
type AuthService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
	GenerateToken(userID string, email string) (string, error)
	ValidateToken(token string) (string, error) // Returns userID
}
