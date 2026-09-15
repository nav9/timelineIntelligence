package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"timeline-intelligence/backend/internal/domain"
	"timeline-intelligence/backend/internal/repository"
)

// Sentinel errors returned by AuthService.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDeactivated = errors.New("account deactivated")
	ErrEmailTaken         = errors.New("email address is already registered")
	ErrWeakPassword       = errors.New("password does not meet requirements")
	ErrRateLimited        = errors.New("too many failed attempts, please wait")
)

// AuthService coordinates user registration, login, logout, and account deactivation.
type AuthService struct {
	userRepo    repository.UserRepository
	sessionSvc  *SessionService
	passwordSvc *PasswordService
	auditRepo   repository.AuditRepository
	rateLimiter *RateLimitService
}

// NewAuthService creates an AuthService with all required dependencies injected.
func NewAuthService(
	userRepo repository.UserRepository,
	sessionSvc *SessionService,
	passwordSvc *PasswordService,
	auditRepo repository.AuditRepository,
	rateLimiter *RateLimitService,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionSvc:  sessionSvc,
		passwordSvc: passwordSvc,
		auditRepo:   auditRepo,
		rateLimiter: rateLimiter,
	}
}

// RegisterInput contains validated registration input.
type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

// LoginInput contains validated login credentials.
type LoginInput struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

// LoginResult is returned on successful authentication.
type LoginResult struct {
	User      *domain.User
	RawToken  string // Place this in the HttpOnly session cookie.
}

// Register creates a new user account.
// Returns ErrEmailTaken if the email is already registered.
// Returns ErrWeakPassword with feedback if the password fails validation.
func (a *AuthService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	// Normalize email.
	email := strings.TrimSpace(strings.ToLower(input.Email))
	name := strings.TrimSpace(input.Name)

	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !isValidEmail(email) {
		return nil, fmt.Errorf("invalid email address")
	}

	// Password validation (backend is authoritative).
	result := a.passwordSvc.ValidatePassword(input.Password, name, email)
	if !result.Valid {
		return nil, fmt.Errorf("%w: %s", ErrWeakPassword, strings.Join(result.Feedback, "; "))
	}

	// Email uniqueness check. Do not reveal this result to login callers.
	exists, err := a.userRepo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return nil, ErrEmailTaken
	}

	// Hash password.
	hash, err := a.passwordSvc.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user.
	user, err := a.userRepo.Create(ctx, name, email, hash)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Audit log.
	uid := user.ID
	a.logAudit(ctx, &domain.AuditLog{
		EventType: domain.AuditEventRegistration,
		UserID:    &uid,
		Email:     email,
		Success:   true,
	})

	log.Info().Int64("user_id", user.ID).Str("email", email).Msg("User registered")
	return user, nil
}

// Login authenticates a user and creates a server-side session.
// Uses generic error messages to prevent account enumeration.
// Applies progressive delay on failure — this delay is enforced server-side.
func (a *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	ip := input.IPAddress
	email := strings.TrimSpace(strings.ToLower(input.Email))

	// Rate limiting check.
	if a.rateLimiter.IsRateLimited(ctx, ip) {
		a.logAuthEvent(ctx, nil, email, ip, false, "rate_limited")
		return nil, ErrRateLimited
	}

	// Apply progressive delay BEFORE computing the result (timing-safe).
	delay := a.rateLimiter.DelayForIP(ctx, ip)
	if delay > 0 {
		log.Debug().Dur("delay", delay).Str("ip", ip).Msg("Applying login delay")
		time.Sleep(delay)
	}

	// Look up user. Use a constant-time path regardless of outcome.
	user, err := a.userRepo.FindByEmail(ctx, email)
	if err != nil {
		log.Error().Err(err).Msg("Login: database error")
		a.recordFailure(ctx, nil, email, ip, "db_error")
		return nil, ErrInvalidCredentials
	}

	// If user not found, still perform a dummy verification to prevent timing attacks.
	if user == nil {
		_ = dummyVerify()
		a.recordFailure(ctx, nil, email, ip, "no_account")
		return nil, ErrInvalidCredentials
	}

	// Verify password. This is the same code path for wrong password and no account.
	ok, err := a.passwordSvc.VerifyPassword(input.Password, user.PasswordHash)
	if err != nil {
		log.Error().Err(err).Msg("Login: password verification error")
		a.recordFailure(ctx, &user.ID, email, ip, "verify_error")
		return nil, ErrInvalidCredentials
	}
	if !ok {
		a.recordFailure(ctx, &user.ID, email, ip, "wrong_password")
		return nil, ErrInvalidCredentials
	}

	// Check account status AFTER password verification to avoid revealing whether the account exists.
	if !user.IsActive() {
		a.recordFailure(ctx, &user.ID, email, ip, "account_deactivated")
		return nil, ErrInvalidCredentials // Use generic error — don't reveal deactivation
	}

	// Successful authentication.
	a.rateLimiter.RecordSuccess(ip)

	rawToken, session, err := a.sessionSvc.CreateSession(ctx, user.ID, ip, input.UserAgent)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	_ = session

	// Audit log.
	a.logAudit(ctx, &domain.AuditLog{
		EventType: domain.AuditEventLoginSuccess,
		UserID:    &user.ID,
		Email:     email,
		IPAddress: ip,
		Success:   true,
	})
	a.logAuthEvent(ctx, &user.ID, email, ip, true, "")

	log.Info().Int64("user_id", user.ID).Str("ip", ip).Msg("User logged in")
	return &LoginResult{User: user, RawToken: rawToken}, nil
}

// Logout revokes the current session.
func (a *AuthService) Logout(ctx context.Context, rawToken string, userID int64, ip string) error {
	if err := a.sessionSvc.RevokeSession(ctx, rawToken); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	a.logAudit(ctx, &domain.AuditLog{
		EventType: domain.AuditEventLogout,
		UserID:    &userID,
		IPAddress: ip,
		Success:   true,
	})

	log.Info().Int64("user_id", userID).Str("ip", ip).Msg("User logged out")
	return nil
}

// DeactivateAccount marks the account as deactivated and revokes all sessions.
// The database record is preserved (not deleted).
func (a *AuthService) DeactivateAccount(ctx context.Context, userID int64, rawToken, ip string) error {
	// Revoke all sessions including the current one.
	if err := a.sessionSvc.RevokeAllSessionsForUser(ctx, userID); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}

	// Mark account deactivated.
	if err := a.userRepo.Deactivate(ctx, userID, time.Now().UTC()); err != nil {
		return fmt.Errorf("deactivate account: %w", err)
	}

	a.logAudit(ctx, &domain.AuditLog{
		EventType: domain.AuditEventAccountDeactivated,
		UserID:    &userID,
		IPAddress: ip,
		Success:   true,
	})

	log.Info().Int64("user_id", userID).Str("ip", ip).Msg("Account deactivated")
	return nil
}

// GetUserByID retrieves a user by ID.
func (a *AuthService) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	return a.userRepo.FindByID(ctx, userID)
}

// --- private helpers ---

func (a *AuthService) recordFailure(ctx context.Context, userID *int64, email, ip, reason string) {
	a.rateLimiter.RecordFailure(ip)
	a.logAuthEvent(ctx, userID, email, ip, false, reason)
	a.logAudit(ctx, &domain.AuditLog{
		EventType: domain.AuditEventLoginFailure,
		UserID:    userID,
		Email:     email,
		IPAddress: ip,
		Success:   false,
		Reason:    reason,
	})
}

func (a *AuthService) logAudit(ctx context.Context, entry *domain.AuditLog) {
	if err := a.auditRepo.CreateAuditLog(ctx, entry); err != nil {
		log.Error().Err(err).Msg("Failed to write audit log")
	}
}

func (a *AuthService) logAuthEvent(ctx context.Context, userID *int64, email, ip string, success bool, reason string) {
	if err := a.auditRepo.CreateAuthEvent(ctx, &domain.AuthEvent{
		UserID:    userID,
		Email:     email,
		IPAddress: ip,
		Success:   success,
		Reason:    reason,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to write auth event")
	}
}

// isValidEmail performs a basic structural email validation.
// Backend validation only — no DNS lookup.
func isValidEmail(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	local, domain := parts[0], parts[1]
	if len(local) == 0 || len(domain) == 0 {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}
	return true
}

// dummyVerify performs a dummy Argon2id operation to normalize timing when
// the user account does not exist. This prevents timing-based account enumeration.
func dummyVerify() error {
	// Use a hardcoded obviously-wrong hash to waste similar CPU time.
	const dummyHash = "$argon2id$v=19$m=65536,t=2,p=4$c29tZXNhbHRoZXJl$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	svc := NewPasswordService(DefaultPasswordParams())
	_, _ = svc.VerifyPassword("__dummy__", dummyHash)
	return nil
}
