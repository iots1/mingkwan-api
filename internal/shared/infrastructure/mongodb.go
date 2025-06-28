package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoClient wraps the MongoDB client and provides connection management
type MongoClient struct {
	client *mongo.Client
	uri    string
	dbName string
}

// NewMongoClient creates a new MongoClient instance.
// Ensure this function name starts with a capital 'N' to be exported.
func NewMongoClient(uri, dbName string) *MongoClient {
	return &MongoClient{
		uri:    uri,
		dbName: dbName,
	}
}

// Connect establishes a connection to the MongoDB database.
func (mc *MongoClient) Connect(ctx context.Context) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mc.uri)
	var err error
	mc.client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = mc.client.Ping(pingCtx, readpref.Primary()); err != nil {
		if disconnectErr := mc.client.Disconnect(context.Background()); disconnectErr != nil {
			log.Printf("Error disconnecting client after failed ping: %v", disconnectErr)
		}
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected to MongoDB!")
	return mc.client, nil
}

// GetDatabase returns a specific database instance.
func (mc *MongoClient) GetDatabase() *mongo.Database {
	if mc.client == nil {
		log.Fatal("MongoDB client is not connected. Call Connect() first.")
		return nil
	}
	return mc.client.Database(mc.dbName)
}

// Disconnect closes the MongoDB connection.
func (mc *MongoClient) Disconnect(ctx context.Context) {
	if mc.client != nil {
		if err := mc.client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		} else {
			log.Println("Disconnected from MongoDB.")
		}
	}
}
