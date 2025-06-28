package modules

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	// Import zap for utils.Logger
	"github.com/iots1/mingkwan-api/internal/shared/event"
	"github.com/iots1/mingkwan-api/internal/shared/utils" // Import utils for utils.Logger
	"github.com/iots1/mingkwan-api/internal/user/adapters"
	"github.com/iots1/mingkwan-api/internal/user/delivery"
	"github.com/iots1/mingkwan-api/internal/user/usecase"
)

func SetupUserModule(
	appCtx context.Context,
	db *mongo.Database,
	lowPublisher event.Publisher,
	highPublisher event.Publisher,
	inMemPubSub *event.InMemPubSub,
) *delivery.UserHandler {
	// ใช้ Logger.Info โดยรวบรวมข้อมูลทั้งหมดใน Field
	utils.Logger.Info("========== Setup User Module ==========")

	// 1. Initialize Repositories
	repo := adapters.NewMongoUserRepository(db, "users")
	utils.Logger.Debug("User module: User repository initialized.")

	// 2. Initialize Use Cases
	userUsecase := usecase.NewUserService(
		repo,
		lowPublisher,
		highPublisher,
	)
	utils.Logger.Debug("User module: User use case initialized.")

	// 3. Initialize and Start In-Memory Event Subscribers
	userInMemorySubscribers := delivery.NewUserInmemoryEventSubscribers(inMemPubSub)
	userInMemorySubscribers.StartAllSubscribers(appCtx)
	utils.Logger.Debug("User module: User in-memory event subscribers started.")

	// 4. Initialize Delivery Handlers (HTTP Handlers)
	userHandler := delivery.NewUserHandler(userUsecase)
	utils.Logger.Debug("User module: User HTTP handler initialized.")

	// บันทึก log สุดท้ายเมื่อตั้งค่าโมดูลเสร็จสมบูรณ์
	utils.Logger.Info("========== User module setup complete. ==========")
	return userHandler
}
