"""
Terminal helpers for the build menu.
"""

from __future__ import annotations

import sys


def info(msg: str) -> None:
    print(f"[INFO] {msg}")


def ok(msg: str) -> None:
    print(f"[ OK ] {msg}")


def warn(msg: str) -> None:
    print(f"[WARN] {msg}")


def fail(msg: str) -> None:
    print(f"[FAIL] {msg}", file=sys.stderr)


def header(title: str) -> None:
    print()
    print("=" * 60)
    print(f" {title}")
    print("=" * 60)


def pause() -> None:
    try:
        input("\nPress Enter to return to the menu… ")
    except EOFError:
        pass


def confirm(prompt: str, default: bool = False) -> bool:
    suffix = " [Y/n] " if default else " [y/N] "
    try:
        answer = input(prompt + suffix).strip().lower()
    except EOFError:
        return default
    if not answer:
        return default
    return answer in ("y", "yes")
