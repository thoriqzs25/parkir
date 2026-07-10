"""
Shared Pydantic models for Gate Controller <-> Operator App communication.
These contracts are shared across the gate-controller and operator-app directories.
"""

from enum import Enum
from pydantic import BaseModel
from typing import Optional


class GateCommandEnum(str, Enum):
    """Valid commands that can be sent to the gate controller."""
    OPEN = "open"
    CLOSE = "close"
    STATUS = "status"


class GateStateEnum(str, Enum):
    """Physical state of the barrier gate."""
    OPEN = "OPEN"
    CLOSED = "CLOSED"
    MOVING = "MOVING"
    FAULT = "FAULT"


class GateCommand(BaseModel):
    """Request model for sending a command to the gate."""
    command: GateCommandEnum


class GateStatus(BaseModel):
    """Current state of the gate."""
    state: GateStateEnum
    vehicle_present: bool
    uptime_seconds: float
    message: Optional[str] = None


class SafetyState(BaseModel):
    """Active safety interlock states."""
    loop_sensor_override_active: bool
    timeout_auto_close_active: bool
    anti_crush_triggered: bool


class GateStatusResponse(BaseModel):
    """Full response from /gate/status including safety info."""
    gate: GateStatus
    safety: SafetyState


class HealthResponse(BaseModel):
    """Health check response from /health."""
    status: str  # "ok" or "degraded"
    interfaces: str  # "mock" or "real"
    version: str
    details: Optional[dict] = None
