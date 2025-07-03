package config

import (
	"fmt"
	"log"
	"os"

	"github.com/chuks/PAYBOTGO/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Read from env
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Check that no value is empty
	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		log.Fatalf("Missing one or more DB environment variables")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&models.Student{},&models.MonoSession{})
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	DB = db
	fmt.Println("✅ Connected to database")
}
