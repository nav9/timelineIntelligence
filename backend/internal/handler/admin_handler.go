package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"timeline-intelligence/backend/internal/repository"
)

// AdminHandler handles administrative API endpoints.
//
// IMPORTANT: Admin functionality is in an early state.
// The authorization boundary (RequireAdmin middleware) is enforced,
// but a full admin UI has not been implemented yet.
//
// TODO (future): Admin user management, full audit log pagination,
// data retention controls.
type AdminHandler struct {
	auditRepo repository.AuditRepository
}

// NewAdminHandler creates an AdminHandler.
func NewAdminHandler(auditRepo repository.AuditRepository) *AdminHandler {
	return &AdminHandler{auditRepo: auditRepo}
}

// ListAuditLogs handles GET /api/admin/logs
// Only accessible to users with the admin role (enforced by RequireAdmin middleware).
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	logs, err := h.auditRepo.ListAuditLogs(c.Request.Context(), 200)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list audit logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve logs"})
		return
	}

	// Convert to API-safe response (all fields are already safe — no passwords or tokens).
	type logEntry struct {
		ID        int64  `json:"id"`
		EventType string `json:"event_type"`
		UserID    *int64 `json:"user_id,omitempty"`
		Email     string `json:"email,omitempty"`
		IPAddress string `json:"ip_address,omitempty"`
		Success   bool   `json:"success"`
		Reason    string `json:"reason,omitempty"`
		CreatedAt string `json:"created_at"`
	}

	entries := make([]logEntry, 0, len(logs))
	for _, l := range logs {
		entries = append(entries, logEntry{
			ID:        l.ID,
			EventType: string(l.EventType),
			UserID:    l.UserID,
			Email:     l.Email,
			IPAddress: l.IPAddress,
			Success:   l.Success,
			Reason:    l.Reason,
			CreatedAt: l.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  entries,
		"count": len(entries),
		"note":  "Showing up to 200 most recent audit log entries.",
	})
}
