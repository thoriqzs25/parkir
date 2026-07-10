"""
Safety interlock logic for the Gate Controller.

Interlocks implemented in this phase:
1. Loop Sensor Override: block close if vehicle is present.
2. Timeout Auto-Close: if gate open > N seconds and loop clear, auto-close.

Interlocks deferred to Phase 2:
3. Failsafe Close: hardware-level (NC relay wiring).
4. Anti-Crush: requires secondary sensor (photocell / IR).
"""

import threading
import time
import logging
from typing import Optional, Callable

from shared.models import GateStateEnum, SafetyState

logger = logging.getLogger(__name__)


class SafetyManager:
    """
    Manages safety interlocks for the barrier gate.
    Must be instantiated with a relay and loop_sensor interface.
    """

    def __init__(
        self,
        relay,
        loop_sensor,
        auto_close_timeout: float = 30.0,
    ) -> None:
        self._relay = relay
        self._loop_sensor = loop_sensor
        self._auto_close_timeout = auto_close_timeout
        self._auto_close_timer: Optional[threading.Timer] = None
        self._lock = threading.Lock()

    def can_open(self) -> bool:
        """Determine if the gate is allowed to open."""
        # In Phase 1, there are no restrictions on opening.
        # Future: could add whitelist checks, maintenance mode, etc.
        return True

    def can_close(self) -> bool:
        """
        Determine if the gate is allowed to close.
        Interlock 1: Loop Sensor Override.
        """
        if self._loop_sensor.is_vehicle_present():
            logger.warning("SafetyManager: close blocked — vehicle still on loop sensor")
            return False
        return True

    def open_gate(self) -> GateStateEnum:
        """
        Open the gate and start the auto-close timer.
        Returns the resulting state.
        """
        with self._lock:
            if not self.can_open():
                logger.warning("SafetyManager: open blocked by safety check")
                return self._relay.get_state()

            state = self._relay.open_gate()
            logger.info("SafetyManager: gate opened, starting auto-close timer (%ss)", self._auto_close_timeout)
            self._start_auto_close_timer()
            return state

    def close_gate(self) -> GateStateEnum:
        """
        Close the gate if allowed.
        Returns the resulting state. If blocked, returns FAULT.
        """
        with self._lock:
            if not self.can_close():
                logger.warning("SafetyManager: close blocked — vehicle on loop sensor")
                return GateStateEnum.FAULT

            self._cancel_auto_close_timer()
            state = self._relay.close_gate()
            logger.info("SafetyManager: gate closed")
            return state

    def get_state(self) -> GateStateEnum:
        """Return the current gate state."""
        return self._relay.get_state()

    def get_safety_state(self) -> SafetyState:
        """Return the current active safety states."""
        return SafetyState(
            loop_sensor_override_active=self._loop_sensor.is_vehicle_present(),
            timeout_auto_close_active=self._auto_close_timer is not None and self._auto_close_timer.is_alive(),
            anti_crush_triggered=False,  # Deferred to Phase 2
        )

    def _start_auto_close_timer(self) -> None:
        """Interlock 2: Start the timeout auto-close timer."""
        self._cancel_auto_close_timer()
        self._auto_close_timer = threading.Timer(
            self._auto_close_timeout,
            self._auto_close_handler,
        )
        self._auto_close_timer.daemon = True
        self._auto_close_timer.start()

    def _cancel_auto_close_timer(self) -> None:
        """Cancel any pending auto-close timer."""
        if self._auto_close_timer is not None:
            self._auto_close_timer.cancel()
            self._auto_close_timer = None

    def _auto_close_handler(self) -> None:
        """
        Interlock 2: Timeout Auto-Close.
        Called by the timer. If the gate is still open and loop is clear, close it.
        """
        logger.info("SafetyManager: auto-close timer fired")
        if self._relay.get_state() == GateStateEnum.OPEN and not self._loop_sensor.is_vehicle_present():
            logger.info("SafetyManager: auto-closing gate (timeout)")
            self._relay.close_gate()
        else:
            logger.info(
                "SafetyManager: auto-close skipped — state=%s, vehicle_present=%s",
                self._relay.get_state(),
                self._loop_sensor.is_vehicle_present(),
            )
