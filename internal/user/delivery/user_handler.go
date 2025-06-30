package delivery

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	sharedAdapter "github.com/iots1/mingkwan-api/internal/shared/adapters"
	sharedModel "github.com/iots1/mingkwan-api/internal/shared/models"
	"github.com/iots1/mingkwan-api/internal/shared/utils"
	userDomain "github.com/iots1/mingkwan-api/internal/user/domain"
	userModel "github.com/iots1/mingkwan-api/internal/user/models"
	userUsecase "github.com/iots1/mingkwan-api/internal/user/usecase"
)

var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type UserHandler struct {
	userUsecase    userUsecase.UserUsecase
	passwordHasher sharedAdapter.PasswordHasher
}

func NewUserHandler(useUsecase userUsecase.UserUsecase, passswordHasher sharedAdapter.PasswordHasher) *UserHandler {
	return &UserHandler{userUsecase: useUsecase, passwordHasher: passswordHasher}
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

func (h *UserHandler) sendSuccessResponse(c *fiber.Ctx, statusCode int, data interface{}, count int) error {
	return c.Status(statusCode).JSON(sharedModel.GenericSuccessResponse{
		Code:    statusCode,
		Success: true,
		Data:    data,
		Count:   count,
	})
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error { // <--- เปลี่ยน Signature ตรงนี้ให้คืนค่าเป็น error เท่านั้น
	var req userModel.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		utils.Logger.Warn("CreateUser: Invalid request body", zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err, nil)
	}

	if err := utils.GetGlobalValidator().Struct(req); err != nil {
		formattedErrors := utils.FormatValidationErrors(err)
		utils.Logger.Warn("CreateUser: Validation failed", zap.Any("validation_details", formattedErrors))
		return h.sendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", nil, formattedErrors)
	}

	// เพิ่มการตรวจสอบผู้ใช้ที่มีอยู่แล้วตามโค้ดที่คุณให้มา
	existingUser, err := h.userUsecase.GetUserByEmail(c.Context(), req.Email) // ใช้ c.Context() เพื่อส่ง context มาตรฐาน
	if err != nil && !errors.Is(err, ErrUserNotFound) {                       // ErrUserNotFound ควรมาจาก domain หรือ usecase
		utils.Logger.Error("Error checking existing user by email", zap.Error(err), zap.String("email", req.Email))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to check existing user", err, nil) // เปลี่ยนการคืนค่าให้ถูกต้อง
	}
	if existingUser != nil {
		utils.Logger.Warn("Registration failed: Email already exists", zap.String("email", req.Email))
		return h.sendErrorResponse(c, fiber.StatusConflict, ErrEmailAlreadyExists.Error(), nil, nil) // เปลี่ยนการคืนค่าให้ถูกต้อง
	}

	hashedPassword, err := h.passwordHasher.HashPassword(req.Password)
	if err != nil {
		utils.Logger.Error("Failed to hash password during registration", zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password", err, nil) // เปลี่ยนการคืนค่าให้ถูกต้อง
	}

	// สร้าง *domain.User object จากข้อมูลที่ผ่านการ validate แล้ว
	// และใส่ Password ที่ถูก Hash แล้ว
	newUser := &userDomain.User{
		ID:        primitive.NewObjectID(), // กำหนด ID ใน Handler หรือ Usecase/Repository
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword, // ใช้รหัสผ่านที่ถูก Hash แล้ว (เปลี่ยนจาก Password เป็น PasswordHash เพื่อความชัดเจน)
		CreatedAt: time.Now(),     // กำหนดเวลาสร้าง
		UpdatedAt: time.Now(),     // กำหนดเวลาอัปเดต
		IsActive:  true,           // กำหนดสถานะ
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	user, err := h.userUsecase.CreateUser(ctx, newUser)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserAlreadyExists) {
			utils.Logger.Info("CreateUser: User already exists", zap.String("email", req.Email))
			return h.sendErrorResponse(c, fiber.StatusConflict, err.Error(), nil, nil)
		}
		utils.Logger.Error("CreateUser: Failed to create user in usecase", zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user", err, nil)
	}

	utils.Logger.Info("User created successfully", zap.String("user_id", user.ID.Hex()), zap.String("email", user.Email))
	// สำหรับการส่ง Success Response, sendSuccessResponse จะจัดการการส่ง JSON กลับไป
	return h.sendSuccessResponse(c, fiber.StatusCreated, userModel.ToUserResponse(user), 1)
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
		if errors.Is(err, userDomain.ErrUserNotFound) {
			utils.Logger.Info("GetUserByID: User not found", zap.String("user_id", id))
			return h.sendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil, nil)
		}
		utils.Logger.Error("GetUserByID: Usecase error", zap.String("user_id", id), zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve user", err, nil)
	}

	utils.Logger.Info("User retrieved successfully", zap.String("user_id", user.ID.Hex()))
	return h.sendSuccessResponse(c, fiber.StatusOK, userModel.ToUserResponse(user), 1)
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	users, err := h.userUsecase.GetAllUsers(ctx)
	if err != nil {
		utils.Logger.Error("GetAllUsers: Usecase error", zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve users", err, nil)
	}

	var userResponses []userModel.UserResponse
	for _, user := range users {
		userRespPtr := userModel.ToUserResponse(&user)
		userResponses = append(userResponses, *userRespPtr)
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

	var req userModel.UpdateUserRequest
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
		if errors.Is(err, userDomain.ErrUserNotFound) {
			utils.Logger.Info("UpdateUser: User not found", zap.String("user_id", id))
			return h.sendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil, nil)
		}
		if errors.Is(err, userDomain.ErrUserAlreadyExists) {
			utils.Logger.Info("UpdateUser: Email already in use", zap.String("email", req.Email))
			return h.sendErrorResponse(c, fiber.StatusConflict, err.Error(), nil, nil)
		}
		utils.Logger.Error("UpdateUser: Failed to update user in usecase", zap.String("user_id", id), zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user", err, nil)
	}

	utils.Logger.Info("User updated successfully", zap.String("user_id", updatedUser.ID.Hex()))
	return h.sendSuccessResponse(c, fiber.StatusOK, userModel.ToUserResponse(updatedUser), 1)
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
		if errors.Is(err, userDomain.ErrUserNotFound) {
			utils.Logger.Info("DeleteUser: User not found", zap.String("user_id", id))
			return h.sendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil, nil)
		}
		utils.Logger.Error("DeleteUser: Failed to delete user in usecase", zap.String("user_id", id), zap.Error(err))
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete user", err, nil)
	}

	utils.Logger.Info("User deleted successfully", zap.String("user_id", id))
	return c.Status(fiber.StatusNoContent).Send(nil)
}
