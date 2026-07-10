# Gate Controller Implementation Plan

## Goal
Implement a minimal working prototype of the Gate Controller for the MX80 barrier gate on Raspberry Pi 4, with mockable hardware interfaces and a testable HTTP API.

## Architecture Overview

The Gate Controller is a standalone HTTP server running on a Raspberry Pi 4 inside the booth LAN. It controls the barrier gate via a dry-contact relay module and reads a loop sensor for vehicle detection. It exposes a REST API for the Operator App to send commands and query status.

Hardware interfaces are abstracted so the code can be tested on any machine (including non-Pi) using mock implementations. The real GPIO implementations use `gpiozero` and are activated via an environment variable.

## Directory Structure

```
parkir/
├── gate-controller/          # Gate Controller firmware
│   ├── src/gate_controller/
│   │   ├── interfaces/       # Abstract hardware interfaces
│   │   │   ├── relay.py          # RelayInterface + Mock + Pi
│   │   │   └── loop_sensor.py    # LoopSensorInterface + Mock + Pi
│   │   ├── api/              # FastAPI HTTP server
│   │   │   └── server.py
│   │   ├── safety/           # Safety interlock logic
│   │   │   └── interlocks.py
│   │   └── main.py           # Entry point (argparse + uvicorn)
│   ├── tests/
│   │   ├── test_interfaces.py
│   │   ├── test_safety.py
│   │   └── test_api.py
│   ├── docs/
│   │   ├── wiring-guide.md     # Physical wiring when ready
│   │   └── next-steps.md       # Roadmap for future phases
│   ├── requirements.txt
│   ├── README.md
│   └── gate-controller.service # systemd service
│
├── operator-app/              # Mock client + future operator code
│   ├── src/operator_app/
│   │   ├── gate_client.py      # HTTP client to talk to Gate Controller
│   │   └── mock_operator.py    # CLI mock for testing
│   ├── tests/
│   └── requirements.txt
│
└── shared/                    # Shared Pydantic contracts
    ├── src/shared/
    │   └── models.py
    └── tests/
```

## Phase 1 — Shared Contracts

File: `shared/src/shared/models.py`

Pydantic models for all cross-component communication:
- `GateCommand`: `command` enum (open, close, status)
- `GateStatus`: `state` enum (OPEN, CLOSED, MOVING, FAULT) + `vehicle_present` + `uptime_seconds`
- `SafetyState`: `loop_sensor_override_active`, `timeout_auto_close_active`, `anti_crush_triggered`
- `HealthResponse`: `status` (ok/degraded), `interfaces` (mock/real), `version`

## Phase 2 — Hardware Interfaces (Mockable)

File: `gate-controller/src/gate_controller/interfaces/relay.py`

- `RelayInterface` (ABC): `open_gate()`, `close_gate()`, `get_state()` → returns `OPEN | CLOSED | MOVING | FAULT`
- `MockRelay`: in-memory state tracking, no GPIO
- `PiRelay`: `gpiozero.OutputDevice` on a configurable pin (default GPIO 17), drives the relay

File: `gate-controller/src/gate_controller/interfaces/loop_sensor.py`

- `LoopSensorInterface` (ABC): `is_vehicle_present()` → bool
- `MockLoopSensor`: in-memory boolean, can be toggled via `set_vehicle_present(bool)` for testing
- `PiLoopSensor`: `gpiozero.Button` on a configurable pin (default GPIO 18), active low

Configuration switch: `GATE_CONTROLLER_HARDWARE=mock` (default) or `real`. If `real` but `gpiozero` is not available (e.g., not on Pi), fall back to mock with a warning.

## Phase 3 — Safety Interlocks

File: `gate-controller/src/gate_controller/safety/interlocks.py`

The `SafetyManager` enforces the following rules:

1. **Loop Sensor Override**: If `loop_sensor.is_vehicle_present()` is true, `close_gate()` is blocked and returns a `GateStatus` with `FAULT` state.
2. **Timeout Auto-Close**: When `open_gate()` is called, start a `threading.Timer`. If the gate remains open for more than `AUTO_CLOSE_TIMEOUT` seconds (default 30) and the loop sensor is clear, automatically call `close_gate()`. The timer is cancelled if the gate is manually closed before the timeout.
3. **Failsafe Close**: Documented but not implemented in this phase. Requires the relay to be wired as Normally-Closed (NC) so that a power loss to the Pi causes the relay to de-energize and the gate closes. This is a hardware configuration, not software.
4. **Anti-Crush**: Documented but not implemented. Requires a secondary photocell or IR sensor wired to an additional GPIO pin. If triggered during close, the gate must reverse. Deferred to Phase 2.

## Phase 4 — FastAPI HTTP Server

File: `gate-controller/src/gate_controller/api/server.py`

Endpoints:

| Method | Path | Description |
|--------|------|-------------|
| POST | `/gate/open` | Open the gate. Starts auto-close timer. Returns `GateStatus`. |
| POST | `/gate/close` | Close the gate. Blocked if loop sensor detects vehicle. Returns `GateStatus`. |
| GET | `/gate/status` | Return current `GateStatus` + `SafetyState`. |
| GET | `/health` | Return `HealthResponse`. Shows mock vs real mode. |

The server holds a singleton `GateController` class that orchestrates the relay, loop sensor, and safety manager.

Auto-generated OpenAPI docs at `/docs` (Swagger UI) for browser-based testing.

## Phase 5 — Mock Operator Client

File: `operator-app/src/operator_app/gate_client.py`

- `GateClient` class: `open_gate()`, `close_gate()`, `get_status()`, `health()` using `httpx` or `requests`
- Base URL configurable (default `http://localhost:8000`)
- Returns typed Pydantic models from `shared/models.py`

File: `operator-app/src/operator_app/mock_operator.py`

- CLI using `argparse` or `click`:
  ```
  python mock_operator.py open       # send POST /gate/open
  python mock_operator.py close      # send POST /gate/close
  python mock_operator.py status     # send GET /gate/status
  python mock_operator.py health     # send GET /health
  ```
- Prints colored JSON output for readability

## Phase 6 — Tests & Deployment

### Tests

File: `gate-controller/tests/test_interfaces.py`
- Test `MockRelay` state transitions
- Test `MockLoopSensor` toggle
- Test Pi fallback to mock when `gpiozero` is unavailable

File: `gate-controller/tests/test_safety.py`
- Test loop sensor override blocks close
- Test timeout auto-close fires after configured delay
- Test timer cancellation on manual close

File: `gate-controller/tests/test_api.py`
- Use `httpx.TestClient` to spin up the FastAPI app
- Test `/gate/open` returns OPEN
- Test `/gate/close` returns CLOSED
- Test `/gate/close` returns FAULT when loop sensor is active
- Test `/gate/status` returns correct state
- Test `/health` returns mock mode

### Deployment

- `requirements.txt`: `fastapi`, `uvicorn[standard]`, `gpiozero`, `pydantic`, `httpx`, `pytest`, `pytest-asyncio`
- `gate-controller.service`: systemd unit to run `uvicorn main:app --host 0.0.0.0 --port 8000` on boot
- `README.md`: setup instructions, testing on Pi vs non-Pi, hardware wiring guide reference

## Phase 7 — Next Steps (Documented in `docs/next-steps.md`)

- **Phase 2**: Real GPIO wiring + relay module. Add `PiRelay` and `PiLoopSensor` wiring diagram. Test on actual Pi.
- **Phase 3**: Failsafe close (NC relay wiring) + anti-crush sensor. Add `AntiCrushSensorInterface` and wiring.
- **Phase 4**: WebSocket support (`/ws/gate`) for real-time status push to the Operator App.
- **Phase 5**: Whitelist / autonomous mode. Store a local list of plate/session states on the Gate Controller so it can open autonomously without the Operator App.
- **Phase 6**: Real Operator App integration. Electron app calling the Gate Controller API instead of the mock client.
- **Phase 7**: Production hardening. TLS/mTLS, structured logging, Prometheus metrics, health checks, config file (YAML).

## Revert Plan
- All new code is isolated in `gate-controller/`, `operator-app/`, and `shared/` directories.
- No existing spec files (e.g., `specs/18-system-architecture.md`) are modified.
- To revert: delete the three directories. The repository returns to spec-only state.
- To revert a bad commit: `git checkout <previous-branch>` or `git reset --hard HEAD~1`.

## Behavioral Parity Check
- The API endpoints (`POST /gate/open`, `POST /gate/close`, `GET /gate/status`) match Chapter 18.5.2 of the spec.
- Safety interlocks 1 and 2 (Loop Sensor Override, Timeout Auto-Close) match Chapter 18.5.3.
- Safety interlocks 3 and 4 (Failsafe Close, Anti-Crush) are documented as deferred per hardware readiness.
- The mock mode ensures identical behavior whether running on a Pi or a developer's laptop.

## Implementation Order
1. `shared/src/shared/models.py`
2. `gate-controller/src/gate_controller/interfaces/relay.py`
3. `gate-controller/src/gate_controller/interfaces/loop_sensor.py`
4. `gate-controller/src/gate_controller/safety/interlocks.py`
5. `gate-controller/src/gate_controller/api/server.py`
6. `gate-controller/src/gate_controller/main.py`
7. `operator-app/src/operator_app/gate_client.py`
8. `operator-app/src/operator_app/mock_operator.py`
9. All test files
10. `README.md`, `requirements.txt`, `gate-controller.service`
11. `docs/wiring-guide.md`, `docs/next-steps.md`
12. Commit and MR
