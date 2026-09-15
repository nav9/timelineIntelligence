# Development workflow

## First-time setup

```bash
python3 build.py
```

Then:

1. **Detect operating system** (option 1)
2. **Check dependencies** (option 2)
3. **Install missing dependencies** (option 3) — explains packages and asks before installing
4. **Build complete application** (option 8)
5. **Initialize/migrate database** (option 13)
6. **Start application** (option 16)

Open `http://127.0.0.1:5173`.

## Configuration

Copy is automatic: the first build/migrate creates `backend/.env` from `backend/.env.example`
with a generated `SESSION_SECRET`.

Do not commit `backend/.env`.

### Go module path (GOPATH)

Some Linux Go packages leave `GOPATH` unset or equal to `GOROOT`, which breaks `go mod download`.
`build.py` detects this and uses `~/go` (or `.go/` in the project if no home directory exists).
You can also set it yourself:

```bash
export GOPATH="$HOME/go"
mkdir -p "$GOPATH"
```

## Debug vs Release

| Mode | Backend | Frontend |
|------|---------|----------|
| Debug (default) | Gin debug mode, verbose logs | Vite build with sourcemaps (`--mode development`) |
| Release | Gin release mode, stripped binary (`-ldflags -s -w`) | Minified production bundle |

Select mode with menu options 4 and 5. The choice is stored in `.build_mode`.

## Local URLs

- Frontend (Vite): `http://127.0.0.1:5173` — proxies `/api` to the backend
- Backend: `http://127.0.0.1:8080`

## Stopping the app

When using **Start application** (16), press Enter in the build menu to stop both processes.

Managed PIDs are tracked in `.build_pids`. Exit also attempts cleanup.

## Testing

```bash
python3 build.py   # option 12 — Run all tests
```

Or:

```bash
cd backend && go test ./tests/... -count=1
cd frontend && npm test
```

## Database

- Default path: `backend/data/timeline.db` (relative to backend working directory)
- Migrations: embedded from `backend/internal/repository/sqlite/migrations/`
- Also mirrored under `backend/migrations/` for documentation

## Adding code later

- Put business rules in `service/`
- Put SQL in `repository/sqlite/`
- Keep handlers thin
- Keep Svelte components free of security decisions
