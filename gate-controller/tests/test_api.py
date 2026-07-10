"""
Integration tests for the Gate Controller HTTP API.
Uses httpx.TestClient to hit the FastAPI app without starting a real server.
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

from fastapi.testclient import TestClient

from gate_controller.interfaces import MockRelay, MockLoopSensor
from gate_controller.safety import SafetyManager
from gate_controller.api import create_app
from shared.models import GateStateEnum


class TestGateAPI:
    def setup_method(self):
        self.relay = MockRelay()
        self.sensor = MockLoopSensor()
        self.safety = SafetyManager(self.relay, self.sensor, auto_close_timeout=999.0)
        self.app = create_app(self.relay, self.sensor, self.safety, mode="mock")
        self.client = TestClient(self.app)

    def test_open_gate(self):
        resp = self.client.post("/gate/open")
        assert resp.status_code == 200
        data = resp.json()
        assert data["state"] == GateStateEnum.OPEN.value
        assert data["vehicle_present"] is False

    def test_close_gate(self):
        self.client.post("/gate/open")
        resp = self.client.post("/gate/close")
        assert resp.status_code == 200
        data = resp.json()
        assert data["state"] == GateStateEnum.CLOSED.value

    def test_close_gate_blocked_by_vehicle(self):
        self.client.post("/gate/open")
        self.sensor.set_vehicle_present(True)
        resp = self.client.post("/gate/close")
        assert resp.status_code == 409
        assert "vehicle detected" in resp.json()["detail"].lower()

    def test_get_status(self):
        self.client.post("/gate/open")
        resp = self.client.get("/gate/status")
        assert resp.status_code == 200
        data = resp.json()
        assert data["gate"]["state"] == GateStateEnum.OPEN.value
        assert data["safety"]["loop_sensor_override_active"] is False

    def test_health_check(self):
        resp = self.client.get("/health")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "ok"
        assert data["interfaces"] == "mock"
        assert "uptime_seconds" in data["details"]

    def test_auto_close_via_api(self):
        self.relay = MockRelay()
        self.sensor = MockLoopSensor()
        self.safety = SafetyManager(self.relay, self.sensor, auto_close_timeout=0.1)
        self.app = create_app(self.relay, self.sensor, self.safety, mode="mock")
        self.client = TestClient(self.app)
        self.client.post("/gate/open")
        time.sleep(0.2)
        resp = self.client.get("/gate/status")
        assert resp.json()["gate"]["state"] == GateStateEnum.CLOSED.value
