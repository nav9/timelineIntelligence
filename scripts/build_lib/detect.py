"""
Operating-system detection and tool version probing.
"""

from __future__ import annotations

import platform
import shutil
import subprocess
from dataclasses import dataclass, field
from typing import Optional


@dataclass
class ToolInfo:
    name: str
    available: bool
    version: str = ""
    path: str = ""
    notes: str = ""


@dataclass
class SystemInfo:
    os_name: str  # linux | windows | darwin | other
    os_pretty: str
    architecture: str
    python_version: str
    tools: dict[str, ToolInfo] = field(default_factory=dict)


def detect_os() -> str:
    system = platform.system().lower()
    if system == "linux":
        return "linux"
    if system == "windows":
        return "windows"
    if system == "darwin":
        return "darwin"
    return "other"


def _run_version(cmd: list[str]) -> Optional[str]:
    try:
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=15,
            check=False,
        )
        out = (result.stdout or result.stderr or "").strip()
        return out.splitlines()[0] if out else None
    except (OSError, subprocess.TimeoutExpired):
        return None


def probe_tool(name: str, binary: str, version_args: list[str]) -> ToolInfo:
    path = shutil.which(binary)
    if not path:
        return ToolInfo(name=name, available=False, notes=f"'{binary}' not found on PATH")
    version = _run_version([path] + version_args) or "unknown"
    return ToolInfo(name=name, available=True, version=version, path=path)


def gather_system_info() -> SystemInfo:
    os_name = detect_os()
    pretty = {
        "linux": "Linux",
        "windows": "Windows",
        "darwin": "macOS",
        "other": platform.system() or "Unknown",
    }[os_name]

    info = SystemInfo(
        os_name=os_name,
        os_pretty=pretty,
        architecture=platform.machine() or "unknown",
        python_version=platform.python_version(),
    )

    info.tools["python"] = ToolInfo(
        name="Python",
        available=True,
        version=platform.python_version(),
        path=sys_executable(),
    )
    info.tools["go"] = probe_tool("Go", "go", ["version"])
    info.tools["node"] = probe_tool("Node.js", "node", ["--version"])
    info.tools["npm"] = probe_tool("npm", "npm", ["--version"])
    info.tools["gcc"] = probe_tool("gcc", "gcc", ["--version"])
    info.tools["make"] = probe_tool("make", "make", ["--version"])
    info.tools["git"] = probe_tool("git", "git", ["--version"])

    return info


def sys_executable() -> str:
    import sys

    return sys.executable


def required_dependencies() -> list[tuple[str, str]]:
    """Return (tool_key, human_name) for required build/run tools."""
    return [
        ("go", "Go (≥ 1.21)"),
        ("node", "Node.js (≥ 20)"),
        ("npm", "npm (≥ 10)"),
        ("gcc", "gcc (required for go-sqlite3 CGo)"),
    ]


def missing_required(sysinfo: SystemInfo) -> list[str]:
    missing = []
    for key, label in required_dependencies():
        tool = sysinfo.tools.get(key)
        if not tool or not tool.available:
            missing.append(label)
    return missing
