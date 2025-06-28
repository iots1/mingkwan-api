package usecase

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/iots1/mingkwan-api/internal/shared/event"
	"github.com/iots1/mingkwan-api/internal/shared/utils"
	"github.com/iots1/mingkwan-api/internal/user/domain"
	"github.com/iots1/mingkwan-api/internal/user/repository"
)

type UserService struct {
	repo    repository.UserRepository
	lowPub  event.Publisher
	highPub event.Publisher
}

func NewUserService(
	repo repository.UserRepository,
	lowPub event.Publisher,
	highPub event.Publisher,
) *UserService {
	return &UserService{
		repo:    repo,
		lowPub:  lowPub,
		highPub: highPub,
	}
}

func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*domain.User, error) {
	existingUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		utils.Logger.Error("UserService: Error checking for existing user by email", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("error checking for existing user: %w", err)
	}
	if existingUser != nil {
		utils.Logger.Info("UserService: User with this email already exists", zap.String("email", email))
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.Logger.Error("UserService: Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &domain.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	createdUser, err := s.repo.CreateUser(ctx, newUser)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			utils.Logger.Warn("UserService: User already exists after creation attempt", zap.String("email", email))
			return nil, domain.ErrUserAlreadyExists
		}
		utils.Logger.Error("UserService: Failed to save user to database", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to save user to database: %w", err)
	}

	emailPayload := event.SendWelcomeEmailPayload{UserID: createdUser.ID.Hex(), Email: createdUser.Email, Name: createdUser.Name}
	if err := s.highPub.Publish(ctx, event.SendWelcomeEmailTaskName, emailPayload); err != nil {
		utils.Logger.Error("UserService: Failed to publish high importance send welcome email task",
			zap.String("user_email", createdUser.Email), zap.Error(err),
		)
	}

	utils.Logger.Debug("UserService: User created and events published", zap.String("name", name), zap.String("user_id", createdUser.ID.Hex()))
	return createdUser, nil
}

func (s *UserService) GetUserByID(ctx context.Context, idStr string) (*domain.User, error) {
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.Logger.Debug("GetUserByID: Invalid user ID format", zap.String("id_string", idStr))
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, objID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.Logger.Info("GetUserByID: User not found", zap.String("user_id", idStr))
			return nil, domain.ErrUserNotFound
		}
		utils.Logger.Error("GetUserByID: Failed to get user by ID", zap.String("user_id", idStr), zap.Error(err))
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		utils.Logger.Error("GetAllUsers: Failed to get all users", zap.Error(err))
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}

func (s *UserService) UpdateUser(ctx context.Context, idStr, name, email string) (*domain.User, error) {
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.Logger.Debug("UpdateUser: Invalid user ID format", zap.String("id_string", idStr))
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	existingUser, err := s.repo.GetUserByID(ctx, objID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.Logger.Info("UpdateUser: User not found for update", zap.String("user_id", idStr))
			return nil, domain.ErrUserNotFound
		}
		utils.Logger.Error("UpdateUser: Error finding existing user by ID", zap.String("user_id", idStr), zap.Error(err))
		return nil, fmt.Errorf("error finding user for update: %w", err)
	}

	updateMap := make(map[string]interface{})
	if name != "" {
		if existingUser.Name != name {
			updateMap["name"] = name
		}
	}
	if email != "" {
		if existingUser.Email != email {
			existingUserByEmail, err := s.repo.GetUserByEmail(ctx, email)
			if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
				utils.Logger.Error("UpdateUser: Error checking new email for existing user", zap.String("email", email), zap.Error(err))
				return nil, fmt.Errorf("error checking new email: %w", err)
			}
			if existingUserByEmail != nil && existingUserByEmail.ID != objID {
				utils.Logger.Warn("UpdateUser: New email already exists for another user", zap.String("email", email), zap.String("existing_user_id", existingUserByEmail.ID.Hex()))
				return nil, domain.ErrUserAlreadyExists
			}
			updateMap["email"] = email
		}
	}

	if len(updateMap) == 0 {
		utils.Logger.Info("UpdateUser: No fields to update", zap.String("user_id", idStr))
		return existingUser, nil
	}

	utils.Logger.Debug("UpdateUser: Preparing to update user with map",
		zap.String("user_id", objID.Hex()), zap.Any("update_map", updateMap))

	updatedUser, err := s.repo.UpdateUser(ctx, objID, updateMap)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.Logger.Info("UpdateUser: User not found", zap.String("user_id", idStr))
			return nil, domain.ErrUserNotFound
		}
		utils.Logger.Error("UpdateUser: Failed to update user in repository", zap.String("user_id", idStr), zap.Error(err))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

func (s *UserService) DeleteUser(ctx context.Context, idStr string) error {
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.Logger.Debug("DeleteUser: Invalid user ID format", zap.String("id_string", idStr))
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	err = s.repo.DeleteUser(ctx, objID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.Logger.Info("DeleteUser: User not found", zap.String("user_id", idStr))
			return domain.ErrUserNotFound
		}
		utils.Logger.Error("DeleteUser: Failed to delete user from repository", zap.String("user_id", idStr), zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
