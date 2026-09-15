package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"timeline-intelligence/backend/internal/service"
	"timeline-intelligence/backend/internal/repository"
	"timeline-intelligence/backend/internal/repository/sqlite"
)

// SetupRouter configures all routes and middleware for the Gin engine.
// This is the single wiring point for the entire HTTP API surface.
func SetupRouter(
	engine *gin.Engine,
	db *sqlite.DB,
	authSvc *service.AuthService,
	sessionSvc *service.SessionService,
	auditRepo repository.AuditRepository,
	sessionHours int,
	frontendPath string,
) {
	// --- Global middleware ---
	engine.Use(SecurityHeaders())
	engine.Use(gin.Recovery())

	// Create handlers.
	authHandler := NewAuthHandler(authSvc, sessionSvc, sessionHours)
	healthHandler := NewHealthHandler(db)
	adminHandler := NewAdminHandler(auditRepo)

	// --- API routes ---
	api := engine.Group("/api")
	api.Use(CSRFProtect())

	// Health (public — no auth)
	api.GET("/health", healthHandler.Health)

	// Authentication (public)
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	// Authenticated auth routes
	authProtected := api.Group("/auth")
	authProtected.Use(RequireAuth(sessionSvc, authSvc))
	authProtected.POST("/logout", authHandler.Logout)
	authProtected.GET("/me", authHandler.Me)
	authProtected.DELETE("/account", authHandler.DeactivateAccount)

	// Admin routes (requires auth + admin role)
	admin := api.Group("/admin")
	admin.Use(RequireAuth(sessionSvc, authSvc))
	admin.Use(RequireAdmin())
	admin.GET("/logs", adminHandler.ListAuditLogs)

	// --- Frontend SPA ---
	// Serve the built Svelte SPA from the given path.
	// In development, Vite runs its own dev server; this is used in production/release mode.
	if frontendPath != "" {
		engine.Static("/assets", frontendPath+"/assets")
		engine.StaticFile("/favicon.ico", frontendPath+"/favicon.ico")

		// SPA catch-all: serve index.html for all non-API routes.
		engine.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if !strings.HasPrefix(path, "/api/") {
				c.File(frontendPath + "/index.html")
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			}
		})
	}
}
