package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"timeline-intelligence/backend/internal/domain"
)

// SessionRepo is the SQLite implementation of repository.SessionRepository.
type SessionRepo struct {
	db *sql.DB
}

// NewSessionRepo creates a SessionRepo using the provided DB connection.
func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(
	ctx context.Context,
	userID int64,
	tokenHash string,
	expiresAt time.Time,
	ip, userAgent string,
) (*domain.Session, error) {
	const query = `
		INSERT INTO sessions (user_id, token_hash, expires_at, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		userID,
		tokenHash,
		expiresAt.UTC().Format(time.RFC3339Nano),
		ip,
		userAgent,
	)
	if err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get session id: %w", err)
	}
	return r.findByID(ctx, id)
}

func (r *SessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	const query = `
		SELECT id, user_id, token_hash, created_at, expires_at, last_seen_at,
		       revoked_at, ip_address, user_agent
		FROM sessions WHERE token_hash = ?
	`
	row := r.db.QueryRowContext(ctx, query, tokenHash)
	s, err := scanSession(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *SessionRepo) UpdateLastSeen(ctx context.Context, sessionID int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET last_seen_at=? WHERE id=?`,
		at.UTC().Format(time.RFC3339Nano), sessionID,
	)
	return err
}

func (r *SessionRepo) Revoke(ctx context.Context, sessionID int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET revoked_at=? WHERE id=? AND revoked_at IS NULL`,
		at.UTC().Format(time.RFC3339Nano), sessionID,
	)
	return err
}

func (r *SessionRepo) RevokeAllForUser(ctx context.Context, userID int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`,
		at.UTC().Format(time.RFC3339Nano), userID,
	)
	return err
}

func (r *SessionRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE expires_at < ?`,
		before.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *SessionRepo) findByID(ctx context.Context, id int64) (*domain.Session, error) {
	const query = `
		SELECT id, user_id, token_hash, created_at, expires_at, last_seen_at,
		       revoked_at, ip_address, user_agent
		FROM sessions WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return scanSession(row)
}

func scanSession(row *sql.Row) (*domain.Session, error) {
	var s domain.Session
	var createdAt, expiresAt, lastSeenAt string
	var revokedAt sql.NullString

	err := row.Scan(
		&s.ID, &s.UserID, &s.TokenHash,
		&createdAt, &expiresAt, &lastSeenAt,
		&revokedAt,
		&s.IPAddress, &s.UserAgent,
	)
	if err != nil {
		return nil, err
	}

	s.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	s.ExpiresAt, _ = time.Parse(time.RFC3339Nano, expiresAt)
	s.LastSeenAt, _ = time.Parse(time.RFC3339Nano, lastSeenAt)
	if revokedAt.Valid && revokedAt.String != "" {
		t, _ := time.Parse(time.RFC3339Nano, revokedAt.String)
		s.RevokedAt = &t
	}
	return &s, nil
}
