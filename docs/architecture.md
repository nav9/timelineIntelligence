# Architecture

## Overview

This project is a **General Purpose Event & Timeline Intelligence Platform**.
The current stage implements only the authentication foundation and application shell.

```
Frontend (Svelte + TypeScript + Vite)
        ↓  HTTP JSON + cookies
HTTP/API layer (Gin handlers)
        ↓
Application services (auth, session, password, rate limit)
        ↓
Repositories (interfaces)
        ↓
SQLite (initial persistence)
```

## Backend layout

| Path | Role |
|------|------|
| `backend/cmd/server` | HTTP server entry point |
| `backend/cmd/migrate` | One-shot DB migration helper |
| `backend/internal/config` | Environment configuration |
| `backend/internal/domain` | Domain models (User, Session, Audit) |
| `backend/internal/repository` | Persistence interfaces |
| `backend/internal/repository/sqlite` | SQLite implementations + embedded migrations |
| `backend/internal/service` | Business logic |
| `backend/internal/handler` | Thin HTTP adapters + middleware |
| `backend/tests` | Integration/unit tests |

## Frontend layout

| Path | Role |
|------|------|
| `frontend/src/routes` | Pages (login, register, home, placeholders) |
| `frontend/src/components` | Navbar, account menu, password UI |
| `frontend/src/lib` | API client, auth store, client password checks |
| `frontend/src/styles` | Dark theme + forms |

## Key decisions

1. **Server-side sessions, not JWT** — sessions are revocable and stored as token hashes only.
2. **Argon2id** — mature library (`golang.org/x/crypto/argon2`); no custom crypto.
3. **Repository interfaces** — SQLite can later be replaced with PostgreSQL without rewriting services.
4. **CSRF double-submit cookie** — `tl_csrf` cookie + `X-CSRF-Token` header on mutating requests.
5. **Admin log boundary** — `GET /api/admin/logs` requires `role=admin`; UI placeholder enforces the same idea.

## Build orchestration

`build.py` is the developer entry point. Modular helpers live under `scripts/build_lib/`:

- `detect.py` — OS and tool probing
- `install.py` — OS-aware dependency installation (explicit confirmation)
- `operations.py` — build, test, migrate, health
- `process_mgr.py` — start/stop backend and frontend processes

## Out of scope (this stage)

Event ingestion, rule engine, visualization, AI, collaboration, domain-specific engines.
