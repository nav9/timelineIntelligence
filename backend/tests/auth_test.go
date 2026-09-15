// Package tests contains integration tests for the backend.
//
// Tests use an in-memory SQLite database (":memory:") so they are
// self-contained and do not leave test data on disk.
//
// Test naming convention: TestX_Y where X is the feature and Y describes
// the scenario (e.g., TestRegister_DuplicateEmail).
package tests

import (
	"context"
	"testing"
	"time"

	"timeline-intelligence/backend/internal/domain"
	"timeline-intelligence/backend/internal/repository/sqlite"
	"timeline-intelligence/backend/internal/service"
)

// newTestDB opens an in-memory SQLite database for testing.
func newTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("open test DB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// newTestServices creates a complete set of services wired to a test DB.
func newTestServices(t *testing.T) (
	*service.AuthService,
	*service.SessionService,
	*service.PasswordService,
) {
	t.Helper()
	db := newTestDB(t)

	userRepo := sqlite.NewUserRepo(db.DB())
	sessionRepo := sqlite.NewSessionRepo(db.DB())
	auditRepo := sqlite.NewAuditRepo(db.DB())

	passwordSvc := service.NewPasswordService(service.PasswordParams{
		Time:    1,     // Reduced for test speed.
		Memory:  8192,
		Threads: 1,
		KeyLen:  32,
	})
	sessionSvc := service.NewSessionService(sessionRepo, 24*time.Hour)
	rateLimiter := service.NewRateLimitService(auditRepo, 10, 0) // No delay in tests.
	authSvc := service.NewAuthService(userRepo, sessionSvc, passwordSvc, auditRepo, rateLimiter)

	return authSvc, sessionSvc, passwordSvc
}

// --- Registration tests ---

func TestRegister_ValidInput(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	user, err := authSvc.Register(ctx, service.RegisterInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "correct-horse-battery-staple",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("Expected non-zero user ID")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
	if user.Status != domain.UserStatusActive {
		t.Errorf("Expected status active, got %s", user.Status)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	_, err := authSvc.Register(ctx, service.RegisterInput{
		Name:     "Test User",
		Email:    "not-an-email",
		Password: "correct-horse-battery-staple",
	})
	if err == nil {
		t.Error("Expected error for invalid email, got nil")
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	_, err := authSvc.Register(ctx, service.RegisterInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "short",
	})
	if err == nil {
		t.Error("Expected error for weak password, got nil")
	}
}

func TestRegister_CommonPassword(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	_, err := authSvc.Register(ctx, service.RegisterInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Error("Expected error for common password, got nil")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	input := service.RegisterInput{
		Name:     "Test User",
		Email:    "duplicate@example.com",
		Password: "correct-horse-battery-staple",
	}
	if _, err := authSvc.Register(ctx, input); err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	_, err := authSvc.Register(ctx, input)
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

func TestRegister_PasswordNotStoredInPlaintext(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	password := "correct-horse-battery-staple"
	user, err := authSvc.Register(ctx, service.RegisterInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user.PasswordHash == password {
		t.Error("Password stored in plaintext!")
	}
	if len(user.PasswordHash) < 32 {
		t.Errorf("Password hash suspiciously short: %d chars", len(user.PasswordHash))
	}
}

func TestRegister_PasswordHashedWithArgon2id(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	user, err := authSvc.Register(ctx, service.RegisterInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "correct-horse-battery-staple",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify the hash has the Argon2id prefix.
	if len(user.PasswordHash) < 10 || user.PasswordHash[:10] != "$argon2id$" {
		t.Errorf("Expected Argon2id hash prefix, got: %s", user.PasswordHash[:min(20, len(user.PasswordHash))])
	}
}

// --- Login tests ---

func TestLogin_CorrectCredentials(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	registerAndVerify(t, authSvc, "user@example.com", "correct-horse-battery-staple")

	result, err := authSvc.Login(ctx, service.LoginInput{
		Email:     "user@example.com",
		Password:  "correct-horse-battery-staple",
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if result.RawToken == "" {
		t.Error("Expected session token, got empty string")
	}
	if result.User.Email != "user@example.com" {
		t.Errorf("Expected email user@example.com, got %s", result.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	registerAndVerify(t, authSvc, "user@example.com", "correct-horse-battery-staple")

	_, err := authSvc.Login(ctx, service.LoginInput{
		Email:     "user@example.com",
		Password:  "wrongpassword!",
		IPAddress: "127.0.0.2",
		UserAgent: "test",
	})
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}
}

func TestLogin_NonExistentAccount(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	_, err := authSvc.Login(ctx, service.LoginInput{
		Email:     "nobody@example.com",
		Password:  "correct-horse-battery-staple",
		IPAddress: "127.0.0.3",
		UserAgent: "test",
	})
	if err == nil {
		t.Error("Expected error for non-existent account, got nil")
	}
}

func TestLogin_DeactivatedAccount(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	user := registerAndVerify(t, authSvc, "user@example.com", "correct-horse-battery-staple")

	// Deactivate the account.
	if err := authSvc.DeactivateAccount(ctx, user.ID, "", "127.0.0.1"); err != nil {
		t.Fatalf("Deactivate failed: %v", err)
	}

	// Attempt login — should fail with generic error (not revealing deactivation status).
	_, err := authSvc.Login(ctx, service.LoginInput{
		Email:     "user@example.com",
		Password:  "correct-horse-battery-staple",
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	})
	if err == nil {
		t.Error("Expected error for deactivated account login, got nil")
	}
}

func TestLogin_SessionCreated(t *testing.T) {
	authSvc, sessionSvc, _ := newTestServices(t)
	ctx := context.Background()

	registerAndVerify(t, authSvc, "user@example.com", "correct-horse-battery-staple")

	result, err := authSvc.Login(ctx, service.LoginInput{
		Email:     "user@example.com",
		Password:  "correct-horse-battery-staple",
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Validate the session token is real.
	session, err := sessionSvc.ValidateSession(ctx, result.RawToken)
	if err != nil {
		t.Fatalf("ValidateSession error: %v", err)
	}
	if session == nil {
		t.Error("Expected valid session, got nil")
	}
}

// --- Logout tests ---

func TestLogout_SessionBecomesInvalid(t *testing.T) {
	authSvc, sessionSvc, _ := newTestServices(t)
	ctx := context.Background()

	user := registerAndVerify(t, authSvc, "user@example.com", "correct-horse-battery-staple")
	result := loginAndVerify(t, authSvc, "user@example.com", "correct-horse-battery-staple")

	// Logout.
	if err := authSvc.Logout(ctx, result.RawToken, user.ID, "127.0.0.1"); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Session should now be invalid.
	session, err := sessionSvc.ValidateSession(ctx, result.RawToken)
	if err != nil {
		t.Fatalf("ValidateSession error: %v", err)
	}
	if session != nil {
		t.Error("Session should be nil after logout, but it is still valid")
	}
}

// --- Account deactivation tests ---

func TestDeactivate_AccountBecomesInactive(t *testing.T) {
	authSvc, _, _ := newTestServices(t)
	ctx := context.Background()

	user := registerAndVerify(t, authSvc, "user@example.com", "correct-horse-battery-staple")

	if err := authSvc.DeactivateAccount(ctx, user.ID, "", "127.0.0.1"); err != nil {
		t.Fatalf("DeactivateAccount failed: %v", err)
	}

	// Verify deactivated user cannot log in.
	_, err := authSvc.Login(ctx, service.LoginInput{
		Email:     "user@example.com",
		Password:  "correct-horse-battery-staple",
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	})
	if err == nil {
		t.Error("Expected login to fail after deactivation")
	}
}

func TestDeactivate_SessionsRevoked(t *testing.T) {
	authSvc, sessionSvc, _ := newTestServices(t)
	ctx := context.Background()

	user := registerAndVerify(t, authSvc, "user@example.com", "correct-horse-battery-staple")
	result := loginAndVerify(t, authSvc, "user@example.com", "correct-horse-battery-staple")

	// Deactivate — should revoke all sessions.
	if err := authSvc.DeactivateAccount(ctx, user.ID, result.RawToken, "127.0.0.1"); err != nil {
		t.Fatalf("DeactivateAccount failed: %v", err)
	}

	// The session must now be invalid.
	session, err := sessionSvc.ValidateSession(ctx, result.RawToken)
	if err != nil {
		t.Fatalf("ValidateSession error: %v", err)
	}
	if session != nil {
		t.Error("Session should be invalid after account deactivation")
	}
}

// --- helpers ---

func registerAndVerify(t *testing.T, authSvc *service.AuthService, email, password string) *domain.User {
	t.Helper()
	user, err := authSvc.Register(context.Background(), service.RegisterInput{
		Name:     "Test User",
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	return user
}

func loginAndVerify(t *testing.T, authSvc *service.AuthService, email, password string) *service.LoginResult {
	t.Helper()
	result, err := authSvc.Login(context.Background(), service.LoginInput{
		Email:     email,
		Password:  password,
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
