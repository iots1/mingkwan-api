// internal/modules/user_module.go
package modules // <-- สำคัญ: เปลี่ยน package เป็น modules (หรือชื่ออื่นที่คุณต้องการ)

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo" // For mongo.Database

	"github.com/iots1/mingkwan-api/internal/shared/event"
	"github.com/iots1/mingkwan-api/internal/user/adapters"
	"github.com/iots1/mingkwan-api/internal/user/delivery"
	"github.com/iots1/mingkwan-api/internal/user/usecase"
)

// SetupUserModule initializes all components related to the User module.
// It sets up repositories, use cases, event subscribers, and delivery handlers.
// This function acts as a "wire" or "bootstrap" for the user module.
func SetupUserModule(
	appCtx context.Context, // Main application context for graceful shutdown
	db *mongo.Database,
	lowPublisher event.Publisher,
	highPublisher event.Publisher,
	inMemPubSub *event.InMemPubSub,
) *delivery.UserHandler { // Returns the UserHandler to be registered in Fiber
	log.Println("Setting up User module...")

	// --- 1. Initialize Repositories ---
	userRepo := adapters.NewMongoUserRepository(db, "users")
	log.Println("  - User repository initialized.")

	// --- 2. Initialize Use Cases ---
	userUsecase := usecase.NewUserService(
		userRepo,
		lowPublisher,
		highPublisher,
	)
	log.Println("  - User use case initialized.")

	// --- 3. Initialize and Start In-Memory Event Subscribers ---
	// The subscribers still need the raw inMemPubSub to subscribe to events.
	// If userUsecase is needed by these subscribers, pass it here.
	userInMemorySubscribers := delivery.NewUserInmemoryEventSubscribers(inMemPubSub /*, userUsecase */)
	userInMemorySubscribers.StartAllSubscribers(appCtx) // Start the goroutines using appCtx
	log.Println("  - User in-memory event subscribers started.")

	// --- 4. Initialize Delivery Handlers (HTTP Handlers) ---
	userHandler := delivery.NewUserHandler(userUsecase)
	log.Println("  - User HTTP handler initialized.")

	log.Println("User module setup complete.")
	return userHandler
}
