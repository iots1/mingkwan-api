package modules

import (
	"github.com/gofiber/fiber/v2"
	"github.com/iots1/mingkwan-api/internal/shared/infrastructure"
	"github.com/iots1/mingkwan-api/internal/shared/utils"
	"github.com/iots1/mingkwan-api/internal/user/adapters"
	"github.com/iots1/mingkwan-api/internal/user/delivery"
	userUsecase "github.com/iots1/mingkwan-api/internal/user/usecase"
)

func SetupUserModule(
	router fiber.Router,
	deps infrastructure.AppDependencies,
) *userUsecase.UserUsecase {
	utils.Logger.Info("========== Setup User Module ==========")

	repo := adapters.NewMongoUserRepository(deps.DB, "users")
	utils.Logger.Debug("User module: User repository initialized.")

	userUsecase := userUsecase.NewUserUsecase(
		repo,
		deps.LowPub,
		deps.HighPub,
	)
	utils.Logger.Debug("User module: User use case initialized.")

	userInMemorySubscribers := delivery.NewUserInmemoryEventSubscribers(deps.InMemPubSub)
	userInMemorySubscribers.StartAllSubscribers(deps.AppCtx)
	utils.Logger.Debug("User module: User in-memory event subscribers started.")

	if userUsecase == nil {
		utils.Logger.Error("AuthModule: authUsecase is nil, check your dependencies")
		panic("AuthUsecase is nil, check your dependencies")
	}

	userHandler := delivery.NewUserHandler(*userUsecase, deps.PasswordHasher)

	setupRouters(router, userHandler)
	utils.Logger.Info("========== User module setup complete. ==========")

	return userUsecase
}

func setupRouters(router fiber.Router, handler *delivery.UserHandler) {
	userRoutes := router.Group("/users")
	userRoutes.Post("/", handler.CreateUser)
	userRoutes.Get("/:id", handler.GetUserByID)
	userRoutes.Get("/", handler.GetAllUsers)
	userRoutes.Put("/:id", handler.UpdateUser)
	userRoutes.Delete("/:id", handler.DeleteUser)
}
