package service_test

import (
	"errors"
	"testing"

	"github.com/juancruzestevez/auth-service/config"
	"github.com/juancruzestevez/auth-service/dto"
	"github.com/juancruzestevez/auth-service/models"
	"github.com/juancruzestevez/auth-service/service"
	"golang.org/x/crypto/bcrypt"
)

// Mock UserRepository
type MockUserRepository struct {
	CreateFunc         func(user *models.User) error
	FindByEmailFunc    func(email string) (*models.User, error)
	FindByNicknameFunc func(nickname string) (*models.User, error)
	FindByIDFunc       func(id uint) (*models.User, error)
	UpdateFunc         func(user *models.User) error
	DeleteFunc         func(id uint) error
}

func (m *MockUserRepository) Create(user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return nil
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(email)
	}
	return nil, nil
}

func (m *MockUserRepository) FindByNickname(nickname string) (*models.User, error) {
	if m.FindByNicknameFunc != nil {
		return m.FindByNicknameFunc(nickname)
	}
	return nil, nil
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, nil
}

func (m *MockUserRepository) Update(user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(user)
	}
	return nil
}

func (m *MockUserRepository) Delete(id uint) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func TestAuthService_Register(t *testing.T) {
	// Setup config for JWT
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:       "test-secret-key",
			ExpiresHours: 24,
		},
	}

	tests := []struct {
		name      string
		request   dto.RegisterRequest
		mockSetup func() *MockUserRepository
		wantErr   error
	}{
		{
			name: "Registro exitoso",
			request: dto.RegisterRequest{
				FirstName: "Juan",
				LastName:  "Pérez",
				Nickname:  "juanp",
				Email:     "juan@example.com",
				Password:  "password123",
			},
			mockSetup: func() *MockUserRepository {
				return &MockUserRepository{
					FindByEmailFunc: func(email string) (*models.User, error) {
						return nil, nil // Usuario no existe
					},
					FindByNicknameFunc: func(nickname string) (*models.User, error) {
						return nil, nil // Nickname disponible
					},
					CreateFunc: func(user *models.User) error {
						user.ID = 1
						return nil
					},
				}
			},
			wantErr: nil,
		},
		{
			name: "Email ya existe",
			request: dto.RegisterRequest{
				FirstName: "Juan",
				LastName:  "Pérez",
				Nickname:  "juanp",
				Email:     "existing@example.com",
				Password:  "password123",
			},
			mockSetup: func() *MockUserRepository {
				return &MockUserRepository{
					FindByEmailFunc: func(email string) (*models.User, error) {
						return &models.User{Email: email}, nil // Usuario ya existe
					},
				}
			},
			wantErr: service.ErrUserAlreadyExists,
		},
		{
			name: "Nickname ya existe",
			request: dto.RegisterRequest{
				FirstName: "Juan",
				LastName:  "Pérez",
				Nickname:  "existingnick",
				Email:     "new@example.com",
				Password:  "password123",
			},
			mockSetup: func() *MockUserRepository {
				return &MockUserRepository{
					FindByEmailFunc: func(email string) (*models.User, error) {
						return nil, nil // Email disponible
					},
					FindByNicknameFunc: func(nickname string) (*models.User, error) {
						return &models.User{Nickname: nickname}, nil // Nickname existe
					},
				}
			},
			wantErr: service.ErrUserAlreadyExists,
		},
		{
			name: "Campos requeridos faltantes",
			request: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				// Faltan FirstName, LastName, Nickname
			},
			mockSetup: func() *MockUserRepository {
				return &MockUserRepository{}
			},
			wantErr: service.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.mockSetup()
			authService := service.NewAuthService(mockRepo)

			response, err := authService.Register(tt.request)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("Expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if response == nil {
					t.Error("Expected response, got nil")
				}
				if response != nil && response.Token == "" {
					t.Error("Expected token in response")
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	setupTestConfig()

	// Crear password hasheado para tests
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		request   dto.LoginRequest
		mockSetup func() *MockUserRepository
		wantErr   error
	}{
		{
			name: "Login exitoso",
			request: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func() *MockUserRepository {
				return &MockUserRepository{
					FindByEmailFunc: func(email string) (*models.User, error) {
						return &models.User{
							Model:     models.User{}.Model,
							Email:     email,
							Password:  string(hashedPassword),
							FirstName: "Test",
							LastName:  "User",
							Nickname:  "testuser",
						}, nil
					},
				}
			},
			wantErr: nil,
		},
		{
			name: "Usuario no encontrado",
			request: dto.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			mockSetup: func() *MockUserRepository {
				return &MockUserRepository{
					FindByEmailFunc: func(email string) (*models.User, error) {
						return nil, nil // Usuario no existe
					},
				}
			},
			wantErr: service.ErrInvalidCredentials,
		},
		{
			name: "Contraseña incorrecta",
			request: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func() *MockUserRepository {
				return &MockUserRepository{
					FindByEmailFunc: func(email string) (*models.User, error) {
						return &models.User{
							Email:    email,
							Password: string(hashedPassword),
						}, nil
					},
				}
			},
			wantErr: service.ErrInvalidCredentials,
		},
		{
			name: "Campos vacíos",
			request: dto.LoginRequest{
				Email:    "",
				Password: "",
			},
			mockSetup: func() *MockUserRepository {
				return &MockUserRepository{}
			},
			wantErr: service.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.mockSetup()
			authService := service.NewAuthService(mockRepo)

			response, err := authService.Login(tt.request)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("Expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if response == nil {
					t.Error("Expected response, got nil")
				}
				if response != nil && response.Token == "" {
					t.Error("Expected token in response")
				}
			}
		})
	}
}

// Helper para setup de config en tests
func setupTestConfig() {
	// Este import se haría desde config_test o similar
	// Por ahora, asumimos que config.AppConfig está disponible
}
