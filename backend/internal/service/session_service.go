package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"timeline-intelligence/backend/internal/domain"
	"timeline-intelligence/backend/internal/repository"
)

// SessionService manages server-side session lifecycle.
type SessionService struct {
	sessionRepo repository.SessionRepository
	expiry      time.Duration
}

// NewSessionService creates a SessionService.
func NewSessionService(sessionRepo repository.SessionRepository, expiry time.Duration) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		expiry:      expiry,
	}
}

// CreateSession generates a cryptographically random session token, hashes it,
// stores the hash in the database, and returns the raw token (for the cookie).
//
// The raw token is NEVER stored — only its SHA-256 hash.
func (ss *SessionService) CreateSession(
	ctx context.Context,
	userID int64,
	ip, userAgent string,
) (rawToken string, session *domain.Session, err error) {
	// Generate 32 random bytes → 64 hex chars.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", nil, fmt.Errorf("generate session token: %w", err)
	}
	rawToken = hex.EncodeToString(tokenBytes)
	tokenHash := hashToken(rawToken)

	expiresAt := time.Now().UTC().Add(ss.expiry)
	session, err = ss.sessionRepo.Create(ctx, userID, tokenHash, expiresAt, ip, userAgent)
	if err != nil {
		return "", nil, fmt.Errorf("create session record: %w", err)
	}
	return rawToken, session, nil
}

// ValidateSession looks up and validates a session by raw token.
// Returns nil if the session is not found, expired, or revoked.
// Updates last_seen_at on valid sessions.
func (ss *SessionService) ValidateSession(ctx context.Context, rawToken string) (*domain.Session, error) {
	if rawToken == "" {
		return nil, nil
	}
	tokenHash := hashToken(rawToken)
	session, err := ss.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session == nil {
		return nil, nil
	}
	if !session.IsValid(time.Now().UTC()) {
		return nil, nil
	}
	// Update last seen asynchronously — failure is not fatal.
	go func() {
		_ = ss.sessionRepo.UpdateLastSeen(context.Background(), session.ID, time.Now().UTC())
	}()
	return session, nil
}

// RevokeSession revokes a single session by raw token.
func (ss *SessionService) RevokeSession(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	tokenHash := hashToken(rawToken)
	session, err := ss.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	if session == nil {
		return nil // Already gone — not an error.
	}
	return ss.sessionRepo.Revoke(ctx, session.ID, time.Now().UTC())
}

// RevokeAllSessionsForUser revokes every active session for a user (e.g., on deactivation).
func (ss *SessionService) RevokeAllSessionsForUser(ctx context.Context, userID int64) error {
	return ss.sessionRepo.RevokeAllForUser(ctx, userID, time.Now().UTC())
}

// hashToken returns the hex-encoded SHA-256 hash of a session token.
// This is used both when storing and when looking up a session.
func hashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}
