// cmd/app/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redis/go-redis/v9"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"go.uber.org/zap"

	"github.com/iots1/mingkwan-api/config"
	"github.com/iots1/mingkwan-api/internal/modules"
	"github.com/iots1/mingkwan-api/internal/shared/adapters"
	"github.com/iots1/mingkwan-api/internal/shared/cache"
	"github.com/iots1/mingkwan-api/internal/shared/event"
	"github.com/iots1/mingkwan-api/internal/shared/infrastructure"
	"github.com/iots1/mingkwan-api/internal/shared/utils"

	_ "github.com/iots1/mingkwan-api/docs"
)

// @title           Mingkwan API
// @version         1.0
// @description     This is a sample server for Mingkwan API.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	// --- 0. Setup Global Application Context ---
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

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

	passwordHasher := adapters.NewPasswordHasher()
	utils.Logger.Debug("Password hasher initialized.")

	// --- Initialize Infrastructure Connections ---
	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()

	// 2.1. MongoDB
	mongoClient := infrastructure.NewMongoClient(mongoConfig.URI, mongoConfig.DBName)
	var err error
	if _, err = mongoClient.Connect(initCtx); err != nil {
		utils.Logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	db := mongoClient.GetDatabase()
	utils.Logger.Info("Connected to MongoDB", zap.String("database", mongoConfig.DBName))

	redisClientConn := infrastructure.NewRedisClient(redisConfig.Addr, redisConfig.Password, redisConfig.DB)
	var rdb *redis.Client                                        // Declare rdb here
	if rdb, err = redisClientConn.Connect(initCtx); err != nil { // Assign to existing rdb and err
		utils.Logger.Fatal("Failed to connect to Redis", zap.Error(err))
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

	// Initialize In-Memory Event Bus (InMemPubSub)
	inMemPubSub := event.NewInMemoryBus()
	utils.Logger.Info("Initialized In-Memory Event Bus (InMemPubSub).")

	// Initialize Asynq Client
	asynqRedisOpt := event.GetRedisClientOpt(redisConfig.Addr, redisConfig.Password, redisConfig.DB)
	asynqConcreteClient := event.NewAsynqClient(asynqRedisOpt)
	utils.Logger.Info("Initialized Asynq client", zap.String("address", redisConfig.Addr), zap.Int("db", redisConfig.DB))

	// Initialize Low and High Importance Publishers
	lowPublisher := event.NewLowImportancePublisher(inMemPubSub)
	highPublisher := event.NewHighImportancePublisher(asynqConcreteClient)
	utils.Logger.Info("Initialized Low and High Importance Publishers.")

	appDeps := infrastructure.NewAppDependencies(
		appCtx,
		db,
		rdb,
		lowPublisher,
		highPublisher,
		inMemPubSub,
		appConfig,
		passwordHasher,
	)

	app := fiber.New()

	// Enable CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// @Summary Root
	// @Description API Version
	// @Accept json
	// @Router / [get]
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"version": "Mingkwan API v1.0"})
	})

	// API Routes Group
	apiV1 := app.Group("/api/v1")
	userUsecase := modules.SetupUserModule(apiV1, appDeps)
	if userUsecase == nil {
		utils.Logger.Fatal("Failed to setup User Module: userUcase is nil")
	}

	modules.SetupAuthModule(apiV1, appDeps, *userUsecase)

	// Health check endpoint
	// @Summary Health check
	// @Description Checks if the API is up and running.
	// @Tags Health
	// @Accept json
	// @Produce json
	// @Success 200 {string} string "OK"
	// @Router api/v1/health [get]
	apiV1.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Auth Routes
	// auth := apiV1.Group("/auth")
	// auth.Post("/register", authHandler.Register)
	// auth.Post("/login", authHandler.Login)
	// auth.Post("/refresh", authHandler.Refresh)
	// auth.Get("/profile", authHandler.GetProfile)

	// ... other routes and middleware

	// --- 5. Start Server in a Goroutine ---
	go func() {
		port := fmt.Sprintf(":%d", appConfig.Port)
		if err = app.Listen(port); err != nil { // Assign to existing err
			utils.Logger.Fatal("Fiber server failed to start", zap.Error(err))
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
	if err = app.ShutdownWithContext(shutdownCtx); err != nil {
		utils.Logger.Fatal("Fiber server forced to shutdown", zap.Error(err))
	}
	utils.Logger.Info("Fiber server gracefully stopped.")

	// 6.4. Disconnect Infrastructure Clients
	mongoClient.Disconnect(shutdownCtx)
	utils.Logger.Info("MongoDB disconnected.")

	redisClientConn.Disconnect()
	utils.Logger.Info("General Redis client disconnected.")

	// Ensure Asynq client is closed
	if asynqConcreteClient != nil {
		if err = asynqConcreteClient.Close(); err != nil {
			utils.Logger.Error("Error closing Asynq client", zap.Error(err))
		} else {
			utils.Logger.Info("Asynq client disconnected.")
		}
	}

	// Give in-memory goroutines a moment to respond to context cancellation
	time.Sleep(1 * time.Second)

	utils.Logger.Info("Application fully stopped.")
}
