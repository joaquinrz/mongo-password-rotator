package main

import (
	"context"
	"github.com/joaquinrz/mongo-password-rotator/internal/keyvault"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/joaquinrz/mongo-password-rotator/internal/mongodb"

	"github.com/fsnotify/fsnotify"
	"github.com/joaquinrz/mongo-password-rotator/internal/config"
)

func main() {
	log.Println("MongoDB password rotator started.")

	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := watchPasswordFileChanges(cfg); err != nil {
		log.Fatalf("Error watching password file: %v", err)
	}
}

func watchPasswordFileChanges(cfg *config.Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	dirPath := filepath.Dir(cfg.NewPasswordFilePath)
	if err := watcher.Add(dirPath); err != nil {
		return err
	}

	var (
		cooldownPeriod = 2 * time.Second
		lastUpdate     time.Time
		mu             sync.Mutex
	)

	log.Println("Waiting for Password file changes...")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil // Exit if channel is closed.
			}

			if filepath.Base(event.Name) == filepath.Base(cfg.NewPasswordFilePath) && time.Since(lastUpdate) > cooldownPeriod {
				mu.Lock()
				if time.Since(lastUpdate) > cooldownPeriod {
					log.Println("Password file changed, updating MongoDB password...")

					// Attempt to update MongoDB password and check for failure
					if err := updateMongoDBPassword(cfg); err != nil {
						log.Printf("Failed to update MongoDB password: %v", err)
						continue // Skip to the next iteration of the loop
					}

					// If updateMongoDBPassword was successful, proceed to update Azure KeyVault secret
					log.Println("MongoDB password updated successfully, updating Azure Key Vault secret...")
					if err := keyvault.UpdateSecret(cfg); err != nil {
						log.Fatalf("Failed to update Azure Key Vault secret: %v", err)
					} else {
						log.Println("Successfully updated Current Azure Key Vault secret with new MongoDB password.")
					}

					lastUpdate = time.Now()
					log.Println("Waiting for Password file changes...")
				}
				mu.Unlock()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil // Exit if channel is closed.
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func updateMongoDBPassword(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongodb.NewClient(ctx, cfg)
	if err != nil {
		log.Printf("Failed to initialize MongoDB client: %v", err)
		return err
	}
	defer mongoClient.Disconnect(ctx)

	if err := mongoClient.UpdatePassword(ctx, cfg); err != nil {
		log.Printf("Failed to update MongoDB password: %v", err)
		return err
	}

	log.Println("MongoDB password updated successfully.")
	return nil
}
