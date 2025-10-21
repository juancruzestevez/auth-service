package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/juancruzestevez/auth-service/config"
	"github.com/juancruzestevez/auth-service/dto"
	"github.com/juancruzestevez/auth-service/handlers"
	"github.com/juancruzestevez/auth-service/models"
	"github.com/juancruzestevez/auth-service/repository"
	"github.com/juancruzestevez/auth-service/router"
	"github.com/juancruzestevez/auth-service/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupIntegrationTest configura una base de datos SQLite en memoria para tests de integración
func setupIntegrationTest(t *testing.T) (*gorm.DB, func()) {
	// Configurar config para JWT
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:       "test-secret-key-integration",
			ExpiresHours: 24,
		},
		Server: config.ServerConfig{
			Env: "test",
		},
	}

	// Crear DB SQLite en memoria
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrar esquema
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return db, cleanup
}

func TestIntegration_RegisterAndLogin(t *testing.T) {
	db, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Inicializar capas de la aplicación
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)
	r := router.SetupRouter(authHandler)

	// Test 1: Registrar un usuario
	t.Run("Register new user", func(t *testing.T) {
		registerReq := dto.RegisterRequest{
			FirstName: "Integration",
			LastName:  "Test",
			Nickname:  "inttest",
			Email:     "integration@test.com",
			Password:  "password123",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var authResp dto.AuthResponse
		if err := json.NewDecoder(rr.Body).Decode(&authResp); err != nil {
			t.Fatalf("Error decoding response: %v", err)
		}

		if authResp.Token == "" {
			t.Error("Expected token in response")
		}

		if authResp.User.Email != "integration@test.com" {
			t.Errorf("Expected email integration@test.com, got %s", authResp.User.Email)
		}
	})

	// Test 2: Login con el usuario registrado
	t.Run("Login with registered user", func(t *testing.T) {
		loginReq := dto.LoginRequest{
			Email:    "integration@test.com",
			Password: "password123",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var authResp dto.AuthResponse
		if err := json.NewDecoder(rr.Body).Decode(&authResp); err != nil {
			t.Fatalf("Error decoding response: %v", err)
		}

		if authResp.Token == "" {
			t.Error("Expected token in response")
		}
	})

	// Test 3: Intentar registrar con email duplicado
	t.Run("Register with duplicate email", func(t *testing.T) {
		registerReq := dto.RegisterRequest{
			FirstName: "Duplicate",
			LastName:  "User",
			Nickname:  "duplicate",
			Email:     "integration@test.com", // Email ya existe
			Password:  "password123",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("Expected status 409, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	// Test 4: Login con credenciales incorrectas
	t.Run("Login with wrong password", func(t *testing.T) {
		loginReq := dto.LoginRequest{
			Email:    "integration@test.com",
			Password: "wrongpassword",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	// Test 5: Login con usuario que no existe
	t.Run("Login with non-existent user", func(t *testing.T) {
		loginReq := dto.LoginRequest{
			Email:    "notfound@test.com",
			Password: "password123",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})
}

func TestIntegration_ValidationErrors(t *testing.T) {
	db, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Inicializar aplicación
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)
	r := router.SetupRouter(authHandler)

	t.Run("Register with missing fields", func(t *testing.T) {
		registerReq := dto.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			// Faltan FirstName, LastName, Nickname
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Login with empty credentials", func(t *testing.T) {
		loginReq := dto.LoginRequest{
			Email:    "",
			Password: "",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})
}
