package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig holds general application-wide settings.
type AppConfig struct {
	Port        int
	Environment string // e.g., "development", "production", "testing"
	// Add other app-specific settings here
}

// MongoConfig holds MongoDB connection settings.
type MongoConfig struct {
	URI    string
	DBName string
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr     string // Host:Port combination
	Password string
	DB       int // Redis DB number
}

// LoadAppConfig loads application configuration from environment variables.
func LoadAppConfig() AppConfig {
	portStr := os.Getenv("APP_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil || port == 0 {
		port = 8080 // Default port
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development" // Default environment
	}

	return AppConfig{
		Port:        port,
		Environment: env,
	}
}

// LoadMongoConfig loads MongoDB connection configuration from environment variables.
func LoadMongoConfig() MongoConfig {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		fmt.Println("WARNING: MONGO_URI not set. Using default: mongodb://localhost:27017")
		uri = "mongodb://localhost:27017" // Default for development
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		fmt.Println("WARNING: MONGO_DB_NAME not set. Using default: mingkwan_db")
		dbName = "mingkwan_db" // Default for development
	}

	return MongoConfig{
		URI:    uri,
		DBName: dbName,
	}
}

// LoadRedisConfig loads Redis connection configuration from environment variables.
func LoadRedisConfig() RedisConfig {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost" // Default host
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379" // Default port
	}

	password := os.Getenv("REDIS_PASSWORD") // Can be empty if no password

	dbStr := os.Getenv("REDIS_DB")
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0 // Default DB number
	}

	return RedisConfig{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	}
}

// InitConfig loads all configurations.
func InitConfig() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v. Proceeding without .env file.\n", err)
	}
	fmt.Println("Configuration loaded from environment variables.")
}
