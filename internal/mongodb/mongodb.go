package mongodb

import (
	"context"
	"fmt"
	"github.com/joaquinrz/mongo-password-rotator/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"log"
	"os"
)

// Client wraps a MongoDB client with additional functionality.
type Client struct {
	MongoDB *mongo.Client
}

// NewClient establishes a new connection to MongoDB using the provided configuration.
// It first tries to connect with the current password, and upon failure, retries with the new password.
func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	passwords := make([]string, 0, 2)

	// Read the current password from the specified file.
	currentPasswordBytes, err := os.ReadFile(cfg.CurrentPasswordFilePath)
	if err != nil {
		log.Printf("Error reading current password from file: %v", err)
		// If the current password cannot be read, it's not immediately fatal; try the new password.
	} else {
		passwords = append(passwords, string(currentPasswordBytes))
	}

	// Try reading the new password as a fallback or additional option.
	newPasswordBytes, err := os.ReadFile(cfg.NewPasswordFilePath)
	if err != nil {
		log.Printf("Error reading new password from file: %v", err)
		if len(passwords) == 0 {
			// If neither password could be read, error out.
			return nil, fmt.Errorf("failed to read both current and new passwords: %v", err)
		}
	} else {
		passwords = append(passwords, string(newPasswordBytes))
	}

	for i, password := range passwords {
		clientOptions := options.Client().ApplyURI(cfg.MongoDBConnectionString).SetAuth(options.Credential{
			Username: cfg.MongoDBUsername,
			Password: password,
		}).SetWriteConcern(writeconcern.Majority())

		// Attempt to connect to MongoDB.
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil || client.Ping(ctx, nil) != nil {
			log.Printf("Attempt %d of %d: Error authenticating with MongoDB using password", i+1, len(passwords))
			continue // Try the next password if this one fails.
		}

		// Successful connection
		log.Printf("Successfully authenticated with MongoDB on attempt %d of %d.", i+1, len(passwords))
		return &Client{MongoDB: client}, nil
	}

	// If this point is reached, both passwords have failed.
	return nil, fmt.Errorf("failed to authenticate with MongoDB using both current and new passwords")
}

// UpdatePassword changes the MongoDB user's password based on configuration settings.
func (c *Client) UpdatePassword(ctx context.Context, cfg *config.Config) error {
	// Read the new password from the specified file.
	newPasswordBytes, err := os.ReadFile(cfg.NewPasswordFilePath)
	if err != nil {
		log.Printf("Error reading new password from file: %v", err)
		return err
	}
	newPassword := string(newPasswordBytes)

	// Execute the command to update the user's password in MongoDB.
	db := c.MongoDB.Database(cfg.MongoDBDBName) // Reference the configured database.
	cmd := bson.D{{Key: "updateUser", Value: cfg.MongoDBUsername}, {Key: "pwd", Value: newPassword}}
	result := db.RunCommand(ctx, cmd)
	if err := result.Err(); err != nil {
		log.Printf("Error updating MongoDB password: %v", err)
		return err
	}

	return nil
}

// Disconnect terminates the connection to the MongoDB database.
func (c *Client) Disconnect(ctx context.Context) error {
	if err := c.MongoDB.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
		return err
	}

	log.Println("Disconnected from MongoDB.")
	return nil
}
