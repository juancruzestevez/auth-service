package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/juancruzestevez/auth-service/auth"
	"github.com/juancruzestevez/auth-service/config"
)

func setupTestConfig() {
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:       "test-secret-key",
			ExpiresHours: 24,
		},
	}
}

func TestGenerateToken(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name    string
		userID  uint
		email   string
		wantErr bool
	}{
		{
			name:    "Token generado exitosamente",
			userID:  1,
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "Token con ID diferente",
			userID:  999,
			email:   "another@example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiresAt, err := auth.GenerateToken(tt.userID, tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verificar que el token no esté vacío
				if token == "" {
					t.Error("Expected non-empty token")
				}

				// Verificar que expiresAt sea en el futuro
				if expiresAt <= time.Now().Unix() {
					t.Error("Expected expiresAt to be in the future")
				}

				// Verificar que el token sea válido
				claims, err := auth.ValidateToken(token)
				if err != nil {
					t.Errorf("Generated token is not valid: %v", err)
				}

				// Verificar que los claims contengan la información correcta
				if claims.UserID != tt.userID {
					t.Errorf("Expected userID %d, got %d", tt.userID, claims.UserID)
				}

				if claims.Email != tt.email {
					t.Errorf("Expected email %s, got %s", tt.email, claims.Email)
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	setupTestConfig()

	// Generar un token válido para las pruebas
	validToken, _, _ := auth.GenerateToken(1, "test@example.com")

	// Generar un token expirado
	secret := []byte("test-secret-key")
	expiredClaims := auth.Claims{
		UserID: 2,
		Email:  "expired@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "auth-service",
		},
	}
	expiredTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredToken, _ := expiredTokenObj.SignedString(secret)

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		wantEmail string
	}{
		{
			name:      "Token válido",
			token:     validToken,
			wantErr:   false,
			wantEmail: "test@example.com",
		},
		{
			name:    "Token expirado",
			token:   expiredToken,
			wantErr: true,
		},
		{
			name:    "Token inválido",
			token:   "invalid.token.here",
			wantErr: true,
		},
		{
			name:    "Token vacío",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Token con firma incorrecta",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20ifQ.wrong_signature",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := auth.ValidateToken(tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if claims == nil {
					t.Error("Expected claims to be non-nil")
					return
				}

				if claims.Email != tt.wantEmail {
					t.Errorf("Expected email %s, got %s", tt.wantEmail, claims.Email)
				}
			}
		})
	}
}
