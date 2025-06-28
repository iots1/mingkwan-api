package infrastructure

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"

	"github.com/iots1/mingkwan-api/internal/shared/utils" // นำเข้า Zap logger ของเรา
)

type MongoClient struct {
	client *mongo.Client
	uri    string
	dbName string
}

func NewMongoClient(uri, dbName string) *MongoClient {
	return &MongoClient{
		uri:    uri,
		dbName: dbName,
	}
}

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
			utils.Logger.Error("Error disconnecting MongoDB client after failed ping", zap.Error(disconnectErr))
		}
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	utils.Logger.Info("Successfully connected to MongoDB!")
	return mc.client, nil
}

func (mc *MongoClient) GetDatabase() *mongo.Database {
	if mc.client == nil {
		utils.Logger.Fatal("MongoDB client is not connected. Call Connect() first.")
		return nil
	}
	return mc.client.Database(mc.dbName)
}

func (mc *MongoClient) Disconnect(ctx context.Context) {
	if mc.client != nil {
		if err := mc.client.Disconnect(ctx); err != nil {
			utils.Logger.Error("Error disconnecting from MongoDB", zap.Error(err))
		} else {
			utils.Logger.Info("Disconnected from MongoDB.")
		}
	}
}
