package config

import (
	"log"
	"os"
)

// Config holds the configuration for the MongoDB password rotator
type Config struct {
	MongoDBConnectionString string
	MongoDBUsername         string
	MongoDBDBName           string
	NewPasswordFilePath     string
	CurrentPasswordFilePath string
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() (*Config, error) {
	mongodbConnectionString := os.Getenv("MONGODB_CONNECTION_STRING")
	mongodbUsername := os.Getenv("MONGODB_USERNAME")
	mongodbDBName := os.Getenv("MONGODB_DBNAME")
	newPasswordFilePath := os.Getenv("NEW_PASSWORD_FILE_PATH")
	currentPasswordFilePath := os.Getenv("CURRENT_PASSWORD_FILE_PATH")

	// Simple validation to ensure required configurations are set
	if mongodbConnectionString == "" || mongodbUsername == "" || mongodbDBName == "" {
		log.Fatal("Required MongoDB configuration is missing")
	}
	if newPasswordFilePath == "" || currentPasswordFilePath == "" {
		log.Fatal("Required file path configuration is missing")
	}

	return &Config{
		MongoDBConnectionString: mongodbConnectionString,
		MongoDBUsername:         mongodbUsername,
		MongoDBDBName:           mongodbDBName,
		NewPasswordFilePath:     newPasswordFilePath,
		CurrentPasswordFilePath: currentPasswordFilePath,
	}, nil
}
