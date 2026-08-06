# PARKIR v2 — Spec Update Summary
**Created:** 2026-08-06

This document summarizes the changes needed to update all numbered specs (01-17) from v1 (manual operator system) to v2 (automated gate system).

---

## Key Changes in v2

### 1. No Operators
- **v1:** Operators manually check-in/check-out vehicles at gates
- **v2:** Fully automated gates, no operators. Drivers self-service entry/exit.

### 2. Three-Tier Architecture
- **v1:** Desktop app (operator) → Backend API → Database
- **v2:** 
  - Gate App (mini PC at gate) → Server Room App (mini PC in server room) → Cloud Backend
  - Gate app is stateless (hardware interface only)
  - Server room app has local DB, handles business logic, syncs to cloud
  - Cloud backend is central source of truth

### 3. Hardware Integration
- **v1:** No hardware integration (operator uses desktop app)
- **v2:** Full hardware integration:
  - Entry gate: vehicle loop sensor, ticket button, Epson thermal printer, gate motor
  - Exit gate: QR scanner (Panda PRJ-777, USB HID), payment terminal (e-money/Flazz), driver-facing monitor, gate motor
  - All hardware connected via central interface hub

### 4. Communication
- **v1:** Desktop app → Backend API (HTTP)
- **v2:**
  - Gate app ↔ Server room app: HTTP + mDNS (LAN)
  - Server room app ↔ Cloud backend: HTTP (sync every 1 min)
  - Gate app is stateless, receives commands from server room app

### 5. Offline Handling
- **v1:** Desktop app queues operations, syncs when online
- **v2:** Server room app has full local DB (SQLite), continues operating without internet. Gates still work if internet is down. Transactions queued and synced when online.

### 6. Config Versioning
- **v1:** No versioning
- **v2:** Rate configs and shift configs are versioned. Transactions record which config version was used (audit trail).

### 7. Shift Management
- **v1:** Shifts track operator work periods
- **v2:** Shift numbers increment continuously (shift_1, shift_2, ...). Default: 3 shifts/day. Assigned by server room app based on time.

### 8. Payment
- **v1:** Cash or digital payment collected by operator
- **v2:** Tap-to-go payment (e-money/Flazz) via payment terminal. Payment vendor handles terminal, PARKIR receives completion signal.

### 9. First Tenant
- **v1:** General-purpose system
- **v2:** AMB (single company, 20+ locations grouped by city)

### 10. User Roles
- **v1:** operator, manager, admin, owner
- **v2:** staff (server room), manager (dashboard), admin, owner (superadmin). No operators.

---

## Spec-by-Spec Update Guide

### ✅ 01-overview.md — UPDATED
**Status:** Updated to v2
**Changes:** New architecture, no operators, three-tier system, hardware integration

### 02-system-goals.md — NEEDS UPDATE
**Key Changes:**
- Goal: Excellent end-to-end automated parking system (not operator-assisted)
- First tenant: AMB (20+ locations)
- Business model: Per-location monthly fee
- Replace expensive third-party system

**Update Points:**
- Remove operator-related goals
- Add automation goals (entry/exit flow time, hardware reliability)
- Add multi-location goals (20+ locations, central management)
- Add offline resilience goals

### ✅ 03-users-and-roles.md — UPDATED
**Status:** Updated to v2
**Changes:** No operators, new role (staff), gate app/server room app access control

### 04-locations.md — NEEDS UPDATE
**Key Changes:**
- Locations have gates (entry/exit)
- Locations have server room app
- Locations grouped by city (for AMB)
- Locations have gate configurations (vehicle type per gate)

**Update Points:**
- Add gate management (entry/exit gates per location)
- Add server room app per location
- Add city grouping (for dashboard filtering)
- Add gate configuration (vehicle type, gate type)

### 05-vehicle-management.md — MINOR UPDATE
**Key Changes:**
- Vehicle types: car, motorcycle, truck (same as v1)
- Vehicle type is configured per gate (not selected by operator)

**Update Points:**
- Gate has fixed vehicle type (configured via dashboard)
- No operator selection of vehicle type
- Vehicle type stored in gate configuration

### 06-parking-sessions.md — NEEDS UPDATE
**Key Changes:**
- Entry flow: loop sensor → button press → ticket dispensed → gate opens
- Exit flow: QR scan → fee calculation → payment → gate opens
- Session created by server room app (not operator)
- QR code contains: session_id, location_id, timestamp, vehicle_type
- Session has shift_number (assigned by server room app)
- Session has rate_config_version and shift_config_version

**Update Points:**
- Remove operator check-in/check-out
- Add automated entry flow (hardware triggers)
- Add automated exit flow (QR scan, payment)
- Add QR code generation and validation
- Add shift_number assignment
- Add config version tracking
- Session states: ACTIVE → PENDING_PAYMENT → CLOSED (same as v1)

### 07-payment-and-billing.md — NEEDS UPDATE
**Key Changes:**
- Payment: tap-to-go (e-money/Flazz) via payment terminal
- Payment vendor handles terminal, PARKIR receives completion signal
- No cash payments (v2)
- Fee calculated by server room app (not cloud)
- Fee uses local rate config (polled from cloud every 1 min)

**Update Points:**
- Remove cash payment flow
- Add tap-to-go payment flow
- Add payment vendor integration (TBD)
- Fee calculation: server room app calculates using local rate config
- Add config version tracking (which rate version was used)
- Payment methods: emoney, flazz (v2)

### 08-receipt.md — MINOR UPDATE
**Key Changes:**
- Entry ticket: thermal receipt with QR code (timestamp, vehicle type, warning text)
- Exit receipt: optional (driver presses receipt button), shows check-in/out time, fee, vehicle type, shift number

**Update Points:**
- Entry ticket: QR code contains session data
- Exit receipt: optional, printed on demand
- Add shift number to receipt
- Add config version to receipt (for audit)

### 09-platform-and-interface.md — NEEDS UPDATE
**Key Changes:**
- Three platforms: Gate App, Server Room App, Dashboard (web)
- Gate app: stateless, hardware interface, HTTP server, mDNS
- Server room app: local DB, business logic, sync
- Dashboard: multi-location management, reporting, gate configuration

**Update Points:**
- Remove operator desktop app
- Add gate app (stateless, hardware interface)
- Add server room app (local DB, business logic)
- Dashboard: add gate management, multi-location support
- Add hardware interfaces (printer, scanner, payment terminal, sensors)

### 10-reports-and-analytics.md — MINOR UPDATE
**Key Changes:**
- Reports grouped by city (for AMB)
- No operator activity reports (no operators)
- Add gate health reports
- Add hardware failure reports

**Update Points:**
- Add city grouping
- Remove operator activity reports (or mark as N/A)
- Add gate health/uptime reports
- Add hardware failure reports
- Keep revenue, occupancy, transaction reports (same as v1)

### 11-incident-management.md — MINOR UPDATE
**Key Changes:**
- Incidents triggered by hardware failures, payment failures, QR unreadable
- Driver presses alert button → staff notified
- Staff handles incidents (no operator)

**Update Points:**
- Incident types: hardware failure, payment failure, QR unreadable, printer jam, gate motor failure
- Alert flow: driver presses button → server room app → audio alarm → staff responds
- Staff handles incident (runs offline SOP if needed)
- Remove operator-related incidents

### 12-manual-adjustments.md — NEEDS UPDATE
**Key Changes:**
- No manual adjustments by operators
- Staff can void transactions (requires manager PIN)
- Adjustments are exception handling (hardware failures, payment issues)

**Update Points:**
- Remove operator manual adjustments
- Add staff exception handling (void transactions, resolve incidents)
- Void requires manager PIN (same as v1)
- Add offline SOP reference (staff handles offline)

### 13-system-observability.md — NEEDS UPDATE
**Key Changes:**
- Monitoring: Loki (logs), Prometheus (metrics), Grafana (dashboards)
- Alerting: email/Telegram for developer, audio alarm for staff
- Health checks: gate app (15s), server room app (20s)
- Gate metrics collected by server room app

**Update Points:**
- Add Loki for log aggregation
- Add Prometheus + Grafana for metrics
- Add alert routing (email/Telegram for developer, audio for staff)
- Add gate health monitoring (server room app pings gate every 15s)
- Add server room app health monitoring (cloud pings every 20s)
- Add sync monitoring (queue length, sync success rate)

### 14-data-model.md — NEEDS UPDATE
**Key Changes:**
- New entities: Gate, ConfigVersion
- Session: add shift_number, rate_config_version, shift_config_version
- Transaction: add shift_number, config_version
- Remove operator_id from session/transaction (no operators)

**Update Points:**
- Add Gate entity (gate_id, location_id, gate_type, vehicle_type, status, config)
- Add ConfigVersion entity (config_type, config_id, version, effective_at, data)
- Session: add shift_number, rate_config_version, shift_config_version, remove operator_id
- Transaction: add shift_number, config_version, remove operator_id
- Add sync tracking (synced_at, sync_queue)

### 15-out-of-scope.md — NEEDS UPDATE
**Key Changes:**
- v2 out of scope: manual operator flows, license plate recognition, mobile app
- v2 in scope: hardware integration, offline operation, multi-location

**Update Points:**
- Add v2 out of scope: license plate recognition (LPR), mobile app, monthly subscriptions
- Add v2 in scope: hardware integration, offline operation, multi-location management
- Clarify what's automated vs manual

### 16-open-questions.md — NEEDS UPDATE
**Key Changes:**
- Many v1 questions resolved
- New v2 questions: payment vendor integration, hardware specifics, offline SOP details

**Update Points:**
- Remove resolved v1 questions
- Add v2 open questions (from PRD):
  - Payment vendor (TBD)
  - Hardware specifics (some known, some TBD)
  - Offline SOP (staff handles, not our concern)
  - Config version format (TBD)
  - Audio alert implementation (TBD)

### 17-shift-management.md — NEEDS UPDATE
**Key Changes:**
- Shift numbers increment continuously (shift_1, shift_2, shift_3, shift_4, ...)
- Default: 3 shifts/day (morning, afternoon, evening)
- Shift assigned by server room app based on check-in/check-out time
- No operator shifts (no operators)

**Update Points:**
- Shift number increments continuously across days
- Default shift config: 3 shifts/day
- Shift assigned by server room app (not operator)
- Shift stored in session and transaction records
- Remove operator shift tracking
- Add shift config versioning

---

## Priority Updates

**High Priority (critical changes):**
1. ✅ 01-overview.md — Done
2. ✅ 03-users-and-roles.md — Done
3. 06-parking-sessions.md — Automated entry/exit flow
4. 09-platform-and-interface.md — Three-tier architecture
5. 14-data-model.md — New entities (Gate, ConfigVersion)

**Medium Priority (significant changes):**
6. 07-payment-and-billing.md — Tap-to-go payment
7. 13-system-observability.md — Monitoring stack (Loki, Prometheus, Grafana)
8. 17-shift-management.md — Continuous shift numbers

**Low Priority (minor changes):**
9. 02-system-goals.md — Update goals
10. 04-locations.md — Add gates
11. 05-vehicle-management.md — Gate-based vehicle type
12. 08-receipt.md — QR code, optional receipt
13. 10-reports-and-analytics.md — City grouping, gate reports
14. 11-incident-management.md — Hardware incidents
15. 12-manual-adjustments.md — Exception handling
16. 15-out-of-scope.md — Update scope
17. 16-open-questions.md — Update questions

---

## Reference Documents

For detailed information, see:
- `PRD-v2-automated-gate-system.md` — Full product requirements
- `BACKEND-TIMELINE-v2.md` — Backend implementation plan
- `FRONTEND-TIMELINE-v2.md` — Frontend implementation plan
- `INFRASTRUCTURE-TIMELINE-v2.md` — Infrastructure implementation plan
- `QA-TIMELINE-v2.md` — QA implementation plan
- `DESIGN-TIMELINE-v2.md` — Design implementation plan
- `IMPLEMENTATION-PLAN-v2.md` — Overall implementation timeline

---

## Next Steps

1. Review this summary
2. Prioritize which specs to update first
3. Update specs incrementally (start with high priority)
4. Review updated specs with team
5. Archive v1 specs (keep for reference)

---

*End of Spec Update Summary*
