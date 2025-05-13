package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"openshield-manager/internal/models"
)

var DB *gorm.DB

func ConnectDatabase() {
	env := os.Getenv("ENVIRONMENT")

	var err error
	switch env {
	case "production":
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect to PostgreSQL: %v", err)
		}
		log.Println("Connected to PostgreSQL (production)")

	default:
		DB, err = gorm.Open(sqlite.Open("openshield.db"), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect to SQLite: %v", err)
		}
		log.Println("Connected to SQLite (development)")
	}

	err = DB.AutoMigrate(&models.Agent{}, &models.Job{}, &models.Task{})
	if err != nil {
		log.Fatalf("auto-migration failed: %v", err)
	}
}
