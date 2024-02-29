package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/fsnotify/fsnotify"
	"github.com/joaquinrz/mongo-password-rotator/internal/config"
    "github.com/joaquinrz/mongo-password-rotator/internal/logger"
	"github.com/joaquinrz/mongo-password-rotator/internal/mongodb"
    "go.mongodb.org/mongo-driver/mongo"
)

func main() {
    // Load and validate configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        logger.Fatal("Failed to load configuration:", err)
    }

    // Initialize MongoDB client
    mongoClient, err := mongodb.NewClient(cfg.MongoDBURI, cfg.MongoDBUsername, cfg.MongoDBPassword)
    if err != nil {
        logger.Fatal("Failed to initialize MongoDB client:", err)
    }
    defer mongoClient.Disconnect(context.Background())

    // Start watching the password file for changes
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        logger.Fatal("Error creating file watcher:", err)
    }
    defer watcher.Close()

    done := make(chan bool)
    go func() {
        for {
            select {
            case event := <-watcher.Events:
                if event.Op&fsnotify.Write == fsnotify.Write {
                    logger.Info("Detected change in password file")
                    newPassword, err := os.ReadFile(cfg.PasswordFilePath)
                    if err != nil {
                        logger.Error("Failed to read new password from file:", err)
                        continue
                    }

                    if err := mongoClient.UpdatePassword(string(newPassword)); err != nil {
                        logger.Error("Failed to update MongoDB password:", err)
                    } else {
                        logger.Info("MongoDB password updated successfully")
                    }
                }
            case err := <-watcher.Errors:
                logger.Error("Watcher error:", err)
            }
        }
    }()

    err = watcher.Add(cfg.PasswordFilePath)
    if err != nil {
        logger.Fatal("Failed to add password file to watcher:", err)
    }

    // Block the main goroutine until the application is terminated
    <-done
}
