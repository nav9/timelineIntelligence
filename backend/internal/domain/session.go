package domain

import "time"

// Session represents an authenticated server-side session.
//
// The raw session token is never stored in the database. Only a cryptographic
// hash of the token is persisted. The raw token is stored in the client's
// HttpOnly cookie and is hashed for each lookup.
type Session struct {
	ID          int64
	UserID      int64
	TokenHash   string    // SHA-256 hash of the raw session token
	CreatedAt   time.Time
	ExpiresAt   time.Time
	LastSeenAt  time.Time
	RevokedAt   *time.Time // nil unless explicitly revoked
	IPAddress   string
	UserAgent   string
}

// IsValid returns true if the session has not expired and has not been revoked.
func (s *Session) IsValid(now time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	return now.Before(s.ExpiresAt)
}

// IsRevoked returns true if the session has been explicitly revoked.
func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}
