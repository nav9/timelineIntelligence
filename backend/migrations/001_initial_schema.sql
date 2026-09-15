-- ============================================================
-- Migration: 001_initial_schema
-- Description: Create initial tables for users, sessions,
--              authentication events, and audit logs.
-- ============================================================

PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

-- ============================================================
-- users
-- Stores registered user accounts.
-- Passwords are never stored here — only the Argon2id hash.
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT    NOT NULL,
    email           TEXT    NOT NULL,
    password_hash   TEXT    NOT NULL,
    status          TEXT    NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'deactivated')),
    role            TEXT    NOT NULL DEFAULT 'user'
                    CHECK (role IN ('user', 'admin')),
    created_at      DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at      DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    deactivated_at  DATETIME
);

-- Email uniqueness: enforce case-insensitive uniqueness via collation.
-- Using UNIQUE INDEX rather than inline UNIQUE to control collation.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email COLLATE NOCASE);

CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);

-- ============================================================
-- sessions
-- Server-side sessions. The raw session token is NEVER stored.
-- Only a SHA-256 hash of the token is stored in token_hash.
-- ============================================================
CREATE TABLE IF NOT EXISTS sessions (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash   TEXT    NOT NULL UNIQUE,
    created_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    expires_at   DATETIME NOT NULL,
    last_seen_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    revoked_at   DATETIME,
    ip_address   TEXT    NOT NULL DEFAULT '',
    user_agent   TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id   ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions (token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);

-- ============================================================
-- auth_events
-- Records individual login attempts (successes and failures).
-- Kept separate from audit_logs for efficient rate-limit queries.
-- Passwords are never stored here.
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER REFERENCES users(id) ON DELETE SET NULL,
    email       TEXT    NOT NULL DEFAULT '',
    ip_address  TEXT    NOT NULL DEFAULT '',
    success     INTEGER NOT NULL DEFAULT 0 CHECK (success IN (0, 1)),
    reason      TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_auth_events_ip_address  ON auth_events (ip_address, created_at);
CREATE INDEX IF NOT EXISTS idx_auth_events_user_id     ON auth_events (user_id);
CREATE INDEX IF NOT EXISTS idx_auth_events_success     ON auth_events (success);

-- ============================================================
-- audit_logs
-- Records security and account lifecycle events.
-- Passwords, hashes, and raw session tokens are never stored here.
-- ============================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type  TEXT    NOT NULL,
    user_id     INTEGER REFERENCES users(id) ON DELETE SET NULL,
    email       TEXT    NOT NULL DEFAULT '',
    ip_address  TEXT    NOT NULL DEFAULT '',
    success     INTEGER NOT NULL DEFAULT 1 CHECK (success IN (0, 1)),
    reason      TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs (event_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id    ON audit_logs (user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at);

-- ============================================================
-- schema_migrations
-- Tracks which migration files have been applied.
-- ============================================================
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     TEXT    NOT NULL PRIMARY KEY,
    applied_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT OR IGNORE INTO schema_migrations (version) VALUES ('001_initial_schema');
