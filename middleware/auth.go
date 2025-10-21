package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/juancruzestevez/auth-service/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const EmailKey contextKey = "email"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener el header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Verificar que tenga el formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validar el token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Agregar la información del usuario al context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)

		// Continuar con el siguiente handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(r *http.Request) uint {
	if userID, ok := r.Context().Value(UserIDKey).(uint); ok {
		return userID
	}
	return 0
}

func GetEmailFromContext(r *http.Request) string {
	if email, ok := r.Context().Value(EmailKey).(string); ok {
		return email
	}
	return ""
}
