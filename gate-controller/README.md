# PAS Gate Controller

Minimal working prototype of the Gate Controller for the MX80 barrier gate on Raspberry Pi 4.

## Features

- **HTTP API** for open, close, status, and health (FastAPI)
- **Mockable hardware interfaces** — test on any machine without a Pi
- **Safety interlocks** (Phase 1):
  - Loop Sensor Override: blocks close if vehicle is present
  - Timeout Auto-Close: closes gate after configurable timeout if loop is clear
- **Auto-generated API docs** at `/docs` for browser-based testing
- **Mock Operator CLI** for manual testing

## Hardware Interfaces

| Interface | Mock | Real (Pi GPIO) |
|-----------|------|----------------|
| Relay (gate motor) | `MockRelay` | `PiRelay` via `gpiozero.OutputDevice` (pin 17) |
| Loop Sensor | `MockLoopSensor` | `PiLoopSensor` via `gpiozero.Button` (pin 18) |

Mode is selected via `GATE_CONTROLLER_MODE` (mock / real / auto). Auto tries real, falls back to mock with a warning.

## Setup

### Install dependencies

```bash
cd gate-controller
pip install -r requirements.txt
```

### Run in mock mode (any machine)

```bash
python src/gate_controller/main.py --mode mock
# API docs: http://localhost:8000/docs
```

### Run on Raspberry Pi with real GPIO

```bash
# Install gpiozero (usually pre-installed on Raspberry Pi OS)
pip install gpiozero

# Run with real hardware
python src/gate_controller/main.py --mode real --relay-pin 17 --loop-pin 18
```

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GATE_CONTROLLER_MODE` | `auto` | `mock`, `real`, or `auto` |
| `GATE_CONTROLLER_PORT` | `8000` | HTTP server port |
| `GATE_CONTROLLER_HOST` | `0.0.0.0` | HTTP server host |
| `GATE_CONTROLLER_RELAY_PIN` | `17` | GPIO pin for relay |
| `GATE_CONTROLLER_LOOP_PIN` | `18` | GPIO pin for loop sensor |
| `GATE_CONTROLLER_AUTO_CLOSE_TIMEOUT` | `30.0` | Auto-close timeout in seconds |

### Run as a systemd service

```bash
sudo cp gate-controller.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable gate-controller
sudo systemctl start gate-controller
```

## Testing

```bash
cd gate-controller
pytest tests/
```

All tests use mock interfaces — no Pi or GPIO required.

## Mock Operator CLI

From the `operator-app` directory:

```bash
python src/operator_app/mock_operator.py open
python src/operator_app/mock_operator.py close
python src/operator_app/mock_operator.py status
python src/operator_app/mock_operator.py health
python src/operator_app/mock_operator.py cycle
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/gate/open` | Open the gate |
| POST | `/gate/close` | Close the gate (409 if vehicle present) |
| GET | `/gate/status` | Gate status + safety interlock states |
| GET | `/health` | Health check (shows mock vs real mode) |
| GET | `/docs` | Swagger UI (auto-generated) |

## Next Steps

See `docs/next-steps.md` for the full roadmap.

## License

Same as the parent repo.
