package usecase

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	authAdapter "github.com/iots1/mingkwan-api/internal/auth/adapters"
	authModel "github.com/iots1/mingkwan-api/internal/auth/models"

	"github.com/iots1/mingkwan-api/internal/shared/event"
	"github.com/iots1/mingkwan-api/internal/shared/utils"
	userDomain "github.com/iots1/mingkwan-api/internal/user/domain"
	userUsecase "github.com/iots1/mingkwan-api/internal/user/usecase"

	sharedAdapter "github.com/iots1/mingkwan-api/internal/shared/adapters"
)

var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type AuthUsecase struct {
	userUsecase    userUsecase.UserUsecase
	jwtGenerator   authAdapter.JWTTokenGenerator
	passwordHasher sharedAdapter.PasswordHasher
	lowPublisher   event.Publisher
	highPublisher  event.Publisher
}

func NewAuthUsecase(
	userUsecase userUsecase.UserUsecase,
	jwtGenerator authAdapter.JWTTokenGenerator,
	passwordHasher sharedAdapter.PasswordHasher,
	inMemPubSub event.Publisher,
	asynqClient event.Publisher,
) *AuthUsecase {

	return &AuthUsecase{
		userUsecase:    userUsecase,
		jwtGenerator:   jwtGenerator,
		passwordHasher: passwordHasher,
		lowPublisher:   inMemPubSub,
		highPublisher:  asynqClient,
	}
}

// Register creates a new user.
func (s *AuthUsecase) Register(ctx context.Context, data *userDomain.User) (*authModel.AuthResponse, error) {

	createdUser, err := s.userUsecase.CreateUser(ctx, data)
	if err != nil {
		utils.Logger.Error("Failed to create user in database", zap.Error(err), zap.String("email", data.Email))
		return nil, errors.New("failed to create user")
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwtGenerator.GenerateTokens(createdUser.ID.Hex())
	if err != nil {
		utils.Logger.Error("Failed to generate tokens after registration", zap.Error(err), zap.String("userID", createdUser.ID.Hex()))
		return nil, errors.New("failed to generate tokens")
	}

	utils.Logger.Info("User registered successfully", zap.String("userID", createdUser.ID.Hex()), zap.String("email", createdUser.Email))

	return &authModel.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login authenticates a user and generates tokens.
func (s *AuthUsecase) Login(ctx context.Context, req *authModel.LoginRequest) (*authModel.AuthResponse, error) {
	utils.Logger.Info("Attempting user login", zap.String("email", req.Email))

	// Find user by email
	user, err := s.userUsecase.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			utils.Logger.Warn("Login failed: User not found", zap.String("email", req.Email))
			return nil, ErrInvalidCredentials
		}
		utils.Logger.Error("Error finding user by email during login", zap.Error(err), zap.String("email", req.Email))
		return nil, err
	}

	// Check password
	if !s.passwordHasher.CheckPasswordHash(req.Password, user.Password) {
		utils.Logger.Warn("Login failed: Invalid password", zap.String("email", req.Email))
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwtGenerator.GenerateTokens(user.ID.Hex())
	if err != nil {
		utils.Logger.Error("Failed to generate tokens after login", zap.Error(err), zap.String("userID", user.ID.Hex()))
		return nil, errors.New("failed to generate tokens")
	}

	// Publish event (e.g., UserLoggedInEvent)
	// s.highPublisher.Publish(ctx, event.NewUserLoggedInEvent(user.ID.Hex()))
	utils.Logger.Info("User logged in successfully", zap.String("userID", user.ID.Hex()), zap.String("email", user.Email))

	return &authModel.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshTokens refreshes access and refresh tokens.
func (s *AuthUsecase) RefreshTokens(ctx context.Context, req *authModel.RefreshRequest) (*authModel.AuthResponse, error) {
	utils.Logger.Info("Attempting to refresh tokens")

	// Parse and validate refresh token
	claims, err := s.jwtGenerator.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		utils.Logger.Warn("Refresh token invalid or expired", zap.Error(err))
		return nil, ErrInvalidToken
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		utils.Logger.Warn("Invalid user ID format in refresh token", zap.String("userID", claims.UserID), zap.Error(err))
		return nil, ErrInvalidToken
	}

	// Check if user exists (optional, but good practice for security)
	user, err := s.userUsecase.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			utils.Logger.Warn("Refresh failed: User not found for token", zap.String("userID", claims.UserID))
			return nil, ErrInvalidToken
		}
		utils.Logger.Error("Error finding user for refresh token", zap.Error(err), zap.String("userID", claims.UserID))
		return nil, err
	}

	// Generate new tokens
	newAccessToken, newRefreshToken, err := s.jwtGenerator.GenerateTokens(user.ID.Hex())
	if err != nil {
		utils.Logger.Error("Failed to generate new tokens during refresh", zap.Error(err), zap.String("userID", user.ID.Hex()))
		return nil, errors.New("failed to generate new tokens")
	}

	utils.Logger.Info("Tokens refreshed successfully", zap.String("userID", user.ID.Hex()))
	return &authModel.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthUsecase) GetProfile(ctx context.Context, userID string) (*authModel.ProfileResponse, error) {
	utils.Logger.Info("Attempting to retrieve user profile", zap.String("userID", userID))

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.Logger.Warn("Invalid user ID format", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInvalidToken
	}

	user, err := s.userUsecase.GetUserByID(ctx, oid)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			utils.Logger.Warn("Profile retrieval failed: User not found", zap.String("userID", userID))
			return nil, ErrUserNotFound
		}
		utils.Logger.Error("Error finding user by ID for profile", zap.Error(err), zap.String("userID", userID))
		return nil, err
	}

	return &authModel.ProfileResponse{
		ID:    user.ID.Hex(),
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
