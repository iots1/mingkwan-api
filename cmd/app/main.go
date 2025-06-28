// cmd/app/main.go
package main

import (
	"context"
	"fmt" // Keep for fmt.Sprintf
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap" // For zap.Error, zap.String etc.

	"github.com/iots1/mingkwan-api/config"
	"github.com/iots1/mingkwan-api/internal/modules"
	"github.com/iots1/mingkwan-api/internal/shared/cache"
	"github.com/iots1/mingkwan-api/internal/shared/event"
	"github.com/iots1/mingkwan-api/internal/shared/infrastructure"
	"github.com/iots1/mingkwan-api/internal/shared/utils" // This package now contains the zap logger
)

func main() {
	// --- 0. Setup Global Application Context ---
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// --- 1. Initialize Configuration ---
	config.InitConfig()
	appConfig := config.LoadAppConfig()
	mongoConfig := config.LoadMongoConfig()
	redisConfig := config.LoadRedisConfig()
	loggerLevel := config.LoadLoggerConfig()

	// --- Initialize Zap Logger FIRST ---
	utils.InitLogger(appConfig.Environment, loggerLevel)
	defer utils.SyncLogger()

	utils.Logger.Info("Application is starting up...")

	// Initialize the global validator
	v := validator.New()
	utils.SetGlobalValidator(v)
	utils.Logger.Debug("Global validator initialized and set.")

	// --- 2. Initialize Infrastructure Connections ---
	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()

	// 2.1. MongoDB
	mongoClient := infrastructure.NewMongoClient(mongoConfig.URI, mongoConfig.DBName)
	// Use _ to discard the return value, but declare err in the main scope
	var err error                                          // Declare err here in the main() function's scope
	if _, err = mongoClient.Connect(initCtx); err != nil { // Assign to the declared err
		utils.Logger.Fatal("Failed to connect to MongoDB", zap.Error(err)) // Use Zap Fatal
	}
	db := mongoClient.GetDatabase()
	utils.Logger.Info("Connected to MongoDB", zap.String("database", mongoConfig.DBName))

	// 2.2. Redis (for general use, e.g., caching, sessions)
	redisClientConn := infrastructure.NewRedisClient(redisConfig.Addr, redisConfig.Password, redisConfig.DB)
	var rdb *redis.Client                                        // Declare rdb here
	if rdb, err = redisClientConn.Connect(initCtx); err != nil { // Assign to existing rdb and err
		utils.Logger.Fatal("Failed to connect to Redis", zap.Error(err)) // Use Zap Fatal
	}
	utils.Logger.Info("Connected to general Redis client", zap.String("address", redisConfig.Addr), zap.Int("db", redisConfig.DB))

	cacheManager := cache.NewCacheManager(rdb)
	testKey := "my_test_key"
	testValue := "hello from redis cache"
	if err = cacheManager.Set(initCtx, testKey, testValue, 1*time.Minute); err != nil { // Assign to existing err
		utils.Logger.Warn("Failed to set test key in Redis cache", zap.Error(err))
	} else {
		if val, getErr := cacheManager.Get(initCtx, testKey); getErr == nil {
			utils.Logger.Debug("Retrieved test key from Redis cache", zap.String("key", testKey), zap.String("value", val))
		}
	}

	// 2.3. Initialize In-Memory Event Bus (InMemPubSub)
	inMemPubSub := event.NewInMemoryBus()
	utils.Logger.Info("Initialized In-Memory Event Bus (InMemPubSub).")

	// 2.4. Initialize Asynq Client
	asynqRedisOpt := event.GetRedisClientOpt(redisConfig.Addr, redisConfig.Password, redisConfig.DB)
	asynqConcreteClient := event.NewAsynqClient(asynqRedisOpt)
	utils.Logger.Info("Initialized Asynq client", zap.String("address", redisConfig.Addr), zap.Int("db", redisConfig.DB))

	// 2.5. Initialize Low and High Importance Publishers
	lowPublisher := event.NewLowImportancePublisher(inMemPubSub)
	highPublisher := event.NewHighImportancePublisher(asynqConcreteClient)
	utils.Logger.Info("Initialized Low and High Importance Publishers.")

	// --- 3. Setup User Module ---
	userHandler := modules.SetupUserModule(appCtx, db, lowPublisher, highPublisher, inMemPubSub)

	// --- 4. Setup Fiber App ---
	app := fiber.New()

	// Register routes for User module
	userRoutes := app.Group("/users")
	userRoutes.Post("/", userHandler.CreateUser)
	userRoutes.Get("/:id", userHandler.GetUserByID)
	userRoutes.Get("/", userHandler.GetAllUsers)
	userRoutes.Put("/:id", userHandler.UpdateUser)
	userRoutes.Delete("/:id", userHandler.DeleteUser)

	// ... other routes and middleware

	// --- 5. Start Server in a Goroutine ---
	go func() {
		port := fmt.Sprintf(":%d", appConfig.Port)
		if err = app.Listen(port); err != nil { // Assign to existing err
			utils.Logger.Fatal("Fiber server failed to start", zap.Error(err)) // Use Zap Fatal
		}
	}()
	utils.Logger.Info("Fiber server listening", zap.Int("port", appConfig.Port), zap.String("environment", appConfig.Environment))

	// --- 6. Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until OS signal is received (SIGINT or SIGTERM)

	utils.Logger.Info("Shutting down application...")

	// 6.1. Cancel the main application context to signal goroutines to stop
	appCancel()

	// 6.2. Create a context for services graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 6.3. Shut down Fiber app
	if err = app.ShutdownWithContext(shutdownCtx); err != nil { // Assign to existing err
		utils.Logger.Fatal("Fiber server forced to shutdown", zap.Error(err)) // Use Zap Fatal
	}
	utils.Logger.Info("Fiber server gracefully stopped.")

	// 6.4. Disconnect Infrastructure Clients
	mongoClient.Disconnect(shutdownCtx)
	utils.Logger.Info("MongoDB disconnected.")

	redisClientConn.Disconnect()
	utils.Logger.Info("General Redis client disconnected.")

	// Ensure Asynq client is closed
	if asynqConcreteClient != nil {
		if err = asynqConcreteClient.Close(); err != nil { // Assign to existing err
			utils.Logger.Error("Error closing Asynq client", zap.Error(err))
		} else {
			utils.Logger.Info("Asynq client disconnected.")
		}
	}

	// Give in-memory goroutines a moment to respond to context cancellation
	time.Sleep(1 * time.Second)

	utils.Logger.Info("Application fully stopped.")
}
