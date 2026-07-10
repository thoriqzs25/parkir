"""
Entry point for the Gate Controller.

Usage:
    python -m gate_controller.main --mode mock --port 8000
    python -m gate_controller.main --mode real --port 8000 --relay-pin 17 --loop-pin 18

Environment variables:
    GATE_CONTROLLER_MODE: mock | real | auto (default: auto)
    GATE_CONTROLLER_PORT: HTTP port (default: 8000)
    GATE_CONTROLLER_RELAY_PIN: GPIO pin for relay (default: 17)
    GATE_CONTROLLER_LOOP_PIN: GPIO pin for loop sensor (default: 18)
"""

import sys
from pathlib import Path

# Add shared package to path
_SHARED_SRC = Path(__file__).parent.parent.parent.parent / "shared" / "src"
if str(_SHARED_SRC) not in sys.path:
    sys.path.insert(0, str(_SHARED_SRC))

# Add gate-controller src to path
_GC_SRC = Path(__file__).parent.parent
if str(_GC_SRC) not in sys.path:
    sys.path.insert(0, str(_GC_SRC))

import argparse
import os
import logging

from gate_controller.interfaces import create_relay, create_loop_sensor
from gate_controller.safety import SafetyManager
from gate_controller.api import create_app

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger(__name__)


def main():
    parser = argparse.ArgumentParser(description="PAS Gate Controller")
    parser.add_argument(
        "--mode",
        choices=["mock", "real", "auto"],
        default=os.environ.get("GATE_CONTROLLER_MODE", "auto"),
        help="Hardware mode: mock (software-only), real (GPIO), auto (try real, fallback to mock)",
    )
    parser.add_argument(
        "--port",
        type=int,
        default=int(os.environ.get("GATE_CONTROLLER_PORT", "8000")),
        help="HTTP server port",
    )
    parser.add_argument(
        "--relay-pin",
        type=int,
        default=int(os.environ.get("GATE_CONTROLLER_RELAY_PIN", "17")),
        help="GPIO pin for relay control",
    )
    parser.add_argument(
        "--loop-pin",
        type=int,
        default=int(os.environ.get("GATE_CONTROLLER_LOOP_PIN", "18")),
        help="GPIO pin for loop sensor input",
    )
    parser.add_argument(
        "--auto-close-timeout",
        type=float,
        default=float(os.environ.get("GATE_CONTROLLER_AUTO_CLOSE_TIMEOUT", "30.0")),
        help="Seconds before auto-close timer fires",
    )
    parser.add_argument(
        "--host",
        default=os.environ.get("GATE_CONTROLLER_HOST", "0.0.0.0"),
        help="HTTP server host",
    )
    args = parser.parse_args()

    logger.info("Starting Gate Controller")
    logger.info("  mode: %s", args.mode)
    logger.info("  host: %s", args.host)
    logger.info("  port: %d", args.port)
    logger.info("  relay_pin: %d", args.relay_pin)
    logger.info("  loop_pin: %d", args.loop_pin)
    logger.info("  auto_close_timeout: %.1fs", args.auto_close_timeout)

    # Create hardware interfaces
    relay = create_relay(mode=args.mode, pin=args.relay_pin)
    loop_sensor = create_loop_sensor(mode=args.mode, pin=args.loop_pin)

    # Detect what mode we actually ended up in
    actual_mode = "real" if relay.__class__.__name__ == "PiRelay" else "mock"

    # Create safety manager
    safety = SafetyManager(
        relay=relay,
        loop_sensor=loop_sensor,
        auto_close_timeout=args.auto_close_timeout,
    )

    # Create FastAPI app
    app = create_app(relay, loop_sensor, safety, mode=actual_mode)

    # Start server
    import uvicorn
    logger.info("Gate Controller ready. API docs at http://%s:%d/docs", args.host, args.port)
    uvicorn.run(app, host=args.host, port=args.port)


if __name__ == "__main__":
    main()
