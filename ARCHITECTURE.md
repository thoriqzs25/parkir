# PARKIR Multi-Lane Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         LAN SWITCH                               │
│                      (Single Subnet)                             │
└─────────────────────────────────────────────────────────────────┘
         │        │        │        │              │
         │        │        │        │              │
    ┌────┴───┐ ┌──┴───┐ ┌──┴───┐ ┌──┴───┐   ┌────┴────┐
    │ Lane 1 │ │Lane 2│ │Lane 3│ │ ...  │   │ Lane 8  │
    │Mini PC │ │Mini PC│ │Mini PC│ │      │   │Mini PC  │
    └────────┘ └──────┘ └──────┘ └──────┘   └─────────┘
                                                      │
                                                      │ HTTP POST
                                                      ▼
                                              ┌──────────────┐
                                              │Main Desktop  │
                                              │  (Gateway)   │
                                              │              │
                                              │ ┌──────────┐ │
                                              │ │  SQLite  │ │
                                              │ │ (buffer) │ │
                                              │ └──────────┘ │
                                              └──────┬───────┘
                                                     │
                                                     │ Batch Upload (5 min)
                                                     ▼
                                              ┌──────────────┐
                                              │  Cloud API   │
                                              │  (Go + Pg)   │
                                              └──────────────┘
```

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Discovery | mDNS | Zero-config, works on single subnet, Go libraries available |
| Transport | HTTP POST | Simple, no broker dependency, existing Go backend stack |
| Reliability | Queue & retry | Bounded buffer, no data loss during outages |
| Main Desktop Role | Gateway/Proxy | Aggregates LAN traffic, forwards to cloud |
| Sync Strategy | Batch upload | Every 5 minutes, reduces API load |
| Local Storage | SQLite | Lightweight, zero-config, perfect for buffering |
| Cloud Unreachable | Keep buffering | No data loss, SQLite can hold millions of events |
| Dashboard Data | Local SQLite | Real-time, works offline, shows sync status |

## Component Details

### 1. Mini PCs (Lane Controllers)

**Hardware**:
- Gate/barrier controller (GPIO/serial)
- Camera (USB/IP, LPR/plate recognition)
- Display screen (optional)
- Mini PC (Raspberry Pi 4, Intel NUC, or similar)

**Software Stack**:
- Go service (lightweight, same language as backend)
- SQLite (local event queue)
- mDNS client (discover main desktop)
- HTTP client (event delivery)
- Retry logic with exponential backoff

**Behavior**:
1. Boot → discover main desktop via mDNS (`_parkir-gateway._tcp.local.`)
2. Capture event (entry/exit/gate status/heartbeat)
3. Store image locally (`/var/parkir/lane1/images/`)
4. Serialize to JSON, insert into SQLite queue
5. POST to main desktop REST API
6. On 200 OK → remove from queue
7. On failure → retry (5s → 10s → 20s → 40s → 80s, cap 5min)
8. Send heartbeat every 30s

**Event Schema**:
```typescript
interface ParkingEvent {
  event_id: string;        // UUID, for deduplication
  lane_id: string;         // "lane-1", "lane-2", etc.
  event_type: "ENTRY" | "EXIT" | "GATE_STATUS" | "HEARTBEAT";
  timestamp: string;       // ISO 8601 with timezone
  plate_number?: string;   // Text from LPR
  image_path?: string;     // Relative path on mini PC
  gate_status?: "OPEN" | "CLOSED" | "FAULT";
  metadata: Record<string, any>;
}
```

### 2. Main Desktop (Gateway + Dashboard)

**Software Stack**:
- Electron + React + TypeScript (dashboard UI)
- Go gateway service (receives events from mini PCs)
- SQLite (local buffer before cloud sync)
- Cloud sync worker (batch upload every 5 minutes)
- mDNS announcer

**Responsibilities**:
- Announce presence via mDNS
- Receive events from mini PCs via HTTP POST
- Store in local SQLite with `sync_status` (synced/unsynced)
- Batch upload to cloud every 5 minutes
- Mark events as "synced" after successful cloud upload
- Display real-time dashboard from local SQLite
- Show sync status per event (synced/unsynced indicator)

**Local SQLite Schema**:
```sql
CREATE TABLE events (
  id TEXT PRIMARY KEY,           -- UUID
  lane_id TEXT NOT NULL,
  event_type TEXT NOT NULL,      -- ENTRY, EXIT, GATE_STATUS, HEARTBEAT
  timestamp TEXT NOT NULL,
  plate_number TEXT,
  image_path TEXT,
  gate_status TEXT,
  metadata TEXT,                 -- JSON
  sync_status TEXT DEFAULT 'unsynced',  -- synced | unsynced | failed
  synced_at TEXT,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sync_status ON events(sync_status);
CREATE INDEX idx_timestamp ON events(timestamp);
```

**API Endpoints**:
```
POST /api/v1/gateway/events
  Headers: X-Lane-ID: lane-1
  Body: {
    "event_id": "uuid",
    "event_type": "ENTRY",
    "timestamp": "2026-08-02T14:23:45+07:00",
    "plate_number": "ABC123",
    "image_path": "/var/parkir/lane1/images/20260802_142345.jpg",
    "metadata": {}
  }
  Response: 200 OK { "ack": true }
```

**Cloud Sync Worker**:
```go
// Runs every 5 minutes
func syncToCloud() {
  events := db.Query("SELECT * FROM events WHERE sync_status = 'unsynced' LIMIT 1000")
  
  resp := cloudAPI.BatchUpload(events)
  if resp.OK {
    db.Exec("UPDATE events SET sync_status = 'synced', synced_at = NOW() WHERE id IN (?)", ids)
  } else {
    // Keep as unsynced, retry next cycle
    log.Warn("Cloud sync failed, will retry")
  }
}
```

**Dashboard Features**:
- Real-time lane status (active/inactive/fault)
- Recent events list (from local SQLite)
- Sync status indicator per event (green check = synced, orange clock = unsynced)
- Unsynced count warning if > 10,000 events

### 3. Cloud API (Existing Go Backend)

**New Endpoint**:
```
POST /api/v1/gateway/batch-events
  Headers: Authorization: Bearer <gateway-token>
  Body: {
    "gateway_id": "main-desktop-001",
    "events": [
      {
        "id": "uuid",
        "lane_id": "lane-1",
        "event_type": "ENTRY",
        "timestamp": "2026-08-02T14:23:45+07:00",
        "plate_number": "ABC123",
        "image_path": "/var/parkir/lane1/images/20260802_142345.jpg",
        "metadata": {}
      },
      // ... up to 1000 events
    ]
  }
  Response: 200 OK { "accepted": 1000 }
```

**Cloud Database**:
- Same schema as local SQLite, but PostgreSQL
- Stores all events from all sites (if multi-site in future)
- Source of truth for billing/reporting

## Data Flow

```
┌─────────────┐
│   Camera    │──── plate scan ────┐
└─────────────┘                    │
                                   ▼
┌─────────────┐              ┌──────────────┐
│ Gate Ctrl   │──── status ──┤  Mini PC     │
└─────────────┘              │  (Lane 1)    │
                             └──────────────┘
                                    │
                                    ├─ 1. Save image locally
                                    ├─ 2. Create event JSON
                                    ├─ 3. Insert into SQLite (sync_status=unsynced)
                                    ├─ 4. POST to main desktop
                                    │
                                    ▼
                             ┌──────────────┐
                             │Main Desktop  │
                             │  (Gateway)   │
                             │              │
                             │  ┌────────┐  │
                             │  │ SQLite │  │
                             │  │(buffer)│  │
                             │  └────────┘  │
                             └──────┬───────┘
                                    │
                                    ├─ 5. Store in local SQLite
                                    ├─ 6. Return 200 OK to mini PC
                                    ├─ 7. Every 5 min: batch upload to cloud
                                    ├─ 8. Mark as synced in SQLite
                                    │
                                    ▼
                             ┌──────────────┐
                             │  Dashboard   │
                             │  (Electron)  │
                             │              │
                             │  Shows:      │
                             │  - Live events
                             │  - Sync status
                             └──────────────┘
```

## Failure Scenarios

| Scenario | Behavior |
|----------|----------|
| **Mini PC → Main Desktop fails** | Mini PC retries with backoff, queue grows in mini PC SQLite |
| **Main Desktop offline** | Mini PCs buffer locally, reconnect when main desktop back |
| **Internet down (cloud unreachable)** | Main desktop keeps buffering in local SQLite, sync status stays "unsynced" |
| **Cloud sync fails** | Events stay "unsynced", retry next 5-min cycle |
| **Main Desktop restarts** | SQLite persists, resumes sync worker, replays unsynced events |
| **Mini PC restarts** | SQLite persists, resumes queue processing |
| **Duplicate events** | Cloud API uses `event_id` (UUID) for idempotency |

## Discovery Protocol

**mDNS Service**:
- Main desktop announces: `_parkir-gateway._tcp.local.` → port 8080
- Mini PCs query on boot, cache result
- Re-query if HTTP connection fails

**POC Simplification**:
- Skip mDNS for POC, hardcode `http://localhost:8080`
- Add mDNS in production deployment

## POC on Single Machine (MacBook)

**Architecture**:
```
┌─────────────────────────────────────────────┐
│              MacBook (localhost)             │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │ Main Desktop (Electron + Go Gateway) │  │
│  │   - Port 8080 (gateway API)          │  │
│  │   - Port 3000 (dashboard)            │  │
│  │   - SQLite (local buffer)            │  │
│  └──────────────────────────────────────┘  │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │ Mini PC Simulators (8 containers)    │  │
│  │   - lane-1 through lane-8            │  │
│  │   - Each with unique lane_id         │  │
│  │   - POST to http://localhost:8080    │  │
│  └──────────────────────────────────────┘  │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │ PostgreSQL (cloud simulation)        │  │
│  │   - Port 5432                        │  │
│  │   - Database: parkir_cloud           │  │
│  └──────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

**Docker Compose**:
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: parkir_cloud
      POSTGRES_USER: parkir
      POSTGRES_PASSWORD: parkir
    ports:
      - "5432:5432"
  
  mini-pc-1:
    build: ./mini-pc-simulator
    environment:
      LANE_ID: lane-1
      GATEWAY_URL: http://host.docker.internal:8080
      SIMULATION_SPEED: 1  # events per minute
  
  mini-pc-2:
    build: ./mini-pc-simulator
    environment:
      LANE_ID: lane-2
      GATEWAY_URL: http://host.docker.internal:8080
      SIMULATION_SPEED: 1
  
  # ... repeat for mini-pc-3 through mini-pc-8
```

**POC Success Criteria**:
1. ✅ Main desktop gateway receives events from 8 mini PC simulators
2. ✅ Events stored in local SQLite with sync_status
3. ✅ Batch sync worker uploads to cloud PostgreSQL every 5 minutes
4. ✅ Dashboard displays live events with sync status indicators
5. ✅ Kill cloud DB → restart → main desktop resumes sync, no data loss
6. ✅ Simulate internet outage → events buffer locally → sync when back

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Mini PC | Go + SQLite + HTTP client |
| Main Desktop | Electron + React + TypeScript + Go gateway |
| Local Storage | SQLite (main desktop + mini PCs) |
| Cloud API | Go + Gin + PostgreSQL (existing) |
| Discovery | mDNS (production), hardcoded (POC) |
| Transport | HTTP POST + JSON |

## Implementation Order

1. **Mini PC Simulator** - Docker container that generates fake events
2. **Main Desktop Gateway** - Go service that receives events + SQLite buffer
3. **Cloud Sync Worker** - Batch upload to cloud PostgreSQL
4. **Dashboard UI** - Electron app showing events + sync status

## Scale Considerations

**Current**: 8 mini PCs, single subnet, ~10 MB/day data volume

**Extensibility**:
- mDNS works up to ~50 devices on single subnet
- For multi-subnet: add discovery relay or static IP config
- For 100+ devices: consider load balancing or sharding
- Image storage: implement retention policy (delete after 30 days)
