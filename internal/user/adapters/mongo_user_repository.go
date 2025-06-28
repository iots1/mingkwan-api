package adapters

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/iots1/mingkwan-api/internal/user/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database, collectionName string) *MongoUserRepository {
	return &MongoUserRepository{
		collection: db.Collection(collectionName),
	}
}

// Create inserts a new user into MongoDB.
func (r *MongoUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

func (r *MongoUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var user domain.User
	filter := bson.M{"_id": id}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found with ID: %s", id.Hex())
		}
		log.Printf("Error finding user by ID: %v", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	filter := bson.M{"email": email}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		log.Printf("Error finding user by email: %v", err)
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

func (r *MongoUserRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	user.UpdatedAt = time.Now()
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	if result.ModifiedCount == 0 {
		return nil, fmt.Errorf("user not found or no changes made with ID: %s", user.ID.Hex())
	}
	return user, nil
}

func (r *MongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found with ID: %s", id.Hex())
	}
	return nil
}

func (r *MongoUserRepository) FindAll(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error finding all users: %v", err)
		return nil, fmt.Errorf("failed to find all users: %w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &users); err != nil {
		log.Printf("Error decoding users from cursor: %v", err)
		return nil, fmt.Errorf("failed to decode all users: %w", err)
	}
	return users, nil
}
