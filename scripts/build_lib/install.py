"""
OS-aware dependency installation.

Does NOT install silently. Explains packages, asks for confirmation,
and uses the appropriate privilege escalation mechanism when needed.
"""

from __future__ import annotations

import shutil
import subprocess
from typing import Optional

from .detect import SystemInfo, gather_system_info, missing_required
from . import ui


InstallPlan = list[tuple[str, list[str]]]  # (description, command)


def _which_pkg_manager(candidates: list[str]) -> Optional[str]:
    for c in candidates:
        if shutil.which(c):
            return c
    return None


def _sudo_prefix() -> list[str]:
    """Return privilege-escalation prefix if needed."""
    if shutil.which("sudo"):
        return ["sudo"]
    return []


def plan_linux(sysinfo: SystemInfo) -> InstallPlan:
    missing_keys = []
    for key, _ in [
        ("go", "Go"),
        ("node", "Node.js"),
        ("npm", "npm"),
        ("gcc", "gcc"),
    ]:
        tool = sysinfo.tools.get(key)
        if not tool or not tool.available:
            missing_keys.append(key)

    if not missing_keys:
        return []

    mgr = _which_pkg_manager(["apt-get", "apt", "dnf", "yum", "pacman", "zypper"])
    if not mgr:
        return []

    sudo = _sudo_prefix()
    packages: list[str] = []

    # Map tools to distro packages (best-effort; package names vary).
    if mgr in ("apt-get", "apt"):
        if "go" in missing_keys:
            packages.append("golang-go")
        if "node" in missing_keys or "npm" in missing_keys:
            packages.extend(["nodejs", "npm"])
        if "gcc" in missing_keys:
            packages.append("build-essential")
        cmd = sudo + [mgr, "update"]
        install = sudo + [mgr, "install", "-y"] + packages
        return [
            ("Update package index", cmd),
            (f"Install packages: {', '.join(packages)}", install),
        ]

    if mgr in ("dnf", "yum"):
        if "go" in missing_keys:
            packages.append("golang")
        if "node" in missing_keys or "npm" in missing_keys:
            packages.extend(["nodejs", "npm"])
        if "gcc" in missing_keys:
            packages.append("gcc")
        return [(f"Install packages: {', '.join(packages)}", sudo + [mgr, "install", "-y"] + packages)]

    if mgr == "pacman":
        if "go" in missing_keys:
            packages.append("go")
        if "node" in missing_keys or "npm" in missing_keys:
            packages.extend(["nodejs", "npm"])
        if "gcc" in missing_keys:
            packages.append("base-devel")
        return [(f"Install packages: {', '.join(packages)}", sudo + [mgr, "-Sy", "--noconfirm"] + packages)]

    if mgr == "zypper":
        if "go" in missing_keys:
            packages.append("go")
        if "node" in missing_keys or "npm" in missing_keys:
            packages.extend(["nodejs", "npm"])
        if "gcc" in missing_keys:
            packages.append("gcc")
        return [(f"Install packages: {', '.join(packages)}", sudo + [mgr, "install", "-y"] + packages)]

    return []


def plan_darwin(sysinfo: SystemInfo) -> InstallPlan:
    missing_keys = []
    for key in ("go", "node", "npm", "gcc"):
        tool = sysinfo.tools.get(key)
        if not tool or not tool.available:
            missing_keys.append(key)
    if not missing_keys:
        return []

    brew = shutil.which("brew")
    if not brew:
        return []

    plan: InstallPlan = []
    if "go" in missing_keys:
        plan.append(("Install Go via Homebrew", [brew, "install", "go"]))
    if "node" in missing_keys or "npm" in missing_keys:
        plan.append(("Install Node.js (includes npm) via Homebrew", [brew, "install", "node"]))
    # gcc on macOS: Xcode CLT — brew install gcc is optional; recommend xcode-select
    if "gcc" in missing_keys:
        plan.append(
            (
                "Install Xcode Command Line Tools (provides clang/gcc-compatible toolchain)",
                ["xcode-select", "--install"],
            )
        )
    return plan


def plan_windows(sysinfo: SystemInfo) -> InstallPlan:
    missing_keys = []
    for key in ("go", "node", "npm"):
        tool = sysinfo.tools.get(key)
        if not tool or not tool.available:
            missing_keys.append(key)
    if not missing_keys:
        return []

    # Prefer winget, then chocolatey, then scoop.
    winget = shutil.which("winget")
    choco = shutil.which("choco")
    scoop = shutil.which("scoop")

    plan: InstallPlan = []
    if winget:
        if "go" in missing_keys:
            plan.append(("Install Go via winget", [winget, "install", "-e", "--id", "GoLang.Go"]))
        if "node" in missing_keys or "npm" in missing_keys:
            plan.append(("Install Node.js LTS via winget", [winget, "install", "-e", "--id", "OpenJS.NodeJS.LTS"]))
        return plan

    if choco:
        if "go" in missing_keys:
            plan.append(("Install Go via Chocolatey", [choco, "install", "golang", "-y"]))
        if "node" in missing_keys or "npm" in missing_keys:
            plan.append(("Install Node.js via Chocolatey", [choco, "install", "nodejs-lts", "-y"]))
        return plan

    if scoop:
        if "go" in missing_keys:
            plan.append(("Install Go via Scoop", [scoop, "install", "go"]))
        if "node" in missing_keys or "npm" in missing_keys:
            plan.append(("Install Node.js via Scoop", [scoop, "install", "nodejs-lts"]))
        return plan

    return []


def manual_instructions(sysinfo: SystemInfo) -> list[str]:
    lines = [
        "Automatic installation could not be planned for this environment.",
        "Please install the following manually, then re-run 'Check dependencies':",
        "",
    ]
    for label in missing_required(sysinfo):
        lines.append(f"  • {label}")
    lines.extend(
        [
            "",
            "Suggested official sources:",
            "  • Go:    https://go.dev/dl/",
            "  • Node:  https://nodejs.org/",
            "  • gcc:   use your OS package manager or build-essential / Xcode CLT",
        ]
    )
    if sysinfo.os_name == "linux":
        lines.append("  Linux example: sudo apt install golang-go nodejs npm build-essential")
    elif sysinfo.os_name == "darwin":
        lines.append("  macOS example: brew install go node && xcode-select --install")
    elif sysinfo.os_name == "windows":
        lines.append("  Windows example: winget install GoLang.Go OpenJS.NodeJS.LTS")
    return lines


def run_command(cmd: list[str]) -> int:
    ui.info("Running: " + " ".join(cmd))
    try:
        result = subprocess.run(cmd, check=False)
        return result.returncode
    except OSError as exc:
        ui.fail(f"Could not run command: {exc}")
        return 1


def install_missing_dependencies() -> bool:
    """
    Detect missing tools, explain the install plan, confirm with the user,
    then execute. Returns True if environment looks ready afterward.
    """
    sysinfo = gather_system_info()
    missing = missing_required(sysinfo)

    if not missing:
        ui.ok("All required dependencies are already installed.")
        return True

    ui.warn("Missing required dependencies:")
    for m in missing:
        print(f"  • {m}")
    print()

    if sysinfo.os_name == "linux":
        plan = plan_linux(sysinfo)
    elif sysinfo.os_name == "darwin":
        plan = plan_darwin(sysinfo)
    elif sysinfo.os_name == "windows":
        plan = plan_windows(sysinfo)
    else:
        plan = []

    if not plan:
        for line in manual_instructions(sysinfo):
            print(line)
        return False

    ui.header("Proposed installation steps")
    for i, (desc, cmd) in enumerate(plan, 1):
        print(f"  {i}. {desc}")
        print(f"     Command: {' '.join(cmd)}")
    print()
    ui.warn("Administrator privileges may be required.")
    ui.warn("Nothing will be installed unless you confirm.")

    if not ui.confirm("Proceed with installation?", default=False):
        ui.info("Installation cancelled by user.")
        return False

    for desc, cmd in plan:
        ui.info(desc)
        code = run_command(cmd)
        if code != 0:
            ui.fail(f"Step failed with exit code {code}: {desc}")
            ui.warn("Fix the issue above, then retry 'Install missing dependencies'.")
            return False
        ui.ok(f"Completed: {desc}")

    # Re-check.
    after = gather_system_info()
    still = missing_required(after)
    if still:
        ui.warn("Some dependencies are still missing after installation:")
        for m in still:
            print(f"  • {m}")
        ui.info("You may need to open a new terminal so PATH updates take effect.")
        return False

    ui.ok("All required dependencies are now available.")
    return True
