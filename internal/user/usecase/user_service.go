// internal/user/usecase/user_service.go
package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/iots1/mingkwan-api/internal/shared/event"
	"github.com/iots1/mingkwan-api/internal/user/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// UserService implements the business logic for user management.
type UserService struct {
	repo    domain.UserRepository // Correct field name
	lowPub  event.Publisher       // Correct field name for low importance publisher
	highPub event.Publisher       // Correct field name for high importance publisher
	// inMemoryBus *event.InMemPubSub // <-- REMOVED: No longer needed here as lowPub handles it
}

// NewUserService creates a new UserService instance.
func NewUserService(
	repo domain.UserRepository,
	lowPub event.Publisher, // Correct parameter name and type
	highPub event.Publisher, // Correct parameter name and type
	// inMemoryBus *event.InMemPubSub, // <-- REMOVED: No longer a parameter for UserService
) *UserService {
	return &UserService{
		repo:    repo,
		lowPub:  lowPub,
		highPub: highPub,
		// inMemoryBus: inMemoryBus, // <-- REMOVED: No longer initialized here
	}
}

// CreateUser handles creating a new user and publishing relevant events.
func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*domain.User, error) {
	log.Printf("UserService: Attempting to create user %s with email %s\n", name, email)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &domain.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	// Corrected: Use s.repo
	createdUser, err := s.repo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to save user to database: %w", err)
	}

	// 1. Publish a low-importance event (In-Memory, e.g., for internal logging/metrics or other in-process features)
	// Corrected: Use s.lowPub
	inMemoryPayload := event.UserCreatedPayload{UserID: createdUser.ID, Name: createdUser.Name, Email: createdUser.Email} // Use ObjectID for UserID
	if err := s.lowPub.Publish(ctx, string(event.UserCreatedInMemoryEvent), inMemoryPayload); err != nil {
		log.Printf("Warning: Failed to publish low importance user created event: %v\n", err)
	}

	// 2. Publish a high-importance task (Asynq, e.g., send welcome email)
	// Corrected: Use s.highPub
	emailPayload := event.SendWelcomeEmailPayload{UserID: createdUser.ID.Hex(), Email: createdUser.Email, Name: createdUser.Name}
	if err := s.highPub.Publish(ctx, event.SendWelcomeEmailTaskName, emailPayload); err != nil {
		log.Printf("Error: Failed to publish high importance send welcome email task for user %s: %v\n", createdUser.Email, err)
		// Depending on criticality, you might return an error to the client here
		// If this error means the user creation itself failed, return it.
		// If it's just a background task failure that can be retried, you might just log.
		// For this example, let's allow user creation to proceed even if email queue fails initially.
		// return nil, fmt.Errorf("failed to queue welcome email: %w", err) // Uncomment if you want to fail user creation
	}

	log.Printf("UserService: User %s created (ID: %s) and events published.\n", name, createdUser.ID.Hex())
	return createdUser, nil
}

// GetUserByID retrieves a user by ID.
func (s *UserService) GetUserByID(ctx context.Context, idStr string) (*domain.User, error) {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	// Corrected: Use s.repo
	return s.repo.FindByID(ctx, id)
}

// GetAllUsers retrieves all users.
func (s *UserService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	// Corrected: Use s.repo
	return s.repo.FindAll(ctx)
}

// UpdateUser updates an existing user and publishes events.
func (s *UserService) UpdateUser(ctx context.Context, idStr, name, email string) (*domain.User, error) {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Corrected: Use s.repo
	existingUser, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err // User not found
	}

	existingUser.Name = name
	existingUser.Email = email
	// Password update should be handled separately

	// Corrected: Use s.repo
	updatedUser, err := s.repo.Update(ctx, existingUser)
	if err != nil {
		return nil, fmt.Errorf("failed to update user in database: %w", err)
	}

	// Publish a low-importance event for user update
	// Corrected: Use s.lowPub
	inMemoryPayload := event.UserUpdatedPayload{UserID: updatedUser.ID, Name: updatedUser.Name, Email: updatedUser.Email} // Use ObjectID for UserID
	if err := s.lowPub.Publish(ctx, string(event.UserUpdatedInMemoryEvent), inMemoryPayload); err != nil {
		log.Printf("Warning: Failed to publish low importance user updated event: %v\n", err)
	}

	log.Printf("UserService: User %s updated (ID: %s) and events published.\n", name, updatedUser.ID.Hex())
	return updatedUser, nil
}

// DeleteUser deletes a user and publishes events.
func (s *UserService) DeleteUser(ctx context.Context, idStr string) error {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	// Corrected: Use s.repo
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user from database: %w", err)
	}

	// Publish a low-importance event for user deletion
	// Corrected: Use s.lowPub
	inMemoryPayload := event.UserDeletedPayload{UserID: id} // Use ObjectID for UserID
	if err := s.lowPub.Publish(ctx, string(event.UserDeletedInMemoryEvent), inMemoryPayload); err != nil {
		log.Printf("Warning: Failed to publish low importance user deleted event: %v\n", err)
	}

	log.Printf("UserService: User ID %s deleted and event published.\n", id.Hex())
	return nil
}
