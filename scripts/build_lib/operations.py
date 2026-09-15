"""
Build, clean, test, migrate, and health-check operations.
"""

from __future__ import annotations

import os
import secrets
import shutil
import subprocess
import sys
import urllib.error
import urllib.request
from pathlib import Path

from . import ui
from .detect import gather_system_info, missing_required
from .paths import (
    BACKEND_BIN,
    BACKEND_DIR,
    DATA_DIR,
    ENV_EXAMPLE,
    ENV_FILE,
    FRONTEND_DIR,
    FRONTEND_DIST,
    ROOT,
    STATE,
    env_with_build_mode,
    go_build_env,
)


def _run(
    cmd: list[str],
    cwd: Path | None = None,
    env: dict[str, str] | None = None,
) -> int:
    ui.info(" ".join(cmd))
    try:
        result = subprocess.run(
            cmd,
            cwd=str(cwd) if cwd else None,
            env=env,
            check=False,
        )
        return result.returncode
    except OSError as exc:
        ui.fail(str(exc))
        return 1


def ensure_env_file() -> bool:
    """Create backend/.env from example if missing, with a generated SESSION_SECRET."""
    if ENV_FILE.exists():
        return True
    if not ENV_EXAMPLE.exists():
        ui.fail(f"Missing {ENV_EXAMPLE}")
        return False

    content = ENV_EXAMPLE.read_text(encoding="utf-8")
    secret = secrets.token_hex(32)
    content = content.replace(
        "CHANGE_THIS_TO_A_STRONG_RANDOM_SECRET_AT_LEAST_32_CHARS",
        secret,
    )
    # Prefer absolute-ish path under backend/data for local runs.
    content = content.replace("DB_PATH=./data/timeline.db", "DB_PATH=./data/timeline.db")
    ENV_FILE.write_text(content, encoding="utf-8")
    ui.ok(f"Created {ENV_FILE} with a generated SESSION_SECRET")
    ui.warn("Do not commit backend/.env — it contains secrets.")
    return True


def show_os_info() -> None:
    info = gather_system_info()
    ui.header("Operating system")
    print(f"  OS:           {info.os_pretty}")
    print(f"  Architecture: {info.architecture}")
    print(f"  Python:       {info.python_version}")
    print()
    print("  Tools:")
    for key in ("go", "node", "npm", "gcc", "make", "git"):
        tool = info.tools.get(key)
        if not tool:
            continue
        if tool.available:
            print(f"    {tool.name:10} {tool.version}")
            print(f"               {tool.path}")
        else:
            print(f"    {tool.name:10} NOT FOUND")


def check_dependencies() -> bool:
    info = gather_system_info()
    ui.header("Dependency check")
    all_ok = True
    for key, label in [
        ("go", "Go"),
        ("node", "Node.js"),
        ("npm", "npm"),
        ("gcc", "gcc"),
    ]:
        tool = info.tools[key]
        if tool.available:
            ui.ok(f"{label}: {tool.version}")
        else:
            ui.fail(f"{label}: missing")
            all_ok = False

    missing = missing_required(info)
    if missing:
        print()
        ui.warn("Use menu option 3 to install missing dependencies.")
    else:
        print()
        ui.ok("Environment looks ready.")
    return all_ok


def select_debug_mode() -> None:
    STATE.set_mode("debug")
    ui.ok("Build mode set to DEBUG")
    print("  Backend: Gin debug mode, verbose logging")
    print("  Frontend: Vite development build with sourcemaps")


def select_release_mode() -> None:
    STATE.set_mode("release")
    ui.ok("Build mode set to RELEASE")
    print("  Backend: Gin release mode, production logging defaults")
    print("  Frontend: minified production bundle without sourcemaps")


def build_backend() -> bool:
    info = gather_system_info()
    if not info.tools["go"].available:
        ui.fail("Go is not installed. Use option 3 to install dependencies.")
        return False
    if not info.tools["gcc"].available:
        ui.warn("gcc not found — go-sqlite3 requires a C compiler (CGo).")

    ensure_env_file()
    bin_dir = BACKEND_BIN.parent
    bin_dir.mkdir(parents=True, exist_ok=True)

    original_gopath = os.environ.get("GOPATH", "").strip()
    original_goroot = os.environ.get("GOROOT", "").strip()
    env = go_build_env()
    if (
        not original_gopath
        or (original_goroot and os.path.normpath(original_gopath) == os.path.normpath(original_goroot))
    ):
        ui.info(f"Using GOPATH={env['GOPATH']} (system GOPATH was missing or invalid)")
    # Fetch modules (generates go.sum if needed).
    if _run(["go", "mod", "download"], cwd=BACKEND_DIR, env=env) != 0:
        ui.fail("go mod download failed")
        return False
    if _run(["go", "mod", "tidy"], cwd=BACKEND_DIR, env=env) != 0:
        ui.warn("go mod tidy reported issues (continuing)")

    ldflags = []
    if STATE.build_mode == "release":
        ldflags = ["-ldflags", "-s -w"]

    cmd = ["go", "build", "-o", str(BACKEND_BIN)] + ldflags + ["./cmd/server"]
    if _run(cmd, cwd=BACKEND_DIR, env=env) != 0:
        ui.fail("Backend build failed")
        return False

    ui.ok(f"Backend built → {BACKEND_BIN}")
    return True


def build_frontend() -> bool:
    info = gather_system_info()
    if not info.tools["npm"].available:
        ui.fail("npm is not installed. Use option 3 to install dependencies.")
        return False

    if _run(["npm", "install"], cwd=FRONTEND_DIR) != 0:
        ui.fail("npm install failed")
        return False

    if STATE.build_mode == "release":
        cmd = ["npm", "run", "build"]
        # vite uses --mode production by default for build
    else:
        # Debug: build with development mode (sourcemaps on)
        cmd = ["npm", "run", "build", "--", "--mode", "development"]

    if _run(cmd, cwd=FRONTEND_DIR) != 0:
        ui.fail("Frontend build failed")
        return False

    ui.ok(f"Frontend built → {FRONTEND_DIST}")
    return True


def build_complete() -> bool:
    ok_be = build_backend()
    ok_fe = build_frontend()
    if ok_be and ok_fe:
        ui.ok("Complete application build succeeded")
        return True
    ui.fail("Complete build finished with errors")
    return False


def clean_build() -> bool:
    removed = []
    for path in (
        BACKEND_BIN,
        BACKEND_BIN.parent,
        FRONTEND_DIST,
        FRONTEND_DIR / "node_modules" / ".vite",
    ):
        if path.exists():
            if path.is_file():
                path.unlink()
            elif path.name == "bin" and path.is_dir():
                shutil.rmtree(path, ignore_errors=True)
            else:
                shutil.rmtree(path, ignore_errors=True)
            removed.append(str(path))

    # Go cache for this module is left alone; only project artifacts.
    ui.ok("Clean complete")
    if removed:
        for r in removed:
            print(f"  removed: {r}")
    else:
        print("  (nothing to remove)")
    return True


def run_backend_tests() -> bool:
    info = gather_system_info()
    if not info.tools["go"].available:
        ui.fail("Go is not installed.")
        return False
    env = go_build_env()
    # Ensure modules are present.
    _run(["go", "mod", "download"], cwd=BACKEND_DIR, env=env)
    code = _run(["go", "test", "./tests/...", "-count=1", "-timeout", "120s"], cwd=BACKEND_DIR, env=env)
    if code == 0:
        ui.ok("Backend tests passed")
        return True
    ui.fail("Backend tests failed")
    return False


def run_frontend_tests() -> bool:
    info = gather_system_info()
    if not info.tools["npm"].available:
        ui.fail("npm is not installed.")
        return False
    if not (FRONTEND_DIR / "node_modules").exists():
        if _run(["npm", "install"], cwd=FRONTEND_DIR) != 0:
            return False
    code = _run(["npm", "run", "test"], cwd=FRONTEND_DIR)
    if code == 0:
        ui.ok("Frontend tests passed")
        return True
    ui.fail("Frontend tests failed")
    return False


def run_all_tests() -> bool:
    be = run_backend_tests()
    fe = run_frontend_tests()
    if be and fe:
        ui.ok("All tests passed")
        return True
    ui.fail("One or more test suites failed")
    return False


def init_database() -> bool:
    """
    Initialize/migrate the SQLite database by running a short Go program path:
    opening the DB via the server binary's migrate-on-open, or a one-shot approach.

    Strategy: ensure .env, create data dir, run `go run` of a tiny migrate helper
    embedded as importing sqlite.Open — simplest reliable approach is to start
    the server briefly is heavy; instead run tests' Open via go run ./cmd/migrate
    if present, else use go run with the Open from a migrate command.

    For this foundation we provide cmd/migrate.
    """
    ensure_env_file()
    DATA_DIR.mkdir(parents=True, exist_ok=True)

    info = gather_system_info()
    if not info.tools["go"].available:
        ui.fail("Go is not installed.")
        return False

    migrate_main = BACKEND_DIR / "cmd" / "migrate" / "main.go"
    if not migrate_main.exists():
        ui.fail("Missing backend/cmd/migrate — cannot initialize database")
        return False

    env = go_build_env()
    # Load .env into environment for the child process.
    _load_dotenv_into(env, ENV_FILE)

    code = _run(["go", "run", "./cmd/migrate"], cwd=BACKEND_DIR, env=env)
    if code == 0:
        ui.ok("Database initialized / migrations applied")
        print(f"  DB path: see DB_PATH in {ENV_FILE}")
        return True
    ui.fail("Database initialization failed")
    return False


def _load_dotenv_into(env: dict[str, str], path: Path) -> None:
    if not path.exists():
        return
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, val = line.partition("=")
        key = key.strip()
        val = val.strip().strip('"').strip("'")
        # Do not override an already-set environment variable.
        if key and key not in os.environ:
            env[key] = val
        elif key:
            env.setdefault(key, val)


def health_check() -> bool:
    url = f"http://127.0.0.1:8080/api/health"
    ui.info(f"GET {url}")
    try:
        with urllib.request.urlopen(url, timeout=5) as resp:
            body = resp.read().decode("utf-8", errors="replace")
            print(f"  HTTP {resp.status}")
            print(f"  {body}")
            if resp.status == 200:
                ui.ok("Health check passed")
                return True
            ui.fail("Health check returned non-OK status")
            return False
    except urllib.error.URLError as exc:
        ui.fail(f"Backend not reachable: {exc}")
        ui.info("Start the backend first (menu option 14 or 16).")
        return False


def show_project_info() -> None:
    ui.header("Project / build information")
    print(f"  Root:       {ROOT}")
    print(f"  Build mode: {STATE.build_mode}")
    print(f"  Backend:    {BACKEND_DIR}")
    print(f"  Frontend:   {FRONTEND_DIR}")
    print(f"  Binary:     {BACKEND_BIN} ({'exists' if BACKEND_BIN.exists() else 'not built'})")
    print(f"  FE dist:    {FRONTEND_DIST} ({'exists' if FRONTEND_DIST.exists() else 'not built'})")
    print(f"  Env file:   {ENV_FILE} ({'exists' if ENV_FILE.exists() else 'missing — will be created on build/migrate'})")
    print(f"  Python:     {sys.version.split()[0]}")
    info = gather_system_info()
    print(f"  OS:         {info.os_pretty} / {info.architecture}")
    print()
    print("  Debug vs Release:")
    print("    debug   — Gin debug, console logs, frontend sourcemaps")
    print("    release — Gin release, stripped binary, minified frontend")
