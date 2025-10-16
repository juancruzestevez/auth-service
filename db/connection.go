package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func DBConnection() {
	// Read from environment variables (set via .env or docker-compose)
	dbHost := getenv("POSTGRES_HOST", "localhost")
	dbPort := getenv("POSTGRES_PORT", "5432")
	dbUser := getenv("POSTGRES_USER", "postgres")
	dbPassword := getenv("POSTGRES_PASSWORD", "")
	dbName := getenv("POSTGRES_DB", "postgres")

	DSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	var err error
	DB, err = gorm.Open(postgres.Open(DSN), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database (host=%s db=%s user=%s): %v", dbHost, dbName, dbUser, err)
		log.Fatal("Database connection failed!")
	}

	log.Println("Successfully connected to PostgreSQL database")
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
