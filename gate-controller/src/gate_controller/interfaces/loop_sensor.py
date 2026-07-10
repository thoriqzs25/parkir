"""
Abstract loop sensor interface and implementations.
Supports both mock (software toggle) and real Raspberry Pi GPIO via gpiozero.
"""

from abc import ABC, abstractmethod
import logging

logger = logging.getLogger(__name__)


class LoopSensorInterface(ABC):
    """
    Abstract base class for the inductive loop sensor.
    Detects vehicle presence at the gate.
    """

    @abstractmethod
    def is_vehicle_present(self) -> bool:
        """Return True if a vehicle is currently over the loop sensor."""
        pass


class MockLoopSensor(LoopSensorInterface):
    """
    Software-only loop sensor for testing.
    Vehicle presence can be toggled via set_vehicle_present().
    """

    def __init__(self, initial_state: bool = False) -> None:
        self._vehicle_present = initial_state

    def is_vehicle_present(self) -> bool:
        return self._vehicle_present

    def set_vehicle_present(self, present: bool) -> None:
        """Toggle the mock sensor state. Useful in tests."""
        self._vehicle_present = present
        logger.info("MockLoopSensor: vehicle_present = %s", present)


class PiLoopSensor(LoopSensorInterface):
    """
    Raspberry Pi GPIO loop sensor using gpiozero Button.
    Assumes active-low: vehicle present pulls the pin LOW.
    """

    def __init__(self, pin: int = 18, pull_up: bool = True) -> None:
        try:
            from gpiozero import Button
        except ImportError as e:
            raise ImportError(
                "gpiozero is required for PiLoopSensor. Install it with: pip install gpiozero"
            ) from e

        # When pull_up=True, the pin is pulled up via internal resistor.
        # A vehicle present would pull it LOW (active low).
        self._button = Button(pin, pull_up=pull_up)
        logger.info("PiLoopSensor initialized on GPIO pin %d (pull_up=%s)", pin, pull_up)

    def is_vehicle_present(self) -> bool:
        # is_pressed returns True when the pin is pulled to the active state
        return self._button.is_pressed


def create_loop_sensor(mode: str = "auto", pin: int = 18) -> LoopSensorInterface:
    """
    Factory function to create the appropriate loop sensor implementation.

    Args:
        mode: "mock" | "real" | "auto"
            - mock: always returns MockLoopSensor
            - real: always returns PiLoopSensor (raises if gpiozero unavailable)
            - auto: tries PiLoopSensor, falls back to MockLoopSensor with warning
        pin: GPIO pin number for PiLoopSensor (default 18)

    Returns:
        LoopSensorInterface instance
    """
    if mode == "mock":
        logger.info("Using MockLoopSensor")
        return MockLoopSensor()

    if mode == "real":
        logger.info("Using PiLoopSensor on pin %d", pin)
        return PiLoopSensor(pin=pin)

    if mode == "auto":
        try:
            logger.info("Auto mode: attempting PiLoopSensor on pin %d", pin)
            return PiLoopSensor(pin=pin)
        except Exception as e:
            logger.warning("Auto mode: PiLoopSensor failed (%s). Falling back to MockLoopSensor.", e)
            return MockLoopSensor()

    raise ValueError(f"Unknown loop sensor mode: {mode}. Choose 'mock', 'real', or 'auto'.")
