package mongodb

import (
    "context"
    "fmt"
    "mongodb-password-rotator/internal/config"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Client holds the MongoDB client to interact with the database
type Client struct {
    MongoDB *mongo.Client
}

// NewClient creates a new MongoDB client
func NewClient(cfg *config.Config) (*Client, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOptions := options.Client().ApplyURI(cfg.MongoDBConnectionString).
        SetAuth(options.Credential{
            Username: cfg.MongoDBUsername,
            Password: cfg.MongoDBDBName,
        }).SetWriteConcern(writeconcern.New(writeconcern.WMajority()))

    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
    }

    return &Client{MongoDB: client}, nil
}

// UpdatePassword updates the password for the MongoDB user
func (c *Client) UpdatePassword(newPassword string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Assuming your MongoDB user's database is "admin"
    db := c.MongoDB.Database("admin")

    // Replace "yourUserName" with the actual username or pass it dynamically
    cmd := mongo.NewUpdateUserOptions()
    cmd.SetPwd(newPassword)

    err := db.RunCommand(ctx, cmd).Err()
    if err != nil {
        return fmt.Errorf("failed to update user password: %v", err)
    }

    return nil
}
