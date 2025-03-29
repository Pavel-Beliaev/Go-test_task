package db

import (
	"fmt"
	"os"
	"test-task/internal/domain"
	"test-task/pkg/logging"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() (db *gorm.DB, err error) {
	log := logging.GetLogger()
	log.Info("Initializing database connection")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Errorf("Error connecting to DB: %v", err)
		return nil, err
	}

	log.Info("Running migrations")
	if err := db.AutoMigrate(&domain.Song{}); err != nil {
		log.Errorf("Error during migration: %v", err)
		return nil, err
	}

	log.Info("Database initialized successfully")
	return db, nil
}
