// Package config loads and provides application configuration from environment
// variables and an optional .env file.
//
// Configuration is loaded once at startup. The Config struct is immutable
// after initialization and is passed via dependency injection.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// BuildMode represents the application build mode.
type BuildMode string

const (
	BuildModeDebug   BuildMode = "debug"
	BuildModeRelease BuildMode = "release"
)

// Config holds all application configuration.
type Config struct {
	// Server
	ServerPort string
	ServerHost string

	// Database
	DBPath string

	// Session
	SessionSecret      string
	SessionExpiryHours int

	// Argon2id parameters
	Argon2Time    uint32
	Argon2Memory  uint32
	Argon2Threads uint8
	Argon2KeyLen  uint32

	// Rate limiting
	RateLimitMaxAttempts    int
	RateLimitBaseDelaySeconds int

	// CSRF
	CSRFTokenExpiryMinutes int

	// Build
	BuildMode BuildMode

	// Logging
	LogLevel  string
	LogFormat string
}

// Load reads configuration from the environment, optionally loading a .env file first.
// The .env file is optional — missing file is not an error.
func Load(envFile string) (*Config, error) {
	// Attempt to load .env file; silently continue if not present.
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			log.Debug().Str("file", envFile).Msg("No .env file found, using environment variables only")
		}
	}

	cfg := &Config{}

	// Server
	cfg.ServerPort = getEnvString("SERVER_PORT", "8080")
	cfg.ServerHost = getEnvString("SERVER_HOST", "127.0.0.1")

	// Database
	cfg.DBPath = getEnvString("DB_PATH", "./data/timeline.db")

	// Session
	secret := getEnvString("SESSION_SECRET", "")
	if secret == "" || secret == "CHANGE_THIS_TO_A_STRONG_RANDOM_SECRET_AT_LEAST_32_CHARS" {
		return nil, fmt.Errorf("SESSION_SECRET must be set to a strong random value; see .env.example")
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("SESSION_SECRET must be at least 32 characters long")
	}
	cfg.SessionSecret = secret
	cfg.SessionExpiryHours = getEnvInt("SESSION_EXPIRY_HOURS", 24)

	// Argon2id
	cfg.Argon2Time = uint32(getEnvInt("ARGON2_TIME", 2))
	cfg.Argon2Memory = uint32(getEnvInt("ARGON2_MEMORY", 65536))
	cfg.Argon2Threads = uint8(getEnvInt("ARGON2_THREADS", 4))
	cfg.Argon2KeyLen = uint32(getEnvInt("ARGON2_KEY_LEN", 32))

	// Rate limiting
	cfg.RateLimitMaxAttempts = getEnvInt("RATE_LIMIT_MAX_ATTEMPTS", 5)
	cfg.RateLimitBaseDelaySeconds = getEnvInt("RATE_LIMIT_BASE_DELAY_SECONDS", 1)

	// CSRF
	cfg.CSRFTokenExpiryMinutes = getEnvInt("CSRF_TOKEN_EXPIRY_MINUTES", 60)

	// Build mode
	modeStr := strings.ToLower(getEnvString("BUILD_MODE", "debug"))
	switch modeStr {
	case "release":
		cfg.BuildMode = BuildModeRelease
	default:
		cfg.BuildMode = BuildModeDebug
	}

	// Logging
	cfg.LogLevel = getEnvString("LOG_LEVEL", "info")
	cfg.LogFormat = getEnvString("LOG_FORMAT", "console")

	return cfg, nil
}

// SessionExpiry returns the session expiry duration.
func (c *Config) SessionExpiry() time.Duration {
	return time.Duration(c.SessionExpiryHours) * time.Hour
}

// ServerAddress returns the formatted server listen address.
func (c *Config) ServerAddress() string {
	return c.ServerHost + ":" + c.ServerPort
}

// IsDebug returns true if the application is running in debug mode.
func (c *Config) IsDebug() bool {
	return c.BuildMode == BuildModeDebug
}

// --- helpers ---

func getEnvString(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
