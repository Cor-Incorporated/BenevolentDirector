"""Intelligence Worker entrypoint with graceful shutdown."""

from __future__ import annotations

import signal
import sys
import threading
from typing import NoReturn

import structlog

logger = structlog.get_logger()

_shutdown_event = threading.Event()


def _handle_signal(signum: int, _frame: object) -> None:
    """Handle termination signals for graceful shutdown.

    Args:
        signum: The signal number received.
        _frame: The current stack frame (unused).
    """
    sig_name = signal.Signals(signum).name
    logger.info("signal_received", signal=sig_name)
    _shutdown_event.set()


def main() -> NoReturn:
    """Start the intelligence worker and block until shutdown signal."""
    structlog.configure(
        processors=[
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.add_log_level,
            structlog.dev.ConsoleRenderer(),
        ],
    )

    signal.signal(signal.SIGINT, _handle_signal)
    signal.signal(signal.SIGTERM, _handle_signal)

    logger.info("intelligence-worker starting")

    _shutdown_event.wait()

    logger.info("intelligence-worker shutting down")
    sys.exit(0)


if __name__ == "__main__":
    main()
