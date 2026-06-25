# PARKIR v1 — Implementation Plan

## 1. Overview

PARKIR is a multi-location Parking Administration System (PAS) for managing vehicle parking operations across one or more physical locations. It is operated by parking attendants at booths/gates via a desktop app and supervised by facility managers through a web-based dashboard.

This plan captures all final decisions and the implementation roadmap for v1.

---

## 2. Locked Decisions

| Area | Decision |
|------|----------|
| Cloud provider | Tencent Cloud |
| Region | Jakarta |
| Server model | Single-server MVP |
| Database | Self-hosted PostgreSQL on the same VM |
| Backend language/framework | Go + Gin |
| Database access | Raw SQL with pgx |
| Web dashboard | Next.js 14+ (App Router), TypeScript, Tailwind CSS |
| Desktop app | Electron + React/TypeScript |
| Auth | JWT access tokens, 8-hour expiry |
| Currency | Indonesian Rupiah (IDR) |
| Vehicle types | CAR, MOTO, TRUCK |
| Rate model | `first_hour_rate` + `subsequent_hourly_rate` + `daily_flat_rate` |
| RBAC model | Hybrid: role-location assignment + independent permission grants |
| Digital payments | Mock/manual confirmation in v1 (no gateway integration) |
| Offline mode | Included in v1 MVP |
| Shift system | Full shift tracking with cash handover + discrepancy flagging |
| Peak load assumption | 120 vehicles/hour at busiest location |
| Concurrent terminals | 1–3 per location |
| Locations year 1 | >2 |

---

## 3. Target Load & Non-Functional Targets

| Metric | Target |
|--------|--------|
| API response time (p95) | < 300ms for session/payment operations |
| Dashboard report load | < 2s |
| Offline sync | All offline sessions synced within 60s of reconnect |
| Receipt print time | < 3s from payment confirmation |
| Audit log retention | 2 years, append-only, non-deletable |
| Concurrent operator sessions | 50 per location |
| Backend uptime | 99.5% |

---

## 4. Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Tencent Cloud                         │
│                          Jakarta                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Single VM                                │   │
│  │  ┌───────────────┐        ┌──────────────────────┐  │   │
│  │  │  Go Backend   │◄──────►│  PostgreSQL          │  │   │
│  │  │  API (Gin)    │        │  Self-hosted         │  │   │
│  │  └───────┬───────┘        └──────────────────────┘  │   │
│  │          │                                           │   │
│  │          ▼                                           │   │
│  │  ┌─────────────────────────────────────────────┐    │   │
│  │  │  Next.js Dashboard (served via backend)     │    │   │
│  │  └─────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────┘   │
│                            ▲                                 │
└────────────────────────────┼─────────────────────────────────┘
                             │ HTTPS
              ┌──────────────┴──────────────┐
              │                             │
       ┌──────▼──────┐              ┌───────▼──────┐
       │  Electron   │              │   Manager    │
       │  Desktop    │              │   Browser    │
       │  (Operator) │              │  (Dashboard) │
       └─────────────┘              └──────────────┘
```

---

## 5. Project Structure

```
PARKIR/
├── backend/                  # Go + Gin API
│   ├── cmd/api/              # Application entrypoint
│   ├── internal/
│   │   ├── config/           # Env config
│   │   ├── db/               # DB connection, migrations runner
│   │   ├── middleware/       # Auth, CORS, logging, RBAC
│   │   ├── auth/             # JWT, password hashing, PIN hashing
│   │   ├── domain/           # One package per domain
│   │   │   ├── users/
│   │   │   ├── roles/
│   │   │   ├── locations/
│   │   │   ├── rates/
│   │   │   ├── sessions/
│   │   │   ├── transactions/
│   │   │   ├── shifts/
│   │   │   ├── incidents/
│   │   │   ├── adjustments/
│   │   │   ├── reports/
│   │   │   ├── auditlogs/
│   │   │   ├── alerts/
│   │   │   └── health/
│   │   └── permissions/      # Permission resolution logic
│   ├── migrations/           # SQL migration files
│   └── Makefile
├── dashboard/                # Next.js App Router web dashboard
│   ├── app/
│   │   ├── (dashboard)/      # Dashboard layout group
│   │   ├── login/
│   │   └── api/              # Next.js API routes if needed
│   ├── components/
│   ├── lib/
│   │   ├── api.ts            # API client
│   │   └── permissions.ts    # Client-side permission helpers
│   ├── hooks/
│   ├── types/
│   └── package.json
├── desktop/                  # Electron desktop app
│   ├── src/
│   │   ├── main/             # Electron main process
│   │   ├── renderer/         # React UI
│   │   ├── lib/
│   │   └── stores/           # Offline local storage
│   └── package.json
├── db/
│   └── schema.sql            # Canonical v1 schema
├── docker-compose.yml        # Local dev stack
├── Makefile                  # Common commands
└── README.md
```

---

## 6. Database Schema

### Core Tables

```sql
-- Physical parking locations
locations
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
  name            VARCHAR(150) NOT NULL
  code            VARCHAR(20) UNIQUE NOT NULL
  address         TEXT
  city            VARCHAR(100)
  status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                    CHECK (status IN ('ACTIVE', 'INACTIVE'))
  capacity        JSONB  -- { "CAR": 100, "MOTO": 50, "TRUCK": 20 }
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()

-- Rate configuration per location and vehicle type
location_rates
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid()
  location_id             UUID NOT NULL REFERENCES locations(id)
  vehicle_type            VARCHAR(10) NOT NULL
                            CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK'))
  first_hour_rate         NUMERIC(12,2) NOT NULL
  subsequent_hourly_rate  NUMERIC(12,2) NOT NULL
  daily_flat_rate         NUMERIC(12,2) NOT NULL
  effective_from          DATE NOT NULL
  effective_until         DATE
  created_by              UUID REFERENCES users(id)
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now()

-- Named permission sets
roles
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
  name            VARCHAR(100) UNIQUE NOT NULL
  permissions     JSONB NOT NULL DEFAULT '[]'
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()

-- System users
users
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
  name            VARCHAR(100) NOT NULL
  email           VARCHAR(255) UNIQUE NOT NULL
  password_hash   VARCHAR NOT NULL
  pin_hash        VARCHAR              -- 6-digit manager PIN (hashed)
  role_id         UUID NOT NULL REFERENCES roles(id)
  status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                    CHECK (status IN ('ACTIVE', 'DEACTIVATED'))
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()

-- Role applies at these locations
user_role_locations
  user_id         UUID NOT NULL REFERENCES users(id)
  location_id     UUID NOT NULL REFERENCES locations(id)
  PRIMARY KEY (user_id, location_id)

-- Independent permission grants (additive)
user_permission_grants
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
  user_id         UUID NOT NULL REFERENCES users(id)
  location_id     UUID REFERENCES locations(id)  -- NULL = global
  permission      VARCHAR(100) NOT NULL
  granted_by      UUID NOT NULL REFERENCES users(id)
  granted_at      TIMESTAMPTZ NOT NULL DEFAULT now()
  expires_at      TIMESTAMPTZ
  revoked_at      TIMESTAMPTZ
  revoked_by      UUID REFERENCES users(id)
  UNIQUE (user_id, location_id, permission)

-- Operator work periods
shifts
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid()
  operator_id             UUID NOT NULL REFERENCES users(id)
  location_id             UUID NOT NULL REFERENCES locations(id)
  status                  VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                            CHECK (status IN ('OPEN', 'CLOSED', 'FLAGGED', 'RESOLVED', 'FORCE_CLOSED'))
  started_at              TIMESTAMPTZ NOT NULL DEFAULT now()
  ended_at                TIMESTAMPTZ
  expected_cash           NUMERIC(12,2)
  cash_handover_amount    NUMERIC(12,2)
  discrepancy             NUMERIC(12,2)
  discrepancy_notes       TEXT
  force_closed_by         UUID REFERENCES users(id)
  force_closed_reason     TEXT
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()

-- Parking sessions
sessions
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
  location_id     UUID NOT NULL REFERENCES locations(id)
  operator_id     UUID NOT NULL REFERENCES users(id)
  shift_id        UUID REFERENCES shifts(id)  -- check-in shift
  plate           VARCHAR(20) NOT NULL
  city_code       VARCHAR(10) NOT NULL DEFAULT 'UNKNOWN'
  vehicle_type    VARCHAR(10) NOT NULL
                    CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK'))
  state           VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                    CHECK (state IN ('ACTIVE', 'PENDING_PAYMENT', 'CLOSED', 'VOIDED'))
  check_in_at     TIMESTAMPTZ NOT NULL
  check_out_at    TIMESTAMPTZ
  fee_amount      NUMERIC(12,2)
  rate_snapshot   JSONB
  offline_sync    BOOLEAN NOT NULL DEFAULT false
  sync_conflict   BOOLEAN NOT NULL DEFAULT false
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()

-- Payment records
transactions
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid()
  session_id              UUID UNIQUE NOT NULL REFERENCES sessions(id)
  location_id             UUID NOT NULL REFERENCES locations(id)
  shift_id                UUID NOT NULL REFERENCES shifts(id)  -- payment shift
  operator_id             UUID NOT NULL REFERENCES users(id)    -- payment collector
  vehicle_type            VARCHAR(10) NOT NULL
  plate                   VARCHAR(20) NOT NULL
  check_in_at             TIMESTAMPTZ NOT NULL
  check_out_at            TIMESTAMPTZ NOT NULL
  duration_hours          INTEGER NOT NULL
  rate_first_hour         NUMERIC(12,2) NOT NULL
  rate_subsequent_hourly  NUMERIC(12,2) NOT NULL
  rate_daily              NUMERIC(12,2) NOT NULL
  fee_amount              NUMERIC(12,2) NOT NULL
  payment_method          VARCHAR(10) NOT NULL
                            CHECK (payment_method IN ('CASH', 'DIGITAL'))
  amount_tendered         NUMERIC(12,2)
  change_amount           NUMERIC(12,2)
  payment_reference       VARCHAR(100)
  receipt_number          VARCHAR(50) UNIQUE NOT NULL
  voided                  BOOLEAN NOT NULL DEFAULT false
  voided_at               TIMESTAMPTZ
  voided_by               UUID REFERENCES users(id)
  void_reason             TEXT
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now()

-- Operational incidents
incidents
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid()
  location_id       UUID NOT NULL REFERENCES locations(id)
  type              VARCHAR(30) NOT NULL
                      CHECK (type IN ('STUCK_AT_GATE', 'PAYMENT_DISPUTE', 'OPERATOR_ERROR', 'SYSTEM_DOWNTIME'))
  state             VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                      CHECK (state IN ('OPEN', 'IN_PROGRESS', 'RESOLVED'))
  session_id        UUID REFERENCES sessions(id)
  reported_by       UUID NOT NULL REFERENCES users(id)
  reported_at       TIMESTAMPTZ NOT NULL
  description       TEXT NOT NULL
  resolved_by       UUID REFERENCES users(id)
  resolved_at       TIMESTAMPTZ
  resolution_notes  TEXT
  offline_sync      BOOLEAN NOT NULL DEFAULT false
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()

incident_notes
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid()
  incident_id   UUID NOT NULL REFERENCES incidents(id)
  author_id     UUID NOT NULL REFERENCES users(id)
  note          TEXT NOT NULL
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()

-- Immutable audit trail
audit_logs
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid()
  action        VARCHAR(100) NOT NULL
  actor_id      UUID REFERENCES users(id)
  actor_role    VARCHAR(50)
  entity_type   VARCHAR(50) NOT NULL
  entity_id     UUID NOT NULL
  location_id   UUID REFERENCES locations(id)
  ip_address    INET
  metadata      JSONB
  timestamp     TIMESTAMPTZ NOT NULL DEFAULT now()

-- Anomaly alerts
alerts
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid()
  code              VARCHAR(50) NOT NULL
  location_id       UUID REFERENCES locations(id)
  state             VARCHAR(20) NOT NULL DEFAULT 'TRIGGERED'
                      CHECK (state IN ('TRIGGERED', 'ACKNOWLEDGED', 'RESOLVED'))
  entity_type       VARCHAR(50)
  entity_id         UUID
  triggered_at      TIMESTAMPTZ NOT NULL DEFAULT now()
  acknowledged_by   UUID REFERENCES users(id)
  acknowledged_at   TIMESTAMPTZ
  resolved_by       UUID REFERENCES users(id)
  resolved_at       TIMESTAMPTZ
  resolution_notes  TEXT
  metadata          JSONB
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()

alert_configs
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid()
  location_id   UUID REFERENCES locations(id)  -- NULL = global default
  code          VARCHAR(50) NOT NULL
  enabled       BOOLEAN NOT NULL DEFAULT true
  threshold     JSONB
  updated_by    UUID REFERENCES users(id)
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
  UNIQUE (location_id, code)
```

---

## 7. Permission Model

Effective permissions for a user at a location are computed as:

```
permissions = []

# 1. Role permissions if role applies at this location
if location in user.role_locations:
    permissions += user.role.permissions

# 2. Active independent grants for this location
permissions += user.grants.where(location_id = location, revoked_at IS NULL, expires_at IS NULL OR expires_at > now())

# 3. Active global grants
permissions += user.grants.where(location_id IS NULL, revoked_at IS NULL, expires_at IS NULL OR expires_at > now())

return unique(permissions)
```

Key rules:
- Owners have all permissions everywhere.
- Admins have all permissions except `finance:*` everywhere.
- Only owners can grant `finance:*` permissions.

---

## 8. Fee Calculation

```
duration_hours = CEIL((check_out_at - check_in_at) in seconds / 3600)
if duration_hours == 0: duration_hours = 1

if duration_hours == 1:
    raw_fee = first_hour_rate
else:
    raw_fee = first_hour_rate + (duration_hours - 1) * subsequent_hourly_rate

fee = MIN(raw_fee, daily_flat_rate)
```

Rate lookup uses the active rate for the vehicle type at the location on the `check_in_at` date. The applied rate is stored in `sessions.rate_snapshot` at check-out for auditability.

---

## 9. Session Lifecycle

```
CHECK-IN → ACTIVE → PENDING_PAYMENT → CLOSED
                ↘        ↘              ↓
                 VOIDED ←───────────────┘
```

- `ACTIVE`: vehicle is parked
- `PENDING_PAYMENT`: check-out initiated, fee calculated, awaiting payment
- `CLOSED`: payment confirmed, receipt printed
- `VOIDED`: terminal state; excluded from revenue

---

## 10. Receipt Number Format

```
[LOCATION_CODE]-[YYYYMMDD]-[SEQUENCE]
Example: GMP01-20250315-00042
```

Sequence resets daily per location. Use a per-location daily sequence table with row-level locking to avoid races.

---

## 11. Offline Mode Behavior

- Operator app detects connectivity loss every 30s.
- Sessions, transactions, and incidents are stored locally.
- Rates are cached locally with 24h TTL.
- On reconnect, unsynced records are pushed in `check_in_at` order.
- Duplicate active plates cause `sync_conflict = true`; manager resolves in dashboard.
- Offline receipts use temporary format `[CODE]-OFFLINE-[SEQ]` until synced.

---

## 12. API Modules

| Module | Endpoints |
|--------|-----------|
| Auth | `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout` |
| Users | `GET /users`, `POST /users`, `GET /users/:id`, `PATCH /users/:id`, `POST /users/:id/deactivate` |
| Roles | `GET /roles`, `POST /roles`, `GET /roles/:id`, `PATCH /roles/:id` |
| Locations | `GET /locations`, `POST /locations`, `GET /locations/:id`, `PATCH /locations/:id`, `POST /locations/:id/assign-operator` |
| Rates | `GET /locations/:id/rates`, `POST /locations/:id/rates`, `PATCH /rates/:id` |
| Sessions | `POST /sessions/check-in`, `POST /sessions/:id/check-out`, `GET /sessions`, `GET /sessions/:id` |
| Payments | `POST /payments/cash`, `POST /payments/digital`, `GET /payments/calculate-fee` |
| Transactions | `GET /transactions`, `GET /transactions/:id`, `POST /transactions/:id/void` |
| Shifts | `POST /shifts/start`, `POST /shifts/:id/end`, `POST /shifts/:id/force-close`, `GET /shifts` |
| Incidents | `GET /incidents`, `POST /incidents`, `GET /incidents/:id`, `PATCH /incidents/:id`, `POST /incidents/:id/notes` |
| Adjustments | `POST /adjustments/void-transaction`, `POST /adjustments/reassign-session` |
| Reports | `GET /reports/daily-revenue`, `GET /reports/occupancy`, `GET /reports/vehicle-breakdown`, `GET /reports/operator-activity` |
| Audit Logs | `GET /audit-logs` |
| Alerts | `GET /alerts`, `POST /alerts/:id/acknowledge`, `POST /alerts/:id/resolve`, `GET /alert-configs`, `PATCH /alert-configs/:id` |
| Health | `GET /health` |

---

## 13. Realistic 16-Week Milestone Plan

**Assumptions:**
- Focused team of 2–3 engineers (1 backend-heavy, 1 frontend-heavy, 1 flexible/full-stack).
- Backend is built before/in parallel with frontend, but API contracts are finalized early.
- Desktop app reuses dashboard components and API client where possible.
- Each milestone includes unit tests, integration tests where applicable, and a review.

### Milestone 0 — Foundation (Week 1)
**Focus:** Get all three projects running locally with a shared dev environment.

- [ ] Repository structure: `backend/`, `dashboard/`, `desktop/`, `db/`
- [ ] Docker Compose for local dev (Postgres + Go backend + Next.js dashboard)
- [ ] Go module with Gin + pgx + `cmd/api/main.go`
- [ ] Next.js App Router project scaffold with Tailwind
- [ ] Electron desktop app scaffold (main + renderer)
- [ ] Makefile with `dev`, `migrate`, `test`, `lint`, `build`
- [ ] `.gitignore` for Go, Node, Electron
- [ ] Initial `README.md` with setup instructions
- [ ] CI/CD skeleton (lint, test, build)

**Definition of done:** `make dev` spins up Postgres, backend, and dashboard; health endpoint returns OK.

---

### Milestone 1 — Backend Core Entities (Weeks 2–4) ✅ COMPLETE
**Focus:** Build the data layer and auth/RBAC foundation. Do not build business logic yet.

- [x] Database migrations for `locations`, `location_rates`, `roles`, `users`, `user_role_locations`, `user_permission_grants`
- [x] Repository pattern with raw SQL/pgx
- [x] Password hashing (bcrypt) and PIN hashing (bcrypt)
- [x] JWT login/refresh/logout
- [x] RBAC permission resolution (`getPermissions(user, location)`)
- [x] Auth middleware: validate JWT + extract user + permissions
- [x] CRUD endpoints: users, roles, locations, rates
- [x] Seed script with default roles (`operator`, `manager`, `admin`, `owner`) and a root owner
- [x] Audit log writes for mutations

**Definition of done:** API endpoints for users/roles/locations/rates manually tested; RBAC middleware blocks unauthorized requests; README documents endpoints.

---

### Milestone 2 — Backend Business Logic (Weeks 5–7) ✅ COMPLETE
**Focus:** Core parking operations — sessions, payments, receipts, shifts.

- [x] `sessions` table migration
- [x] Check-in endpoint with plate normalization and duplicate active plate warning
- [x] Check-out endpoint with fee calculation engine
- [x] `transactions` table migration
- [x] Cash payment recording + change calculation
- [x] Digital (mock) payment recording
- [x] Receipt number sequence generation (per location/day, race-safe)
- [x] `shifts` table migration + start/end shift endpoints
- [x] Shift cash handover + discrepancy calculation
- [x] Cross-shift attribution: `sessions.shift_id` (check-in) vs `transactions.shift_id` (payment)
- [x] Audit log writes for session, transaction, and shift events
- [x] Integration test scaffolding and CI

**Definition of done:** Full check-in → check-out → payment → receipt flow works via API; shift discrepancy is calculated correctly; integration tests pass.

---

### Milestone 3 — Web Dashboard Foundation (Weeks 6–8) ✅ COMPLETE
**Focus:** Manager-facing management and operational pages. Starts in parallel with Milestone 2 once API contracts are stable.

- [x] Login page
- [x] Layout with navigation and location selector
- [x] Users management page (CRUD, reset password/PIN, role assignment)
- [x] Roles & permissions page
- [x] Locations management page (assign/remove operators)
- [x] Rates configuration page
- [x] Active sessions page
- [x] Session history page with filters
- [x] Session detail page with linked transaction
- [x] Transactions list page
- [x] Shifts list page
- [x] Shifts detail page with transactions and cash summary
- [x] CI: backend tests, dashboard build, dashboard type-check

**Definition of done:** Managers can configure the system and view sessions/transactions/shifts entirely through the dashboard; backend integration tests pass; dashboard builds and type-checks successfully.

---

### Milestone 4 — Desktop App Online Mode (Weeks 9–10)
**Focus:** Operator day-to-day flows in the Electron app while online.

- [ ] Electron main process + secure API client
- [ ] Operator login + active location selection
- [ ] Home screen with quick actions and active session counts
- [ ] Check-in screen (plate + vehicle type, duplicate warning)
- [ ] Check-out / payment screen (search session, fee display, cash/digital)
- [ ] Session search screen
- [ ] Incident report screen
- [ ] Settings screen (printer config, location switch)
- [ ] Thermal receipt generation and ESC/POS printing

**Definition of done:** An operator can complete the full online check-in → check-out → payment → receipt print flow from the desktop app.

---

### Milestone 5 — Offline Mode & Sync (Weeks 11–12)
**Focus:** Make the desktop app resilient to connectivity loss.

- [ ] Connectivity detection (heartbeat polling)
- [ ] Local SQLite/IndexedDB store for sessions, transactions, incidents
- [ ] Local rate cache with 24h TTL
- [ ] Offline check-in, check-out, payment, receipt printing
- [ ] Sync queue ordered by `check_in_at`
- [ ] Backend sync endpoint that accepts batched offline records
- [ ] Sync conflict detection (duplicate active plate)
- [ ] Dashboard sync conflict resolution UI
- [ ] Auto-retry and operator-facing sync status

**Definition of done:** Operator can work fully offline for an extended period; on reconnect, data syncs and conflicts are surfaced to managers.

---

### Milestone 6 — Incidents, Adjustments & Observability (Weeks 13–14)
**Focus:** Exception handling, corrections, audit, and monitoring.

- [ ] `incidents` and `incident_notes` migrations
- [ ] File incident endpoint (operator)
- [ ] Incident list/detail/resolve workflow (manager dashboard)
- [ ] Manual adjustments: void transaction + manager PIN
- [ ] Manual adjustments: reassign session + manager PIN
- [ ] Audit log page in dashboard with filters
- [ ] System health dashboard (API, DB, payment gateway, printers)
- [ ] Alert rules: `LONG_SESSION`, `UNPAID_EXIT`, `HIGH_VOID_RATE`, `SYNC_FAILURE`, `GATEWAY_FAILURE`
- [ ] Alert list and configuration page

**Definition of done:** Managers can resolve incidents and adjustments; audit log is queryable; health and alerts are visible.

---

### Milestone 7 — Reports & Polish (Week 15)
**Focus:** Data-driven insights and end-user polish.

- [ ] Daily revenue summary report (cards + bar chart + table)
- [ ] Occupancy over time report (heatmap + line chart)
- [ ] Per-vehicle-type breakdown report
- [ ] Operator activity log report
- [ ] CSV export for reports
- [ ] UI/UX consistency pass
- [ ] Error boundaries and loading states
- [ ] End-to-end smoke tests for critical flows

**Definition of done:** All four reports render correctly; core flows work end-to-end without obvious bugs.

---

### Milestone 8 — Testing & Deploy (Week 16)
**Focus:** Production readiness.

- [ ] Backend integration test suite complete
- [ ] Dashboard component and integration tests
- [ ] Desktop app smoke tests
- [ ] Tencent Cloud Jakarta VM provisioning
- [ ] PostgreSQL install + backup config (daily dumps)
- [ ] Go backend deploy with systemd or Docker
- [ ] Next.js dashboard deploy (served by Go backend or standalone)
- [ ] Electron auto-update channel setup
- [ ] Environment-specific config (dev/staging/prod)
- [ ] Production smoke tests
- [ ] Deployment runbook

**Definition of done:** System is live on Tencent Cloud Jakarta; operators and managers can log in and perform core flows.

---

## 14. Parallel Work Strategy

| Week | Backend | Dashboard | Desktop |
|------|---------|-----------|---------|
| 1 | Foundation | Foundation | Foundation |
| 2–4 | Core entities + RBAC | Login + management pages | — |
| 5–7 | Sessions/payments/shifts | Operations pages | — |
| 8 | Sessions/payments wrap-up | Operations pages wrap-up | Scaffold |
| 9–10 | API support for desktop | Minor dashboard polish | Online mode |
| 11–12 | Sync endpoint + conflicts | Sync conflict UI | Offline mode |
| 13–14 | Incidents/adjustments/observability | Incident/adjustment/audit/health pages | Incident filing |
| 15 | Reports APIs | Reports UI | Reports read-only |
| 16 | Deploy | Deploy | Deploy + auto-update |

---

## 15. First Tasks (Implementation Start)

Start with **Milestone 0 — Foundation**. The goal is a working local dev environment, not feature code.

1. Create `backend/` Go module with Gin + pgx + `cmd/api/main.go`
2. Create `dashboard/` Next.js App Router project
3. Create `desktop/` Electron project skeleton
4. Create `docker-compose.yml` with Postgres, backend, and dashboard services
5. Create `db/schema.sql` with full v1 schema
6. Create `Makefile` with `dev`, `migrate`, `test`, `lint`, `build` commands
7. Add `.gitignore` for Go, Node, and Electron
8. Add initial `README.md` with setup instructions
9. Add `GET /health` endpoint returning `{"status":"ok"}`

**Success criteria:** Run `make dev` and verify Postgres, backend, and dashboard are all running; `curl http://localhost:8080/health` returns OK.

---

## 16. Risks & Notes

1. **Offline mode complexity** — Requires robust local storage and conflict resolution. Consider iterative delivery: basic offline first, then conflict UI.
2. **Receipt printing** — ESC/POS over USB/serial/network can be finicky. Test with target printer models early.
3. **Receipt number races** — Use database-level locking (advisory lock or sequence table) for per-location daily sequences.
4. **Permission resolution performance** — Cache effective permissions per user/location in JWT or Redis if needed.
5. **Audit log append-only** — Enforce at DB level by revoking UPDATE/DELETE on `audit_logs` from application role.
6. **Shift cross-over sessions** — A vehicle may check in during one shift and pay during another. `sessions.shift_id` tracks check-in shift; `transactions.shift_id` tracks payment shift for cash reconciliation.

---

## 17. Out of Scope (v1)

- Parking slot-level tracking
- Monthly subscriptions / pass-based billing
- Driver-facing portal or mobile app
- Gate/barrier hardware integration
- Push notifications (email/SMS/WhatsApp)
- Mixed payment (cash + digital)
- Multi-currency support
- EV charging slot management
- Self-service payment kiosks
- Automated digital payment gateway callbacks

---

*Plan finalized. Ready to implement.*
