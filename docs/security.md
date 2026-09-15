# Security notes (v0.1)

## Passwords

- Hashed with **Argon2id** (`golang.org/x/crypto/argon2`)
- Random 16-byte salt per password
- Encoded hash stores algorithm parameters for verification
- Never logged, never returned in API responses

## Sessions

- Cryptographically random 32-byte token (hex-encoded)
- Only SHA-256 hash stored in SQLite
- Cookie: `tl_session`, HttpOnly, SameSite=Strict
- Set `Secure` when serving over HTTPS in production
- Logout and deactivation revoke sessions server-side

## Login hardening

- Generic error message (no account enumeration)
- Progressive delay on failures (server-side; cannot be bypassed by the frontend)
- IP-based rate limiting after repeated failures
- Failed and successful attempts recorded in `auth_events` and `audit_logs`

## CSRF

- Double-submit cookie pattern (`tl_csrf` + `X-CSRF-Token`)
- Applied to state-changing `/api` methods

## HTTP headers

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Content-Security-Policy` (restrictive; allows `'unsafe-inline'` styles for Svelte)
- `Referrer-Policy`, `Permissions-Policy`

## Authorization

- Authenticated routes require a valid session
- Admin logs require `role=admin` (`RequireAdmin` middleware)
- Ordinary users see a Logs placeholder stating access is denied

## Secrets

- `SESSION_SECRET` must be set (≥ 32 characters)
- Provided via environment / `.env` — never hard-coded
- `.gitignore` excludes `.env` and database files

## Retention / deactivation

Account deactivation marks the user inactive and revokes sessions.
Records are retained; permanent deletion is a future concern driven by
applicable legal and operational requirements — no fixed retention period is claimed.
