package main

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joaquinrz/mongo-password-rotator/internal/config"
	"github.com/joaquinrz/mongo-password-rotator/internal/mongodb"
)

func main() {
	log.Println("Starting MongoDB password rotator.")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(filepath.Dir(cfg.NewPasswordFilePath))
	if err != nil {
		log.Fatalf("Failed to watch the directory for password file changes: %v", err)
	}

	log.Println("Monitoring for password changes...")

	listenForEvents(watcher, cfg)

	// Block main goroutine indefinitely.
	select {}
}

func listenForEvents(watcher *fsnotify.Watcher, cfg *config.Config) {
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return // Exit if the channel is closed.
				}
				handleEvent(event, cfg)
			case err, ok := <-watcher.Errors:
				if !ok {
					return // Exit if the channel is closed.
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()
}

func handleEvent(event fsnotify.Event, cfg *config.Config) {
	// Check if the event is related to the NewPasswordFilePath.
	if filepath.Clean(event.Name) == filepath.Clean(cfg.NewPasswordFilePath) {
		switch {
		case event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create:
			log.Printf("Detected change in password file: %s", event.Name)
			updateMongoDBPassword(cfg)
		case event.Op&fsnotify.Remove == fsnotify.Remove:
			log.Printf("Password file was deleted: %s", event.Name)
		}
	}
}

func updateMongoDBPassword(cfg *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongodb.NewClient(ctx, cfg)
	if err != nil {
		log.Printf("Failed to initialize MongoDB client: %v", err)
		return
	}
	defer mongoClient.Disconnect(ctx)

	if err := mongoClient.UpdatePassword(ctx, cfg); err != nil {
		log.Printf("Failed to update MongoDB password: %v", err)
	} else {
		log.Println("MongoDB password updated successfully.")
	}
}
