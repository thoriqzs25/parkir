"""
Unit tests for hardware interfaces.
Runs with mocked interfaces — no GPIO or Pi required.
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
import pytest

from gate_controller.interfaces import MockRelay, MockLoopSensor, create_relay, create_loop_sensor
from shared.models import GateStateEnum


class TestMockRelay:
    def test_initial_state_is_closed(self):
        relay = MockRelay()
        assert relay.get_state() == GateStateEnum.CLOSED

    def test_open_changes_state_to_open(self):
        relay = MockRelay()
        state = relay.open_gate()
        assert state == GateStateEnum.OPEN
        assert relay.get_state() == GateStateEnum.OPEN

    def test_close_changes_state_to_closed(self):
        relay = MockRelay()
        relay.open_gate()
        state = relay.close_gate()
        assert state == GateStateEnum.CLOSED
        assert relay.get_state() == GateStateEnum.CLOSED

    def test_open_when_already_open_is_noop(self):
        relay = MockRelay()
        relay.open_gate()
        state = relay.open_gate()
        assert state == GateStateEnum.OPEN

    def test_close_when_already_closed_is_noop(self):
        relay = MockRelay()
        state = relay.close_gate()
        assert state == GateStateEnum.CLOSED

    def test_uptime_increases(self):
        relay = MockRelay()
        t1 = relay.get_uptime()
        time.sleep(0.01)
        t2 = relay.get_uptime()
        assert t2 > t1


class TestMockLoopSensor:
    def test_initial_state_false(self):
        sensor = MockLoopSensor()
        assert sensor.is_vehicle_present() is False

    def test_initial_state_true(self):
        sensor = MockLoopSensor(initial_state=True)
        assert sensor.is_vehicle_present() is True

    def test_toggle(self):
        sensor = MockLoopSensor()
        sensor.set_vehicle_present(True)
        assert sensor.is_vehicle_present() is True
        sensor.set_vehicle_present(False)
        assert sensor.is_vehicle_present() is False


class TestCreateRelay:
    def test_create_mock(self):
        relay = create_relay(mode="mock")
        assert isinstance(relay, MockRelay)

    def test_create_auto_fallback(self):
        # On non-Pi hardware, auto should fall back to mock
        relay = create_relay(mode="auto")
        assert isinstance(relay, MockRelay)

    def test_create_real_raises_on_non_pi(self):
        # On non-Pi hardware, "real" should raise ImportError
        with pytest.raises(ImportError):
            create_relay(mode="real")


class TestCreateLoopSensor:
    def test_create_mock(self):
        sensor = create_loop_sensor(mode="mock")
        assert isinstance(sensor, MockLoopSensor)

    def test_create_auto_fallback(self):
        sensor = create_loop_sensor(mode="auto")
        assert isinstance(sensor, MockLoopSensor)

    def test_create_real_raises_on_non_pi(self):
        with pytest.raises(ImportError):
            create_loop_sensor(mode="real")
