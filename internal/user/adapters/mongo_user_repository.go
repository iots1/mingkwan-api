package adapters

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/iots1/mingkwan-api/internal/shared/utils"
	"github.com/iots1/mingkwan-api/internal/user/domain"
	"github.com/iots1/mingkwan-api/internal/user/repository"
)

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database, collectionName string) *MongoUserRepository {
	return &MongoUserRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *MongoUserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	existingUser, err := r.GetUserByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		utils.Logger.Error("MongoUserRepository: Error checking for existing user by email during creation",
			zap.String("email", user.Email), zap.Error(err))
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}
	if existingUser != nil {
		utils.Logger.Info("MongoUserRepository: User with this email already exists", zap.String("email", user.Email))
		return nil, domain.ErrUserAlreadyExists
	}

	res, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, we := range writeException.WriteErrors {
				if we.Code == 11000 {
					utils.Logger.Warn("MongoUserRepository: Duplicate email found during insert", zap.String("email", user.Email))
					return nil, domain.ErrUserAlreadyExists
				}
			}
		}
		utils.Logger.Error("MongoUserRepository: Failed to insert new user", zap.Error(err))
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid
	} else {
		utils.Logger.Warn("MongoUserRepository: Could not convert InsertedID to ObjectID", zap.Any("inserted_id", res.InsertedID))
		return nil, fmt.Errorf("failed to retrieve inserted ID")
	}

	return user, nil
}

func (r *MongoUserRepository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	return &user, nil
}

func (r *MongoUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

func (r *MongoUserRepository) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get users cursor: %w", err)
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}
	return users, nil
}

func (r *MongoUserRepository) UpdateUser(ctx context.Context, id primitive.ObjectID, update map[string]interface{}) (*domain.User, error) {
	filter := bson.M{"_id": id}
	update["updated_at"] = time.Now()
	updateDoc := bson.M{"$set": update}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedUser domain.User
	err := r.collection.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&updatedUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return &updatedUser, nil
}

func (r *MongoUserRepository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

var _ repository.UserRepository = (*MongoUserRepository)(nil)
