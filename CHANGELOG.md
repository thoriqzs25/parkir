# Changelog

## 2026-07-10

### Added
- **Gate Display App** (`gate-display/`) — standalone Electron app for parking entrance displays
  - State machine: IDLE → VEHICLE_DETECTED → TICKET_PRESSED → TICKET_READY → GATE_OPENING → GATE_OPEN → VEHICLE_EXITED
  - Display components: Header, WelcomeSign, CameraFeed, LoopIndicator, TicketButton, InstructionText, RatesTable, GateBarrier, DebugPanel
  - Mock controller with hardware abstraction (`ControllerInterface`)
  - Registration flow: mDNS announcement + HTTP server on port 9800
  - Persistent device identity
  - 18 unit tests on the state machine
- **Backend gates infrastructure** — new `gates` DB table + CRUD API + public `/gate/:id/info` endpoint
  - Permissions: `gates:view`, `gates:register`, `gates:edit`, `gates:delete`
  - 9 integration tests
- **Desktop Gate Setup** (`desktop/src/renderer/screens/GateSetup.tsx`) — IP-based gate registration flow
- **Dashboard Gates page** — CRUD table for managing gates per location
- **Plans** (`plans/`) — detailed milestone plans for the gate display system

### Changed
- **Fee calculation** — switched from single-block cap to recurring 24-hour block model
  - Each 24h block: `min(initial_fee + (block_hours-1) * per_hour_fee, daily_fixed_fee)`
  - Loops for subsequent 24-hour periods
- **Vehicle types** — now dynamic via `vehicle_types` DB table instead of hardcoded CHECK constraints
  - CRUD API + dashboard management page
  - Removed `oneof=CAR MOTO TRUCK` validation from sessions, rates, and sync handlers
  - Permissions: `vehicle-types:view`, `vehicle-types:create`, `vehicle-types:edit`, `vehicle-types:delete`

### Fixed
- Registration screen now shows device ID and IP address
- TID: placeholder registration screen replaced with full implementation
