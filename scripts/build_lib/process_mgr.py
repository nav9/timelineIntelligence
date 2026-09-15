"""
Process management for starting/stopping backend and frontend.
Avoids orphaned processes where practical via a PID file.
"""

from __future__ import annotations

import os
import signal
import subprocess
import sys
import time
from pathlib import Path

from . import ui
from .operations import ensure_env_file, _load_dotenv_into, build_backend
from .paths import (
    BACKEND_BIN,
    BACKEND_DIR,
    BACKEND_HOST,
    BACKEND_PORT,
    ENV_FILE,
    FRONTEND_DIR,
    FRONTEND_PORT,
    PID_FILE,
    ROOT,
    STATE,
    env_with_build_mode,
)


def _read_pids() -> dict[str, int]:
    result: dict[str, int] = {}
    if not PID_FILE.exists():
        return result
    for line in PID_FILE.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or "=" not in line:
            continue
        name, _, pid_s = line.partition("=")
        try:
            result[name.strip()] = int(pid_s.strip())
        except ValueError:
            continue
    return result


def _write_pids(pids: dict[str, int]) -> None:
    lines = [f"{k}={v}" for k, v in pids.items()]
    PID_FILE.write_text("\n".join(lines) + ("\n" if lines else ""), encoding="utf-8")


def _pid_alive(pid: int) -> bool:
    if pid <= 0:
        return False
    try:
        os.kill(pid, 0)
        return True
    except OSError:
        return False


def stop_managed_processes() -> None:
    pids = _read_pids()
    if not pids:
        return
    for name, pid in list(pids.items()):
        if _pid_alive(pid):
            ui.info(f"Stopping {name} (pid {pid})")
            try:
                os.kill(pid, signal.SIGTERM)
            except OSError:
                pass
            # Brief wait, then SIGKILL if needed.
            for _ in range(20):
                if not _pid_alive(pid):
                    break
                time.sleep(0.1)
            if _pid_alive(pid):
                try:
                    os.kill(pid, signal.SIGKILL)
                except OSError:
                    pass
        del pids[name]
    _write_pids(pids)
    if PID_FILE.exists() and not pids:
        PID_FILE.unlink(missing_ok=True)


def _spawn(name: str, cmd: list[str], cwd: Path, env: dict[str, str]) -> bool:
    pids = _read_pids()
    if name in pids and _pid_alive(pids[name]):
        ui.warn(f"{name} already running (pid {pids[name]})")
        return True

    log_dir = ROOT / "logs"
    log_dir.mkdir(parents=True, exist_ok=True)
    log_path = log_dir / f"{name}.log"
    log_f = open(log_path, "a", encoding="utf-8")

    ui.info(f"Starting {name}: {' '.join(cmd)}")
    try:
        # New session so we can manage the process group.
        kwargs = {
            "cwd": str(cwd),
            "env": env,
            "stdout": log_f,
            "stderr": subprocess.STDOUT,
        }
        if os.name == "nt":
            kwargs["creationflags"] = subprocess.CREATE_NEW_PROCESS_GROUP  # type: ignore[attr-defined]
            proc = subprocess.Popen(cmd, **kwargs)
        else:
            proc = subprocess.Popen(cmd, start_new_session=True, **kwargs)
    except OSError as exc:
        log_f.close()
        ui.fail(f"Failed to start {name}: {exc}")
        return False

    pids[name] = proc.pid
    _write_pids(pids)
    STATE.managed_pids.append(proc.pid)
    ui.ok(f"{name} started (pid {proc.pid}), log: {log_path}")
    return True


def start_backend() -> bool:
    ensure_env_file()
    if not BACKEND_BIN.exists():
        ui.info("Backend binary missing — building first…")
        if not build_backend():
            return False

    env = env_with_build_mode()
    _load_dotenv_into(env, ENV_FILE)
    # Ensure BUILD_MODE from menu wins.
    env["BUILD_MODE"] = STATE.build_mode

    ok = _spawn("backend", [str(BACKEND_BIN)], BACKEND_DIR, env)
    if ok:
        print(f"  Backend URL: http://{BACKEND_HOST}:{BACKEND_PORT}")
        print("  Stop: menu option after start prompts Ctrl+C, or kill via Exit cleanup.")
    return ok


def start_frontend() -> bool:
    env = env_with_build_mode()
    # Prefer npx vite / npm run dev
    npm = "npm.cmd" if os.name == "nt" else "npm"
    if not (FRONTEND_DIR / "node_modules").exists():
        ui.info("Installing frontend dependencies…")
        code = subprocess.run([npm, "install"], cwd=str(FRONTEND_DIR), check=False).returncode
        if code != 0:
            ui.fail("npm install failed")
            return False

    ok = _spawn(
        "frontend",
        [npm, "run", "dev", "--", "--host", "127.0.0.1", "--port", str(FRONTEND_PORT)],
        FRONTEND_DIR,
        env,
    )
    if ok:
        print(f"  Frontend URL: http://127.0.0.1:{FRONTEND_PORT}")
        print("  API requests are proxied to the backend by Vite.")
    return ok


def start_application() -> bool:
    """Start backend + frontend and wait until the user presses Enter to stop."""
    stop_managed_processes()  # Clean slate
    if not start_backend():
        return False
    time.sleep(1.0)
    if not start_frontend():
        stop_managed_processes()
        return False

    print()
    ui.ok("Application running")
    print(f"  Open http://127.0.0.1:{FRONTEND_PORT} in your browser")
    print(f"  Backend API: http://{BACKEND_HOST}:{BACKEND_PORT}/api/health")
    print()
    print("  Press Enter to stop both processes and return to the menu.")
    try:
        input()
    except EOFError:
        pass
    stop_managed_processes()
    ui.ok("Application stopped")
    return True
