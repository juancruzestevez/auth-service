package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/juancruzestevez/auth-service/config"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken genera un JWT token para el usuario
func GenerateToken(userID uint, email string) (string, int64, error) {
	cfg := config.AppConfig
	secret := []byte(cfg.JWT.Secret)

	expirationTime := time.Now().Add(time.Duration(cfg.JWT.ExpiresHours) * time.Hour)

	// Crear los claims
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
		},
	}

	// Crear el token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar el token
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime.Unix(), nil
}

// ValidateToken valida y parsea un JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.AppConfig
	secret := []byte(cfg.JWT.Secret)

	// Parsear el token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el método de firma sea HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	// Verificar que el token sea válido y extraer los claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}
