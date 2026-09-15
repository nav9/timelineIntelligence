# Authentication API (v0.1)

Base URL (dev): `http://127.0.0.1:8080/api`

All mutating requests must include:

- Cookie: session cookie when authenticated (`tl_session`)
- Header: `X-CSRF-Token` matching the `tl_csrf` cookie

Credentials (cookies) must be included (`credentials: 'include'`).

## `GET /health`

Public health check. No sensitive internals.

```json
{
  "status": "ok",
  "timestamp": "2026-01-01T00:00:00Z",
  "components": { "database": "ok", "api": "ok" }
}
```

## `POST /auth/register`

Body:

```json
{ "name": "Ada", "email": "ada@example.com", "password": "long-passphrase-here" }
```

Responses:

- `201` — created (`user` without password hash)
- `409` — email taken
- `422` — weak password (`feedback` array)
- `400` — validation error

## `POST /auth/login`

Body:

```json
{ "email": "ada@example.com", "password": "long-passphrase-here" }
```

Responses:

- `200` — sets `tl_session` cookie; returns user
- `401` — `{"error":"invalid credentials"}` (generic)
- `429` — rate limited

## `POST /auth/logout`

Requires authentication. Revokes session and clears cookie.

## `GET /auth/me`

Requires authentication. Returns current user.

## `DELETE /auth/account`

Requires authentication. Deactivates account, revokes all sessions, clears cookie.

## `GET /admin/logs`

Requires authentication **and** `role=admin`. Returns recent audit log entries.
