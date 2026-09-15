package tests

import (
	"context"
	"testing"
	"time"

	"timeline-intelligence/backend/internal/repository/sqlite"
	"timeline-intelligence/backend/internal/service"
)

func newTestSessionService(t *testing.T) (*service.SessionService, *sqlite.DB, int64) {
	t.Helper()
	db := newTestDB(t)

	// Sessions reference users(id) — create a stub user for FK integrity.
	res, err := db.DB().ExecContext(context.Background(),
		`INSERT INTO users (name, email, password_hash, status, role)
		 VALUES ('Session Test', 'session-test@example.com', '$argon2id$test', 'active', 'user')`)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("user id: %v", err)
	}

	sessionRepo := sqlite.NewSessionRepo(db.DB())
	sessionSvc := service.NewSessionService(sessionRepo, 24*time.Hour)
	return sessionSvc, db, userID
}

func TestSession_CreateAndValidate(t *testing.T) {
	sessionSvc, _, userID := newTestSessionService(t)
	ctx := context.Background()

	rawToken, session, err := sessionSvc.CreateSession(ctx, userID, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if rawToken == "" {
		t.Error("Expected non-empty raw token")
	}
	if session == nil {
		t.Error("Expected non-nil session")
	}

	validated, err := sessionSvc.ValidateSession(ctx, rawToken)
	if err != nil {
		t.Fatalf("ValidateSession error: %v", err)
	}
	if validated == nil {
		t.Error("Expected valid session, got nil")
	}
}

func TestSession_RevokeInvalidates(t *testing.T) {
	sessionSvc, _, userID := newTestSessionService(t)
	ctx := context.Background()

	rawToken, _, err := sessionSvc.CreateSession(ctx, userID, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if err := sessionSvc.RevokeSession(ctx, rawToken); err != nil {
		t.Fatalf("RevokeSession failed: %v", err)
	}

	validated, err := sessionSvc.ValidateSession(ctx, rawToken)
	if err != nil {
		t.Fatalf("ValidateSession error: %v", err)
	}
	if validated != nil {
		t.Error("Session should be nil after revocation")
	}
}

func TestSession_EmptyTokenReturnsNil(t *testing.T) {
	sessionSvc, _, _ := newTestSessionService(t)
	ctx := context.Background()

	session, err := sessionSvc.ValidateSession(ctx, "")
	if err != nil {
		t.Fatalf("ValidateSession error: %v", err)
	}
	if session != nil {
		t.Error("Expected nil for empty token")
	}
}

func TestSession_RawTokenNotEqualToHash(t *testing.T) {
	sessionSvc, db, userID := newTestSessionService(t)
	ctx := context.Background()

	rawToken, _, err := sessionSvc.CreateSession(ctx, userID, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	var storedHash string
	err = db.DB().QueryRowContext(ctx, `SELECT token_hash FROM sessions WHERE user_id=?`, userID).Scan(&storedHash)
	if err != nil {
		t.Fatalf("DB query error: %v", err)
	}

	if storedHash == rawToken {
		t.Error("Raw session token stored in DB — it should only be stored as a hash!")
	}
}

func TestSession_RevokeAllForUser(t *testing.T) {
	sessionSvc, _, userID := newTestSessionService(t)
	ctx := context.Background()

	token1, _, err := sessionSvc.CreateSession(ctx, userID, "127.0.0.1", "agent1")
	if err != nil {
		t.Fatalf("CreateSession 1 failed: %v", err)
	}
	token2, _, err := sessionSvc.CreateSession(ctx, userID, "127.0.0.2", "agent2")
	if err != nil {
		t.Fatalf("CreateSession 2 failed: %v", err)
	}

	if err := sessionSvc.RevokeAllSessionsForUser(ctx, userID); err != nil {
		t.Fatalf("RevokeAllSessionsForUser failed: %v", err)
	}

	s1, _ := sessionSvc.ValidateSession(ctx, token1)
	s2, _ := sessionSvc.ValidateSession(ctx, token2)

	if s1 != nil || s2 != nil {
		t.Error("All sessions should be revoked after RevokeAllSessionsForUser")
	}
}
