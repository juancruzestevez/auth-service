package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/juancruzestevez/auth-service/db"
	"github.com/juancruzestevez/auth-service/models"
	"github.com/juancruzestevez/auth-service/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	db.DBConnection()

	db.DB.AutoMigrate(models.User{})

	r := mux.NewRouter()

	r.HandleFunc("/users", routes.GetUserHandler).Methods("GET")
	r.HandleFunc("/users/{id}", routes.GetUserHandler).Methods("GET")
	r.HandleFunc("/users", routes.PostUsersHandler).Methods("POST")
	r.HandleFunc("/users/{id}", routes.DeleteUsersHandler).Methods("DELETE")

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
