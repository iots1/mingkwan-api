package utils

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger() {
	var config zap.Config
	env := os.Getenv("APP_ENV") // Assuming you have an APP_ENV environment variable
	if env == "" {
		env = "development" // Default to development if not set
	}

	switch env {
	case "production":
		config = zap.NewProductionConfig()
		// Production logs typically go to files or a centralized logging system
		// and are often in JSON format for easier parsing.
		// You might want to configure output paths here.
		// Example: config.OutputPaths = []string{"stdout", "/var/log/your-app/app.log"}
		// Ensure logs are JSON formatted for production
		config.Encoding = "json"
	case "development":
		config = zap.NewDevelopmentConfig()
		// Development logs usually go to stdout/stderr for human readability.
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add colors to level
		config.EncoderConfig.EncodeTime = customTimeEncoder                 // Custom time format
		config.EncoderConfig.TimeKey = "timestamp"                          // Key for time field
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder      // Show short file path
		config.Encoding = "console"                                         // Human-readable console output
	default:
		// Fallback for unknown environments, maybe a mix of production and development
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = customTimeEncoder
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		config.Encoding = "console"
	}

	// Set initial log level from environment, defaulting to Info
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "info" // Default to Info
	}
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevelStr)); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Invalid LOG_LEVEL '%s', defaulting to INFO\n", logLevelStr)
		level = zapcore.InfoLevel
	}
	config.Level.SetLevel(level)

	var err error
	Logger, err = config.Build(zap.AddCallerSkip(1)) // Skip 1 caller frame to get the actual call site
	if err != nil {
		// If logger creation fails, we can't really log, so panic.
		panic(fmt.Sprintf("Failed to initialize Zap logger: %v", err))
	}
	Logger.Info("Zap logger initialized successfully.", zap.String("environment", env), zap.String("log_level", level.String()))
}

// customTimeEncoder formats time as YYYY-MM-DD HH:MM:SS (UTC).
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format("2006-01-02 15:04:05 UTC"))
}

// Ensure the logger is initialized when the package is loaded.
func init() {
	// You can choose to call InitLogger here or explicitly in main.go
	// Calling here ensures Logger is always ready, but main.go gives more control.
	InitLogger()
}

// SyncLogger flushes any buffered logs. Should be called before application exits.
func SyncLogger() {
	if Logger != nil {
		err := Logger.Sync()                                                 // Flushes buffer, if any
		if err != nil && err.Error() != "sync /dev/null: invalid argument" { // Ignore common harmless error on some systems
			fmt.Fprintf(os.Stderr, "Error syncing Zap logger: %v\n", err)
		}
	}
}

// WithContext adds a context field to the logger.
func WithContext(ctx string) *zap.Logger {
	return Logger.With(zap.String("context", ctx))
}

// You can add more helper functions here if needed, e.g., for specific contexts
// func HttpRequestLogger(req *http.Request) *zap.Logger {
//     return Logger.With(
//         zap.String("method", req.Method),
//         zap.String("path", req.URL.Path),
//         zap.String("remote_ip", req.RemoteAddr),
//     )
// }
