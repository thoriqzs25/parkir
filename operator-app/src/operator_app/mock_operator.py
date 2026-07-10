"""
CLI mock operator for testing the Gate Controller API.

Usage:
    python mock_operator.py open       # Open the gate
    python mock_operator.py close      # Close the gate
    python mock_operator.py status     # Get gate status
    python mock_operator.py health     # Health check
    python mock_operator.py cycle      # Open, wait 5s, close
    python mock_operator.py --base-url http://192.168.1.50:8000 open
"""

import sys
from pathlib import Path

# Add shared package to path
_SHARED_SRC = Path(__file__).parent.parent.parent.parent / "shared" / "src"
if str(_SHARED_SRC) not in sys.path:
    sys.path.insert(0, str(_SHARED_SRC))

import argparse
import json
import time
import logging

from operator_app.gate_client import GateClient

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)


def _print_json(data) -> None:
    """Pretty-print a Pydantic model as JSON."""
    print(json.dumps(data.model_dump(), indent=2, default=str))


def run_cli() -> None:
    parser = argparse.ArgumentParser(description="Mock Operator — Gate Controller CLI")
    parser.add_argument(
        "action",
        choices=["open", "close", "status", "health", "cycle"],
        help="Action to perform",
    )
    parser.add_argument(
        "--base-url",
        default="http://localhost:8000",
        help="Gate Controller base URL",
    )
    args = parser.parse_args()

    client = GateClient(base_url=args.base_url)

    try:
        if args.action == "open":
            result = client.open_gate()
            _print_json(result)

        elif args.action == "close":
            result = client.close_gate()
            _print_json(result)

        elif args.action == "status":
            result = client.get_status()
            _print_json(result)

        elif args.action == "health":
            result = client.health()
            _print_json(result)

        elif args.action == "cycle":
            print("Opening gate...")
            _print_json(client.open_gate())
            print("Waiting 5 seconds...")
            time.sleep(5)
            print("Closing gate...")
            _print_json(client.close_gate())

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    finally:
        client.close()


if __name__ == "__main__":
    run_cli()
