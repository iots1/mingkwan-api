package delivery

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/iots1/mingkwan-api/internal/shared/models"
	"github.com/iots1/mingkwan-api/internal/shared/utils"
	"github.com/iots1/mingkwan-api/internal/user/domain"
)

type UserUsecase interface {
	CreateUser(ctx context.Context, name, email, password string) (*domain.User, error)
	GetUserByID(ctx context.Context, idStr string) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	UpdateUser(ctx context.Context, idStr, name, email string) (*domain.User, error)
	DeleteUser(ctx context.Context, idStr string) error
}

type UserHandler struct {
	userUsecase UserUsecase
}

func NewUserHandler(usecase UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: usecase}
}

func toUserResponse(user *domain.User) UserResponse {
	if user == nil {
		return UserResponse{}
	}
	return UserResponse{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *UserHandler) sendErrorResponse(c *fiber.Ctx, statusCode int, message string, err error, validationErrors map[string][]string) error {
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

	return c.Status(statusCode).JSON(models.CommonErrorResponse{
		Success:   false,
		Timestamp: time.Now().UTC(),
		Message:   message,
		Errors:    validationErrors,
		Code:      statusCode * 1000,
		Method:    c.Method(),
		Path:      c.Path(),
	})
}

func (h *UserHandler) sendSuccessResponse(c *fiber.Ctx, statusCode int, data interface{}, count int) error {
	return c.Status(statusCode).JSON(models.GenericSuccessResponse{
		Code:    statusCode,
		Success: true,
		Data:    data,
		Count:   count,
	})
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		utils.Logger.Warn("CreateUser: Invalid request body", zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err, nil)
	}

	if err := utils.GetGlobalValidator().Struct(req); err != nil {
		formattedErrors := utils.FormatValidationErrors(err)
		utils.Logger.Warn("CreateUser: Validation failed", zap.Any("validation_details", formattedErrors))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", nil, formattedErrors)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	user, err := h.userUsecase.CreateUser(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			utils.Logger.Info("CreateUser: User already exists", zap.String("email", req.Email))
			return h.sendErrorResponse(c, fiber.StatusConflict, err.Error(), nil, nil)
		}
		utils.Logger.Error("CreateUser: Failed to create user in usecase", zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user", err, nil)
	}

	utils.Logger.Info("User created successfully", zap.String("user_id", user.ID.Hex()), zap.String("email", user.Email))
	return h.sendSuccessResponse(c, fiber.StatusCreated, toUserResponse(user), 1)
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		utils.Logger.Warn("GetUserByID: User ID is empty in request params")
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "User ID is required", nil, nil)
	}
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		utils.Logger.Warn("GetUserByID: Invalid user ID format", zap.String("id", id), zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", err, nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	user, err := h.userUsecase.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.Logger.Info("GetUserByID: User not found", zap.String("user_id", id))
			return h.sendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil, nil)
		}
		utils.Logger.Error("GetUserByID: Usecase error", zap.String("user_id", id), zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve user", err, nil)
	}

	utils.Logger.Info("User retrieved successfully", zap.String("user_id", user.ID.Hex()))
	return h.sendSuccessResponse(c, fiber.StatusOK, toUserResponse(user), 1)
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	users, err := h.userUsecase.GetAllUsers(ctx)
	if err != nil {
		utils.Logger.Error("GetAllUsers: Usecase error", zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve users", err, nil)
	}

	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, toUserResponse(&user))
	}
	utils.Logger.Info("All users retrieved successfully", zap.Int("count", len(userResponses)))
	return h.sendSuccessResponse(c, fiber.StatusOK, userResponses, len(userResponses))
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		utils.Logger.Warn("UpdateUser: User ID is empty in request params")
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "User ID is required", nil, nil)
	}
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		utils.Logger.Warn("UpdateUser: Invalid user ID format", zap.String("id", id), zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", err, nil)
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		utils.Logger.Warn("UpdateUser: Invalid request body", zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err, nil)
	}

	if err := utils.GetGlobalValidator().Struct(req); err != nil {
		formattedErrors := utils.FormatValidationErrors(err)
		utils.Logger.Warn("UpdateUser: Validation failed", zap.Any("validation_details", formattedErrors))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", nil, formattedErrors)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	updatedUser, err := h.userUsecase.UpdateUser(ctx, id, req.Name, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.Logger.Info("UpdateUser: User not found", zap.String("user_id", id))
			return h.sendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil, nil)
		}
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			utils.Logger.Info("UpdateUser: Email already in use", zap.String("email", req.Email))
			return h.sendErrorResponse(c, fiber.StatusConflict, err.Error(), nil, nil)
		}
		utils.Logger.Error("UpdateUser: Failed to update user in usecase", zap.String("user_id", id), zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user", err, nil)
	}

	utils.Logger.Info("User updated successfully", zap.String("user_id", updatedUser.ID.Hex()))
	return h.sendSuccessResponse(c, fiber.StatusOK, toUserResponse(updatedUser), 1)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		utils.Logger.Warn("DeleteUser: User ID is empty in request params")
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "User ID is required", nil, nil)
	}
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		utils.Logger.Warn("DeleteUser: Invalid user ID format", zap.String("id", id), zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", err, nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	err := h.userUsecase.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.Logger.Info("DeleteUser: User not found", zap.String("user_id", id))
			return h.sendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil, nil)
		}
		utils.Logger.Error("DeleteUser: Failed to delete user in usecase", zap.String("user_id", id), zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete user", err, nil)
	}

	utils.Logger.Info("User deleted successfully", zap.String("user_id", id))
	return c.Status(fiber.StatusNoContent).Send(nil)
}
