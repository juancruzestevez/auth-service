package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/juancruzestevez/auth-service/dto"
	"github.com/juancruzestevez/auth-service/handlers"
	"github.com/juancruzestevez/auth-service/service"
)

// Mock AuthService
type MockAuthService struct {
	RegisterFunc func(req dto.RegisterRequest) (*dto.AuthResponse, error)
	LoginFunc    func(req dto.LoginRequest) (*dto.AuthResponse, error)
}

func (m *MockAuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(req)
	}
	return nil, nil
}

func (m *MockAuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(req)
	}
	return nil, nil
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func() *MockAuthService
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "Registro exitoso",
			requestBody: dto.RegisterRequest{
				FirstName: "Juan",
				LastName:  "Pérez",
				Nickname:  "juanp",
				Email:     "juan@example.com",
				Password:  "password123",
			},
			mockSetup: func() *MockAuthService {
				return &MockAuthService{
					RegisterFunc: func(req dto.RegisterRequest) (*dto.AuthResponse, error) {
						return &dto.AuthResponse{
							Token:     "fake-jwt-token",
							ExpiresAt: 1234567890,
							User: dto.UserResponse{
								ID:        1,
								FirstName: req.FirstName,
								LastName:  req.LastName,
								Nickname:  req.Nickname,
								Email:     req.Email,
							},
						}, nil
					},
				}
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var authResp dto.AuthResponse
				if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
					t.Errorf("Error decoding response: %v", err)
				}
				if authResp.Token == "" {
					t.Error("Expected token in response")
				}
				if authResp.User.Email != "juan@example.com" {
					t.Errorf("Expected email juan@example.com, got %s", authResp.User.Email)
				}
			},
		},
		{
			name: "Email ya existe",
			requestBody: dto.RegisterRequest{
				FirstName: "Juan",
				LastName:  "Pérez",
				Nickname:  "juanp",
				Email:     "existing@example.com",
				Password:  "password123",
			},
			mockSetup: func() *MockAuthService {
				return &MockAuthService{
					RegisterFunc: func(req dto.RegisterRequest) (*dto.AuthResponse, error) {
						return nil, service.ErrUserAlreadyExists
					},
				}
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				if !bytes.Contains(resp.Body.Bytes(), []byte("already exists")) {
					t.Error("Expected 'already exists' error message")
				}
			},
		},
		{
			name: "Campos inválidos",
			requestBody: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "123456",
			},
			mockSetup: func() *MockAuthService {
				return &MockAuthService{
					RegisterFunc: func(req dto.RegisterRequest) (*dto.AuthResponse, error) {
						return nil, service.ErrInvalidInput
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				if !bytes.Contains(resp.Body.Bytes(), []byte("required")) {
					t.Error("Expected 'required' error message")
				}
			},
		},
		{
			name:        "JSON inválido",
			requestBody: "invalid json",
			mockSetup: func() *MockAuthService {
				return &MockAuthService{}
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := tt.mockSetup()
			handler := handlers.NewAuthHandler(mockService)

			// Crear request
			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			// Ejecutar handler
			handler.Register(rr, req)

			// Verificar status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			// Verificar respuesta adicional
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func() *MockAuthService
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "Login exitoso",
			requestBody: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func() *MockAuthService {
				return &MockAuthService{
					LoginFunc: func(req dto.LoginRequest) (*dto.AuthResponse, error) {
						return &dto.AuthResponse{
							Token:     "fake-jwt-token",
							ExpiresAt: 1234567890,
							User: dto.UserResponse{
								ID:        1,
								FirstName: "Test",
								LastName:  "User",
								Nickname:  "testuser",
								Email:     req.Email,
							},
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var authResp dto.AuthResponse
				if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
					t.Errorf("Error decoding response: %v", err)
				}
				if authResp.Token == "" {
					t.Error("Expected token in response")
				}
				if authResp.User.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %s", authResp.User.Email)
				}
			},
		},
		{
			name: "Credenciales inválidas",
			requestBody: dto.LoginRequest{
				Email:    "wrong@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func() *MockAuthService {
				return &MockAuthService{
					LoginFunc: func(req dto.LoginRequest) (*dto.AuthResponse, error) {
						return nil, service.ErrInvalidCredentials
					},
				}
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				if !bytes.Contains(resp.Body.Bytes(), []byte("Invalid credentials")) {
					t.Error("Expected 'Invalid credentials' error message")
				}
			},
		},
		{
			name: "Campos vacíos",
			requestBody: dto.LoginRequest{
				Email:    "",
				Password: "",
			},
			mockSetup: func() *MockAuthService {
				return &MockAuthService{
					LoginFunc: func(req dto.LoginRequest) (*dto.AuthResponse, error) {
						return nil, service.ErrInvalidInput
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:        "JSON inválido",
			requestBody: "invalid json",
			mockSetup: func() *MockAuthService {
				return &MockAuthService{}
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name: "Error interno del servidor",
			requestBody: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func() *MockAuthService {
				return &MockAuthService{
					LoginFunc: func(req dto.LoginRequest) (*dto.AuthResponse, error) {
						return nil, errors.New("database connection failed")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := tt.mockSetup()
			handler := handlers.NewAuthHandler(mockService)

			// Crear request
			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			// Ejecutar handler
			handler.Login(rr, req)

			// Verificar status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			// Verificar respuesta adicional
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}
