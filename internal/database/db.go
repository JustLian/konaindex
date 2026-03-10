package database

import (
	"fmt"
	"konaindex/internal/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dsn string) {
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB.Exec("CREATE EXTENSION IF NOT EXISTS vector;")

	err = DB.AutoMigrate(&models.Post{}, &models.PostColor{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("Database connected and migrated successfully")
}
