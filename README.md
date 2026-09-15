# General Purpose Event & Timeline Intelligence Platform

> **Early development version — foundation only.**

## Purpose

A self-hosted, offline-capable platform for importing, transforming, correlating, querying, and visualizing chronological event data. Domain-independent: suitable for logs, personal records, research data, or any timestamped information.

The initial version establishes the secure foundation: user accounts, authentication, and project structure. Event processing, visualization, and analysis belong to later development stages.

## Current Scope (v0.1)

- User registration with Argon2id password hashing
- Login / Logout with server-side sessions (no JWT)
- Account deactivation
- Basic authenticated landing page
- Navigation bar with account menu
- Placeholder pages: Settings, Help, FAQ, Logs, About
- SQLite database with schema migrations
- Python 3 build/dev menu (`build.py`)
- Security: CSRF protection, rate limiting, progressive login delays, audit logging

## Architecture

```
frontend/          Svelte + TypeScript + Vite (SPA)
backend/           Go + Gin HTTP server
  cmd/server/      Entry point
  internal/
    config/        Configuration
    domain/        Domain models (User, Session, AuditLog)
    repository/    Database interfaces + SQLite implementations
    service/       Business logic (auth, password, session, rate limiting)
    handler/       HTTP handlers + middleware
  migrations/      SQL migration files
  tests/           Go integration/unit tests
scripts/           Helper shell scripts
docs/              Architecture and API documentation
```

**Key design decisions:**
- Server-side sessions stored in SQLite (not JWT) — sessions are revocable
- Session token hashed before storage — raw token never in database
- Argon2id with OWASP-recommended parameters
- Repository interfaces isolate SQLite — PostgreSQL can be substituted later
- Frontend is purely presentational; all security decisions are backend-side

See [`docs/architecture.md`](docs/architecture.md) for details.

## Prerequisites

| Dependency | Version   | Notes                        |
|------------|-----------|------------------------------|
| Go         | ≥ 1.21    | Backend compiler             |
| Node.js    | ≥ 20 LTS  | Frontend build               |
| npm        | ≥ 10      | Frontend package manager     |
| Python 3   | ≥ 3.8     | Build orchestration          |
| gcc        | any       | Required for go-sqlite3 CGo  |

Use `python3 build.py` → `Check dependencies` to verify your environment.

## Installation

```bash
git clone <repo-url>
cd timelineIntelligence
python3 build.py
```

Then from the menu:
1. **Check dependencies** — verify Go, Node.js, npm are present
2. **Install missing dependencies** — guided installation if needed
3. **Build complete application** — builds backend + frontend
4. **Initialize/migrate database** — creates the SQLite database
5. **Start application** — starts backend + frontend dev server

## Build.py Usage

```
python3 build.py
```

Interactive menu — stays open after each operation until you choose **Exit**.

| Option | Description |
|--------|-------------|
| 1  | Detect operating system |
| 2  | Check dependencies |
| 3  | Install missing dependencies |
| 4  | Select Debug mode |
| 5  | Select Release mode |
| 6  | Build backend |
| 7  | Build frontend |
| 8  | Build complete application |
| 9  | Clean build |
| 10 | Run backend tests |
| 11 | Run frontend tests |
| 12 | Run all tests |
| 13 | Initialize/migrate database |
| 14 | Start backend |
| 15 | Start frontend |
| 16 | Start application |
| 17 | Run health checks |
| 18 | Show project/build information |
| 19 | Exit |

## Running

After building and database initialization:

```bash
python3 build.py   # then choose option 16 (Start application)
```

- Backend: `http://localhost:8080`
- Frontend: `http://localhost:5173`

Open `http://localhost:5173` in your browser.

## Testing

```bash
python3 build.py   # then choose option 12 (Run all tests)
```

Or individually:
```bash
# Backend tests
cd backend && go test ./tests/... -v

# Frontend unit tests
cd frontend && npm run test
```

## Database

- Location: `backend/data/timeline.db` (created on first run / migration)
- Engine: SQLite 3
- Migration files: `backend/migrations/`
- Initialize: `python3 build.py` → option 13

Do **not** commit `.db` files containing real user data. The `.gitignore` excludes them.

## Development Workflow

1. `python3 build.py` — open the menu
2. `Check dependencies` — ensure environment is ready
3. `Build complete application` — compile backend, bundle frontend
4. `Initialize/migrate database` — set up SQLite schema
5. `Start application` — start both servers
6. Edit code in `backend/` or `frontend/src/`
7. Backend: rebuild and restart to apply changes
8. Frontend: Vite hot-reloads automatically

See [`docs/development.md`](docs/development.md) for detailed workflow.

## Security Notes

- Passwords are hashed with Argon2id — plaintext is never stored or logged
- Sessions are server-side; the cookie contains only an opaque random token
- Failed login attempts are rate-limited with progressive delays
- All authentication events are recorded in the database
- See [`docs/security.md`](docs/security.md) for full details

## License

See `LICENSE`.
