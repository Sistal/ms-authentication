package usecases

import (
	"context"
	"errors"
	"github.com/Sistal/ms-authentication/internal/domain/entities"
	"github.com/Sistal/ms-authentication/internal/domain/ports"
	"github.com/google/uuid"
	"time"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound = errors.New("user not found")
)

// UserUseCase handles user-related business logic
type UserUseCase struct {
	userRepo    ports.UserRepository
	authService ports.AuthService
}

// NewUserUseCase creates a new UserUseCase
func NewUserUseCase(userRepo ports.UserRepository, authService ports.AuthService) *UserUseCase {
	return &UserUseCase{
		userRepo:    userRepo,
		authService: authService,
	}
}

// RegisterUser creates a new user account
func (uc *UserUseCase) RegisterUser(ctx context.Context, email, password, name string) (*entities.User, error) {
	// Check if user already exists
	existingUser, _ := uc.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := uc.authService.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &entities.User{
		ID:        uuid.New().String(),
		Email:     email,
		Password:  hashedPassword,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a token
func (uc *UserUseCase) Login(ctx context.Context, email, password string) (string, *entities.User, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Compare password
	if err := uc.authService.ComparePassword(user.Password, password); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Generate token
	token, err := uc.authService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// GetUserByID retrieves a user by ID
func (uc *UserUseCase) GetUserByID(ctx context.Context, id string) (*entities.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
