# Chapter 9 — Platform & Interface

## 9.1 Overview

The system has three distinct interfaces serving different purposes in a three-tier architecture:

| Interface | Target Users | Platform | Primary Use |
|-----------|-------------|----------|-------------|
| **Gate Desktop App** | Drivers (indirect) | Mini PC at gate | Hardware control, ticket dispensing, payment processing |
| **Server Room Desktop App** | Server room staff | Mini PC in server room | Local DB, fee calculation, gate management, sync to cloud |
| **Web Dashboard** | AMB admins, managers | Browser | Multi-location management, reporting, configuration, monitoring |

All three interfaces communicate via HTTP APIs. Gate app and server room app run on local LAN. Dashboard and cloud backend run on internet.

---

## 9.2 Gate Desktop App

### 9.2.1 Platform

- **Target OS:** Ubuntu 22.04 LTS (lightweight Linux).
- **Tech:** **Tauri** (preferred) or **Electron** — web-based desktop app packaged as lightweight binary.
- **Hardware:** Mini PC at each gate (entry or exit).
- **Screen:** Driver-facing monitor (exit gate only) or no monitor (entry gate).
- **Input:** Hardware events (button press, QR scan, vehicle detection) via USB/serial.
- **Output:** Hardware control (printer, gate motor) via USB/serial.
- **Network:** LAN only (communicates with server room app via HTTP + mDNS).
- **Stateless:** No local database, no business logic. Receives commands from server room app.

### 9.2.2 Core Functionality

#### Entry Gate
- **Vehicle Detection:** Listen to loop sensor (via hub).
- **Button Press:** Listen to ticket button (via hub).
- **Session Creation:** Call server room app to create session.
- **Ticket Printing:** Generate QR code, print via Epson thermal printer.
- **Gate Control:** Open/close gate motor (via hub).
- **Alert:** Listen to alert button, send alert to server room app.

#### Exit Gate
- **QR Scanning:** Read QR code via USB HID scanner (Panda PRJ-777).
- **Fee Display:** Show fee, check-in time, duration on driver-facing monitor.
- **Payment Processing:** Interface with payment terminal (e-money/Flazz).
- **Receipt Printing:** Optional receipt on button press.
- **Gate Control:** Open/close gate motor (via hub).
- **Alert:** Listen to alert button, send alert to server room app.

### 9.2.3 User Interface (Driver-Facing Monitor)

#### Entry Gate (No Monitor or Simple Status)
- Optional: "Press button for ticket" display.
- No complex UI (driver interaction is minimal).

#### Exit Gate (Driver-Facing Monitor)
- **Idle Screen:** "Scan your ticket" (large text, QR scanner icon).
- **Fee Display Screen:**
  - Fee amount (large, bold).
  - Check-in time.
  - Duration (e.g., "2 hours 15 minutes").
  - Vehicle type.
  - Instruction: "Please tap your e-money card".
- **Payment Processing Screen:** "Processing payment..." (loading spinner).
- **Payment Success Screen:** "Payment successful" (checkmark, green), "Gate opening...".
- **Payment Failed Screen:** "Payment failed" (X, red), "Insufficient balance, please topup".
- **Error Screen:** "Out of service" (warning, yellow), "Please press alert button for help".
- **Receipt Screen:** (optional) Receipt printed on button press.

### 9.2.4 Hardware Abstraction Layer (HAL)

Gate app uses HAL interfaces to support different hardware vendors:

```typescript
interface Printer {
  printTicket(data: TicketData): Promise<void>;
  printReceipt(data: ReceiptData): Promise<void>;
}

interface GateMotor {
  open(): Promise<void>;
  close(): Promise<void>;
  isOpen(): boolean;
}

interface Scanner {
  readQR(): Promise<string>;
}

interface LoopSensor {
  isVehiclePresent(): boolean;
}

interface Button {
  waitForPress(): Promise<ButtonEvent>;
}
```

**Implementations:**
- Epson thermal printer (ESC/POS commands via USB/serial).
- Gate motor (via hub, protocol TBD).
- QR scanner (USB HID, read keyboard input).
- Loop sensor (via hub).
- Ticket button (via hub).
- Alert button (via hub).

### 9.2.5 Communication

**Protocol:** HTTP + mDNS (LAN)

- **mDNS Announcement:** Gate app announces itself as `parking-gate-{gate_id}.local:8080`.
- **HTTP Server:** Gate app runs HTTP server (receives commands from server room app).
- **HTTP Client:** Gate app calls server room app (create session, close session, send alert).

**Endpoints (Gate App HTTP Server):**
- `GET /info` — Return gate info (gate_id, gate_type, vehicle_type).
- `POST /command` — Receive command from server room app (open_gate, print_ticket, etc.).
- `GET /health` — Health check (for server room app to ping every 15s).

**Endpoints (Server Room App HTTP Server):**
- `POST /gate/session/create` — Create session (entry).
- `POST /gate/session/{id}/calculate-fee` — Calculate fee (exit).
- `POST /gate/session/{id}/close` — Close session (exit).
- `POST /gate/alert` — Send alert (hardware failure, exception).

### 9.2.6 Bootstrapping

**New Gate Setup:**
1. Gate mini PC boots → Gate app starts.
2. Gate app auto-detects `gate_id` from hardware serial/MAC.
3. Gate app announces via mDNS: `parking-gate-{gate_id}.local:8080`.
4. Server room app discovers gate via mDNS.
5. Server room app registers gate in local DB (status: `unregistered`).
6. Server room app syncs to cloud → cloud knows about unregistered gate.
7. Admin configures gate via dashboard (vehicle_type, gate_type, location).
8. Config flows: Dashboard → Cloud → Server Room App → Gate App.
9. Gate app saves minimal config locally: `server_room_url`, `gate_id`.
10. Gate is now operational (status: `operational`).

---

## 9.3 Server Room Desktop App

### 9.3.1 Platform

- **Target OS:** Ubuntu 22.04 LTS.
- **Tech:** **Tauri** (preferred) or **Electron** — web-based desktop app.
- **Hardware:** Mini PC in server room (1 per location).
- **Screen:** Staff monitor (shows gate status, alerts).
- **Network:** LAN (gate apps) + Internet (cloud backend).
- **Local DB:** SQLite (stores all configs, sessions, transactions).
- **Stateful:** Full business logic, fee calculation, sync management.

### 9.3.2 Core Functionality

#### Local Database
- SQLite database (local, portable).
- Stores: gates, rate_configs, shift_configs, sessions, transactions, sync_queue.
- Synced to cloud every 1 minute (when internet connected).
- Local snapshot backup daily (stored in local directory).

#### Config Sync
- Polls cloud backend every 1 minute for config updates.
- Configs: rates, shifts, gates, locations.
- Config versioning: local configs always overwritten by cloud.
- Manual refresh button (trigger immediate poll).

#### Data Sync
- Syncs sessions and transactions to cloud every 1 minute.
- Sync queue: tracks unsynced data.
- Exponential backoff retry on failure (1s, 2s, 4s, 8s, max 30s).
- Sync failure alerting (log error, alert developer).
- Manual refresh button (trigger immediate sync).

#### Gate Management
- mDNS discovery (discover gate apps on LAN).
- Gate registry: track connected gates (gate_id, gate_type, vehicle_type, status, last_seen_at).
- Health check: ping each gate every 15s.
- Status reporting: send gate statuses to cloud every 20s.

#### Fee Calculation
- Fee calculation engine (multi-day stays, daily caps).
- Uses local rate config (polled from cloud).
- Assigns shift numbers (from shift config).
- Config version tracking (audit trail).

#### Alert Management
- Receives alerts from gate apps (hardware failures, exceptions).
- Plays audio alert: "Gate {gate_id} needs assistance".
- Visual alert on staff monitor (gate status grid).

### 9.3.3 User Interface (Staff Monitor)

#### Main Monitoring Screen
- **Gate Status Grid:** Cards for each gate (gate_id, status, health indicator, last_seen_at).
- **Alert Banner:** (top of screen, red background) "Gate {gate_id} needs assistance".
- **Offline Banner:** (top of screen, yellow background) "System offline" (internet disconnected).
- **Manual Refresh Buttons:** (header) Refresh configs, refresh data.

#### Alert Detail Modal
- Gate ID.
- Alert type (hardware failure, payment failure, QR unreadable).
- Timestamp.
- Action buttons: acknowledge, resolve.

#### Settings Page
- Cloud backend URL (configurable).
- Sync interval (default: 60s).
- Gate health check interval (default: 15s).
- Audio alert volume.
- Theme (light/dark).

### 9.3.4 Offline Mode

**Internet Outage:**
- Gates still operate (using local DB).
- Fee calculation uses local rate config.
- Transactions queued in sync_queue.
- Sync when internet restored.
- Offline banner visible on staff monitor.

**LAN Outage (Gate ↔ Server Room):**
- Gate stops working.
- Audio alarm plays for staff.
- Staff runs offline SOP.

### 9.3.5 Communication

**With Gate Apps (LAN):**
- HTTP server (receives requests from gate apps).
- mDNS discovery (discover gate apps).
- Health check (ping gate apps every 15s).

**With Cloud Backend (Internet):**
- HTTP client (poll configs, push data).
- Sync every 1 minute (configs down, data up).
- Status reporting every 20s (gate statuses).
- Manual refresh (trigger immediate sync).

---

## 9.4 Web Dashboard

### 9.4.1 Platform

- **Browser support:** Chrome, Firefox, Edge (latest 2 major versions).
- **Tech:** Next.js 14 (App Router, TypeScript, Tailwind CSS).
- **Responsive:** Desktop-first; tablet usable; mobile not required.
- **Authentication:** Email + password; JWT token (httpOnly cookie); 8-hour session with refresh.
- **Hosting:** Cloud backend (served as static files or SSR).

### 9.4.2 Navigation Structure

```
Web Dashboard
├── Overview (home)
│   ├── Revenue today (live)
│   ├── Alert counts
│   └── Active sessions count
├── Gates
│   ├── Gate Status (grid view)
│   ├── Unregistered Gates (configure new gates)
│   └── Gate Detail (health history)
├── Sessions
│   ├── Active Sessions
│   └── Session History
├── Transactions
│   └── Transaction List (with filters)
├── Reports
│   ├── Daily Revenue (chart + table + CSV export)
│   ├── Occupancy Over Time (chart + table + CSV export)
│   ├── Vehicle Type Breakdown (pie chart + table)
│   └── Transaction Report (table + filters + CSV export)
├── Configuration
│   ├── Locations (grouped by city)
│   ├── Rates (per location, versioned)
│   ├── Shift Configs (per location, versioned)
│   └── Vehicle Types
├── Users & Roles
│   ├── Users (list, create, edit, deactivate)
│   └── Roles (list, create, edit, permissions)
├── Monitoring
│   ├── System Health
│   ├── Audit Logs (view + CSV export)
│   └── Alerts (view, acknowledge, resolve)
└── Settings
    └── Manual Refresh (configs, data)
```

### 9.4.3 Location Selector

- Dropdown in header (grouped by city).
- Switch location → update URL (`/[locationId]/...`).
- Persist selected location in context.
- Owner/admin can view all locations (no filtering).

### 9.4.4 Key Pages

#### Gate Status Page
- Grid of gate cards (gate_id, status badge, health indicator, last_seen_at).
- Unregistered gates section (configure new gates).
- Filter by status (online, offline, all).
- Auto-refresh every 30s.

#### Gate Configuration Modal
- Gate type (radio: entry/exit).
- Vehicle type (select: car, motorcycle, truck, all).
- Location (read-only, from URL).
- Submit → PATCH /api/v1/gates/:id.

#### Rate Management Page
- Rate list (table: vehicle_type, effective_date, rates, version).
- Create rate modal (form with validation).
- Edit rate modal (pre-filled form).
- Version history (read-only).

#### Daily Revenue Report
- Date range picker.
- Revenue chart (line chart, daily totals).
- Revenue table (date, total_revenue, transaction_count, avg_transaction).
- Filter by vehicle_type.
- Export to CSV.

### 9.4.5 Authentication

- Login page (email + password form).
- Auth context (user state, permissions, login/logout functions).
- Protected routes (redirect to /login if not authenticated).
- JWT cookie handling (httpOnly, secure, sameSite).
- Auto-refresh token (refresh before expiry).

---

## 9.5 Design Decisions

**Why three-tier architecture?**
- **Gate app (edge):** Stateless, simple, easy to replace. If mini PC dies, swap it.
- **Server room app (local):** Offline resilience. Gates work without internet.
- **Cloud backend (central):** Multi-location management, central reporting, config source of truth.

**Why Tauri over Electron?**
- Smaller binary size (10-20 MB vs 100+ MB).
- Lower memory usage (uses system WebView).
- Better performance (Rust backend).
- Easier to deploy on mini PCs (limited resources).

**Why HTTP + mDNS (not WebSocket)?**
- Simpler auto-recovery (mDNS handles re-discovery).
- No persistent connections to manage.
- Gate app is stateless (no connection state).
- HTTP request/response is sufficient for command/response pattern.

**Why SQLite (not PostgreSQL) for server room app?**
- Portable (single file, no server process).
- Easy backup (copy file).
- Sufficient for single location (2-10 gates).
- No DBA needed (field technicians can manage).

**Why no operator desktop app?**
- Fully automated gates eliminate need for operators.
- Reduces operational costs.
- Faster throughput (no human delay).
- Staff only handle exceptions (server room app provides visibility).

---

*End of Chapter 9 — Platform & Interface (v2)*
