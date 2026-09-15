package sqlite

import (
	"context"
	"database/sql"
	"time"

	"timeline-intelligence/backend/internal/domain"
)

// AuditRepo is the SQLite implementation of repository.AuditRepository.
type AuditRepo struct {
	db *sql.DB
}

// NewAuditRepo creates an AuditRepo using the provided DB connection.
func NewAuditRepo(db *sql.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) CreateAuditLog(ctx context.Context, entry *domain.AuditLog) error {
	successInt := 0
	if entry.Success {
		successInt = 1
	}
	var userID interface{}
	if entry.UserID != nil {
		userID = *entry.UserID
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs (event_type, user_id, email, ip_address, success, reason)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		string(entry.EventType), userID, entry.Email, entry.IPAddress, successInt, entry.Reason,
	)
	return err
}

func (r *AuditRepo) CreateAuthEvent(ctx context.Context, event *domain.AuthEvent) error {
	successInt := 0
	if event.Success {
		successInt = 1
	}
	var userID interface{}
	if event.UserID != nil {
		userID = *event.UserID
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO auth_events (user_id, email, ip_address, success, reason)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, event.Email, event.IPAddress, successInt, event.Reason,
	)
	return err
}

func (r *AuditRepo) CountRecentFailuresByIP(ctx context.Context, ip string, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM auth_events
		 WHERE ip_address=? AND success=0 AND created_at >= ?`,
		ip, since.UTC().Format(time.RFC3339Nano),
	).Scan(&count)
	return count, err
}

func (r *AuditRepo) ListAuditLogs(ctx context.Context, limit int) ([]*domain.AuditLog, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, event_type, user_id, email, ip_address, success, reason, created_at
		 FROM audit_logs ORDER BY created_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var entry domain.AuditLog
		var userID sql.NullInt64
		var successInt int
		var createdAt string

		err := rows.Scan(
			&entry.ID, &entry.EventType, &userID,
			&entry.Email, &entry.IPAddress, &successInt, &entry.Reason, &createdAt,
		)
		if err != nil {
			return nil, err
		}
		if userID.Valid {
			id := userID.Int64
			entry.UserID = &id
		}
		entry.Success = successInt == 1
		entry.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		logs = append(logs, &entry)
	}
	return logs, rows.Err()
}
