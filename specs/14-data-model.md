# Chapter 14 — Data Model

## 14.1 Overview

This chapter defines the complete conceptual and physical data model for the Parking Administration System. All tables use UUID primary keys. Timestamps are stored in UTC (`TIMESTAMP WITH TIME ZONE`). Currency values use `NUMERIC(12,2)`.

---

## 14.2 Entity Relationship Overview

```
locations ──────────────────────────────────────────┐
    │                                                │
    ├── location_rates (per vehicle type)            │
    │                                                │
    ├── user_locations ── users ── roles             │
    │                                                │
    ├── sessions ──────────────────────────────────  │
    │       │                                        │
    │       ├── transactions                         │
    │       │                                        │
    │       └── incidents ── incident_notes          │
    │                                                │
    ├── alerts ── alert_configs                      │
    │                                                │
    └── audit_logs ──────────────────────────────────┘
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
  capacity        JSONB,  -- { "CAR": 100, "MOTO": 50, "TRUCK": 20 }
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

### `location_rates`
Rate configuration per location and vehicle type.

```sql
CREATE TABLE location_rates (
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_id             UUID NOT NULL REFERENCES locations(id),
  vehicle_type            VARCHAR(10) NOT NULL CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK')),
  first_hour_rate         NUMERIC(12,2) NOT NULL,
  subsequent_hourly_rate  NUMERIC(12,2) NOT NULL,
  daily_flat_rate         NUMERIC(12,2) NOT NULL,
  effective_from          DATE NOT NULL,
  effective_until         DATE,
  created_by              UUID REFERENCES users(id),
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT no_overlap UNIQUE (location_id, vehicle_type, effective_from)
);

CREATE INDEX idx_rates_location_type ON location_rates (location_id, vehicle_type);
```

---

### `roles`
Named permission sets assigned to users.

```sql
CREATE TABLE roles (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            VARCHAR(100) UNIQUE NOT NULL,
  permissions     JSONB NOT NULL DEFAULT '[]',
  -- e.g. ["sessions:view", "sessions:create", "payments:collect_cash"]
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

### `users`
System users (operators, managers, admins).

```sql
CREATE TABLE users (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            VARCHAR(100) NOT NULL,
  email           VARCHAR(255) UNIQUE NOT NULL,
  password_hash   VARCHAR NOT NULL,
  pin_hash        VARCHAR,  -- 6-digit manager PIN (hashed)
  role_id         UUID NOT NULL REFERENCES roles(id),
  status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                    CHECK (status IN ('ACTIVE', 'DEACTIVATED')),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role ON users (role_id);
```

---

### `user_locations`
Maps users to the locations they are authorized to operate at.

```sql
CREATE TABLE user_locations (
  user_id         UUID NOT NULL REFERENCES users(id),
  location_id     UUID NOT NULL REFERENCES locations(id),
  PRIMARY KEY (user_id, location_id)
);
```

---

### `sessions`
Core record of a single vehicle's parking event.

```sql
CREATE TABLE sessions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_id     UUID NOT NULL REFERENCES locations(id),
  operator_id     UUID NOT NULL REFERENCES users(id),
  plate    VARCHAR(10) NOT NULL,  -- normalized: A-1234-BCD
  city_code       VARCHAR(4) NOT NULL,   -- extracted from plate prefix
  vehicle_type    VARCHAR(10) NOT NULL CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK')),
  state           VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                    CHECK (state IN ('ACTIVE', 'PENDING_PAYMENT', 'CLOSED', 'VOIDED')),
  check_in_at     TIMESTAMPTZ NOT NULL,
  check_out_at    TIMESTAMPTZ,
  fee_amount      NUMERIC(12,2),
  rate_snapshot   JSONB,
  -- { "first_hour_rate": 5000, "subsequent_hourly_rate": 3000, "daily_flat_rate": 30000, "vehicle_type": "CAR", "effective_from": "2025-01-01" }
  offline_sync    BOOLEAN NOT NULL DEFAULT false,
  sync_conflict   BOOLEAN NOT NULL DEFAULT false,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_plate        ON sessions (plate);
CREATE INDEX idx_sessions_city_code    ON sessions (city_code);
CREATE INDEX idx_sessions_loc_state    ON sessions (location_id, state);
CREATE INDEX idx_sessions_check_in     ON sessions (check_in_at);
CREATE INDEX idx_sessions_operator     ON sessions (operator_id);
```

---

### `transactions`
Payment record produced when a session is closed.

```sql
CREATE TABLE transactions (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id          UUID UNIQUE NOT NULL REFERENCES sessions(id),
  location_id         UUID NOT NULL REFERENCES locations(id),
  shift_id            UUID NOT NULL REFERENCES shifts(id),  -- payment collection shift
  operator_id         UUID NOT NULL REFERENCES users(id),   -- payment collector
  vehicle_type        VARCHAR(10) NOT NULL,
  plate        VARCHAR(20) NOT NULL,
  check_in_at         TIMESTAMPTZ NOT NULL,
  check_out_at        TIMESTAMPTZ NOT NULL,
  duration_hours      INTEGER NOT NULL,
  rate_first_hour         NUMERIC(12,2) NOT NULL,
  rate_subsequent_hourly  NUMERIC(12,2) NOT NULL,
  rate_daily              NUMERIC(12,2) NOT NULL,
  fee_amount          NUMERIC(12,2) NOT NULL,
  payment_method      VARCHAR(10) NOT NULL CHECK (payment_method IN ('CASH', 'DIGITAL')),
  amount_tendered     NUMERIC(12,2),
  change_amount       NUMERIC(12,2),
  payment_reference   VARCHAR(100),
  receipt_number      VARCHAR(50) UNIQUE NOT NULL,
  voided              BOOLEAN NOT NULL DEFAULT false,
  voided_at           TIMESTAMPTZ,
  voided_by           UUID REFERENCES users(id),
  void_reason         TEXT,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_loc_date   ON transactions (location_id, check_out_at);
CREATE INDEX idx_transactions_shift      ON transactions (shift_id);
CREATE INDEX idx_transactions_operator   ON transactions (operator_id);
CREATE INDEX idx_transactions_plate      ON transactions (plate);
CREATE INDEX idx_transactions_receipt    ON transactions (receipt_number);
CREATE INDEX idx_transactions_active     ON transactions (voided) WHERE voided = false;
```

---

### `incidents`
Operational problems reported by operators.

```sql
CREATE TABLE incidents (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_id       UUID NOT NULL REFERENCES locations(id),
  type              VARCHAR(30) NOT NULL
                      CHECK (type IN ('STUCK_AT_GATE','PAYMENT_DISPUTE','OPERATOR_ERROR','SYSTEM_DOWNTIME')),
  state             VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                      CHECK (state IN ('OPEN', 'IN_PROGRESS', 'RESOLVED')),
  session_id        UUID REFERENCES sessions(id),
  reported_by       UUID NOT NULL REFERENCES users(id),
  reported_at       TIMESTAMPTZ NOT NULL,
  description       TEXT NOT NULL,
  resolved_by       UUID REFERENCES users(id),
  resolved_at       TIMESTAMPTZ,
  resolution_notes  TEXT,
  offline_sync      BOOLEAN NOT NULL DEFAULT false,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_incidents_location_state ON incidents (location_id, state);
```

---

### `incident_notes`
Timestamped notes added to an incident by managers.

```sql
CREATE TABLE incident_notes (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  incident_id   UUID NOT NULL REFERENCES incidents(id),
  author_id     UUID NOT NULL REFERENCES users(id),
  note          TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

### `audit_logs`
Immutable record of every state-changing action.

```sql
CREATE TABLE audit_logs (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  action        VARCHAR(100) NOT NULL,
  actor_id      UUID REFERENCES users(id),
  actor_role    VARCHAR(50),
  entity_type   VARCHAR(50) NOT NULL,
  entity_id     UUID NOT NULL,
  location_id   UUID REFERENCES locations(id),
  ip_address    INET,
  metadata      JSONB,
  timestamp     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Append-only: revoke UPDATE and DELETE from application DB user
REVOKE UPDATE, DELETE ON audit_logs FROM app_user;

CREATE INDEX idx_audit_logs_actor      ON audit_logs (actor_id);
CREATE INDEX idx_audit_logs_entity     ON audit_logs (entity_type, entity_id);
CREATE INDEX idx_audit_logs_timestamp  ON audit_logs (timestamp);
CREATE INDEX idx_audit_logs_action     ON audit_logs (action);
```

---

### `alerts`
Triggered anomaly alerts.

```sql
CREATE TABLE alerts (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code              VARCHAR(50) NOT NULL,
  location_id       UUID REFERENCES locations(id),
  state             VARCHAR(20) NOT NULL DEFAULT 'TRIGGERED'
                      CHECK (state IN ('TRIGGERED', 'ACKNOWLEDGED', 'RESOLVED')),
  entity_type       VARCHAR(50),
  entity_id         UUID,
  triggered_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  acknowledged_by   UUID REFERENCES users(id),
  acknowledged_at   TIMESTAMPTZ,
  resolved_by       UUID REFERENCES users(id),
  resolved_at       TIMESTAMPTZ,
  resolution_notes  TEXT,
  metadata          JSONB,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

### `alert_configs`
Per-location configurable alert thresholds.

```sql
CREATE TABLE alert_configs (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_id   UUID REFERENCES locations(id),  -- NULL = global default
  code          VARCHAR(50) NOT NULL,
  enabled       BOOLEAN NOT NULL DEFAULT true,
  threshold     JSONB,  -- e.g. { "hours": 24 } or { "count": 5 }
  updated_by    UUID REFERENCES users(id),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (location_id, code)
);
```

---

### `shifts`
Operator work periods with cash handover tracking.

```sql
CREATE TABLE shifts (
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  operator_id             UUID NOT NULL REFERENCES users(id),
  location_id             UUID NOT NULL REFERENCES locations(id),
  status                  VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                            CHECK (status IN ('OPEN', 'CLOSED', 'FLAGGED', 'RESOLVED', 'FORCE_CLOSED')),
  started_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at                TIMESTAMPTZ,
  expected_cash           NUMERIC(12,2),
  cash_handover_amount    NUMERIC(12,2),
  discrepancy             NUMERIC(12,2),
  discrepancy_notes       TEXT,
  force_closed_by         UUID REFERENCES users(id),
  force_closed_reason     TEXT,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shifts_operator ON shifts (operator_id);
CREATE INDEX idx_shifts_location ON shifts (location_id, started_at);
CREATE INDEX idx_shifts_status   ON shifts (status);
```

> **Note:** `sessions.shift_id UUID REFERENCES shifts(id)` is added to the sessions table to link each session to the shift during which it was created.

---

## 14.4 Key Constraints Summary

| Rule | Implementation |
|------|---------------|
| Audit logs are never deleted | REVOKE DELETE on `audit_logs` |
| Rate snapshot stored at billing | `sessions.rate_snapshot` JSONB |
| Receipt number is globally unique | UNIQUE on `transactions.receipt_number` |
| One transaction per session | UNIQUE on `transactions.session_id` |
| Operator must be assigned to location | Enforced at API layer via `user_locations` |
| Void is terminal | Enforced at API layer (no un-void) |
