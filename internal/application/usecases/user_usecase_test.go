package usecases

import (
	"context"
	"testing"

	"github.com/Sistal/ms-authentication/internal/domain/entities"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	users map[string]*entities.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*entities.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	delete(m.users, id)
	return nil
}

// MockAuthService is a mock implementation of AuthService for testing
type MockAuthService struct{}

func NewMockAuthService() *MockAuthService {
	return &MockAuthService{}
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	return "hashed_" + password, nil
}

func (m *MockAuthService) ComparePassword(hashedPassword, password string) error {
	if hashedPassword == "hashed_"+password {
		return nil
	}
	return ErrInvalidCredentials
}

func (m *MockAuthService) GenerateToken(userID string, email string) (string, error) {
	return "mock_token_" + userID, nil
}

func (m *MockAuthService) ValidateToken(token string) (string, error) {
	return "mock_user_id", nil
}

// TestRegisterUser tests the user registration use case
func TestRegisterUser(t *testing.T) {
	repo := NewMockUserRepository()
	authService := NewMockAuthService()
	useCase := NewUserUseCase(repo, authService)

	user, err := useCase.RegisterUser(context.Background(), "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name Test User, got %s", user.Name)
	}
}

// TestRegisterUserAlreadyExists tests registering a user that already exists
func TestRegisterUserAlreadyExists(t *testing.T) {
	repo := NewMockUserRepository()
	authService := NewMockAuthService()
	useCase := NewUserUseCase(repo, authService)

	// Register first user
	_, err := useCase.RegisterUser(context.Background(), "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Try to register the same user again
	_, err = useCase.RegisterUser(context.Background(), "test@example.com", "password123", "Test User")
	if err != ErrUserAlreadyExists {
		t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
	}
}

// TestLogin tests successful login
func TestLogin(t *testing.T) {
	repo := NewMockUserRepository()
	authService := NewMockAuthService()
	useCase := NewUserUseCase(repo, authService)

	// Register a user first
	user, _ := useCase.RegisterUser(context.Background(), "test@example.com", "password123", "Test User")

	// Try to login
	token, loginUser, err := useCase.Login(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected token to be generated")
	}

	if loginUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, loginUser.ID)
	}
}

// TestLoginInvalidCredentials tests login with invalid credentials
func TestLoginInvalidCredentials(t *testing.T) {
	repo := NewMockUserRepository()
	authService := NewMockAuthService()
	useCase := NewUserUseCase(repo, authService)

	// Register a user first
	_, _ = useCase.RegisterUser(context.Background(), "test@example.com", "password123", "Test User")

	// Try to login with wrong password
	_, _, err := useCase.Login(context.Background(), "test@example.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}
