package router

import (
	"github.com/gorilla/mux"
	"github.com/juancruzestevez/auth-service/handlers"
	"github.com/juancruzestevez/auth-service/middleware"
)

// SetupRouter configura todas las rutas de la aplicación
func SetupRouter(authHandler *handlers.AuthHandler) *mux.Router {
	r := mux.NewRouter()

	// API prefix
	api := r.PathPrefix("/api/v1").Subrouter()

	// Rutas públicas de autenticación
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", authHandler.Register).Methods("POST")
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Rutas protegidas (ejemplo para futuras rutas)
	protected := api.PathPrefix("/users").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	// protected.HandleFunc("/me", userHandler.GetProfile).Methods("GET")
	// protected.HandleFunc("/me", userHandler.UpdateProfile).Methods("PUT")

	return r
}
