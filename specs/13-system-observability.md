# Chapter 13 — System Observability

## 13.1 Overview

The observability system provides real-time awareness of system health across all three tiers (gate apps, server room apps, cloud backend), a complete and immutable audit trail, centralized logging, metrics collection, and automated alerting. The monitoring stack uses **Loki** (logs), **Prometheus** (metrics), and **Grafana** (visualization).

---

## 13.2 Monitoring Stack

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Monitoring Stack                          │
│                                                              │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────┐  │
│  │  Loki    │    │  Prometheus  │    │  Grafana         │  │
│  │  (Logs)  │    │  (Metrics)   │    │  (Visualization) │  │
│  └────┬─────┘    └──────┬───────┘    └────────┬─────────┘  │
│       │                 │                      │            │
│       └─────────────────┴──────────────────────┘            │
│                     ▲         ▲              ▲              │
│                     │         │              │              │
│  ┌──────────────────┴───┐  ┌─┴──────────┐  ┌┴──────────┐  │
│  │  Promtail (Agent)    │  │  Exporters │  │  Alerts   │  │
│  │  - Gate apps         │  │  - node    │  │  - Email  │  │
│  │  - Server room apps  │  │  - postgres│  │  - Telegram│ │
│  │  - Cloud backend     │  │  - custom  │  │           │  │
│  └──────────────────────┘  └────────────┘  └───────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Components

| Component | Purpose | Location |
|-----------|---------|----------|
| **Loki** | Log aggregation | Cloud server |
| **Promtail** | Log collector (agent) | All servers (cloud, server room) |
| **Prometheus** | Metrics collection | Cloud server |
| **node_exporter** | System metrics (CPU, memory, disk) | All servers |
| **postgres_exporter** | Database metrics | Cloud server |
| **Custom exporters** | Application metrics | Cloud backend, server room apps |
| **Grafana** | Visualization and dashboards | Cloud server |
| **Alertmanager** | Alert routing (email, Telegram) | Cloud server |

---

## 13.3 Logging

### Log Levels

| Level | Description | Example |
|-------|-------------|---------|
| **ERROR** | Critical errors requiring immediate attention | Database connection failed, payment processing error |
| **WARN** | Warnings that may indicate problems | High memory usage, slow query, sync retry |
| **INFO** | Informational messages about system operations | Session created, transaction synced, gate connected |
| **DEBUG** | Detailed debugging information (disabled in production) | Request/response payloads, internal state |

### Log Format

All logs use structured JSON format:

```json
{
  "timestamp": "2026-08-06T10:30:00Z",
  "level": "INFO",
  "service": "server-room-app",
  "location_id": "uuid",
  "gate_id": "GATE-ENTRY-01",
  "message": "Session created",
  "session_id": "uuid",
  "duration_ms": 45
}
```

### Log Sources

| Source | Logs To | Collection |
|--------|---------|------------|
| **Gate App** | Loki | Promtail (on server room PC) or direct to Loki |
| **Server Room App** | Loki | Promtail (on server room PC) |
| **Cloud Backend** | Loki | Promtail (on cloud server) |
| **Nginx** | Loki | Promtail (on cloud server) |
| **PostgreSQL** | Loki (optional) | Promtail (on cloud server) |

### Log Retention

- **Cloud backend:** 30 days (configurable)
- **Server room apps:** 7 days (local), 30 days (Loki)
- **Gate apps:** Forwarded to server room app or Loki directly

---

## 13.4 Metrics

### Metrics Collection

**Prometheus** scrapes metrics from all components:

| Target | Endpoint | Frequency |
|--------|----------|-----------|
| Cloud Backend | `:8080/metrics` | 15s |
| Server Room App | `:8080/metrics` (via internet) | 15s |
| node_exporter (cloud) | `:9100/metrics` | 15s |
| node_exporter (server room) | `:9100/metrics` (via server room relay) | 15s |
| postgres_exporter | `:9187/metrics` | 15s |

### Key Metrics

#### Cloud Backend
- `http_requests_total` — Total HTTP requests (labels: method, path, status)
- `http_request_duration_seconds` — Request duration histogram
- `db_query_duration_seconds` — Database query duration
- `db_connections_active` — Active database connections
- `sync_transactions_total` — Total transactions synced from server room apps
- `sync_duration_seconds` — Sync operation duration

#### Server Room App
- `gate_health_checks_total` — Total gate health checks (labels: gate_id, status)
- `gate_response_time_seconds` — Gate app response time
- `sync_queue_length` — Number of unsynced transactions
- `sync_retries_total` — Total sync retries
- `fee_calculations_total` — Total fee calculations
- `fee_calculation_duration_seconds` — Fee calculation duration
- `config_poll_duration_seconds` — Config poll duration

#### Gate App (Relayed via Server Room App)
- `gate_entry_operations_total` — Total entry operations
- `gate_exit_operations_total` — Total exit operations
- `gate_hardware_errors_total` — Hardware errors (labels: type)
- `gate_uptime_seconds` — Gate app uptime

#### System (node_exporter)
- `node_cpu_seconds_total` — CPU usage
- `node_memory_available_bytes` — Available memory
- `node_filesystem_avail_bytes` — Available disk space
- `node_network_receive_bytes_total` — Network traffic

#### Database (postgres_exporter)
- `pg_stat_activity_count` — Active connections
- `pg_stat_database_tup_fetched` — Rows fetched
- `pg_stat_database_tup_inserted` — Rows inserted
- `pg_stat_bgwriter_buffers_checkpoint` — Buffers checkpointed

### Grafana Dashboards

#### System Overview
- CPU, memory, disk usage (all servers)
- Network traffic
- Database connections and query performance

#### Backend API
- Request rate, latency, error rate
- Endpoint breakdown
- Database performance

#### Server Room Apps
- Gate health status (per location)
- Sync queue length and retry rate
- Fee calculation performance
- Config poll success rate

#### Gate Apps
- Entry/exit operation rate
- Hardware error rate
- Uptime (per gate)

#### Sync Monitoring
- Transactions synced per minute
- Sync queue length over time
- Sync failure rate
- Retry attempts

---

## 13.5 Health Checks

### Cloud Backend Health

**Endpoint:** `GET /health/ready`

**Check:**
- Backend API is running.
- Database connection is healthy (`SELECT 1`).

**Response:**
```json
{
  "status": "ok",
  "database": "connected",
  "uptime_seconds": 86400
}
```

**Monitoring:**
- Prometheus scrapes `/health/ready` every 15s.
- Alert if status != "ok" for 5 minutes.

### Server Room App Health

**Reporting:**
- Server room app reports status to cloud every 20 seconds.
- Includes: gate statuses (online/offline), sync queue length, uptime.

**Cloud Monitoring:**
- Cloud tracks server room app last_seen timestamp.
- Alert if last_seen > 60 seconds (server room app down).

### Gate App Health

**Health Check:**
- Server room app pings each gate app every 15 seconds.
- Gate app responds with status (online, hardware status).

**Monitoring:**
- Server room app tracks gate last_seen timestamp.
- If gate last_seen > 30 seconds → gate offline.
- Alert: audio alarm in server room ("Gate {gate_id} needs assistance").
- Cloud dashboard shows gate status (via server room app reporting).

### Status Levels

| Status | Color | Description |
|--------|-------|-------------|
| **HEALTHY** | Green | Component is operating normally |
| **DEGRADED** | Yellow | Component is reachable but showing elevated errors or latency |
| **DOWN** | Red | Component is unreachable or returning critical errors |
| **UNKNOWN** | Grey | No recent health data available |

---

## 13.6 Alerting

### Alert Channels

| Channel | Purpose | Recipients |
|---------|---------|------------|
| **Audio Alarm** | Gate hardware failures, exceptions | Server room staff (on-site) |
| **Email** | Critical system failures, sync failures | Developers |
| **Telegram** | Critical system failures, sync failures | Developers |
| **Dashboard** | All alerts (visual) | AMB admins, managers |

### Alert Rules

#### Critical (Immediate Action)

| Alert | Condition | Channel | Action |
|-------|-----------|---------|--------|
| Cloud backend down | `/health/ready` fails for 5 min | Email + Telegram | Developer investigates |
| Server room app down | last_seen > 60s | Email + Telegram | Developer investigates |
| Database down | Connection fails for 2 min | Email + Telegram | Developer investigates |
| Disk space low | < 10% free space | Email + Telegram | Developer investigates |

#### Warning (Monitor Closely)

| Alert | Condition | Channel | Action |
|-------|-----------|---------|--------|
| High CPU usage | > 80% for 5 min | Dashboard | Monitor, investigate if persistent |
| High memory usage | > 90% for 5 min | Dashboard | Monitor, investigate if persistent |
| Sync queue growing | > 100 unsynced transactions | Dashboard | Monitor, check internet connectivity |
| Sync failures | > 10 failures in 1 hour | Email + Telegram | Developer investigates |

#### Gate Alerts (On-Site)

| Alert | Condition | Channel | Action |
|-------|-----------|---------|--------|
| Gate hardware failure | Gate app reports error | Audio alarm | Staff walks to gate, handles via offline SOP |
| Gate offline | Gate last_seen > 30s | Audio alarm | Staff investigates |
| Printer jammed | Printer error | Audio alarm | Staff refills/fixes printer |
| QR scanner error | Scanner error | Audio alarm | Staff handles via offline SOP |

### Audio Alert Format

```
"Gate {gate_id} needs assistance"
```

Example:
```
"Gate GATE-EXIT-01 needs assistance"
```

**Implementation:**
- Text-to-speech (TTS) or pre-recorded audio.
- Played on server room PC speakers.
- Volume configurable.

---

## 13.7 Audit Log

### Purpose
Maintain a complete, immutable, and queryable record of every state-changing action in the system. Used for accountability, investigation, and compliance.

### Audit Log Principles
- **Immutable:** Audit log entries can never be modified or deleted by any user.
- **Complete:** Every action that changes system state must produce an audit log entry.
- **Contextual:** Each entry captures enough context to reconstruct what happened.

### Audited Actions

| Module | Actions Logged |
|--------|---------------|
| Sessions | SESSION_CREATED, SESSION_CLOSED, SESSION_VOIDED |
| Transactions | TRANSACTION_CREATED, TRANSACTION_VOIDED |
| Gates | GATE_REGISTERED, GATE_CONFIGURED, GATE_DEACTIVATED |
| Incidents | INCIDENT_FILED, INCIDENT_RESOLVED |
| Users | USER_CREATED, USER_UPDATED, USER_DEACTIVATED, USER_LOGIN, USER_LOGIN_FAILED |
| Roles | ROLE_CREATED, ROLE_UPDATED, ROLE_DELETED |
| Locations | LOCATION_CREATED, LOCATION_UPDATED, LOCATION_DEACTIVATED |
| Rates | RATE_CREATED, RATE_UPDATED |
| Shifts | SHIFT_CONFIG_CREATED, SHIFT_CONFIG_UPDATED |
| Alerts | ALERT_TRIGGERED, ALERT_ACKNOWLEDGED, ALERT_RESOLVED |
| Configs | CONFIG_SYNCED (cloud → server room app) |

### Audit Log Entry

```json
{
  "id": "uuid",
  "user_id": "uuid",  // null for system actions
  "action": "SESSION_CREATED",
  "entity_type": "session",
  "entity_id": "uuid",
  "location_id": "uuid",
  "metadata": {
    "gate_id": "GATE-ENTRY-01",
    "vehicle_type": "motorcycle",
    "shift_number": 15
  },
  "ip_address": "192.168.1.100",
  "created_at": "2026-08-06T10:30:00Z"
}
```

### Audit Log Retention
- Minimum 2 years.
- Non-deletable (even by system admins).
- Exportable to CSV (for compliance).

---

## 13.8 Dashboard Integration

### System Health Page

**Layout:**
```
System Health                           Last updated: 14:32:01

┌──────────────────┬────────────────┬──────────────────┐
│  Cloud Backend   │  Server Rooms  │  Gates           │
│  ● HEALTHY       │  20/20 ONLINE  │  38/40 ONLINE    │
└──────────────────┴────────────────┴──────────────────┘

Cloud Backend:
  API               ● HEALTHY
  Database          ● HEALTHY

Server Room Apps:
  Location A        ● HEALTHY  [Last seen: 14:32:01]
  Location B        ● HEALTHY  [Last seen: 14:32:00]
  Location C        ✕ DOWN     [Last seen: 14:30:15]

Gates (Location A):
  GATE-ENTRY-01     ● ONLINE   [Last seen: 14:31:55]
  GATE-EXIT-01      ● ONLINE   [Last seen: 14:31:50]
  GATE-ENTRY-02     ✕ OFFLINE  [Last seen: 14:28:30]
```

### Alert History Page

- List of all alerts (triggered, acknowledged, resolved).
- Filter by: location, gate, severity, date range.
- Click alert → view details (gate_id, error, timestamp, resolution).

### Audit Log Page

- List of all audit log entries.
- Filter by: action, user, entity_type, location, date range.
- Export to CSV.

---

## 13.9 Design Decisions

**Why Loki + Prometheus + Grafana?**
- Industry standard (widely used, well-documented).
- Loki for logs (lightweight, integrates with Grafana).
- Prometheus for metrics (pull-based, reliable).
- Grafana for visualization (powerful, flexible).

**Why not ELK (Elasticsearch, Logstash, Kibana)?**
- Loki is lighter weight (lower resource usage).
- Loki integrates better with Prometheus/Grafana.
- Simpler to deploy and maintain.

**Why audio alerts for gate failures?**
- Staff is on-site (server room).
- Immediate awareness (no need to check dashboard).
- Faster response time.

**Why email + Telegram for critical alerts?**
- Email: formal record, easier to track.
- Telegram: instant notification (mobile), faster response.
- Redundancy (if one fails, other works).

**Why gate metrics relayed via server room app?**
- Gate apps are on LAN (not directly accessible from cloud).
- Server room app has internet access.
- Server room app collects gate metrics (via health checks) and relays to cloud.

**Why 15s health check interval?**
- Balance between responsiveness and network overhead.
- 15s is fast enough to detect failures quickly.
- Not so fast that it creates network congestion.

---

*End of Chapter 13 — System Observability (v2)*
