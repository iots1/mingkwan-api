package delivery

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	authAdapter "github.com/iots1/mingkwan-api/internal/auth/adapters"
	authModel "github.com/iots1/mingkwan-api/internal/auth/models"
	authUsecase "github.com/iots1/mingkwan-api/internal/auth/usecase"
	sharedModel "github.com/iots1/mingkwan-api/internal/shared/models"
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

type AuthHandler struct {
	authUsecase    authUsecase.AuthUsecase
	userUsecase    userUsecase.UserUsecase
	jwtGenerator   authAdapter.JWTTokenGenerator
	passwordHasher sharedAdapter.PasswordHasher
}

func NewAuthHandler(authUsecase authUsecase.AuthUsecase, userUsecase userUsecase.UserUsecase, jwtGenerator authAdapter.JWTTokenGenerator, passwordHasher sharedAdapter.PasswordHasher) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase, userUsecase: userUsecase, jwtGenerator: jwtGenerator, passwordHasher: passwordHasher}
}

func (h *AuthHandler) sendErrorResponse(c *fiber.Ctx, statusCode int, message string, err error, validationErrors map[string][]string) error {
	logFields := []zap.Field{
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.Int("status_code", statusCode),
		zap.String("message", message),
	}
	if err != nil {
		logFields = append(logFields, zap.Error(err))
	}
	if validationErrors != nil {
		logFields = append(logFields, zap.Any("validation_errors", validationErrors))
	}
	utils.Logger.Error("API Error", logFields...)

	return c.Status(statusCode).JSON(sharedModel.CommonErrorResponse{
		Success:   false,
		Timestamp: time.Now().UTC(),
		Message:   message,
		Errors:    validationErrors,
		Code:      statusCode * 1000,
		Method:    c.Method(),
		Path:      c.Path(),
	})
}

func (h *AuthHandler) sendSuccessResponse(c *fiber.Ctx, statusCode int, data interface{}, count int) error {
	return c.Status(statusCode).JSON(sharedModel.GenericSuccessResponse{
		Code:    statusCode,
		Success: true,
		Data:    data,
		Count:   count,
	})
}

func (s *AuthHandler) Register(c *fiber.Ctx) error {
	var req authModel.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		utils.Logger.Warn("Register: Invalid request body", zap.Error(err))
		return s.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err, nil)
	}

	if err := utils.GetGlobalValidator().Struct(req); err != nil {
		formattedErrors := utils.FormatValidationErrors(err)
		utils.Logger.Warn("Register: Validation failed", zap.Any("validation_details", formattedErrors))
		return s.sendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", nil, formattedErrors)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	existingUser, err := s.userUsecase.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		utils.Logger.Error("Error checking existing user by email", zap.Error(err), zap.String("email", req.Email))
		return s.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to check existing user", err, nil)
	}
	if existingUser != nil {
		utils.Logger.Warn("Registration failed: Email already exists", zap.String("email", req.Email))
		return s.sendErrorResponse(c, fiber.StatusConflict, ErrEmailAlreadyExists.Error(), nil, nil)
	}

	hashedPassword, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		utils.Logger.Error("Failed to hash password during registration", zap.Error(err))
		return s.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password", err, nil)
	}

	newUser := &userDomain.User{
		ID:        primitive.NewObjectID(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	createdUser, err := s.userUsecase.CreateUser(ctx, newUser)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserAlreadyExists) {
			utils.Logger.Info("Register: User already exists", zap.String("email", req.Email))
			return s.sendErrorResponse(c, fiber.StatusConflict, err.Error(), nil, nil)
		}
		utils.Logger.Error("Failed to create user in database", zap.Error(err), zap.String("email", req.Email))
		return s.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user", err, nil)
	}

	accessToken, refreshToken, err := s.jwtGenerator.GenerateTokens(createdUser.ID.Hex())
	if err != nil {
		utils.Logger.Error("Failed to generate tokens after registration", zap.Error(err), zap.String("userID", createdUser.ID.Hex()))
		return s.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate tokens", err, nil)
	}

	// Publish event (e.g., UserRegisteredEvent)
	// s.lowPublisher.Publish(ctx, event.NewUserRegisteredEvent(createdUser.ID.Hex(), createdUser.Email))
	utils.Logger.Info("User registered successfully", zap.String("userID", createdUser.ID.Hex()), zap.String("email", createdUser.Email))

	return s.sendSuccessResponse(c, fiber.StatusCreated, &authModel.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, 1)
}

func (s *AuthHandler) Login(ctx context.Context, req *authModel.LoginRequest) (*authModel.AuthResponse, error) {
	utils.Logger.Debug("Attempting user login", zap.String("email", req.Email))

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

func (s *AuthHandler) RefreshTokens(ctx context.Context, req *authModel.RefreshRequest) (*authModel.AuthResponse, error) {
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

	user, err := s.userUsecase.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			utils.Logger.Warn("Refresh failed: User not found for token", zap.String("userID", claims.UserID))
			return nil, ErrInvalidToken
		}
		utils.Logger.Error("Error finding user for refresh token", zap.Error(err), zap.String("userID", claims.UserID))
		return nil, err
	}

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

func (s *AuthHandler) GetProfile(ctx context.Context, userID primitive.ObjectID) (*authModel.ProfileResponse, error) {
	user, err := s.userUsecase.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			utils.Logger.Warn("Profile retrieval failed: User not found", zap.String("userID", userID.String()))
			return nil, ErrUserNotFound
		}
		utils.Logger.Error("Error finding user by ID for profile", zap.Error(err), zap.String("userID", userID.String()))
		return nil, err
	}

	return &authModel.ProfileResponse{
		ID:    user.ID.Hex(),
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
