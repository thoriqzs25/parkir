"""
HTTP client for the Operator App to communicate with the Gate Controller.
Uses httpx (or requests) and returns typed Pydantic models.
"""

import sys
from pathlib import Path

# Add shared package to path
_SHARED_SRC = Path(__file__).parent.parent.parent.parent / "shared" / "src"
if str(_SHARED_SRC) not in sys.path:
    sys.path.insert(0, str(_SHARED_SRC))

import json
import logging
from typing import Optional

try:
    import httpx
except ImportError:
    raise ImportError("httpx is required. Install it with: pip install httpx")

from shared.models import GateStatus, GateStatusResponse, HealthResponse

logger = logging.getLogger(__name__)


class GateClient:
    """
    Client for the Gate Controller HTTP API.
    """

    def __init__(self, base_url: str = "http://localhost:8000") -> None:
        self._base_url = base_url.rstrip("/")
        self._client = httpx.Client(base_url=self._base_url, timeout=10.0)

    def open_gate(self) -> GateStatus:
        """Send POST /gate/open and return the resulting status."""
        logger.info("GateClient: opening gate")
        resp = self._client.post("/gate/open")
        resp.raise_for_status()
        return GateStatus(**resp.json())

    def close_gate(self) -> GateStatus:
        """Send POST /gate/close and return the resulting status."""
        logger.info("GateClient: closing gate")
        resp = self._client.post("/gate/close")
        resp.raise_for_status()
        return GateStatus(**resp.json())

    def get_status(self) -> GateStatusResponse:
        """Send GET /gate/status and return the full status."""
        logger.info("GateClient: fetching status")
        resp = self._client.get("/gate/status")
        resp.raise_for_status()
        return GateStatusResponse(**resp.json())

    def health(self) -> HealthResponse:
        """Send GET /health and return the health response."""
        logger.info("GateClient: health check")
        resp = self._client.get("/health")
        resp.raise_for_status()
        return HealthResponse(**resp.json())

    def close(self) -> None:
        """Close the underlying HTTP client."""
        self._client.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()
