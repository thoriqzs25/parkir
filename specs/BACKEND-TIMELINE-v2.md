# PARKIR v2 — Backend Timeline & Implementation
**Created:** 2026-08-06

---

## Backend Components

### 1. Cloud Backend (Go/Gin)
- Central API server
- PostgreSQL database
- Authentication (JWT RS256)
- Multi-location management
- Config versioning
- Sync aggregation

### 2. Server Room App (Go or Rust)
- Local API server
- SQLite database
- Config sync from cloud
- Data sync to cloud
- Gate management (mDNS, HTTP)
- Fee calculation engine
- Shift management

### 3. Gate App (Go or Rust)
- HTTP server (receive commands)
- mDNS announcement
- Hardware abstraction layer (HAL)
- Communication with server room app

---

## Week 1-2: Cloud Backend Foundation

### Database Design (3 days)
**Deliverables:**
- [ ] Schema design (locations, gates, configs, sessions, transactions, users, roles)
- [ ] Migration files (PostgreSQL)
- [ ] Index strategy (performance optimization)
- [ ] Foreign key relationships
- [ ] Config versioning schema

**Schema Overview:**
```sql
-- Core entities
locations (id, name, code, city, address, capacity, status)
gates (id, location_id, gate_id, gate_type, vehicle_type, status, config, last_seen_at)
users (id, email, password_hash, role_id, status)
roles (id, name, permissions)

-- Configs (versioned)
rate_configs (id, location_id, vehicle_type, effective_date, first_hour_rate, subsequent_hourly_rate, daily_flat_rate, version)
shift_configs (id, location_id, shift_code, shift_number, start_time, end_time, is_overnight, version)

-- Operational data
sessions (id, location_id, gate_id, vehicle_type, check_in_at, check_out_at, fee_amount, rate_config_version, shift_config_version, shift_number, state, qr_data, synced_at)
transactions (id, session_id, location_id, gate_id, amount, payment_method, payment_reference, config_version, shift_number, created_at, synced_at, voided, voided_at, voided_by, void_reason)

-- Sync tracking
sync_queue (id, entity_type, entity_id, action, payload, retry_count, last_attempt_at, status)
```

### API Structure (2 days)
**Deliverables:**
- [ ] Project structure (cmd/api, internal/domain, internal/store, internal/middleware)
- [ ] Router setup (Gin)
- [ ] Middleware (CORS, logging, recovery, request ID)
- [ ] Error handling (standard error responses)
- [ ] Response envelope ({data, error, meta})

**Project Structure:**
```
backend/
├── cmd/api/
│   └── main.go
├── internal/
│   ├── domain/
│   │   ├── auth/
│   │   ├── locations/
│   │   ├── gates/
│   │   ├── configs/
│   │   ├── sessions/
│   │   ├── transactions/
│   │   ├── sync/
│   │   └── health/
│   ├── store/
│   │   ├── postgres/
│   │   └── store.go (interface)
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logging.go
│   └── db/
│       ├── pool.go
│       └── migrate.go
└── migrations/
```

### Authentication (2 days)
**Deliverables:**
- [ ] JWT RS256 implementation (generate keys, sign tokens, verify tokens)
- [ ] Login endpoint (POST /api/v1/auth/login)
- [ ] Logout endpoint (POST /api/v1/auth/logout)
- [ ] Refresh token endpoint (POST /api/v1/auth/refresh)
- [ ] Get current user endpoint (GET /api/v1/auth/me)
- [ ] Auth middleware (extract JWT, validate, inject user context)
- [ ] Superadmin bootstrap script (create initial superadmin user)

**Auth Flow:**
```
POST /api/v1/auth/login
  → Validate email + password
  → Generate JWT (access_token, 8h expiry)
  → Set httpOnly cookie (access_token)
  → Return user + permissions

GET /api/v1/auth/me
  → Extract JWT from cookie
  → Validate token
  → Return user + permissions
```

### Location Management API (1 day)
**Deliverables:**
- [ ] GET /api/v1/locations (list all locations, filter by city)
- [ ] GET /api/v1/locations/:id (get location by ID)
- [ ] POST /api/v1/locations (create location)
- [ ] PATCH /api/v1/locations/:id (update location)
- [ ] POST /api/v1/locations/:id/deactivate (deactivate location)

### Config Management API (2 days)
**Deliverables:**
- [ ] Rate config endpoints:
  - GET /api/v1/rates (list rates, filter by location_id)
  - POST /api/v1/rates (create rate, auto-increment version)
  - PATCH /api/v1/rates/:id (update rate, auto-increment version)
- [ ] Shift config endpoints:
  - GET /api/v1/locations/:id/shift-configs (list shift configs)
  - POST /api/v1/locations/:id/shift-configs (create shift config)
  - PATCH /api/v1/locations/:id/shift-configs/:code (update shift config)
- [ ] Config versioning logic (auto-increment version on create/update)
- [ ] Overlap prevention (PostgreSQL triggers for rate/shift date ranges)

### Gate Management API (1 day)
**Deliverables:**
- [ ] POST /api/v1/gates/register (register new gate, status: unregistered)
- [ ] PATCH /api/v1/gates/:id (update gate config, status: registered)
- [ ] GET /api/v1/gates (list all gates, filter by location_id, status)
- [ ] GET /api/v1/gates/:id (get gate by ID)
- [ ] Gate status tracking (last_seen_at, health status)

### Sync API (1 day)
**Deliverables:**
- [ ] POST /api/v1/sync/batch (receive batch of sessions + transactions from server room apps)
- [ ] Duplicate detection (check entity_id, skip if exists)
- [ ] Sync status tracking (mark as synced)
- [ ] Response: list of successfully synced entity_ids

### Health Check (0.5 day)
**Deliverables:**
- [ ] GET /health (basic status)
- [ ] GET /health/ready (DB connectivity check)

---

## Week 3-4: Server Room App Foundation

### Local Database Design (2 days)
**Deliverables:**
- [ ] Schema design (SQLite)
- [ ] Migration system
- [ ] Tables: gates, rate_configs, shift_configs, sessions, transactions, sync_queue

**Schema Overview:**
```sql
-- Same structure as cloud, but local
-- Plus sync tracking
sync_queue (
  id, entity_type, entity_id, action, payload, 
  retry_count, last_attempt_at, status, error_message
)
```

### Cloud Sync Client (3 days)
**Deliverables:**
- [ ] HTTP client (poll cloud for configs, push data to cloud)
- [ ] Config polling (every 1 min, fetch latest configs)
- [ ] Config version tracking (compare local vs cloud version)
- [ ] Data sync (push sessions, transactions to cloud)
- [ ] Sync queue management (queue unsynced data, retry on failure)
- [ ] Exponential backoff retry (1s, 2s, 4s, 8s, max 30s)
- [ ] Sync failure alerting (log error, increment counter)

**Sync Flow:**
```
Every 1 minute:
  1. Poll cloud for configs (GET /api/v1/configs?version={local_version})
  2. If cloud has newer version → update local configs
  3. Push unsynced data (POST /api/v1/sync/batch)
  4. Mark synced data as synced
  5. Retry failed data with exponential backoff
```

### mDNS Discovery (2 days)
**Deliverables:**
- [ ] mDNS announcement (server room app announces itself)
- [ ] mDNS discovery (discover gate apps on LAN)
- [ ] Gate registration (add discovered gates to local DB, status: unregistered)
- [ ] Gate re-discovery (handle gate restarts, re-register)

**mDNS Flow:**
```
Server room app starts:
  1. Announce via mDNS: _parking-server._tcp.local
  2. Start discovering: _parking-gate._tcp.local
  3. When gate discovered → add to local DB (status: unregistered)
  4. When gate sends /register → update status to registered
```

### HTTP Server (2 days)
**Deliverables:**
- [ ] HTTP server (listen on port 8080 or configurable)
- [ ] Endpoints for gate apps:
  - POST /gate/register (gate app registers with server room app)
  - POST /gate/session/create (gate app creates session)
  - POST /gate/session/:id/close (gate app closes session)
  - POST /gate/alert (gate app sends alert)
  - GET /gate/health (health check)
- [ ] Gate registry (track connected gates, last_seen_at)
- [ ] Health check loop (ping each gate every 15s)

### Status Reporting (1 day)
**Deliverables:**
- [ ] Report status to cloud (every 20s)
- [ ] Include: gate statuses (online/offline), server room app health
- [ ] Endpoint: POST /api/v1/server-room/status (cloud receives status)

---

## Week 3-4: Gate App Foundation

### HTTP Server (1 day)
**Deliverables:**
- [ ] HTTP server (listen on port 8080)
- [ ] Endpoints:
  - GET /info (return gate info: gate_id, gate_type, vehicle_type)
  - POST /command (receive command from server room app: open_gate, print_ticket, etc.)
  - GET /health (health check)

### mDNS Announcement (1 day)
**Deliverables:**
- [ ] mDNS announcement (parking-gate-{gate_id}.local:8080)
- [ ] Auto-detect gate_id from hardware serial/MAC
- [ ] Handle mDNS conflicts (if gate_id already exists, append suffix)

### Hardware Abstraction Layer (HAL) (3 days)
**Deliverables:**
- [ ] HAL interfaces (printer, gate_motor, scanner, loop_sensor, button)
- [ ] Mock implementations (for testing without hardware)
- [ ] Real implementations (platform-specific):
  - Epson thermal printer (ESC/POS commands via USB/serial)
  - Gate motor (via hub, protocol TBD)
  - QR scanner (USB HID, read keyboard input)
  - Vehicle loop sensor (via hub)
  - Ticket button (via hub)
  - Alert button (via hub)

**HAL Interfaces:**
```go
type Printer interface {
  PrintTicket(data TicketData) error
  PrintReceipt(data ReceiptData) error
}

type GateMotor interface {
  Open() error
  Close() error
  IsOpen() bool
}

type Scanner interface {
  ReadQR() (string, error)
}

type LoopSensor interface {
  IsVehiclePresent() bool
}

type Button interface {
  WaitForPress() (ButtonEvent, error)
}
```

### Communication with Server Room App (1 day)
**Deliverables:**
- [ ] HTTP client (send requests to server room app)
- [ ] Endpoints called:
  - POST /gate/session/create (create session on entry)
  - POST /gate/session/:id/close (close session on exit)
  - POST /gate/alert (send alert)
- [ ] Error handling (retry on failure)

### Config Storage (0.5 day)
**Deliverables:**
- [ ] Minimal config storage (JSON file)
- [ ] Fields: gate_id, server_room_url, gate_type, vehicle_type
- [ ] Load on startup, save on config update

---

## Week 3: Entry Flow Backend

### Server Room App: Session Creation (2 days)
**Deliverables:**
- [ ] POST /gate/session/create handler
- [ ] Session creation logic:
  - Generate session_id (UUID)
  - Determine current shift (from shift_configs)
  - Assign shift_number (increment daily)
  - Save to local DB (status: ACTIVE)
  - Add to sync_queue (for cloud sync)
- [ ] Return: session_id, location_id, timestamp, vehicle_type, shift_number

### Cloud Backend: Session Sync (1 day)
**Deliverables:**
- [ ] Receive sessions from server room app sync
- [ ] Save to cloud DB
- [ ] Duplicate detection (skip if session_id exists)

### Gate App: Entry Sequence (2 days)
**Deliverables:**
- [ ] Entry sequence logic:
  1. Wait for vehicle in loop (loop_sensor.IsVehiclePresent())
  2. Wait for button press (button.WaitForPress())
  3. If vehicle in loop → call server room app: POST /gate/session/create
  4. Receive session data
  5. Generate QR code (encode: session_id, location_id, timestamp, vehicle_type)
  6. Print ticket (printer.PrintTicket)
  7. Open gate (gate_motor.Open())
  8. Wait for vehicle to enter (loop_sensor.IsVehiclePresent() == false)
  9. Close gate (gate_motor.Close())
- [ ] Error handling (printer jam, gate failure → stay closed, send alert)

---

## Week 4: Exit Flow Backend

### Server Room App: Fee Calculation (3 days)
**Deliverables:**
- [ ] Fee calculation engine:
  - Lookup rate config (by location_id, vehicle_type, check_in_date)
  - Calculate duration (check_out - check_in, in hours)
  - Calculate fee:
    - For each 24-hour block:
      - fee = min(first_hour + (hours-1) * subsequent, daily_flat_rate)
    - Sum fees for all blocks
  - Handle edge cases (duration < 1 hour → 1 hour)
- [ ] Session lookup (find session by session_id from QR)
- [ ] Session update (set check_out_at, fee_amount, shift_number, state: PENDING_PAYMENT)
- [ ] Return: fee_amount, check_in_time, duration, vehicle_type, shift_number

### Server Room App: Transaction Creation (1 day)
**Deliverables:**
- [ ] POST /gate/session/:id/close handler
- [ ] Transaction creation:
  - Record payment (amount, method, reference)
  - Mark session as CLOSED
  - Add to sync_queue
- [ ] Return: transaction_id

### Cloud Backend: Transaction Sync (1 day)
**Deliverables:**
- [ ] Receive transactions from server room app sync
- [ ] Save to cloud DB
- [ ] Duplicate detection

### Gate App: Exit Sequence (2 days)
**Deliverables:**
- [ ] Exit sequence logic:
  1. Wait for QR scan (scanner.ReadQR())
  2. Parse QR data (extract session_id)
  3. Call server room app: POST /gate/session/create (get fee)
  4. Display fee on monitor (driver-facing UI)
  5. Wait for payment (payment_terminal.WaitForPayment())
  6. If payment success → call server room app: POST /gate/session/:id/close
  7. Open gate
  8. Wait for vehicle to exit
  9. Close gate
  10. Optional: wait for receipt button press → print receipt
- [ ] Error handling (QR unreadable, payment failed → stay closed, send alert)

---

## Week 5-6: Dashboard Backend APIs

### Cloud Backend: Dashboard APIs (5 days)
**Deliverables:**
- [ ] Gate status API:
  - GET /api/v1/gates (list gates with status)
  - GET /api/v1/gates/unregistered (list unregistered gates)
- [ ] Session API:
  - GET /api/v1/sessions (list sessions, filters: location_id, state, date_range)
  - GET /api/v1/sessions/:id (get session by ID)
- [ ] Transaction API:
  - GET /api/v1/transactions (list transactions, filters)
  - GET /api/v1/transactions/:id (get transaction by ID)
- [ ] Reports API:
  - GET /api/v1/reports/daily-revenue (aggregate revenue by day)
  - GET /api/v1/reports/occupancy (occupancy by time bucket)
  - GET /api/v1/reports/vehicle-breakdown (revenue by vehicle type)
  - GET /api/v1/reports/export (CSV/Excel export)
- [ ] User management API:
  - GET /api/v1/users (list users)
  - POST /api/v1/users (create user)
  - PATCH /api/v1/users/:id (update user)
- [ ] Role/permission API:
  - GET /api/v1/roles (list roles)
  - POST /api/v1/roles (create role)
  - PATCH /api/v1/roles/:id (update role)

---

## Week 6: Offline & Sync Backend

### Server Room App: Offline Mode (2 days)
**Deliverables:**
- [ ] Offline detection (check internet connectivity)
- [ ] Sync queue (store unsynced transactions in local DB)
- [ ] Sync loop (push queued data to cloud)
- [ ] Exponential backoff retry
- [ ] Sync failure alerting

### Cloud Backend: Batch Sync (1 day)
**Deliverables:**
- [ ] POST /api/v1/sync/batch (receive multiple transactions)
- [ ] Duplicate detection (check transaction_id)
- [ ] Return list of successfully synced IDs

---

## Week 7: Monitoring Backend

### Logging (1 day)
**Deliverables:**
- [ ] Structured logging (JSON format)
- [ ] Log levels (error, warn, info, debug)
- [ ] Log to stdout (Loki collects from stdout)
- [ ] Request logging (middleware: log all HTTP requests)

### Metrics (2 days)
**Deliverables:**
- [ ] Prometheus metrics (expose /metrics endpoint)
- [ ] Metrics:
  - HTTP request count, duration, status codes
  - DB query count, duration
  - Sync queue length
  - Gate health status
  - Transaction count
- [ ] Custom metrics (business metrics: transactions per hour, revenue per day)

---

## Week 8: Testing Backend

### Unit Tests (3 days)
**Deliverables:**
- [ ] Cloud backend unit tests (handlers, store logic)
- [ ] Server room app unit tests (fee calculation, sync logic)
- [ ] Gate app unit tests (HAL mocks, sequence logic)
- [ ] Test coverage: 80%+ for critical paths

### Integration Tests (2 days)
**Deliverables:**
- [ ] Cloud backend integration tests (API endpoints, DB operations)
- [ ] Server room app integration tests (sync with cloud, gate communication)
- [ ] Gate app integration tests (hardware mocks, sequence logic)

### Load Tests (1 day)
**Deliverables:**
- [ ] Cloud backend load test (simulate 20+ locations, 100+ gates)
- [ ] Measure: API response time, DB query time, sync throughput
- [ ] Identify bottlenecks, optimize

---

## Backend Timeline Summary

| Week | Cloud Backend | Server Room App | Gate App |
|------|---------------|-----------------|----------|
| **Week 1-2** | Foundation (DB, API, auth, configs, gates, sync) | - | - |
| **Week 3-4** | Session/transaction sync | Foundation (DB, sync, mDNS, HTTP, status) | Foundation (HTTP, mDNS, HAL, config) |
| **Week 3** | Entry flow sync | Entry flow (session creation) | Entry flow (entry sequence) |
| **Week 4** | Exit flow sync | Exit flow (fee calculation, transaction) | Exit flow (exit sequence) |
| **Week 5-6** | Dashboard APIs (gates, sessions, transactions, reports, users) | Offline mode, sync queue | - |
| **Week 6** | Batch sync API | Offline sync loop | - |
| **Week 7** | - | Logging, metrics | Logging, metrics |
| **Week 8** | Unit + integration tests | Unit + integration tests | Unit + integration tests |

---

## Backend Tech Stack

### Cloud Backend
- **Language:** Go 1.22+
- **Framework:** Gin (HTTP router)
- **Database:** PostgreSQL 15
- **ORM:** pgx (raw SQL, no ORM)
- **Auth:** JWT RS256
- **Logging:** Structured JSON → Loki
- **Metrics:** Prometheus client_golang
- **Testing:** Go testing + httptest

### Server Room App
- **Language:** Go or Rust (Go preferred for consistency)
- **Database:** SQLite (local, portable)
- **mDNS:** github.com/miekg/dns or hashicorp/mdns
- **HTTP:** net/http (stdlib)
- **Logging:** Structured JSON → Loki
- **Metrics:** Prometheus client
- **Testing:** Go testing

### Gate App
- **Language:** Go or Rust (Go preferred for consistency)
- **HTTP:** net/http (stdlib)
- **mDNS:** github.com/miekg/dns or hashicorp/mdns
- **Hardware:** Platform-specific libraries (serial, USB, GPIO)
- **Logging:** Structured JSON → Loki
- **Metrics:** Prometheus client
- **Testing:** Go testing + hardware mocks

---

## Backend Risks & Mitigation

| Risk | Mitigation |
|------|------------|
| Hardware integration complexity | Start HAL early (Week 3-4), use mocks for testing |
| Sync conflicts (duplicate transactions) | Use transaction_id as unique key, cloud detects duplicates |
| Offline sync delays | Exponential backoff, alert on persistent failures |
| Performance at scale (20+ locations) | Load test (Week 8), optimize DB queries, add indexes |
| mDNS discovery issues | Fallback to manual IP config, log discovery events |
| Fee calculation bugs | Unit test fee calculation thoroughly, edge cases (multi-day, overnight) |

---

*End of Backend Timeline*
