package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"timeline-intelligence/backend/internal/domain"
)

// UserRepo is the SQLite implementation of repository.UserRepository.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo creates a UserRepo using the provided DB connection.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, name, email, passwordHash string) (*domain.User, error) {
	const query = `
		INSERT INTO users (name, email, password_hash, status, role, created_at, updated_at)
		VALUES (?, ?, ?, 'active', 'user', strftime('%Y-%m-%dT%H:%M:%fZ','now'), strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	`
	result, err := r.db.ExecContext(ctx, query, name, email, passwordHash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, fmt.Errorf("email already registered")
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *UserRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	const query = `
		SELECT id, name, email, password_hash, status, role,
		       created_at, updated_at, deactivated_at
		FROM users WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)
	u, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, name, email, password_hash, status, role,
		       created_at, updated_at, deactivated_at
		FROM users WHERE email = ? COLLATE NOCASE
	`
	row := r.db.QueryRowContext(ctx, query, email)
	u, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil //nolint:nilnil // interface contract: nil,nil means not found
	}
	return u, err
}

func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM users WHERE email = ? COLLATE NOCASE`, email,
	).Scan(&count)
	return count > 0, err
}

func (r *UserRepo) Deactivate(ctx context.Context, userID int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET status='deactivated', deactivated_at=?, updated_at=? WHERE id=?`,
		at.UTC().Format(time.RFC3339Nano), at.UTC().Format(time.RFC3339Nano), userID,
	)
	return err
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, userID int64, newHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash=?, updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id=?`,
		newHash, userID,
	)
	return err
}

// scanUser scans a single user row into a domain.User.
func scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	var deactivatedAt sql.NullString
	var createdAt, updatedAt string

	err := row.Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash,
		&u.Status, &u.Role,
		&createdAt, &updatedAt, &deactivatedAt,
	)
	if err != nil {
		return nil, err
	}

	u.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	u.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	if deactivatedAt.Valid && deactivatedAt.String != "" {
		t, _ := time.Parse(time.RFC3339Nano, deactivatedAt.String)
		u.DeactivatedAt = &t
	}
	return &u, nil
}
