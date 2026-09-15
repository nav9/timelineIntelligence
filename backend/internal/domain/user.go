// Package domain defines the core domain models for the application.
// These are plain Go structs with no framework dependencies.
//
// Domain models are immutable value objects passed between layers.
// The repository layer is responsible for persistence; the service layer
// for business logic. Domain models must not import repository or handler packages.
package domain

import "time"

// UserStatus represents the lifecycle state of a user account.
type UserStatus string

const (
	// UserStatusActive indicates a normal, usable account.
	UserStatusActive UserStatus = "active"

	// UserStatusDeactivated indicates an account that has been deactivated by the user.
	// The record is retained for data integrity and potential regulatory requirements.
	// Deactivated accounts cannot log in.
	UserStatusDeactivated UserStatus = "deactivated"
)

// UserRole controls what capabilities a user has.
type UserRole string

const (
	// UserRoleUser is a standard authenticated user.
	UserRoleUser UserRole = "user"

	// UserRoleAdmin has access to administrative functions such as viewing audit logs.
	// Admin functionality is not fully implemented in the initial version.
	UserRoleAdmin UserRole = "admin"
)

// User represents an authenticated user of the platform.
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string // Argon2id encoded hash; never transmitted to clients
	Status       UserStatus
	Role         UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeactivatedAt *time.Time // nil unless deactivated
}

// IsActive returns true if the user account is in the active state.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsAdmin returns true if the user has the admin role.
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// SafeUser is a User representation safe to transmit to clients.
// It never includes the password hash or other internal fields.
type SafeUser struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Status    UserStatus `json:"status"`
	Role      UserRole   `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
}

// ToSafe converts a User to a SafeUser suitable for API responses.
func (u *User) ToSafe() SafeUser {
	return SafeUser{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Status:    u.Status,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
