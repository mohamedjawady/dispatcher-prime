package handlers

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// InitializeMongoClient initializes the MongoDB client with a timeout.
func InitializeMongoClient(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	return nil
}

// DisconnectMongoClient disconnects the MongoDB client.
func DisconnectMongoClient() error {
	if Client == nil {
		return nil
	}
	return Client.Disconnect(context.Background())
}
