# Chapter 2 — System Goals

## 2.1 Goal Summary

The system is built around five primary goals. Each goal maps to one or more features and user-facing modules.

| # | Goal | Primary Users | Key Modules |
|---|------|--------------|-------------|
| G1 | Slot Occupancy Tracking | Operators, Managers | Sessions, Dashboard |
| G2 | Payment & Billing | Operators, Managers | Payments, Rates, Receipts |
| G3 | Reporting & Analytics | Managers, Admins | Reports Dashboard |
| G4 | Incident Management | Operators, Managers | Incident Module |
| G5 | System Observability | Admins, Managers | Health Monitor, Audit Log, Alerts |

---

## 2.2 G1 — Slot Occupancy Tracking

### Intent
Provide real-time visibility into how many vehicles are currently parked at each location, broken down by vehicle type.

### Functional Requirements
- The system must record a check-in timestamp when a vehicle enters.
- The system must record a check-out timestamp when a vehicle exits.
- The current occupancy count per location must be derivable at any point in time from active sessions.
- The web dashboard must display live occupancy per location.
- Occupancy must be broken down by vehicle type (CAR, MOTO, TRUCK).
- Operators must be able to look up whether a specific plate is currently checked in.

### Acceptance Criteria
- [ ] Checking in a vehicle increases the active session count for that location by 1.
- [ ] Checking out a vehicle decreases the active session count for that location by 1.
- [ ] Dashboard occupancy count matches the count of `ACTIVE` sessions per location at any given moment.
- [ ] Plate lookup returns the active session details if the vehicle is checked in.

---

## 2.3 G2 — Payment & Billing

### Intent
Accurately calculate parking fees and record payments using flexible rate models and multiple payment methods.

### Functional Requirements
- Fee must be calculated at check-out based on duration and applicable rate (pay-on-exit model).
- The system must support the following rate components: first-hour rate, subsequent hourly rate, and daily flat rate (as a price ceiling).
- Rates must be configurable per location and per vehicle type.
- Daily flat rate must automatically apply when the hourly total exceeds the configured daily cap.
- Payment must be recorded as either cash or digital.
- A transaction record must be created for every closed session.
- Voided transactions must be excluded from all revenue totals.

### Acceptance Criteria
- [ ] A 2.5-hour CAR session at Rp 5,000/hour is billed Rp 15,000 (rounding up to next full hour).
- [ ] A 14-hour session where hourly total exceeds the daily cap applies the daily flat rate instead.
- [ ] A session cannot be closed without a payment method and amount being recorded.
- [ ] Voided transactions do not appear in daily revenue totals.

---

## 2.4 G3 — Reporting & Analytics

### Intent
Give managers and admins data-driven insight into revenue performance, occupancy trends, and operator activity.

### Functional Requirements
- Reports must be filterable by location, date range, and vehicle type.
- The system must provide: daily revenue summary, occupancy over time, per-vehicle-type breakdown, and operator activity log.
- All reports must be viewable in the web dashboard.
- Revenue reports must distinguish between cash and digital payments.
- Operator activity reports must cover: check-ins, check-outs, payments, incidents, and adjustments.

### Acceptance Criteria
- [ ] A daily revenue report for a given date shows total revenue matching the sum of all non-voided transactions that day.
- [ ] Occupancy over time graph shows correct hourly occupancy for the selected period.
- [ ] Per-vehicle-type breakdown correctly attributes sessions and revenue to CAR, MOTO, TRUCK.
- [ ] Operator activity log shows every action by the selected operator in the selected date range.

---

## 2.5 G4 — Incident Management

### Intent
Provide a structured workflow for operators to report problems and for managers to resolve them, with a complete record of every incident.

### Functional Requirements
- Operators must be able to open an incident from the desktop app at any time.
- Incidents must be typed (STUCK_AT_GATE, PAYMENT_DISPUTE, OPERATOR_ERROR, SYSTEM_DOWNTIME).
- Each incident must be linked to a session where applicable.
- Incidents must be visible to managers in the web dashboard.
- Managers must be able to add resolution notes and close incidents.
- Incident history must be retained and searchable.

### Acceptance Criteria
- [ ] An operator can file an incident in under 30 seconds from the desktop app.
- [ ] A filed incident appears in the manager's dashboard immediately (or upon next sync in offline mode).
- [ ] A manager can close an incident with resolution notes; the incident shows `RESOLVED` status.
- [ ] Closed incidents remain visible in the incident history log.

---

## 2.6 G5 — System Observability

### Intent
Give administrators and managers real-time awareness of system health, a
complete audit trail of all actions, and automated alerting when anomalies
occur.

### Functional Requirements
- The system must expose a health status dashboard covering: API, database, printers, payment gateway.
- Every state-changing action must produce an immutable audit log entry.
- Audit logs must be queryable by actor, action type, entity, and date range.
- Automated alerts must fire when configured anomaly thresholds are exceeded.
- Per-operator performance metrics must be available to managers.

### Acceptance Criteria
- [ ] Every session create, close, void, and reassign action produces an audit log entry with actor, timestamp, and entity reference.
- [ ] Audit log entries cannot be deleted by any user, including system admins.
- [ ] An alert fires when a session remains `ACTIVE` for more than 24 hours.
- [ ] An alert fires when a session is closed without a linked transaction (unpaid exit).
- [ ] System health dashboard updates within 30 seconds of a component going down.
