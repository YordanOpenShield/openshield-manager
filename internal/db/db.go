package db

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"openshield-manager/internal/config"
	"openshield-manager/internal/models"
)

var DB *gorm.DB

func ConnectDatabase() {
	config := config.GlobalConfig

	var err error
	switch config.ENVIRONMENT {
	case "production":
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			config.DB_HOST,
			config.DB_USER,
			config.DB_PASSWORD,
			config.DB_NAME,
			config.DB_PORT,
		)
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect to PostgreSQL: %v", err)
		}
		log.Println("Connected to PostgreSQL (production)")

	default:
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			config.DB_HOST,
			config.DB_USER,
			config.DB_PASSWORD,
			config.DB_NAME,
			config.DB_PORT,
		)
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect to PostgreSQL: %v", err)
		}
		log.Println("Connected to PostgreSQL (development)")
	}

	err = DB.AutoMigrate(&models.Agent{}, &models.AgentAddress{}, &models.Job{}, &models.Task{})
	if err != nil {
		log.Fatalf("auto-migration failed: %v", err)
	}
}
