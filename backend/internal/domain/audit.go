package domain

import "time"

// AuditEventType classifies security and account lifecycle events.
type AuditEventType string

const (
	AuditEventRegistration       AuditEventType = "registration"
	AuditEventLoginSuccess       AuditEventType = "login_success"
	AuditEventLoginFailure       AuditEventType = "login_failure"
	AuditEventLogout             AuditEventType = "logout"
	AuditEventAccountDeactivated AuditEventType = "account_deactivated"
	AuditEventSessionExpired     AuditEventType = "session_expired"
	AuditEventSessionRevoked     AuditEventType = "session_revoked"
	AuditEventRateLimited        AuditEventType = "rate_limited"
)

// AuditLog records a security or account lifecycle event.
//
// Never log passwords, password hashes, raw session tokens, or other secrets.
// The user_id may be zero/null for events where the user is not yet authenticated.
type AuditLog struct {
	ID        int64
	EventType AuditEventType
	UserID    *int64  // nil for unauthenticated events (e.g. failed login with unknown email)
	Email     string  // email involved, if known and not considered sensitive in context
	IPAddress string
	Success   bool
	Reason    string  // optional human-readable reason/category
	CreatedAt time.Time
}

// AuthEvent records individual login attempts (both successful and failed).
// Stored separately from AuditLog to allow efficient rate-limit queries by IP.
type AuthEvent struct {
	ID        int64
	UserID    *int64
	Email     string
	IPAddress string
	Success   bool
	Reason    string
	CreatedAt time.Time
}
