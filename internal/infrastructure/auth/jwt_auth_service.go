package auth

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Sistal/ms-authentication/internal/domain/services"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTAuthService implementa AuthService usando JWT
type JWTAuthService struct {
	secretKey []byte
	expiresIn time.Duration
}

// NewJWTAuthService crea una nueva instancia del servicio de autenticación
func NewJWTAuthService(secretKey string) *JWTAuthService {
	return &JWTAuthService{
		secretKey: []byte(secretKey),
		expiresIn: 24 * time.Hour, // 24 horas de expiración según contrato
	}
}

// CustomClaims define los claims personalizados del JWT
type CustomClaims struct {
	UserID         int    `json:"sub"`
	Username       string `json:"nombre_usuario"`
	NombreCompleto string `json:"nombre_completo"`
	RUT            string `json:"rut"`
	Role           int    `json:"id_rol"`
	NombreRol      string `json:"nombre_rol"`
	jwt.RegisteredClaims
}

// GenerateToken genera un JWT para un usuario
func (s *JWTAuthService) GenerateToken(userID int, username string, nombreCompleto string, rut string, role int, nombreRol string) (string, error) {
	slog.Debug("[JWTAuthService.GenerateToken] Construyendo claims del JWT",
		slog.Int("user_id", userID),
		slog.String("username", username),
		slog.Int("role", role),
		slog.String("nombre_rol", nombreRol),
	)

	now := time.Now()
	expiresAt := now.Add(s.expiresIn)

	claims := CustomClaims{
		UserID:         userID,
		Username:       username,
		NombreCompleto: nombreCompleto,
		RUT:            rut,
		Role:           role,
		NombreRol:      nombreRol,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		slog.Error("[JWTAuthService.GenerateToken] Error al firmar el token JWT",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	slog.Debug("[JWTAuthService.GenerateToken] Token JWT firmado y generado correctamente",
		slog.Int("user_id", userID),
		slog.Time("expires_at", expiresAt),
	)
	return signedToken, nil
}

// ValidateToken valida un JWT y retorna los claims
func (s *JWTAuthService) ValidateToken(tokenString string) (*services.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el método de firma sea HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Verificar expiración
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return &services.TokenClaims{
		UserID:         claims.UserID,
		Username:       claims.Username,
		NombreCompleto: claims.NombreCompleto,
		RUT:            claims.RUT,
		Role:           claims.Role,
		NombreRol:      claims.NombreRol,
		IssuedAt:       claims.IssuedAt.Unix(),
		ExpiresAt:      claims.ExpiresAt.Unix(),
	}, nil
}

// HashPassword hashea una contraseña usando bcrypt
func (s *JWTAuthService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// ComparePassword compara una contraseña con su hash
func (s *JWTAuthService) ComparePassword(hashedPassword, password string) error {
	slog.Debug("[JWTAuthService.ComparePassword] Comparando contraseña con hash bcrypt")
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		slog.Warn("[JWTAuthService.ComparePassword] La contraseña no coincide con el hash almacenado")
		return fmt.Errorf("invalid password")
	}
	slog.Debug("[JWTAuthService.ComparePassword] Contraseña válida")
	return nil
}
