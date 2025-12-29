package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Sistal/ms-authentication/config"
	"github.com/Sistal/ms-authentication/internal/application/usecases"
	"github.com/Sistal/ms-authentication/internal/infrastructure/adapters/database"
	"github.com/Sistal/ms-authentication/internal/infrastructure/adapters/http"
)

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
	userRepo := database.NewPostgresUserRepository(db)
	if err := userRepo.InitSchema(context.Background()); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}
	log.Println("Database schema initialized")

	// Initialize services
	authService := database.NewJWTAuthService(cfg.JWT.SecretKey)

	// Initialize use cases
	userUseCase := usecases.NewUserUseCase(userRepo, authService)

	// Initialize HTTP handlers
	userHandler := http.NewUserHandler(userUseCase)

	// Setup router
	router := http.SetupRouter(userHandler, authService)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
