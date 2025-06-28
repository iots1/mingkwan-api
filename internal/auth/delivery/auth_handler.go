package delivery

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	authModel "github.com/iots1/mingkwan-api/internal/auth/models"
	sharedModel "github.com/iots1/mingkwan-api/internal/shared/models"
	"github.com/iots1/mingkwan-api/internal/shared/utils"
	userDomain "github.com/iots1/mingkwan-api/internal/user/domain"
	userModel "github.com/iots1/mingkwan-api/internal/user/models"
)

type AuthUsecase interface {
	Register(ctx context.Context, name, email, password string) (*userDomain.User, string, string, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
}

type AuthHandler struct {
	authUsecase AuthUsecase
}

func NewAuthHandler(authUsecase AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
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

// @Summary Register a new user
// @Description Register a new user with name, email, and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register User"
// @Success 201 {object} AuthResponse "User registered successfully"
// @Failure 400 {object} models.CommonErrorResponse "Bad request or validation error"
// @Failure 409 {object} models.CommonErrorResponse "Email already registered"
// @Failure 500 {object} models.CommonErrorResponse "Internal server error"
// @Router api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req authModel.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		utils.Logger.Warn("AuthHandler: Register - Invalid request body", zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err, nil)
	}

	if err := utils.GetGlobalValidator().Struct(req); err != nil {
		formattedErrors := utils.FormatValidationErrors(err)
		utils.Logger.Warn("AuthHandler: Register - Validation failed", zap.Any("validation_details", formattedErrors))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", nil, formattedErrors)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	user, accessToken, refreshToken, err := h.authUsecase.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserAlreadyExists) {
			utils.Logger.Info("AuthHandler: Register - User with email already exists", zap.String("email", req.Email))
			return h.sendErrorResponse(c, fiber.StatusConflict, "Email already registered", nil, nil)
		}
		utils.Logger.Error("AuthHandler: Register - Failed to register user through usecase", zap.Error(err), zap.String("email", req.Email))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to register user", err, nil)
	}

	accessTokenExpiresIn := int64(15 * 60)

	utils.Logger.Debug("AuthHandler: User registered successfully", zap.String("user_id", user.ID.Hex()), zap.String("email", user.Email))
	return h.sendSuccessResponse(c, fiber.StatusCreated, authModel.AuthResponse{
		User:         userModel.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    accessTokenExpiresIn,
		TokenType:    "Bearer",
	}, 1)
}
