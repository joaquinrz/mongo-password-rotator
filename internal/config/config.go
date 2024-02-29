package config

import (
    "os"
    "log"
)

// Config holds the configuration for the MongoDB password rotator
type Config struct {
    MongoDBConnectionString string
    MongoDBUsername         string
    MongoDBDBName           string
    PasswordFilePath        string
    CurrentPasswordFilePath string
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() (*Config, error) {
    mongodbConnectionString := os.Getenv("MONGODB_CONNECTION_STRING")
    mongodbUsername := os.Getenv("MONGODB_USERNAME")
    mongodbDBName := os.Getenv("MONGODB_DBNAME")
    passwordFilePath := os.Getenv("PASSWORD_FILE_PATH")
    currentPasswordFilePath := os.Getenv("CURRENT_PASSWORD_FILE_PATH")

    // Simple validation to ensure required configurations are set
    if mongodbConnectionString == "" || mongodbUsername == "" || mongodbDBName == "" {
        log.Fatal("Required MongoDB configuration is missing")
    }
    if passwordFilePath == "" || currentPasswordFilePath == "" {
        log.Fatal("Required file path configuration is missing")
    }

    return &Config{
        MongoDBConnectionString: mongodbConnectionString,
        MongoDBUsername:         mongodbUsername,
        MongoDBDBName:           mongodbDBName,
        PasswordFilePath:        passwordFilePath,
        CurrentPasswordFilePath: currentPasswordFilePath,
    }, nil
}
