package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/juancruzestevez/auth-service/dto"
	"github.com/juancruzestevez/auth-service/service"
)

// AuthHandler maneja las peticiones HTTP de autenticación
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler crea una nueva instancia del handler de autenticación
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register maneja el registro de nuevos usuarios
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest

	// Decodificar el JSON del body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Llamar al servicio
	response, err := h.authService.Register(req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Enviar respuesta exitosa
	respondWithJSON(w, http.StatusCreated, response)
}

// Login maneja el login de usuarios
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	// Decodificar el JSON del body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Llamar al servicio
	response, err := h.authService.Login(req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Enviar respuesta exitosa
	respondWithJSON(w, http.StatusOK, response)
}

// handleServiceError mapea errores del servicio a códigos HTTP
func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
	case errors.Is(err, service.ErrUserAlreadyExists):
		respondWithError(w, http.StatusConflict, "User with this email or nickname already exists")
	case errors.Is(err, service.ErrInvalidInput):
		respondWithError(w, http.StatusBadRequest, "All fields are required")
	case errors.Is(err, service.ErrUserNotFound):
		respondWithError(w, http.StatusNotFound, "User not found")
	default:
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// respondWithJSON envía una respuesta JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// respondWithError envía una respuesta de error
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
