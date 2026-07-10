"""
FastAPI HTTP server for the Gate Controller.
Exposes endpoints for open, close, status, and health.
Auto-generated OpenAPI docs at /docs.
"""

import time
import logging
from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse

from shared.models import (
    GateCommand,
    GateStatus,
    SafetyState,
    GateStatusResponse,
    HealthResponse,
    GateStateEnum,
)

logger = logging.getLogger(__name__)


class GateController:
    """
    Singleton gate controller that orchestrates relay, loop sensor, and safety manager.
    """

    def __init__(self, relay, loop_sensor, safety_manager):
        self._relay = relay
        self._loop_sensor = loop_sensor
        self._safety = safety_manager
        self._version = "1.0.0"
        self._start_time = time.time()

    def open(self) -> GateStatus:
        state = self._safety.open_gate()
        return self._build_status(state)

    def close(self) -> GateStatus:
        state = self._safety.close_gate()
        return self._build_status(state)

    def status(self) -> GateStatusResponse:
        state = self._safety.get_state()
        safety = self._safety.get_safety_state()
        return GateStatusResponse(
            gate=self._build_status(state),
            safety=safety,
        )

    def health(self, mode: str) -> HealthResponse:
        return HealthResponse(
            status="ok",
            interfaces=mode,
            version=self._version,
            details={
                "uptime_seconds": round(time.time() - self._start_time, 2),
                "relay_state": self._relay.get_state().value,
                "loop_sensor": self._loop_sensor.is_vehicle_present(),
            },
        )

    def _build_status(self, state: GateStateEnum) -> GateStatus:
        return GateStatus(
            state=state,
            vehicle_present=self._loop_sensor.is_vehicle_present(),
            uptime_seconds=round(self._relay.get_uptime(), 2),
            message=None if state != GateStateEnum.FAULT else "Safety interlock blocked the operation",
        )


def create_app(relay, loop_sensor, safety_manager, mode: str = "mock") -> FastAPI:
    """
    Factory function to create the FastAPI app with injected dependencies.
    """
    controller = GateController(relay, loop_sensor, safety_manager)

    app = FastAPI(
        title="PAS Gate Controller",
        description="HTTP API for the MX80 barrier gate controller.",
        version="1.0.0",
    )

    @app.post("/gate/open", response_model=GateStatus)
    async def gate_open():
        """Open the gate. Starts the auto-close timer."""
        logger.info("API: /gate/open called")
        return controller.open()

    @app.post("/gate/close", response_model=GateStatus)
    async def gate_close():
        """Close the gate. Blocked if the loop sensor detects a vehicle."""
        logger.info("API: /gate/close called")
        result = controller.close()
        if result.state == GateStateEnum.FAULT:
            raise HTTPException(
                status_code=409,
                detail="Close command blocked: vehicle detected on loop sensor",
            )
        return result

    @app.get("/gate/status", response_model=GateStatusResponse)
    async def gate_status():
        """Get the current gate status and active safety interlocks."""
        return controller.status()

    @app.get("/health", response_model=HealthResponse)
    async def health():
        """Health check. Shows mock vs real mode."""
        return controller.health(mode)

    return app
