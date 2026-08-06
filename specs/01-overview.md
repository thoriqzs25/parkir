# Chapter 1 — System Overview

## 1.1 Purpose

PARKIR v2 is a fully automated parking management system designed for high-throughput, operator-free parking facilities. Drivers self-service entry and exit through automated gates, with server room staff handling exceptions only.

The system handles the full lifecycle of a parking session — from vehicle entry (ticket dispensing) to payment (tap-to-go e-money/Flazz) and exit — while providing management tools for billing configuration, monitoring, and operational reporting across multiple locations.

---

## 1.2 Scope

This system covers:

- Multi-location parking management under a single administrative instance (AMB: 20+ locations)
- Fully automated vehicle entry (ticket dispensing with QR code)
- Fully automated vehicle exit (QR scanning, fee calculation, tap-to-go payment)
- Hardware gate integration (barrier, thermal printer, QR scanner, payment terminal, sensors)
- Three-tier architecture (gate app, server room app, cloud backend)
- Offline operation with local database and sync
- Role-based access control for admin dashboard
- System health monitoring and alerting
- Reporting and analytics dashboard

This system does **not** cover (v2):

- Manual operator-assisted check-in/check-out (fully automated)
- Individual parking bay / slot tracking
- Driver-facing mobile app or self-service portal
- Monthly subscription or pass-based billing
- Push notifications via email, SMS, or WhatsApp
- License plate recognition (LPR/ANPR)

---

## 1.3 System Context

```
┌─────────────────────────────────────────────────────────────────────┐
│                        PARKIR v2 System                              │
│                                                                      │
│  ┌──────────────────┐      ┌──────────────────┐      ┌──────────┐  │
│  │  Gate Desktop App │      │ Server Room App  │      │  Cloud   │  │
│  │  (Mini PC/Gate)   │      │ (Mini PC/Server) │      │ Backend  │  │
│  │                   │      │                  │      │          │  │
│  │  - Hardware ctrl  │◄────►│  - Local DB      │◄────►│  - API   │  │
│  │  - QR generation  │ HTTP │  - Fee calc      │ HTTP │  - DB    │  │
│  │  - Payment iface  │mDNS │  - Sync queue    │      │  - Sync  │  │
│  └────────┬──────────┘      └────────┬─────────┘      └────┬─────┘  │
│           │                          │                      │        │
│           │                          │                      │        │
│  ┌────────▼──────────┐      ┌────────▼─────────┐         │        │
│  │  Gate Hardware    │      │   Dashboard Web  │         │        │
│  │  - Barrier        │      │   (AMB Admin)    │         │        │
│  │  - Printer        │      │   (Browser)      │         │        │
│  │  - QR Scanner     │      └──────────────────┘         │        │
│  │  - Payment Term.  │                                    │        │
│  │  - Sensors        │                                    │        │
│  └───────────────────┘                                    │        │
│                                                            │        │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 1.4 Key Concepts

| Term | Definition |
|------|-----------|
| **Location** | A physical parking facility managed by the system. Each location has its own rates, gates, and server room app. |
| **Gate** | A physical entry/exit point with hardware (barrier, printer, scanner, payment terminal). Can be entry-only or exit-only. |
| **Session** | A single parking event: from vehicle entry (ticket dispensed) to exit (payment completed). |
| **Transaction** | The payment record produced when a session is closed. |
| **Gate App** | Desktop application running on mini PC at each gate. Controls hardware, stateless, receives commands from server room app. |
| **Server Room App** | Desktop application running in server room (1 per location). Manages local DB, fee calculation, sync to cloud. |
| **Staff** | Server room personnel who monitor gates and handle exceptions (hardware failures, payment issues). No operators at gates. |
| **Shift** | A time-bounded work period. Shift numbers increment continuously (shift_1, shift_2, ...). Default: 3 shifts/day. |
| **Config Version** | Versioned configuration (rates, shifts) for audit trail. Transactions record which version was used. |
| **Offline Mode** | Operating state where server room app functions without internet connectivity. Gates still operate using local DB. |

### Architecture Tiers

| Tier | Component | Role | State |
|------|-----------|------|-------|
| **Edge** | Gate App | Hardware interface, stateless | No local DB |
| **Local** | Server Room App | Business logic, local DB, sync | Full local DB |
| **Cloud** | Cloud Backend | Central DB, config source, reporting | Central source of truth |

### Data Flow

- **Configs:** Cloud → Server Room App (poll every 1 min)
- **Transactions:** Server Room App → Cloud (sync every 1 min)
- **Gate commands:** Server Room App → Gate App (HTTP + mDNS)

---

## 1.5 Design Principles

1. **Automation first** — No operators at gates. Drivers self-service entry/exit. System must be reliable and handle edge cases gracefully.
2. **Offline resilience** — Server room app must continue operating without internet. Gates must continue operating if server room app loses LAN connectivity (staff handles via offline SOP).
3. **Hardware agnostic** — Gate app uses hardware abstraction layer (HAL) to support different hardware vendors.
4. **Auditability** — Every action (session creation, payment, config change) must be logged with timestamp, gate ID, and config version.
5. **Scalability** — System must support 20+ locations with 2-10 gates each, all syncing to central cloud backend.
6. **Simplicity** — Gate app is stateless and simple. Server room app handles complexity. Staff dashboard provides visibility.

---

## 1.6 Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| Entry flow time | < 5 seconds from button press to gate open |
| Exit flow time | < 10 seconds from QR scan to gate open (with payment) |
| API response time | < 300ms for session and payment operations (p95) |
| Dashboard load time | < 2s for report pages |
| Sync latency | < 5 seconds for transaction sync (p95) |
| Offline operation | Gates continue operating without internet |
| Audit log retention | Minimum 2 years, non-deletable |
| Concurrent gates | Support 2-10 gates per location |
| Uptime target | 95% (hardware-dependent) |
| Receipt print time | < 3 seconds from payment confirmation |
| Config sync | Every 1 minute (automatic) + manual refresh |
