# Chapter 13 — System Observability

## 13.1 Overview

The observability module gives administrators and managers real-time awareness of system health, a complete and immutable audit trail of every action, automated anomaly detection with alerting, and per-operator performance metrics. All observability features are accessible from the web dashboard.

---

## 13.2 System Health Dashboard

### Purpose
Provide a live view of the operational status of every system component, so managers and admins can detect and respond to issues quickly.

### Components Monitored

| Component | Description | Check Method |
|-----------|-------------|-------------|
| Backend API | Core application server | HTTP health endpoint (`/health`) |
| Database | Primary database connectivity | DB ping from API |
| Payment Gateway | Digital payment service availability | Gateway status endpoint or probe |
| Printer (per terminal) | Thermal printer connectivity at each operator workstation | Heartbeat from desktop app |

### Status Levels

| Status | Color | Description |
|--------|-------|-------------|
| `HEALTHY` | Green | Component is operating normally |
| `DEGRADED` | Yellow | Component is reachable but showing elevated errors or latency |
| `DOWN` | Red | Component is unreachable or returning critical errors |
| `UNKNOWN` | Grey | No recent health data available |

### Dashboard Layout

```
System Health                           Last updated: 14:32:01

┌──────────────────┬────────────┬──────────────┬──────────────┐
│  Backend API     │  Database  │  Pay Gateway │  Printers    │
│  ● HEALTHY       │  ● HEALTHY │  ● DEGRADED  │  3/4 ONLINE  │
└──────────────────┴────────────┴──────────────┴──────────────┘

Printer Details:
  Terminal A (Gate 1)   ● ONLINE
  Terminal B (Gate 2)   ● ONLINE
  Terminal C (Gate 3)   ● ONLINE
  Terminal D (Gate 4)   ✕ OFFLINE  [Last seen: 13:55]
```

### Health Check Frequency
- Backend API & Database: polled every 30 seconds.
- Payment Gateway: polled every 60 seconds.
- Printer status: reported by desktop app heartbeat every 60 seconds.

### Status Update SLA
- Dashboard must reflect a component going DOWN within 60 seconds of detection.

---

## 13.3 Audit Log

### Purpose
Maintain a complete, immutable, and queryable record of every state-changing action in the system. Used for accountability, investigation, and compliance.

### Audit Log Principles
- **Immutable:** Audit log entries can never be modified or deleted by any user, including system admins.
- **Complete:** Every action that changes system state must produce an audit log entry.
- **Contextual:** Each entry captures enough context to reconstruct what happened without querying other tables.

### Audited Actions

| Module | Actions Logged |
|--------|---------------|
| Sessions | SESSION_CREATED, SESSION_CHECKEDOUT, SESSION_CLOSED, SESSION_VOIDED, SESSION_REASSIGNED |
| Transactions | TRANSACTION_CREATED, TRANSACTION_VOIDED |
| Incidents | INCIDENT_FILED, INCIDENT_UPDATED, INCIDENT_RESOLVED |
| Users | USER_CREATED, USER_UPDATED, USER_DEACTIVATED, USER_LOGIN, USER_LOGIN_FAILED |
| Roles | ROLE_CREATED, ROLE_UPDATED, ROLE_DELETED |
| Locations | LOCATION_CREATED, LOCATION_UPDATED, LOCATION_DEACTIVATED |
| Rates | RATE_CREATED, RATE_UPDATED |
| Alerts | ALERT_TRIGGERED, ALERT_RESOLVED, ALERT_CONFIG_UPDATED |
| Adjustments | TRANSACTION_VOIDED, SESSION_REASSIGNED |
| Offline Sync | OFFLINE_SESSION_SYNCED, SYNC_CONFLICT_FLAGGED |

### Audit Log Entry Structure

```json
{
  "id": "uuid",
  "action": "SESSION_CLOSED",
  "actor_id": "user-uuid",
  "actor_role": "operator",
  "entity_type": "session",
  "entity_id": "session-uuid",
  "location_id": "location-uuid",
  "ip_address": "192.168.1.10",
  "metadata": {
    "plate": "B 1234 XYZ",
    "vehicle_type": "CAR",
    "fee_amount": 20000,
    "payment_method": "CASH"
  },
  "timestamp": "2025-03-15T14:32:00Z"
}
```

### Audit Log Query Interface (Web Dashboard)

**Filters:**
| Filter | Options |
|--------|---------|
| Date range | Custom range picker |
| Action | Multi-select from action list |
| Actor | Search by user name |
| Entity type | Session / Transaction / Incident / User / Role / etc. |
| Location | Single or all |

**Display columns:**
| Column | Description |
|--------|-------------|
| Timestamp | UTC, displayed in local timezone |
| Action | Human-readable label |
| Actor | User name + role |
| Entity | Type + ID (clickable link to detail view) |
| Location | |
| Details | Expandable metadata preview |

**Retention:** Minimum 2 years. Logs are never deleted.

### Data Model

```
audit_logs
  id                UUID, primary key
  action            VARCHAR(100), not null
  actor_id          UUID, FK → users.id (nullable if system action)
  actor_role        VARCHAR(50)
  entity_type       VARCHAR(50), not null
  entity_id         UUID, not null
  location_id       UUID, FK → locations.id, nullable
  ip_address        INET, nullable
  metadata          JSONB
  timestamp         TIMESTAMP WITH TIME ZONE, not null

-- No UPDATE or DELETE permissions granted on this table for any application role.
-- Append-only enforced at database level (row-level security or trigger).
```

---

## 13.4 Anomaly Alerting

### Purpose
Automatically detect unusual or problematic patterns and surface them to the right people before they become larger issues.

### Alert Rules (Default Configuration)

| Alert Code | Condition | Default Threshold | Target |
|-----------|-----------|-----------------|--------|
| `UNPAID_EXIT` | Session closed without a linked transaction | Any occurrence | Manager |
| `LONG_SESSION` | Session remains `ACTIVE` beyond threshold | > 24 hours | Manager |
| `HIGH_VOID_RATE` | Number of voids by a single operator in a day | > 5 per operator/day | Admin |
| `SYNC_FAILURE` | Offline session fails to sync after reconnect | Any occurrence | Admin |
| `GATEWAY_FAILURE` | Payment gateway returns error consecutively | > 2 consecutive failures | Admin |

### Alert Configuration
- Thresholds are configurable by admins with `observability:manage_alerts` permission.
- Each alert rule can be enabled or disabled per location.
- Custom alert rules are out of scope for v1 (only the above 5 are supported).

### Alert Lifecycle

```
Condition detected
        │
        ▼
  Alert TRIGGERED
  - Stored in alerts table
  - In-app notification sent to target role at the location
        │
        │  Target user reviews and takes action
        ▼
  Alert ACKNOWLEDGED  (optional intermediate state)
        │
        ▼
  Alert RESOLVED
  - Resolved by user with resolution notes
  - Or auto-resolved if condition clears (e.g. session closed)
```

### Alert Display (Web Dashboard)

- Alert badge in the navigation bar showing count of unacknowledged alerts.
- Alerts panel: list of all active alerts with type, location, triggered time, and detail.
- Click an alert to view details and linked entity (session, operator, etc.).

### Data Model

```
alerts
  id                UUID, primary key
  code              VARCHAR(50), not null
  location_id       UUID, FK → locations.id
  state             ENUM('TRIGGERED', 'ACKNOWLEDGED', 'RESOLVED'), default 'TRIGGERED'
  entity_type       VARCHAR(50), nullable  -- e.g. 'session', 'operator'
  entity_id         UUID, nullable
  triggered_at      TIMESTAMP WITH TIME ZONE, not null
  acknowledged_by   UUID, FK → users.id, nullable
  acknowledged_at   TIMESTAMP WITH TIME ZONE, nullable
  resolved_by       UUID, FK → users.id, nullable
  resolved_at       TIMESTAMP WITH TIME ZONE, nullable
  resolution_notes  TEXT, nullable
  metadata          JSONB  -- alert-specific context
  created_at        TIMESTAMP

alert_configs
  id                UUID, primary key
  location_id       UUID, FK → locations.id, nullable  -- null = global default
  code              VARCHAR(50), not null
  enabled           BOOLEAN, default true
  threshold         JSONB  -- e.g. { "hours": 24 } or { "count": 5 }
  updated_by        UUID, FK → users.id
  updated_at        TIMESTAMP
```

---

## 13.5 Per-Operator Performance Metrics

### Purpose
Give managers a quantitative view of each operator's activity for performance review and accountability.

### Metrics Available

| Metric | Description |
|--------|-------------|
| Total Check-ins | Count of `SESSION_CREATED` actions |
| Total Check-outs | Count of `SESSION_CLOSED` actions |
| Total Revenue Collected | Sum of `fee_amount` on closed non-voided transactions |
| Cash Collected | Sum where `payment_method = CASH` |
| Digital Collected | Sum where `payment_method = DIGITAL` |
| Avg Sessions per Hour | Check-outs / active hours |
| Incidents Filed | Count of incidents reported by operator |
| Voids Involved In | Count of voided sessions where operator was original or reassigned |
| Sessions Reassigned Away | Count of sessions reassigned off this operator |

### Filters
- Date range
- Location (if operator is assigned to multiple)

### Display
- Summary cards per operator.
- Ranking table: operators sorted by total sessions or revenue (configurable).
- Drill-down: click operator to see full activity timeline (same as Operator Activity Log in Chapter 10).
