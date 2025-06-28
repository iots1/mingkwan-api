package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	// <-- เพิ่ม import นี้สำหรับ mongo.Database
	"github.com/iots1/mingkwan-api/config"
	"github.com/iots1/mingkwan-api/internal/modules"
	"github.com/iots1/mingkwan-api/internal/shared/cache"
	"github.com/iots1/mingkwan-api/internal/shared/event"
	"github.com/iots1/mingkwan-api/internal/shared/infrastructure"
	"github.com/iots1/mingkwan-api/internal/shared/utils"
	// ยังคงต้องการ delivery เพื่อเข้าถึง UserHandler
)

func main() {
	// --- 0. Setup Global Application Context ---
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// --- 1. Initialize Configuration ---
	config.InitConfig()

	v := validator.New()
	utils.SetGlobalValidator(v)
	log.Println("Global validator initialized and set.")

	appConfig := config.LoadAppConfig()
	mongoConfig := config.LoadMongoConfig()
	redisConfig := config.LoadRedisConfig()

	// --- 2. Initialize Infrastructure Connections ---
	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()

	// 2.1. MongoDB
	mongoClient := infrastructure.NewMongoClient(mongoConfig.URI, mongoConfig.DBName)
	_, err := mongoClient.Connect(initCtx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	db := mongoClient.GetDatabase()

	// 2.2. Redis (for general use, e.g., caching, sessions)
	redisClientConn := infrastructure.NewRedisClient(redisConfig.Addr, redisConfig.Password, redisConfig.DB)
	var rdb *redis.Client
	rdb, err = redisClientConn.Connect(initCtx)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Connected to general Redis client: %s (DB: %d)", redisConfig.Addr, redisConfig.DB)

	cacheManager := cache.NewCacheManager(rdb)
	testKey := "my_test_key"
	testValue := "hello from redis cache"
	if err := cacheManager.Set(initCtx, testKey, testValue, 1*time.Minute); err != nil {
		log.Printf("Failed to set test key in Redis cache: %v", err)
	} else {
		log.Printf("Successfully set '%s' in Redis cache.", testKey)
		if val, getErr := cacheManager.Get(initCtx, testKey); getErr == nil {
			log.Printf("Retrieved '%s' from Redis cache: %s", testKey, val)
		}
	}

	// 2.3. Initialize In-Memory Event Bus (InMemPubSub) - needed for subscription and LowImportancePublisher
	inMemPubSub := event.NewInMemoryBus()
	log.Println("Initialized In-Memory Event Bus (InMemPubSub).")

	// 2.4. Initialize Asynq Client (the concrete implementation of event.AsynqClient interface)
	asynqRedisOpt := event.GetRedisClientOpt(redisConfig.Addr, redisConfig.Password, redisConfig.DB)
	asynqConcreteClient := event.NewAsynqClient(asynqRedisOpt) // Returns *event.AsynqClientImpl
	log.Printf("Initialized Asynq concrete client connecting to Redis: %s (DB: %d)", redisConfig.Addr, redisConfig.DB)

	// <--- NEW: 2.5. Initialize Low and High Importance Publishers
	lowPublisher := event.NewLowImportancePublisher(inMemPubSub)           // In-memory events
	highPublisher := event.NewHighImportancePublisher(asynqConcreteClient) // Asynq tasks
	log.Println("Initialized Low and High Importance Publishers.")

	// --- 3. Setup User Module ---
	// NEW: Call the setup function for the User module
	userHandler := modules.SetupUserModule(appCtx, db, lowPublisher, highPublisher, inMemPubSub)
	log.Println("User module setup complete.")

	// --- 6. Setup Fiber App ---
	app := fiber.New()

	// Register routes
	userRoutes := app.Group("/users")
	userRoutes.Post("/", userHandler.CreateUser)
	userRoutes.Get("/:id", userHandler.GetUserByID)
	userRoutes.Get("/", userHandler.GetAllUsers)
	userRoutes.Put("/:id", userHandler.UpdateUser)
	userRoutes.Delete("/:id", userHandler.DeleteUser)

	// ... other routes and middleware

	// --- 7. Start Server in a Goroutine ---
	go func() {
		port := fmt.Sprintf(":%d", appConfig.Port)
		if err := app.Listen(port); err != nil {
			log.Fatalf("Fiber server failed to start: %v", err)
		}
	}()
	log.Printf("Fiber server listening on port %d in %s environment", appConfig.Port, appConfig.Environment)

	// --- 8. Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until OS signal is received (SIGINT or SIGTERM)

	log.Println("Shutting down application...")

	// 8.1. Cancel the main application context to signal goroutines to stop
	appCancel()

	// 8.2. Create a context for services graceful shutdown (e.g., waiting for connections to close)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 8.3. Shut down Fiber app
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Fatalf("Fiber server forced to shutdown: %v", err)
	}
	log.Println("Fiber server gracefully stopped.")

	// 8.4. Disconnect Infrastructure Clients
	mongoClient.Disconnect(shutdownCtx)
	log.Println("MongoDB disconnected.")

	redisClientConn.Disconnect()
	log.Println("General Redis client disconnected.")

	// Ensure Asynq client is closed
	if asynqConcreteClient != nil { // Use the concrete client instance for closing
		if err := asynqConcreteClient.Close(); err != nil {
			log.Printf("Error closing Asynq client: %v", err)
		} else {
			log.Println("Asynq client disconnected.")
		}
	}

	// Give in-memory goroutines a moment to respond to context cancellation
	time.Sleep(1 * time.Second)

	log.Println("Application fully stopped.")
}
