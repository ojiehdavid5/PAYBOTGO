package database

import (
	"fmt"
	"log"
	"os"

	"github.com/chuks/PAYBOTGO/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect to the database
func Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Failed to connect to database")
		fmt.Println("Error:", err)
	}
	fmt.Println("Connection Opened to Database")

	// Migrate the schemas
	DB.AutoMigrate(&models.Student{})
	fmt.Println("Database Migrated")
	return DB, err
}
