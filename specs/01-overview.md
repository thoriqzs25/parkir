# Chapter 1 — System Overview

## 1.1 Purpose

The Parking Administration System (PAS) is a general-purpose platform for
managing vehicle parking operations across one or more physical locations. It
is designed to be operated by parking attendants at booths or gates, and
supervised by facility managers through a web-based dashboard.

The system handles the full lifecycle of a parking session — from vehicle entry
to payment and receipt — while providing management tools for billing
configuration, incident handling, manual corrections, and operational
reporting.

---

## 1.2 Scope

This system covers:

- Multi-location parking management under a single administrative instance
- Vehicle check-in and check-out with plate and type tracking
- Fee calculation and payment collection (cash and digital)
- Thermal receipt generation
- Role-based access control with custom permissions
- Incident reporting and resolution workflow
- Manual adjustment procedures with authorization and audit trail
- Reporting and analytics dashboard
- System health monitoring and anomaly alerting

This system does **not** cover (v1):

- Individual parking bay / slot tracking
- Driver-facing mobile app or self-service portal
- Hardware gate/barrier integration
- Monthly subscription or pass-based billing
- Push notifications via email, SMS, or WhatsApp

---

## 1.3 System Context

```
┌──────────────────────────────────────────────────────────┐
│                  Parking Admin System                    │
│                                                          │
│  ┌─────────────────┐        ┌──────────────────────┐    │
│  │  Operator        │        │  Web Dashboard        │   │
│  │  Desktop App     │        │  (Manager / Admin)    │   │
│  │  (Windows)       │        │  (Browser)            │   │
│  └────────┬────────┘        └──────────┬───────────┘    │
│           │                            │                  │
│           └──────────┬─────────────────┘                 │
│                      │                                    │
│              ┌───────▼────────┐                          │
│              │   Backend API  │                          │
│              └───────┬────────┘                          │
│                      │                                    │
│         ┌────────────┼──────────────┐                    │
│         │            │              │                     │
│  ┌──────▼──┐  ┌──────▼──┐  ┌───────▼──┐                 │
│  │Database │  │Payment  │  │Thermal   │                  │
│  │         │  │Gateway  │  │Printer   │                  │
│  └─────────┘  └─────────┘  └──────────┘                 │
└──────────────────────────────────────────────────────────┘
```

---

## 1.4 Key Concepts

| Term | Definition |
|------|-----------|
| **Location** | A physical parking facility managed by the system. Each location has its own rates, capacity, and operators. |
| **Session** | A single parking event: from vehicle check-in to check-out and payment closure. |
| **Transaction** | The payment record produced when a session is closed. |
| **Operator** | A staff member (user account) who uses the desktop app to process check-ins, check-outs, and payments. Represents the person. |
| **Shift** | A time-bounded work period during which an operator is actively working at a location. Used for cash accountability and activity tracking. |
| **Manager** | A staff member who uses the web dashboard to oversee operations, run reports, and perform adjustments. |
| **Incident** | A recorded operational problem (e.g. payment dispute, operator error) that requires resolution. |
| **Audit Log** | An immutable chronological record of every action taken in the system. |
| **Offline Mode** | A degraded operating state where the desktop app functions without backend connectivity. |

### Operator vs Shift

| Concept | What It Is | Lifetime | Purpose |
|---------|------------|----------|---------|
| **Operator** | A person (user account) | Permanent | Identity — WHO did something |
| **Shift** | A work period | Hours/day | Time scope — WHEN and WHERE they worked |

- One operator can have many shifts (one per work day)
- Sessions are linked to the operator who checked in the vehicle
- Transactions are linked to the shift during which payment was collected
- Cash reconciliation is calculated per shift, not per operator

---

## 1.5 Design Principles

1. **Operator simplicity** — The desktop app must be fast and minimal. Operators deal with high throughput; every flow must be completable in seconds.
2. **Manager visibility** — The web dashboard must give managers a clear, real-time picture of every location without needing to be on-site.
3. **Auditability** — Every action that changes state (session, payment, adjustment, role) must be logged with actor, timestamp, and context. Nothing is silently deleted.
4. **Resilience** — The operator desktop app must continue to function when the backend is unreachable, syncing when connectivity is restored.
5. **Extensibility** — Rate models, vehicle types, roles, and alert thresholds should be configurable — not hardcoded.

---

## 1.6 Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| API response time | < 300ms for session and payment operations (p95) |
| Dashboard load time | < 2s for report pages |
| Offline sync | All offline sessions synced within 60 seconds of reconnect |
| Audit log retention | Minimum 2 years, non-deletable |
| Concurrent operators | Support at least 50 concurrent operator sessions per location |
| Uptime target | 99.5% for backend API |
| Receipt print time | < 3 seconds from payment confirmation |
