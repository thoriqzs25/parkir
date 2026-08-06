# Chapter 14 — Data Model

## 14.1 Overview

This chapter defines the complete conceptual and physical data model for PARKIR v2. The system uses two databases:

1. **Cloud Backend (PostgreSQL):** Central source of truth for configs, aggregated data, multi-location management.
2. **Server Room App (SQLite):** Local database per location, stores configs, sessions, transactions, sync queue.

Both databases share similar schemas. Cloud backend uses UUID primary keys. Server room app uses UUID for consistency. Timestamps are stored in UTC. Currency values use integers (cents) for precision.

---

## 14.2 Entity Relationship Overview

```
Cloud Backend (PostgreSQL):

locations ─────────────────────────────────────────────┐
    │                                                   │
    ├── gates (entry/exit gates per location)           │
    │                                                   │
    ├── rate_configs (versioned, per vehicle type)      │
    │                                                   │
    ├── shift_configs (versioned, per location)         │
    │                                                   │
    ├── users ── roles                                  │
    │                                                   │
    ├── sessions ────────────────────────────────────   │
    │       │                                           │
    │       ├── transactions                            │
    │       │                                           │
    │       └── incidents ── incident_notes             │
    │                                                   │
    ├── alerts                                          │
    │                                                   │
    └── audit_logs ─────────────────────────────────────┘

Server Room App (SQLite):
- Same structure as cloud, plus sync_queue table
- Synced from cloud (configs) and to cloud (sessions, transactions)
```

---

## 14.3 Table Definitions

### `locations`
Represents a physical parking facility.

```sql
CREATE TABLE locations (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            VARCHAR(150) NOT NULL,
  code            VARCHAR(20) UNIQUE NOT NULL,
  address         TEXT,
  city            VARCHAR(100),
  status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                    CHECK (status IN ('ACTIVE', 'INACTIVE')),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_locations_city ON locations (city);
```

---

### `gates`
Physical entry/exit gates at a location.

```sql
CREATE TABLE gates (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_id     UUID NOT NULL REFERENCES locations(id),
  gate_id         VARCHAR(100) UNIQUE NOT NULL,  -- hardware serial/MAC
  gate_type       VARCHAR(10) NOT NULL CHECK (gate_type IN ('ENTRY', 'EXIT')),
  vehicle_type    VARCHAR(10) NOT NULL CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK', 'ALL')),
  status          VARCHAR(20) NOT NULL DEFAULT 'UNREGISTERED'
                    CHECK (status IN ('UNREGISTERED', 'REGISTERED', 'OPERATIONAL')),
  config          JSONB,  -- gate-specific configuration
  server_room_url VARCHAR(255),  -- learned during bootstrap
  last_seen_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_gates_location ON gates (location_id);
CREATE INDEX idx_gates_status ON gates (status);
```

**Notes:**
- `gate_id` is hardware identifier (serial number or MAC address).
- `status`: UNREGISTERED (discovered but not configured), REGISTERED (configured), OPERATIONAL (active).
- `server_room_url`: gate app saves this for reconnection after restart.
- `config`: JSONB for flexible gate-specific settings.

---

### `rate_configs`
Versioned rate configuration per location and vehicle type.

```sql
CREATE TABLE rate_configs (
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_id             UUID NOT NULL REFERENCES locations(id),
  vehicle_type            VARCHAR(10) NOT NULL CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK')),
  version                 VARCHAR(50) NOT NULL,  -- e.g., "rate_v42"
  first_hour_rate         INTEGER NOT NULL,  -- in cents (e.g., 500000 = Rp 5,000)
  subsequent_hourly_rate  INTEGER NOT NULL,
  daily_flat_rate         INTEGER NOT NULL,
  effective_from          DATE NOT NULL,
  effective_until         DATE,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_rate_configs_location_type ON rate_configs (location_id, vehicle_type);
CREATE INDEX idx_rate_configs_version ON rate_configs (version);
```

**Notes:**
- `version`: auto-incremented on create/update (e.g., "rate_v1", "rate_v2", ...).
- Transactions record `rate_config_version` (audit trail).
- Integer cents for precision (no floating point).

---

### `shift_configs`
Versioned shift configuration per location.

```sql
CREATE TABLE shift_configs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_id     UUID NOT NULL REFERENCES locations(id),
  version         VARCHAR(50) NOT NULL,  -- e.g., "shift_v15"
  shift_code      VARCHAR(20) NOT NULL,  -- e.g., "06-14", "14-22", "22-06"
  shift_number    INTEGER NOT NULL,  -- display number within day
  start_time      TIME NOT NULL,
  end_time        TIME NOT NULL,
  is_overnight    BOOLEAN NOT NULL DEFAULT false,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shift_configs_location ON shift_configs (location_id);
CREATE INDEX idx_shift_configs_version ON shift_configs (version);
```

**Notes:**
- Default: 3 shifts/day (morning, afternoon, evening).
- Shift numbers increment continuously across days (shift_1, shift_2, shift_3, shift_4, ...).
- `is_overnight`: true if shift crosses midnight (e.g., 22:00 - 06:00).

---

### `sessions`
Parking sessions (vehicle entry to exit).

```sql
CREATE TABLE sessions (
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_id             UUID NOT NULL REFERENCES locations(id),
  gate_id                 VARCHAR(100) NOT NULL,  -- entry gate hardware ID
  vehicle_type            VARCHAR(10) NOT NULL CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK')),
  check_in_at             TIMESTAMPTZ NOT NULL,
  check_out_at            TIMESTAMPTZ,
  fee_amount              INTEGER,  -- in cents, nullable until check-out
  rate_config_version     VARCHAR(50),  -- which rate version was used
  shift_config_version    VARCHAR(50),  -- which shift config version was used
  shift_number            INTEGER NOT NULL,  -- assigned at check-in
  state                   VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                            CHECK (state IN ('ACTIVE', 'PENDING_PAYMENT', 'CLOSED', 'VOIDED')),
  qr_data                 TEXT NOT NULL,  -- QR code content (for audit)
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  synced_at               TIMESTAMPTZ  -- when synced to cloud
);

CREATE INDEX idx_sessions_location ON sessions (location_id);
CREATE INDEX idx_sessions_state ON sessions (state);
CREATE INDEX idx_sessions_check_in ON sessions (check_in_at);
CREATE INDEX idx_sessions_synced ON sessions (synced_at);
```

**Notes:**
- No `operator_id` (no operators in v2).
- `gate_id` is hardware ID (string), not a foreign key to gates table.
- `qr_data` stores what was encoded in QR code (for audit/validation).
- `synced_at` tracks when session was synced to cloud (null = not synced yet).

---

### `transactions`
Payment records for closed sessions.

```sql
CREATE TABLE transactions (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id          UUID UNIQUE NOT NULL REFERENCES sessions(id),
  location_id         UUID NOT NULL REFERENCES locations(id),
  gate_id             VARCHAR(100) NOT NULL,  -- exit gate hardware ID
  amount              INTEGER NOT NULL,  -- in cents
  payment_method      VARCHAR(20) NOT NULL CHECK (payment_method IN ('EMONEY', 'FLAZZ', 'CASH', 'OFFLINE_SOP')),
  payment_reference   VARCHAR(100),  -- vendor reference (for e-money/Flazz)
  config_version      VARCHAR(50) NOT NULL,  -- which rate version was used
  shift_number        INTEGER NOT NULL,  -- assigned at check-out
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  synced_at           TIMESTAMPTZ,  -- when synced to cloud
  voided              BOOLEAN NOT NULL DEFAULT false,
  voided_at           TIMESTAMPTZ,
  voided_by           UUID REFERENCES users(id),  -- manager who voided
  void_reason         TEXT
);

CREATE INDEX idx_transactions_session ON transactions (session_id);
CREATE INDEX idx_transactions_location ON transactions (location_id);
CREATE INDEX idx_transactions_created ON transactions (created_at);
CREATE INDEX idx_transactions_synced ON transactions (synced_at);
CREATE INDEX idx_transactions_voided ON transactions (voided);
```

**Notes:**
- 1:1 relationship with sessions (one transaction per closed session).
- `payment_method`: EMONEY, FLAZZ (tap-to-go), CASH (staff taps with own card), OFFLINE_SOP.
- `config_version`: audit trail for which rate was used.
- `shift_number`: may differ from session shift_number if overnight.
- `voided_by`: manager who voided (requires manager PIN).

---

### `sync_queue` (Server Room App only)
Tracks unsynced data for retry.

```sql
CREATE TABLE sync_queue (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  entity_type     VARCHAR(50) NOT NULL,  -- "session", "transaction"
  entity_id       UUID NOT NULL,
  action          VARCHAR(20) NOT NULL,  -- "create", "update"
  payload         JSONB NOT NULL,
  retry_count     INTEGER NOT NULL DEFAULT 0,
  last_attempt_at TIMESTAMPTZ,
  status          VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING', 'IN_PROGRESS', 'SUCCESS', 'FAILED')),
  error_message   TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sync_queue_status ON sync_queue (status);
CREATE INDEX idx_sync_queue_entity ON sync_queue (entity_type, entity_id);
```

**Notes:**
- Server Room App only (not in cloud backend).
- Tracks unsynced sessions and transactions.
- Retry with exponential backoff (1s, 2s, 4s, 8s, max 30s).
- Alert developer if sync fails persistently.

---

### `roles`
Named permission sets assigned to users.

```sql
CREATE TABLE roles (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            VARCHAR(100) UNIQUE NOT NULL,
  permissions     JSONB NOT NULL DEFAULT '[]',
  -- e.g., ["sessions:view", "gates:view", "payments:void"]
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

### `users`
System users (staff, managers, admins, owners).

```sql
CREATE TABLE users (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            VARCHAR(100) NOT NULL,
  email           VARCHAR(255) UNIQUE NOT NULL,
  password_hash   VARCHAR(255) NOT NULL,
  role_id         UUID NOT NULL REFERENCES roles(id),
  status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                    CHECK (status IN ('ACTIVE', 'DEACTIVATED')),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_role ON users (role_id);
CREATE INDEX idx_users_status ON users (status);
```

**Notes:**
- No operators (v2). Users are: staff, managers, admins, owners.
- Initial superadmin created by developer (bootstrap).
- Superadmin creates subsequent users.

---

### `incidents`
Operational incidents (hardware failures, exceptions).

```sql
CREATE TABLE incidents (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_id     UUID NOT NULL REFERENCES locations(id),
  gate_id         VARCHAR(100),  -- hardware ID (if applicable)
  type            VARCHAR(50) NOT NULL CHECK (type IN (
                    'HARDWARE_FAILURE', 'PAYMENT_FAILURE', 'QR_UNREADABLE',
                    'PRINTER_JAM', 'GATE_MOTOR_FAILURE', 'OTHER'
                  )),
  state           VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                    CHECK (state IN ('OPEN', 'IN_PROGRESS', 'RESOLVED')),
  description     TEXT NOT NULL,
  resolved_at     TIMESTAMPTZ,
  resolved_by     UUID REFERENCES users(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_incidents_location ON incidents (location_id);
CREATE INDEX idx_incidents_state ON incidents (state);
```

---

### `audit_logs`
Immutable audit trail.

```sql
CREATE TABLE audit_logs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID REFERENCES users(id),  -- null for system actions
  action          VARCHAR(100) NOT NULL,  -- e.g., "session.create", "transaction.void"
  entity_type     VARCHAR(50) NOT NULL,  -- "session", "transaction", "gate", etc.
  entity_id       UUID,
  location_id     UUID REFERENCES locations(id),
  metadata        JSONB,  -- additional context
  ip_address      VARCHAR(45),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_user ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs (entity_type, entity_id);
CREATE INDEX idx_audit_logs_location ON audit_logs (location_id);
CREATE INDEX idx_audit_logs_created ON audit_logs (created_at);
```

**Notes:**
- Immutable (no updates or deletes).
- Logs all state-changing actions.
- Retain minimum 2 years.

---

## 14.4 Sync Behavior

### Configs (Cloud → Server Room App)
- Server room app polls cloud every 1 minute.
- Fetches latest configs (rates, shifts, gates, locations).
- Compares version numbers.
- If cloud has newer version → update local DB.
- Local configs always overwritten by cloud (cloud is source of truth).

### Data (Server Room App → Cloud)
- Server room app syncs sessions and transactions every 1 minute.
- Batch sync: POST /api/v1/sync/batch.
- Cloud checks for duplicates (by entity_id).
- Cloud marks successfully synced entities.
- Server room app updates `synced_at` timestamp.
- Failed syncs: retry with exponential backoff, alert developer.

### Gate Status (Server Room App → Cloud)
- Server room app reports gate statuses every 20 seconds.
- Includes: gate_id, status (online/offline), last_seen_at.
- Cloud aggregates for dashboard monitoring.

---

## 14.5 Data Types

| Type | PostgreSQL | SQLite | Notes |
|------|-----------|--------|-------|
| UUID | `UUID` | `TEXT` | Use gen_random_uuid() in PostgreSQL, uuid4() in app |
| Timestamp | `TIMESTAMPTZ` | `TEXT` (ISO 8601) | Store in UTC |
| Currency | `INTEGER` (cents) | `INTEGER` (cents) | No floating point |
| JSON | `JSONB` | `TEXT` (JSON string) | Parse in application |
| Boolean | `BOOLEAN` | `INTEGER` (0/1) | SQLite uses integer |

---

## 14.6 Design Decisions

**Why UUID primary keys?**
- Globally unique (no collisions across locations).
- No sequential patterns (security).
- Easy to merge data from multiple locations.

**Why integer cents (not decimal)?**
- No floating point precision issues.
- Faster arithmetic.
- Standard practice for currency.

**Why versioned configs?**
- Audit trail (know which config was used for each transaction).
- Prevents retroactive changes.
- Supports offline operation (use last known config).

**Why gate_id is string (not UUID)?**
- Hardware identifier (serial number or MAC address).
- Assigned by hardware, not database.
- Must be globally unique across all locations.

**Why no operator_id in sessions/transactions?**
- No operators in v2 (fully automated).
- Gates are automated (no human involvement).
- Staff only handle exceptions (logged in audit_logs, not sessions).

**Why sync_queue table?**
- Track unsynced data for retry.
- Exponential backoff (avoid hammering cloud).
- Alert on persistent failures.
- Idempotent sync (cloud checks for duplicates).

---

*End of Chapter 14 — Data Model (v2)*
