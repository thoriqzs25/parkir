# PARKIR v2 — Automated Gate System PRD

**Version:** 2.0
**Status:** Draft
**Last Updated:** 2026-08-06

---

## 1. Problem Statement

AMB currently operates 20+ parking locations using a third-party automated ticketing system. The system works well but is expensive. AMB has requested a similar automated system at a lower monthly cost.

**Goal:** Build an excellent end-to-end automated parking management system for AMB, replacing their current expensive vendor.

**Business model:** Per-location monthly fee.

**First tenant:** AMB (single company, 20+ locations grouped by city).

---

## 2. System Overview

**Fully automated self-service parking system.** No operators at gates. Drivers self-service entry and exit. Server room staff handles exceptions only.

### Architecture

```
[Gate Hardware] ← hub → [Gate Desktop App (mini PC at gate)]
                                ↓ HTTP + mDNS (LAN)
                        [Server Room Desktop App (1 per location)]
                                ↓ sync every 1 min (internet)
                        [Cloud Backend (PARKIR)]
                                ↓
                        [AMB Admin Dashboard (web)]
```

### Components

| Component | Location | Role |
|-----------|----------|------|
| **Gate Desktop App** | Mini PC at each gate | Stateless hardware interface: controls printer, gate motor, QR scanner, payment terminal. Receives commands from server room app. |
| **Server Room Desktop App** | Server room, 1 per location | Business logic, local DB, fee calculation, config cache, transaction queue, sync to cloud. |
| **Cloud Backend** | Cloud server | Central database, config management, reporting, multi-location dashboard. |
| **AMB Admin Dashboard** | Web browser | Central management for all 20+ locations: reporting, config management, monitoring. |

### Gate Communication

- **Protocol:** HTTP + mDNS
- Gate app runs HTTP server, announces via mDNS (`parking-gate-{gate_id}.local:8080`)
- Server room app discovers gates via mDNS
- Server room app sends HTTP commands to gate apps
- Auto-recovery: mDNS handles re-discovery on restart, no explicit reconnection logic

---

## 3. Hardware Stack

### Entry Gate Hardware

| Hardware | Model | Communication | Notes |
|----------|-------|---------------|-------|
| Central Interface Hub | TBD | TBD | Connects: ticket button, vehicle loop sensor, gate motor, Epson printer |
| Vehicle Loop Sensor | Existing | Via hub | Detects vehicle presence |
| Ticket Button | Existing (physical) | Via hub | Driver presses to request ticket |
| Thermal Printer | Epson (existing AMB hardware) | Controlled by gate app | Prints QR ticket |
| Gate Motor | Existing | Via hub | Open/close |

### Exit Gate Hardware

| Hardware | Model | Communication | Notes |
|----------|-------|---------------|-------|
| QR Scanner | Panda PRJ-777 | USB HID (fixed-mount) | Sends decoded QR data as keyboard input |
| Payment Terminal | TBD (vendor-provided) | TBD | e-money + Flazz card reader. Vendor handles hardware + payment processing. |
| Driver-Facing Monitor | TBD | TBD | Displays fee, check-in time, duration, payment status |
| Alert Button | Physical | Via hub | Driver presses for staff assistance |

---

## 4. Core Workflows

### 4.1 Entry Flow

```
1. Vehicle enters loop sensor → Gate app detects vehicle
2. Driver presses ticket button
3. Gate app verifies: vehicle in loop?
   - NO → Ignore button press
   - YES → Continue
4. Gate app → Server room app: "Create session"
5. Server room app creates session in local DB
6. Server room app → Gate app: {session_id, location_id, timestamp, vehicle_type}
7. Gate app prints ticket with QR code containing:
   - session_id
   - location_id
   - timestamp (check-in time)
   - vehicle_type
   - Copy: "Kunci kendaraan anda dengan rapat. Jangan tinggalkan karcis parkir di dalam kendaraan Anda"
8. Gate opens → Vehicle enters → Loop sensor clears
```

**Gate configuration:** Vehicle type is fixed per gate (configured via dashboard). Example: "NORTH-ENTRY-01" = motorcycle only.

**Exceptions:**
- **Printer empty/jammed:** Gate stays closed. Display "Out of service" on monitor. Driver presses alert button → staff handles offline SOP or refills dispenser.
- **No vehicle in loop:** Button press ignored (prevents wasted tickets).

### 4.2 Exit Flow

```
1. Driver scans QR ticket at fixed-mount scanner (USB HID)
2. Gate app reads QR data → sends to server room app
3. Server room app:
   - Looks up session in local DB
   - Calculates fee using local rate config (cached from cloud)
   - Returns to gate app: {fee_amount, check_in_time, duration, vehicle_type}
4. Gate app displays on driver-facing monitor:
   - Fee amount
   - Check-in time
   - Duration
   - "Please tap your e-money card"
5. Driver taps e-money/Flazz card → Payment vendor processes
6. Payment vendor signals success to gate app
7. Gate app → Server room app: "Finalize session"
8. Server room app marks session as CLOSED, records transaction
9. Gate opens → Vehicle exits
```

**Exceptions:**
- **QR unreadable/damaged:** Gate stays closed. Driver presses alert button → staff runs offline SOP.
- **Payment failed (insufficient balance):** Gate closed. Display "Insufficient balance, please topup." If driver asks for help → staff pays with their own emoney (driver pays cash to staff).
- **Payment vendor down:** Staff runs offline SOP.
- **Session not found in local DB:** Should not happen. If it does → staff runs offline SOP.

### 4.3 Alert Flow

```
1. Driver presses alert button (physical) at gate
2. Gate app → Server room app (LAN): "Alert from gate {gate_id}"
3. Server room app plays audio alert: "Gate {gate_id} needs assistance"
4. Staff walks to gate, helps driver
5. If needed, staff runs offline SOP (staff already has SOP)
```

---

## 5. Gate Desktop App

**Runs on:** Mini PC at each gate (entry and exit).

**Role:** Stateless hardware interface. No business logic, no local database.

**Responsibilities:**
- Receive hardware events from hub (button press, QR scan, vehicle detected, alert button)
- Control hardware (printer, gate motor)
- Communicate with server room app via HTTP
- Display messages on driver-facing monitor (exit gate)

**Stateless design:**
- No local database
- No transaction logic
- Minimal persistent state: `gate_id` (hardware serial/MAC) + `server_room_url` (learned during bootstrap)
- All configuration comes from server room app

**Bootstrapping (new gate):**
1. Gate mini PC boots → Gate app announces via mDNS: `parking-gate-{gate_id}.local:8080`
2. Server room app discovers gate → registers in local DB as "unregistered"
3. Server room app syncs to cloud → cloud knows about unregistered gate
4. Dashboard shows "New gate detected: {gate_id}"
5. Admin configures gate via dashboard (vehicle type, gate type, location)
6. Config flows: Dashboard → Cloud → Server room app → Gate app
7. Gate app saves `server_room_url` locally (for reconnection after restart)
8. Gate is now operational

**Health check:** Server room app pings gate app every 15 seconds.

---

## 6. Server Room Desktop App

**Runs on:** 1 mini PC per location, in server room.

**Role:** Business logic + local DB + sync gateway.

**Responsibilities:**
- Run local DB (PostgreSQL or SQLite)
- Store all configs, sessions, transactions locally
- Calculate fees using local rate config
- Create sessions (entry) and finalize sessions (exit)
- Sync transactions to cloud (every 1 min + manual)
- Poll configs from cloud (every 1 min + manual)
- Manage gate connections (HTTP + mDNS)
- Play audio alerts for gate issues
- Health check gate apps (every 15s)
- Report status to cloud (every 20s)

**Local DB stores:**
- Gate configs (synced from cloud)
- Rate configs (synced from cloud, versioned)
- Shift configs (synced from cloud, versioned)
- Sessions (active + closed)
- Transactions (synced to cloud)
- Alerts/incidents
- Sync queue (unsynced transactions)

**Config versioning:**
- Every rate config has a version (`config_id_version`)
- Every shift config has a version
- Transactions record which config version was used (audit trail)
- If rates change while offline, use old rates until reconnected

**Sync behavior:**
- **Configs (cloud → local):** Poll cloud every 1 min. Local always overwritten by cloud. Manual refresh button.
- **Data (local → cloud):** Sync every 1 min. Retry with exponential backoff on failure. Alert developer on failure. Flag unsynced transactions. Manual refresh button.

**Offline operation (no internet):**
- Gates still operate using local DB
- Fee calculation uses local rate config
- Transactions queued locally
- Sync when internet restored

**Gate failure (LAN down):**
- Gate app stops working
- Audio alarm plays for staff
- Staff runs offline SOP

---

## 7. Cloud Backend

**Role:** Central source of truth for configs. Aggregated data for reporting.

**Responsibilities:**
- Store all configs (rates, shifts, gates, locations, users)
- Serve configs to server room apps
- Receive synced data from server room apps
- Aggregate data across all 20+ locations
- Provide API for dashboard
- Monitor server room app health (ping every 20s)

**Data received from server room apps (every 1 min):**
- Transactions (completed sessions)
- Sessions (active + closed)
- Alerts/incidents
- Gate status (online/offline)
- Config versions in use

**Health monitoring:**
- Ping server room app every 20s
- If server room app unhealthy:
  - Alert developer via email/Telegram
  - Show in dashboard
- Gate health reported by server room app (aggregated)

**API endpoints for dashboard:**
- Locations (grouped by city)
- Revenue reports (per location + total)
- Occupancy reports
- Transaction lists
- Gate status (live)
- Config management (rates, shifts, gates)
- User management
- Export (CSV/Excel)

---

## 8. AMB Admin Dashboard

**Platform:** Web browser (Next.js)

**Access:** AMB admin + leaders

**Features:**

### Live Monitoring
- All 20+ locations overview
- Gate status per location (online/offline/busy)
- Active sessions per location
- Revenue today (live)
- Alert counts

### Reports
- Daily revenue (per location + total across all locations)
- Occupancy rates
- Transaction counts
- Exception/alert counts
- Staff activity
- Grouped by city, filterable
- Export to Excel/CSV

### Configuration Management
- Locations (add/edit/deactivate)
- Rates per location (with versioning)
- Shift schedules per location
- Gate configurations (vehicle type, gate type)
- User accounts (staff, leaders)
- Manual refresh button for latest data

### Gate Management
- View all gates across all locations
- Configure new (unregistered) gates
- Gate status (online/offline)
- Gate health history

---

## 9. Offline Handling

### Server room app offline (no internet)
- Gates still operate using local DB
- Fee calculation uses local rate config
- Transactions queued in sync queue
- Sync to cloud when internet restored
- Retry with exponential backoff
- Alert developer if sync fails persistently

### Gate app offline (LAN down)
- Gate stops working
- Audio alarm for staff
- Staff runs offline SOP

### Payment vendor offline
- Staff runs offline SOP

### Config changes while offline
- Use old config until reconnected
- Transaction records which config version was used
- When reconnected, sync transactions with old config version (audit trail preserved)

---

## 10. Deployment & Infrastructure

### Installation
- **Server room app:** Manual USB install on mini PC
- **Gate app:** Manual USB install on gate mini PC
- **Updates:** Manual USB

### Provisioning new location
1. Install server room mini PC (USB)
2. Install gate mini PCs (USB)
3. Connect hardware (hub, sensors, printer, scanner, monitor, payment terminal)
4. Server room app boots, discovers gates via mDNS
5. Admin configures gates via dashboard
6. Location is operational

### Provisioning new gate
1. Install gate mini PC (USB)
2. Gate app boots, announces via mDNS
3. Server room app detects, registers as "unregistered"
4. Dashboard shows "New gate detected"
5. Admin configures gate via dashboard
6. Gate is operational (no deployment needed)

### Backup
- Local DB snapshot daily (server room app, stored in local dir)
- Cloud backup (synced transactions)

### Monitoring
- Server room app pings gate app every 15s
- Cloud pings server room app every 20s
- Gate unhealthy → audio alarm (staff)
- Server room unhealthy → email/Telegram (developer) + dashboard alert

### Uptime SLA
- 95% (hardware-dependent)

---

## 11. Payment Integration — TBD

**Vendor:** TBD

**PARKIR responsibility:**
- Display fee amount on driver-facing monitor
- Wait for payment completion signal from vendor
- Finalize session check-out on payment success
- Open gate

**TBD:**
- Payment terminal brand/model
- Communication protocol (how vendor signals completion)
- Payment failure handling (signal format)
- Receipt printing (vendor terminal or PARKIR printer?)

---

## 12. Open Questions

1. **What DB does server room app use?** PostgreSQL? SQLite?
2. **Gate app tech stack?** Electron? Tauri? Something else?
3. **Server room app tech stack?** Electron? Tauri? Native?
4. **QR code format?** Raw JSON? Encrypted? What encoding?
5. **Audio alert implementation?** Text-to-speech? Pre-recorded? Which OS audio API?
6. **Config version format?** Auto-increment integer? UUID? Timestamp?
7. **Sync protocol?** HTTP polling? WebSocket? gRPC?
8. **Dashboard tech stack?** Continue with Next.js?
9. **Cloud backend tech stack?** Continue with Go/Gin?
10. **Driver-facing monitor:** What resolution? Touchscreen? How does gate app render UI?
11. **Receipt printing:** Does payment vendor terminal print? Or does PARKIR need to print receipt?
12. **Multi-language:** Indonesian only? Or support English too?
13. **Audit logging:** What level of detail? (Already in current codebase)
14. **Shift management:** How does shift work with automated gates? (Already in current codebase)

---

## 13. Data Model (Draft)

### New/Modified Entities

**Gate**
```
{
  gate_id: string (hardware serial/MAC),
  location_id: string,
  gate_type: "entry" | "exit",
  vehicle_type: "car" | "motorcycle" | "truck" | "all",
  status: "unregistered" | "registered" | "operational",
  server_room_url: string,
  last_seen_at: timestamp,
  config: JSONB
}
```

**Session**
```
{
  session_id: UUID,
  location_id: string,
  gate_id: string (entry gate),
  vehicle_type: string,
  check_in_at: timestamp,
  check_out_at: timestamp (nullable),
  fee_amount: integer,
  rate_config_version: string,
  shift_config_version: string,
  state: "ACTIVE" | "PENDING_PAYMENT" | "CLOSED" | "VOIDED",
  qr_data: string (what was encoded in QR),
  synced_at: timestamp (nullable, when synced to cloud)
}
```

**Transaction**
```
{
  transaction_id: UUID,
  session_id: UUID,
  location_id: string,
  gate_id: string (exit gate),
  amount: integer,
  payment_method: "emoney" | "flazz" | "cash" | "offline_sop",
  payment_reference: string (nullable),
  config_version: string,
  created_at: timestamp,
  synced_at: timestamp (nullable),
  voided: boolean,
  voided_at: timestamp (nullable),
  voided_by: string (nullable),
  void_reason: string (nullable)
}
```

**Config Version**
```
{
  config_type: "rate" | "shift",
  config_id: string,
  version: string,
  effective_at: timestamp,
  data: JSONB,
  created_at: timestamp
}
```

---

## 14. Migration Plan

### Phase 1: Core Infrastructure
- Cloud backend: gate management, config versioning, sync API
- Server room app: local DB, gate discovery (mDNS), HTTP API, fee calculation, sync
- Gate app: hardware interface (printer, gate motor, QR scanner), HTTP server, mDNS

### Phase 2: Entry Flow
- Ticket dispensing (QR code generation + printing)
- Session creation
- Loop sensor integration

### Phase 3: Exit Flow
- QR scanning
- Fee calculation + display
- Session finalization

### Phase 4: Payment Integration
- Payment vendor integration (TBD)
- Driver-facing monitor UI

### Phase 5: Dashboard & Monitoring
- AMB admin dashboard (multi-location)
- Live monitoring
- Reports
- Gate bootstrapping UI

### Phase 6: Offline & Sync
- Offline operation
- Sync queue
- Config versioning
- Health checks + alerting

### Phase 7: Deployment
- Manual USB packaging
- Backup system
- Monitoring (email/Telegram alerts)

---

## 15. Differences from v1 (Current System)

| Aspect | v1 (Current) | v2 (Automated) |
|--------|--------------|----------------|
| Entry | Operator manually checks in | Driver self-service (ticket dispenser) |
| Exit | Operator manually checks out + collects payment | Driver self-service (QR scan + tap-to-pay) |
| Operator | Human at gate | No operator at gate |
| Desktop app | Manual operator POS | Gate app (stateless hardware interface) + Server room app (business logic) |
| Payment | Cash/digital via operator | e-money/Flazz via vendor terminal |
| Offline | Desktop app queues operations | Server room app has full local DB |
| Gate hardware | Not integrated | Full integration (printer, scanner, gate motor, loop sensor) |
| Multi-location | Dashboard only | Dashboard + server room app per location |
| Config management | Cloud only | Cloud → server room app (local cache with versioning) |

---

*End of PRD v2 — Automated Gate System*
