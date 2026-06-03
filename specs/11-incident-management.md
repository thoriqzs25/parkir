# Chapter 11 — Incident Management

## 11.1 Overview

The incident management module provides a structured workflow for operators to report operational problems and for managers to review, act on, and close those incidents. Every incident is recorded with full context and remains in the system permanently for audit purposes.

---

## 11.2 Incident Types

| Code | Name | Typical Trigger |
|------|------|----------------|
| `STUCK_AT_GATE` | Vehicle Stuck at Gate | Gate/barrier malfunction; vehicle cannot exit |
| `PAYMENT_DISPUTE` | Payment Dispute | Driver disputes the calculated fee |
| `OPERATOR_ERROR` | Operator Error | Wrong plate or vehicle type entered at check-in |
| `SYSTEM_DOWNTIME` | System Downtime | Backend unreachable; operator working in offline mode |

### 11.2.1 STUCK_AT_GATE Examples

Physical barrier/gate issue preventing vehicle movement.

| Example | What Happened |
|---------|---------------|
| Gate won't open | Driver paid, receipt printed, but the exit barrier is stuck closed. Operator needs to manually lift the gate. |
| Gate won't close | Previous vehicle exited but gate stayed open. Next vehicle enters without check-in being recorded. |
| Vehicle stuck on barrier | Barrier came down on a slow-moving vehicle. |
| Power failure | Gate system lost power, all vehicles stuck at entry/exit. |

**Typical Resolution:** On-site staff manually operates gate; technician called for repair.

### 11.2.2 PAYMENT_DISPUTE Examples

Driver disagrees with the parking fee.

| Example | What Happened |
|---------|---------------|
| Wrong duration | Driver claims they parked for 1 hour but system shows 3 hours. "I just arrived 1 hour ago!" |
| Wrong vehicle type | Driver's motorcycle was recorded as CAR, charged Rp 15,000 instead of Rp 5,000. |
| Already paid | Driver claims they paid at entry (confusion with another parking location). |
| Rate disagreement | Driver says "The sign outside says Rp 3,000/hour, why am I charged Rp 5,000?" |
| Lost ticket | Driver has no proof of check-in time, disputes the default maximum charge. |

**Typical Resolution:** Manager reviews session details, may void and re-issue at corrected rate, or confirms original charge is correct.

### 11.2.3 OPERATOR_ERROR Examples

Operator made a mistake during check-in or payment.

| Example | What Happened |
|---------|---------------|
| Wrong plate | Operator typed "B 1234 XYZ" but actual plate is "B 1234 XYA". Vehicle can't be found at check-out. |
| Wrong vehicle type | Operator selected CAR but vehicle is a TRUCK. Rate is wrong. |
| Wrong location | Operator forgot to switch active location, session recorded at Gate A but vehicle is at Gate B. |
| Double check-in | Operator accidentally submitted check-in twice, created duplicate session. |
| Checked out wrong vehicle | Two similar plates (B 1234 AB vs B 1234 AC), operator closed the wrong session. |
| Cash entry error | Operator entered Rp 50,000 tendered but driver gave Rp 100,000. Wrong change given. |

**Typical Resolution:** Manager voids incorrect session/transaction, operator creates new correct record.

### 11.2.4 SYSTEM_DOWNTIME Examples

Technical failure affecting operations.

| Example | What Happened |
|---------|---------------|
| Backend unreachable | Internet down, desktop app switches to offline mode. |
| Database timeout | API responding slowly or timing out. |
| Printer failure | Thermal printer not printing receipts. |
| Payment gateway down | Digital payments failing, cash-only mode needed. |

**Typical Resolution:** Operator continues in offline mode, IT investigates, data syncs when restored.

---

## 11.3 Incident States

| State | Description |
|-------|-------------|
| `OPEN` | Incident filed; awaiting manager attention |
| `IN_PROGRESS` | Manager has acknowledged and is handling it |
| `RESOLVED` | Incident closed with resolution notes |

---

## 11.4 Incident Lifecycle

```
Operator detects problem
        │
        ▼
  [ File Incident ]
  - Select type
  - Link session (optional)
  - Write description
        │
        ▼
   State: OPEN
        │
        │  Manager reviews in dashboard
        ▼
  State: IN_PROGRESS
        │
        │  Manager takes action (e.g. manual adjustment, rate override)
        │  Manager adds resolution notes
        ▼
   State: RESOLVED
```

---

## 11.5 Filing an Incident (Operator, Desktop App)

**Steps:**
1. Operator taps **Report Incident** from the home screen or session detail view.
2. Selects incident type from dropdown.
3. Optionally links a session (search by plate or session ID).
4. Enters a description (required, min 10 characters).
5. Submits — incident is created with state `OPEN`.
6. Confirmation shown with incident ID.

**In offline mode:**
- Incidents are saved locally and synced when connectivity is restored.
- An offline-created incident is marked `offline_sync = true` until synced.

---

## 11.6 Incident Detail (Manager, Web Dashboard)

Managers see all open incidents for their location(s) in the Incidents panel.

**Incident detail view includes:**
- Incident ID, type, state
- Filed by (operator name), filed at (timestamp)
- Linked session (if any): plate, vehicle type, check-in time, current state
- Description
- Manager notes (editable by any manager with `incidents:resolve` permission)
- Resolution notes (required to close)
- Timeline of state changes

**Manager actions:**
| Action | Condition | Result |
|--------|-----------|--------|
| Mark In Progress | State = OPEN | State → IN_PROGRESS |
| Add Note | Any state | Appends timestamped note |
| Resolve | State = OPEN or IN_PROGRESS | Requires resolution notes; State → RESOLVED |

---

## 11.7 Incident-Specific Handling Guidance

### STUCK_AT_GATE
- Manager should coordinate with on-site staff to manually release the gate.
- If a session is linked, manager may void the transaction if the driver was unable to pay due to the malfunction.
- Resolution notes should document the root cause and corrective action.

### PAYMENT_DISPUTE
- Manager reviews the session and transaction details.
- Manager may void the transaction and re-issue at an adjusted rate (manual adjustment).
- Or manager confirms original charge is correct and informs the operator to communicate this to the driver.
- Resolution notes should record the outcome (voided, upheld, adjusted).

### OPERATOR_ERROR
- Manager determines correction needed:
  - **Wrong plate:** Void session → Operator creates a new correct check-in.
  - **Wrong vehicle type:** Void transaction → Re-issue with corrected type.
- See Chapter 12 for the adjustment procedures.
- Resolution notes should reference the adjustment action taken.

### SYSTEM_DOWNTIME
- Operator switches to offline mode (see Section 11.9).
- Incident is filed to officially log the downtime event.
- Manager investigates root cause using system health dashboard.
- Resolution notes should include: downtime duration, cause, and sync outcome.

---

## 11.8 Incident Notifications

In v1, notifications are **in-app only** (web dashboard):

| Event | Notification Target |
|-------|-------------------|
| New incident filed | All managers at the incident's location |
| Incident moved to IN_PROGRESS | Reporting operator (if they have dashboard access) |
| Incident resolved | Reporting operator (if they have dashboard access) |

Notification badge shown in the dashboard nav. Clicking opens the incident detail.

Email / SMS / WhatsApp notifications are out of scope for v1.

---

## 11.9 Offline Mode (SYSTEM_DOWNTIME Handling)

When the operator desktop app cannot reach the backend:

**Automatic behavior:**
- App detects connectivity loss (polling interval: 30 seconds).
- Offline mode badge appears in the app header.
- A `SYSTEM_DOWNTIME` incident is auto-drafted (operator can edit and submit, or discard).

**Offline capabilities:**
- Check-in: creates session locally with temp ID, flagged `offline_sync = true`.
- Check-out & payment: completed using locally cached rate configuration.
- Receipt printing: uses cached data; receipt number uses offline format `[CODE]-OFFLINE-[SEQ]`.
- Incident filing: saved locally and synced on reconnect.

**Rate cache:**
- Rate data is synced to local storage every time the app connects.
- Cache TTL: 24 hours.
- If cache is expired and app is offline, operator is warned. Rates from last known cache are used with a warning on the receipt.

**On reconnect:**
1. App detects connectivity restored.
2. Auto-sync begins: pushes all offline sessions in `check_in_at` order.
3. Backend assigns official IDs and receipt numbers.
4. If a conflict is detected (e.g. same plate already checked in via another terminal), session is flagged `sync_conflict = true`.
5. Manager reviews sync conflicts in the web dashboard (dedicated view).
6. Operator sees a sync completion notification with count of synced records and any conflicts.

---

## 11.10 Data Model

```
incidents
  id                UUID, primary key
  location_id       UUID, FK → locations.id, not null
  type              ENUM('STUCK_AT_GATE', 'PAYMENT_DISPUTE', 'OPERATOR_ERROR', 'SYSTEM_DOWNTIME'), not null
  state             ENUM('OPEN', 'IN_PROGRESS', 'RESOLVED'), default 'OPEN'
  session_id        UUID, FK → sessions.id, nullable
  reported_by       UUID, FK → users.id, not null
  reported_at       TIMESTAMP WITH TIME ZONE, not null
  description       TEXT, not null
  resolved_by       UUID, FK → users.id, nullable
  resolved_at       TIMESTAMP WITH TIME ZONE, nullable
  resolution_notes  TEXT, nullable
  offline_sync      BOOLEAN, default false
  created_at        TIMESTAMP
  updated_at        TIMESTAMP

incident_notes
  id                UUID, primary key
  incident_id       UUID, FK → incidents.id, not null
  author_id         UUID, FK → users.id, not null
  note              TEXT, not null
  created_at        TIMESTAMP WITH TIME ZONE, not null
```
