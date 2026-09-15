#!/usr/bin/env python3
"""
Timeline Intelligence — interactive development / build menu.

Run:
    python3 build.py

The menu stays open until the user chooses Exit.
Each operation reports success/failure and returns to the menu.

Dependency installation is explicit (option 3) and confirms before changing
the system. This script does not install software unless the user chooses to.
"""

from __future__ import annotations

import sys
from pathlib import Path

# Ensure scripts/ is on sys.path so `build_lib` imports work when run as build.py.
ROOT = Path(__file__).resolve().parent
SCRIPTS = ROOT / "scripts"
if str(SCRIPTS) not in sys.path:
    sys.path.insert(0, str(SCRIPTS))

from build_lib import ui  # noqa: E402
from build_lib import operations as ops  # noqa: E402
from build_lib import process_mgr as procs  # noqa: E402
from build_lib.install import install_missing_dependencies  # noqa: E402
from build_lib.paths import STATE  # noqa: E402


MENU = """
============================================================
 Timeline Intelligence — Build Menu
============================================================
  Build mode: {mode}

  1.  Detect operating system
  2.  Check dependencies
  3.  Install missing dependencies
  4.  Select Debug mode
  5.  Select Release mode
  6.  Build backend
  7.  Build frontend
  8.  Build complete application
  9.  Clean build
  10. Run backend tests
  11. Run frontend tests
  12. Run all tests
  13. Initialize/migrate database
  14. Start backend
  15. Start frontend
  16. Start application
  17. Run health checks
  18. Show project/build information
  19. Exit
============================================================
"""


def dispatch(choice: str) -> bool:
    """
    Execute a menu choice.
    Returns False only when the menu should exit.
    """
    if choice == "1":
        ops.show_os_info()
    elif choice == "2":
        ops.check_dependencies()
    elif choice == "3":
        install_missing_dependencies()
    elif choice == "4":
        ops.select_debug_mode()
    elif choice == "5":
        ops.select_release_mode()
    elif choice == "6":
        ops.build_backend()
    elif choice == "7":
        ops.build_frontend()
    elif choice == "8":
        ops.build_complete()
    elif choice == "9":
        ops.clean_build()
    elif choice == "10":
        ops.run_backend_tests()
    elif choice == "11":
        ops.run_frontend_tests()
    elif choice == "12":
        ops.run_all_tests()
    elif choice == "13":
        ops.init_database()
    elif choice == "14":
        procs.start_backend()
        print()
        ui.info("Backend runs in the background. Logs: logs/backend.log")
        ui.info("Use option 16's stop flow, or Exit, to stop managed processes.")
    elif choice == "15":
        procs.start_frontend()
        print()
        ui.info("Frontend runs in the background. Logs: logs/frontend.log")
    elif choice == "16":
        procs.start_application()
        return True  # already paused inside start_application
    elif choice == "17":
        ops.health_check()
    elif choice == "18":
        ops.show_project_info()
    elif choice == "19":
        return False
    else:
        ui.warn("Invalid choice. Enter a number from 1–19.")
        return True

    return True


def main() -> int:
    if sys.version_info < (3, 8):
        print("Python 3.8+ is required.", file=sys.stderr)
        return 1

    ui.info("Timeline Intelligence build menu")
    ui.info(f"Project root: {ROOT}")

    while True:
        print(MENU.format(mode=STATE.build_mode.upper()))
        try:
            choice = input("Select option [1-19]: ").strip()
        except (EOFError, KeyboardInterrupt):
            print()
            break

        if choice == "19":
            break

        # Option 16 handles its own pause.
        keep_going = dispatch(choice)
        if not keep_going:
            break
        if choice != "16":
            ui.pause()

    procs.stop_managed_processes()
    ui.info("Goodbye.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
