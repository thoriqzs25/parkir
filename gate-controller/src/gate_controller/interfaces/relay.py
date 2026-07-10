"""
Abstract relay interface and implementations for controlling the barrier gate.
Supports both mock (software-only) and real Raspberry Pi GPIO via gpiozero.
"""

from abc import ABC, abstractmethod
import time
import os
import logging

from shared.models import GateStateEnum

logger = logging.getLogger(__name__)


class RelayInterface(ABC):
    """
    Abstract base class for the gate relay/motor controller.
    Implementations must be thread-safe.
    """

    @abstractmethod
    def open_gate(self) -> GateStateEnum:
        """Energize the relay to raise the barrier. Return the resulting state."""
        pass

    @abstractmethod
    def close_gate(self) -> GateStateEnum:
        """De-energize the relay to lower the barrier. Return the resulting state."""
        pass

    @abstractmethod
    def get_state(self) -> GateStateEnum:
        """Return the current physical state of the gate."""
        pass

    @abstractmethod
    def get_uptime(self) -> float:
        """Return seconds since the gate controller started."""
        pass


class MockRelay(RelayInterface):
    """
    Software-only relay implementation for testing and development.
    Tracks state in memory. No GPIO required.
    """

    def __init__(self, auto_close_delay: float = 3.0) -> None:
        self._state = GateStateEnum.CLOSED
        self._start_time = time.time()
        self._auto_close_delay = auto_close_delay
        self._auto_close_timer = None
        self._moving_until = 0.0

    def open_gate(self) -> GateStateEnum:
        if self._state == GateStateEnum.OPEN:
            logger.info("MockRelay: already open, no-op")
            return self._state
        self._state = GateStateEnum.MOVING
        self._moving_until = time.time() + 1.0  # Simulate 1s movement
        # Simulate open completion
        self._state = GateStateEnum.OPEN
        logger.info("MockRelay: gate opened")
        return self._state

    def close_gate(self) -> GateStateEnum:
        if self._state == GateStateEnum.CLOSED:
            logger.info("MockRelay: already closed, no-op")
            return self._state
        self._state = GateStateEnum.MOVING
        self._moving_until = time.time() + 1.0
        self._state = GateStateEnum.CLOSED
        logger.info("MockRelay: gate closed")
        return self._state

    def get_state(self) -> GateStateEnum:
        # If currently in MOVING and time has passed, resolve to next state
        if self._state == GateStateEnum.MOVING and time.time() >= self._moving_until:
            # This is a simplification; in reality the caller would resolve it
            pass
        return self._state

    def get_uptime(self) -> float:
        return time.time() - self._start_time


class PiRelay(RelayInterface):
    """
    Raspberry Pi GPIO relay implementation using gpiozero.
    Requires gpiozero and RPi.GPIO to be installed.
    """

    def __init__(self, pin: int = 17) -> None:
        try:
            from gpiozero import OutputDevice
        except ImportError as e:
            raise ImportError(
                "gpiozero is required for PiRelay. Install it with: pip install gpiozero"
            ) from e

        self._relay = OutputDevice(pin, active_high=True, initial_value=False)
        self._start_time = time.time()
        self._state = GateStateEnum.CLOSED
        logger.info("PiRelay initialized on GPIO pin %d", pin)

    def open_gate(self) -> GateStateEnum:
        self._relay.on()
        self._state = GateStateEnum.OPEN
        logger.info("PiRelay: relay ON (gate opening)")
        return self._state

    def close_gate(self) -> GateStateEnum:
        self._relay.off()
        self._state = GateStateEnum.CLOSED
        logger.info("PiRelay: relay OFF (gate closing)")
        return self._state

    def get_state(self) -> GateStateEnum:
        # gpiozero doesn't directly tell us OPEN vs CLOSED, but we track it
        return self._state

    def get_uptime(self) -> float:
        return time.time() - self._start_time


def create_relay(mode: str = "auto", pin: int = 17) -> RelayInterface:
    """
    Factory function to create the appropriate relay implementation.

    Args:
        mode: "mock" | "real" | "auto"
            - mock: always returns MockRelay
            - real: always returns PiRelay (raises if gpiozero unavailable)
            - auto: tries PiRelay, falls back to MockRelay with warning
        pin: GPIO pin number for PiRelay (default 17)

    Returns:
        RelayInterface instance
    """
    if mode == "mock":
        logger.info("Using MockRelay")
        return MockRelay()

    if mode == "real":
        logger.info("Using PiRelay on pin %d", pin)
        return PiRelay(pin=pin)

    if mode == "auto":
        try:
            logger.info("Auto mode: attempting PiRelay on pin %d", pin)
            return PiRelay(pin=pin)
        except Exception as e:
            logger.warning("Auto mode: PiRelay failed (%s). Falling back to MockRelay.", e)
            return MockRelay()

    raise ValueError(f"Unknown relay mode: {mode}. Choose 'mock', 'real', or 'auto'.")
