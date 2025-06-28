// internal/user/delivery/user_handler.go
package delivery

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iots1/mingkwan-api/internal/shared/models" // Import your shared models (for responses)
	"github.com/iots1/mingkwan-api/internal/shared/utils"  // Import utils (for validator)
	"github.com/iots1/mingkwan-api/internal/user/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserUsecase defines the interface for user business logic.
type UserUsecase interface {
	CreateUser(ctx context.Context, name, email, password string) (*domain.User, error)
	GetUserByID(ctx context.Context, idStr string) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	UpdateUser(ctx context.Context, idStr, name, email string) (*domain.User, error)
	DeleteUser(ctx context.Context, idStr string) error
}

// UserHandler handles HTTP requests related to users.
type UserHandler struct {
	userUsecase UserUsecase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(usecase UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: usecase}
}

// --- Common Error Response Helper ---
// This function creates a standardized error response.
func (h *UserHandler) sendErrorResponse(c *fiber.Ctx, statusCode int, message string, err error, validationErrors map[string][]string) error {
	log.Printf("API Error - Method: %s, Path: %s, Status: %d, Message: %s, Details: %v, ValidationErrors: %+v",
		c.Method(), c.Path(), statusCode, message, err, validationErrors)

	return c.Status(statusCode).JSON(models.CommonErrorResponse{
		Success:   false,
		Timestamp: time.Now().UTC(),
		Message:   message,
		Errors:    validationErrors,  // Will be nil if not a validation error
		Code:      statusCode * 1000, // Example custom error code based on HTTP status
		Method:    c.Method(),
		Path:      c.Path(),
	})
}

// --- Common Success Response Helper ---
// This function creates a standardized success response.
func (h *UserHandler) sendSuccessResponse(c *fiber.Ctx, statusCode int, data interface{}, count int) error {
	return c.Status(statusCode).JSON(models.GenericSuccessResponse{
		Code:    statusCode,
		Success: true,
		Data:    data,
		Count:   count,
	})
}

// CreateUser handles POST /users to create a new user.
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err, nil)
	}

	if err := utils.GetGlobalValidator().Struct(req); err != nil {
		formattedErrors := utils.FormatValidationErrors(err)
		return h.sendErrorResponse(c, http.StatusBadRequest, "Validation failed", nil, formattedErrors)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	user, err := h.userUsecase.CreateUser(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user", err, nil)
	}

	return h.sendSuccessResponse(c, fiber.StatusCreated, toUserResponse(user), 1)
}

// GetUserByID handles GET /users/:id to get a user by ID.
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", err, nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	user, err := h.userUsecase.GetUserByID(ctx, id)
	if err != nil {
		// In a real app, check for "not found" error type from usecase/repo
		return h.sendErrorResponse(c, fiber.StatusNotFound, "User not found", err, nil)
	}

	return h.sendSuccessResponse(c, fiber.StatusOK, toUserResponse(user), 1)
}

// GetAllUsers handles GET /users to get all users.
func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	users, err := h.userUsecase.GetAllUsers(ctx)
	if err != nil {
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve users", err, nil)
	}

	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, toUserResponse(&user))
	}
	return h.sendSuccessResponse(c, fiber.StatusOK, userResponses, len(userResponses))
}

// UpdateUser handles PUT /users/:id to update an existing user.
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", err, nil)
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err, nil)
	}

	// Validation
	if err := utils.GetGlobalValidator().Struct(req); err != nil {
		formattedErrors := utils.FormatValidationErrors(err)
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", nil, formattedErrors)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	updatedUser, err := h.userUsecase.UpdateUser(ctx, id, req.Name, req.Email)
	if err != nil {
		// In a real app, differentiate between "not found" and other errors
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user", err, nil)
	}

	return h.sendSuccessResponse(c, fiber.StatusOK, toUserResponse(updatedUser), 1)
}

// DeleteUser handles DELETE /users/:id to delete a user.
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", err, nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	err := h.userUsecase.DeleteUser(ctx, id)
	if err != nil {
		// In a real app, differentiate between "not found" and other errors
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete user", err, nil)
	}

	// For successful DELETE, StatusNoContent means no body is returned.
	// Fiber's Send(nil) is appropriate here.
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// toUserResponse converts a domain.User to a UserResponse DTO
func toUserResponse(user *domain.User) UserResponse {
	// Add checks for nil user if it's possible for userUsecase to return nil without error
	if user == nil {
		return UserResponse{} // Return empty response or handle error
	}
	return UserResponse{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339), // Format to ISO 8601 string
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339), // Format to ISO 8601 string
	}
}
