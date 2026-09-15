"""
Shared paths and build state for the Timeline Intelligence build system.
"""

from __future__ import annotations

import os
from pathlib import Path

# Project root = repository root (scripts/build_lib/ → ../..)
ROOT = Path(__file__).resolve().parents[2]
BACKEND_DIR = ROOT / "backend"
FRONTEND_DIR = ROOT / "frontend"
SCRIPTS_DIR = ROOT / "scripts"
DOCS_DIR = ROOT / "docs"
DATA_DIR = BACKEND_DIR / "data"
BACKEND_BIN = BACKEND_DIR / "bin" / "server"
FRONTEND_DIST = FRONTEND_DIR / "dist"
ENV_EXAMPLE = BACKEND_DIR / ".env.example"
ENV_FILE = BACKEND_DIR / ".env"
PID_FILE = ROOT / ".build_pids"
BUILD_MODE_FILE = ROOT / ".build_mode"

# Default ports
BACKEND_HOST = "127.0.0.1"
BACKEND_PORT = 8080
FRONTEND_PORT = 5173


class BuildState:
    """Mutable session state for the interactive menu."""

    def __init__(self) -> None:
        self.build_mode: str = self._load_mode()
        self.managed_pids: list[int] = []

    def _load_mode(self) -> str:
        if BUILD_MODE_FILE.exists():
            mode = BUILD_MODE_FILE.read_text(encoding="utf-8").strip().lower()
            if mode in ("debug", "release"):
                return mode
        return "debug"

    def set_mode(self, mode: str) -> None:
        if mode not in ("debug", "release"):
            raise ValueError(f"invalid build mode: {mode}")
        self.build_mode = mode
        BUILD_MODE_FILE.write_text(mode + "\n", encoding="utf-8")

    def is_debug(self) -> bool:
        return self.build_mode == "debug"


STATE = BuildState()


def env_with_build_mode(extra: dict[str, str] | None = None) -> dict[str, str]:
    """Return a copy of os.environ with BUILD_MODE set."""
    env = os.environ.copy()
    env["BUILD_MODE"] = STATE.build_mode
    if extra:
        env.update(extra)
    return env


def go_build_env(extra: dict[str, str] | None = None) -> dict[str, str]:
    """
    Environment for Go commands.

    Some Linux package installs leave GOPATH unset or equal to GOROOT, which
    breaks module download / sumdb. Use a safe default when misconfigured.
    """
    env = env_with_build_mode({"CGO_ENABLED": "1"})

    goroot = env.get("GOROOT", "").strip()
    gopath = env.get("GOPATH", "").strip()

    invalid = not gopath or (goroot and os.path.normpath(gopath) == os.path.normpath(goroot))

    if invalid:
        home_go = Path.home() / "go"
        if home_go.parent.exists():
            default_gopath = home_go
        else:
            default_gopath = ROOT / ".go"
        default_gopath.mkdir(parents=True, exist_ok=True)
        env["GOPATH"] = str(default_gopath)
        # Module cache lives under GOPATH on older Go toolchains.
        if not env.get("GOMODCACHE", "").strip():
            env["GOMODCACHE"] = str(default_gopath / "pkg" / "mod")

    if extra:
        env.update(extra)
    return env
