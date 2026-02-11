package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Sistal/ms-authentication/config"
	_ "github.com/Sistal/ms-authentication/docs"
	"github.com/Sistal/ms-authentication/internal/application/usecases"
	"github.com/Sistal/ms-authentication/internal/infrastructure/auth"
	"github.com/Sistal/ms-authentication/internal/infrastructure/database"
	"github.com/Sistal/ms-authentication/internal/infrastructure/http/handlers"
	"github.com/Sistal/ms-authentication/internal/infrastructure/http/routes"
	_ "github.com/lib/pq"
)

// @title ms-authentication API
// @version 1.0
// @description Microservicio de autenticación con JWT para el sistema Sistal
// @description
// @description Este servicio proporciona endpoints para:
// @description - Registro de usuarios
// @description - Login con JWT
// @description - Validación de tokens
// @description - Gestión de autenticación y autorización
//
// @contact.name Sistal API Support
// @contact.email support@sistal.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Token JWT en formato: Bearer {token}

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Initialize database schema
	usuarioRepo := database.NewPostgresUsuarioRepository(db)
	if err := usuarioRepo.InitSchema(context.Background()); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}
	log.Println("Database schema initialized")

	// Initialize services
	authService := auth.NewJWTAuthService(cfg.JWT.SecretKey)

	// Initialize use cases
	authUseCase := usecases.NewAuthUseCase(usuarioRepo, authService)

	// Initialize HTTP handlers
	authHandler := handlers.NewAuthHandler(authUseCase)

	// Setup router
	router := routes.SetupRouter(authHandler, authService, cfg.CORS.AllowedOrigins)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
