package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/juancruzestevez/auth-service/config"
	"github.com/juancruzestevez/auth-service/db"
	"github.com/juancruzestevez/auth-service/handlers"
	"github.com/juancruzestevez/auth-service/models"
	"github.com/juancruzestevez/auth-service/repository"
	"github.com/juancruzestevez/auth-service/router"
	"github.com/juancruzestevez/auth-service/service"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()
	log.Printf("Starting Auth Service in %s mode", cfg.Server.Env)

	// Conectar a la base de datos
	if err := db.Connect(cfg); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// Ejecutar migraciones
	if err := db.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("✓ Database migrations completed")

	// Inicializar capas de la aplicación
	userRepo := repository.NewUserRepository(db.DB)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// Configurar rutas
	r := router.SetupRouter(authHandler)

	// Iniciar servidor
	log.Printf("✓ Server starting on port %s", cfg.Server.Port)
	log.Printf("✓ API available at http://localhost:%s/api/v1", cfg.Server.Port)

	// Graceful shutdown
	go func() {
		if err := http.ListenAndServe(":"+cfg.Server.Port, r); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Esperar señal de terminación
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
