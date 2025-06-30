package modules

import (
	"github.com/gofiber/fiber/v2"

	authAdapter "github.com/iots1/mingkwan-api/internal/auth/adapters"
	"github.com/iots1/mingkwan-api/internal/auth/delivery"
	authHandler "github.com/iots1/mingkwan-api/internal/auth/delivery"
	authUsecase "github.com/iots1/mingkwan-api/internal/auth/usecase"
	"github.com/iots1/mingkwan-api/internal/shared/infrastructure"
	"github.com/iots1/mingkwan-api/internal/shared/utils"
	"github.com/iots1/mingkwan-api/internal/user/usecase"
)

// SetupAuthModule initializes authentication dependencies and registers routes.
func SetupAuthModule(
	router fiber.Router,
	deps infrastructure.AppDependencies,
	userUsecase usecase.UserUsecase,
) {
	// Initialize JWT Token Generator

	jwtGenerator := authAdapter.NewJWTTokenGenerator(deps.AppConfig.SecretKey)

	authUsecase := authUsecase.NewAuthUsecase(
		userUsecase,
		jwtGenerator,
		deps.PasswordHasher,
		deps.LowPub,
		deps.HighPub,
	)

	if authUsecase == nil {
		utils.Logger.Error("AuthModule: authUsecase is nil, check your dependencies")
		panic("AuthUsecase is nil, check your dependencies")
	}

	authHandler := authHandler.NewAuthHandler(*authUsecase, userUsecase, jwtGenerator, deps.PasswordHasher)
	setupAuthRoutes(router, authHandler)
}

// RegisterAuthRoutes registers authentication routes with a Fiber group.
// This function assumes authHandler has its annotations in delivery layer.
func setupAuthRoutes(router fiber.Router, authHandler *delivery.AuthHandler) {
	auth := router.Group("/auth")
	// @Summary Register a new user
	// @Description Register a new user with name, email, and password
	// @Tags Auth
	// @Accept json
	// @Produce json
	// @Param request body authDelivery.RegisterRequest true "Register User"
	// @Success 201 {object} authDelivery.AuthResponse "User registered successfully"
	// @Failure 400 {object} models.CommonErrorResponse "Bad request or validation error" // Assuming CommonErrorResponse exists in models
	// @Failure 409 {object} models.CommonErrorResponse "Email already registered"
	// @Failure 500 {object} models.CommonErrorResponse "Internal server error"
	// @Router /api/v1/auth/register [post]
	auth.Post("/register", authHandler.Register)

	// @Summary User login
	// @Description Authenticate user and get access and refresh tokens
	// @Tags Auth
	// @Accept json
	// @Produce json
	// @Param request body authDelivery.LoginRequest true "Login Credentials"
	// @Success 200 {object} authDelivery.AuthResponse "Login successful"
	// @Failure 400 {object} models.CommonErrorResponse "Bad request or validation error"
	// @Failure 401 {object} models.CommonErrorResponse "Invalid credentials"
	// @Failure 500 {object} models.CommonErrorResponse "Internal server error"
	// @Router /api/v1/auth/login [post]
	// auth.Post("/login", authHandler.Login)

	// @Summary Refresh access token
	// @Description Use refresh token to get a new access token
	// @Tags Auth
	// @Accept json
	// @Produce json
	// @Param request body authDelivery.RefreshRequest true "Refresh Token"
	// @Success 200 {object} authDelivery.AuthResponse "Tokens refreshed successfully"
	// @Failure 400 {object} models.CommonErrorResponse "Bad request or invalid refresh token"
	// @Failure 401 {object} models.CommonErrorResponse "Unauthorized or expired refresh token"
	// @Failure 500 {object} models.CommonErrorResponse "Internal server error"
	// @Router /api/v1/auth/refresh [post]
	// auth.Post("/refresh", authHandler.Refresh)

	// @Summary Get user profile
	// @Description Get authenticated user's profile
	// @Tags Auth
	// @Security ApiKeyAuth
	// @Accept json
	// @Produce json
	// @Success 200 {object} authDelivery.ProfileResponse "User profile"
	// @Failure 401 {object} models.CommonErrorResponse "Unauthorized"
	// @Failure 500 {object} models.CommonErrorResponse "Internal server error"
	// @Router /api/v1/auth/profile [get]
	// auth.Get("/profile", authHandler.GetProfile) // Uncomment and add actual middleware/logic later
}
