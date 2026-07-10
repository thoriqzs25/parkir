"""
Unit tests for safety interlocks.
"""

import sys
from pathlib import Path

# Add project src directories to path
_SHARED_SRC = Path(__file__).parent.parent.parent / "shared" / "src"
_GC_SRC = Path(__file__).parent.parent / "src"
if str(_SHARED_SRC) not in sys.path:
    sys.path.insert(0, str(_SHARED_SRC))
if str(_GC_SRC) not in sys.path:
    sys.path.insert(0, str(_GC_SRC))

import time
import threading

from gate_controller.interfaces import MockRelay, MockLoopSensor
from gate_controller.safety import SafetyManager
from shared.models import GateStateEnum


class TestLoopSensorOverride:
    """Interlock 1: Loop Sensor Override"""

    def test_can_close_when_no_vehicle(self):
        relay = MockRelay()
        sensor = MockLoopSensor(initial_state=False)
        safety = SafetyManager(relay, sensor, auto_close_timeout=999.0)
        assert safety.can_close() is True

    def test_cannot_close_when_vehicle_present(self):
        relay = MockRelay()
        sensor = MockLoopSensor(initial_state=True)
        safety = SafetyManager(relay, sensor, auto_close_timeout=999.0)
        assert safety.can_close() is False

    def test_close_returns_fault_when_vehicle_present(self):
        relay = MockRelay()
        sensor = MockLoopSensor(initial_state=True)
        safety = SafetyManager(relay, sensor, auto_close_timeout=999.0)
        state = safety.close_gate()
        assert state == GateStateEnum.FAULT

    def test_close_succeeds_when_vehicle_leaves(self):
        relay = MockRelay()
        sensor = MockLoopSensor(initial_state=True)
        safety = SafetyManager(relay, sensor, auto_close_timeout=999.0)
        sensor.set_vehicle_present(False)
        state = safety.close_gate()
        assert state == GateStateEnum.CLOSED


class TestTimeoutAutoClose:
    """Interlock 2: Timeout Auto-Close"""

    def test_auto_close_fires_after_timeout(self):
        relay = MockRelay()
        sensor = MockLoopSensor(initial_state=False)
        safety = SafetyManager(relay, sensor, auto_close_timeout=0.1)
        safety.open_gate()
        assert relay.get_state() == GateStateEnum.OPEN
        time.sleep(0.2)
        assert relay.get_state() == GateStateEnum.CLOSED

    def test_auto_close_cancelled_on_manual_close(self):
        relay = MockRelay()
        sensor = MockLoopSensor(initial_state=False)
        safety = SafetyManager(relay, sensor, auto_close_timeout=0.5)
        safety.open_gate()
        safety.close_gate()
        assert relay.get_state() == GateStateEnum.CLOSED
        time.sleep(0.6)
        assert relay.get_state() == GateStateEnum.CLOSED

    def test_auto_close_skipped_if_vehicle_present(self):
        relay = MockRelay()
        sensor = MockLoopSensor(initial_state=False)
        safety = SafetyManager(relay, sensor, auto_close_timeout=0.1)
        safety.open_gate()
        sensor.set_vehicle_present(True)
        time.sleep(0.2)
        assert relay.get_state() == GateStateEnum.OPEN

    def test_safety_state_shows_timer_active(self):
        relay = MockRelay()
        sensor = MockLoopSensor(initial_state=False)
        safety = SafetyManager(relay, sensor, auto_close_timeout=999.0)
        safety.open_gate()
        sstate = safety.get_safety_state()
        assert sstate.timeout_auto_close_active is True
        assert sstate.loop_sensor_override_active is False
        assert sstate.anti_crush_triggered is False


class TestSafetyState:
    def test_safety_state_reflects_loop_sensor(self):
        relay = MockRelay()
        sensor = MockLoopSensor(initial_state=True)
        safety = SafetyManager(relay, sensor)
        sstate = safety.get_safety_state()
        assert sstate.loop_sensor_override_active is True
