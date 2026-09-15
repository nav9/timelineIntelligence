// Command server is the entry point for the timeline-intelligence backend.
//
// Startup sequence:
//  1. Load configuration from .env and environment variables.
//  2. Configure structured logging.
//  3. Open and migrate the SQLite database.
//  4. Wire all services and repositories via constructor injection.
//  5. Set up the Gin HTTP router.
//  6. Start the server.
//
// All dependencies are wired explicitly here — there is no global state or
// hidden singletons. This makes the dependency graph visible and testable.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"timeline-intelligence/backend/internal/config"
	"timeline-intelligence/backend/internal/handler"
	"timeline-intelligence/backend/internal/repository/sqlite"
	"timeline-intelligence/backend/internal/service"
)

func main() {
	// --- Configuration ---
	cfg, err := config.Load(".env")
	if err != nil {
		// Use basic output before logger is configured.
		log.Fatal().Err(err).Msg("Configuration error")
	}

	// --- Logging ---
	setupLogger(cfg)
	log.Info().
		Str("mode", string(cfg.BuildMode)).
		Str("address", cfg.ServerAddress()).
		Msg("Starting timeline-intelligence backend")

	// --- Database ---
	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", cfg.DBPath).Msg("Failed to open database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing database")
		}
	}()

	// --- Repositories ---
	userRepo := sqlite.NewUserRepo(db.DB())
	sessionRepo := sqlite.NewSessionRepo(db.DB())
	auditRepo := sqlite.NewAuditRepo(db.DB())

	// --- Services ---
	passwordParams := service.PasswordParams{
		Time:    cfg.Argon2Time,
		Memory:  cfg.Argon2Memory,
		Threads: cfg.Argon2Threads,
		KeyLen:  cfg.Argon2KeyLen,
	}
	passwordSvc := service.NewPasswordService(passwordParams)
	sessionSvc := service.NewSessionService(sessionRepo, cfg.SessionExpiry())
	rateLimiter := service.NewRateLimitService(auditRepo, cfg.RateLimitMaxAttempts, cfg.RateLimitBaseDelaySeconds)
	authSvc := service.NewAuthService(userRepo, sessionSvc, passwordSvc, auditRepo, rateLimiter)

	// --- Gin ---
	if cfg.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.Logger())

	// --- Routes ---
	frontendPath := os.Getenv("FRONTEND_DIST_PATH") // Set for production serving.
	handler.SetupRouter(engine, db, authSvc, sessionSvc, auditRepo, cfg.SessionExpiryHours, frontendPath)

	// --- HTTP server ---
	srv := &http.Server{
		Addr:         cfg.ServerAddress(),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background.
	go func() {
		log.Info().Str("address", cfg.ServerAddress()).Msg("Server listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Server error")
		}
	}()

	// --- Graceful shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	}
	log.Info().Msg("Server stopped")
}

// setupLogger configures zerolog based on the build configuration.
func setupLogger(cfg *config.Config) {
	if cfg.LogFormat == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
}
