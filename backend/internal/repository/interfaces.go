// Package repository defines the data-access interfaces for the platform.
//
// These interfaces are the boundary between the application service layer and
// the persistence layer. The SQLite implementations in the sqlite sub-package
// satisfy these interfaces. Alternative implementations (e.g., PostgreSQL)
// can be substituted without changing any service or handler code.
package repository

import (
	"context"
	"time"

	"timeline-intelligence/backend/internal/domain"
)

// UserRepository defines all database operations for user accounts.
type UserRepository interface {
	// Create inserts a new user and returns the assigned ID.
	Create(ctx context.Context, name, email, passwordHash string) (*domain.User, error)

	// FindByID retrieves a user by their primary key.
	FindByID(ctx context.Context, id int64) (*domain.User, error)

	// FindByEmail retrieves a user by email address (case-insensitive).
	// Returns nil, nil if no user exists — do not expose this distinction to clients.
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// EmailExists returns true if the email is already registered.
	// Used for registration validation only — do not expose result to login callers.
	EmailExists(ctx context.Context, email string) (bool, error)

	// Deactivate marks a user account as deactivated and records the timestamp.
	Deactivate(ctx context.Context, userID int64, at time.Time) error

	// UpdatePasswordHash replaces the stored password hash.
	// (Reserved for future password-change functionality.)
	UpdatePasswordHash(ctx context.Context, userID int64, newHash string) error
}

// SessionRepository defines database operations for server-side sessions.
type SessionRepository interface {
	// Create stores a new session and returns it.
	Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, ip, userAgent string) (*domain.Session, error)

	// FindByTokenHash retrieves an active session by its token hash.
	// Returns nil, nil if not found or expired.
	FindByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)

	// UpdateLastSeen updates the last_seen_at timestamp for a session.
	UpdateLastSeen(ctx context.Context, sessionID int64, at time.Time) error

	// Revoke marks a session as revoked.
	Revoke(ctx context.Context, sessionID int64, at time.Time) error

	// RevokeAllForUser revokes all active sessions for a user (used on deactivation).
	RevokeAllForUser(ctx context.Context, userID int64, at time.Time) error

	// DeleteExpired removes sessions that expired before the given time.
	// Called periodically for housekeeping.
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

// AuditRepository defines database operations for audit and auth event records.
type AuditRepository interface {
	// CreateAuditLog records a security or lifecycle event.
	CreateAuditLog(ctx context.Context, log *domain.AuditLog) error

	// CreateAuthEvent records a login attempt.
	CreateAuthEvent(ctx context.Context, event *domain.AuthEvent) error

	// CountRecentFailuresByIP counts failed login attempts from an IP within a duration.
	CountRecentFailuresByIP(ctx context.Context, ip string, since time.Time) (int, error)

	// ListAuditLogs retrieves recent audit log entries (admin only).
	// limit controls the maximum number of entries returned.
	ListAuditLogs(ctx context.Context, limit int) ([]*domain.AuditLog, error)
}
