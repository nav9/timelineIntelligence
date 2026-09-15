package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"timeline-intelligence/backend/internal/repository/sqlite"
)

// HealthHandler handles the health check endpoint.
type HealthHandler struct {
	db *sqlite.DB
}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler(db *sqlite.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health handles GET /api/health
// Returns a simple health status object.
// Never exposes sensitive internal details.
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	dbOK := true
	if err := h.db.Ping(ctx); err != nil {
		dbOK = false
	}

	status := "ok"
	httpStatus := http.StatusOK
	if !dbOK {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"components": gin.H{
			"database": boolStatus(dbOK),
			"api":      "ok",
		},
	})
}

func boolStatus(ok bool) string {
	if ok {
		return "ok"
	}
	return "error"
}
