# Chapter 17 — Shift Management

## 17.1 Overview

The shift system tracks the working periods of operators at each location. A shift has a defined start and end time, and is used to scope operator activity reports, cash collection accountability, and handover records.

Shifts are operator-driven: the operator starts their shift when they begin work and ends it when they finish. A shift is always tied to one operator and one location.

---

## 17.2 Shift Lifecycle

```
Operator logs in
      │
      ▼
  Start Shift
  - Shift record created
  - State: OPEN
      │
      │  Operator works: check-ins, check-outs, payments
      │
      ▼
  End Shift
  - Operator records cash handover amount
  - Shift record closed
  - State: CLOSED
```

---

## 17.3 Shift States

| State | Description |
|-------|-------------|
| `OPEN` | Shift is active; operator is working |
| `CLOSED` | Shift ended; cash handover recorded |

---

## 17.4 Starting a Shift (Operator, Desktop App)

1. After login, the operator is prompted to **Start Shift** before accessing operational functions.
2. Operator confirms their active location (pre-selected from login).
3. System creates a shift record with `started_at = now()`, state = `OPEN`.
4. All subsequent sessions and transactions are linked to this shift.

**Only one shift can be OPEN per operator at a time.**
If the operator logs in and finds an unclosed shift from a previous session (e.g. app crashed), they are prompted to close it first or flag it for manager review.

---

## 17.5 Ending a Shift (Operator, Desktop App)

1. Operator selects **End Shift** from the home screen.
2. System displays shift summary:
   - Shift duration
   - Total check-ins
   - Total check-outs
   - Total sessions closed
   - Total cash collected (sum of cash transactions)
   - Total digital collected (sum of digital transactions)
   - Total revenue
3. Operator enters **cash handover amount** — the physical cash they are handing to the supervisor.
4. System calculates and displays:
   - Expected cash = sum of all cash transactions in the shift
   - Discrepancy = cash_handover_amount − expected_cash
5. Operator submits. Shift record closed with `ended_at = now()`.
6. If discrepancy is non-zero, the shift is flagged and a manager notification is triggered.

---

## 17.6 Shift Summary (Manager, Web Dashboard)

Managers can view shift records per location:

### Shift List View
| Column | Description |
|--------|-------------|
| Operator | Name |
| Location | |
| Started At | |
| Ended At | |
| Duration | |
| Sessions | Total closed sessions |
| Cash Collected | Sum of cash transactions |
| Digital Collected | Sum of digital transactions |
| Total Revenue | |
| Cash Handover | Amount physically handed over |
| Discrepancy | Handover − Expected cash |
| Status | OPEN / CLOSED / FLAGGED |

### Shift Detail View
- Full session list for the shift (plate, vehicle type, in/out times, fee, payment method).
- Breakdown: cash vs digital.
- Discrepancy detail with flag reason.

---

## 17.7 Discrepancy Handling

A shift is flagged (`status = FLAGGED`) when:
- `cash_handover_amount ≠ expected_cash_collected`

Flagged shifts appear in the manager's alert panel alongside anomaly alerts (Chapter 13).

Manager actions on a flagged shift:
- Add a resolution note explaining the discrepancy.
- Mark as resolved (`status = RESOLVED`).

Common causes: missed cash transactions, change errors, partial payments not recorded correctly.

---

## 17.8 Unclosed Shifts

If an operator's shift remains `OPEN` for more than the configured threshold (default: 16 hours):
- An anomaly alert is triggered: `SHIFT_NOT_CLOSED`.
- Manager can forcibly close the shift with a note.
- Force-closed shifts are marked `FORCE_CLOSED` and audit-logged.

---

## 17.9 Shift and Session Linkage

- Every session opened during a shift is linked to that shift via `shift_id`.
- Sessions created in offline mode are linked to the shift that was open at the time of check-in.
- If a shift ends before an offline session syncs, the session is still attributed to that shift based on `check_in_at` timestamp.

### 17.9.1 Cross-Shift Session Handling

When a vehicle is checked in during one shift but checks out during another shift:

```
08:00  Operator A starts shift
08:30  Vehicle B 1234 XYZ checks in → Session created, linked to Shift A
16:00  Operator A ends shift (vehicle still parked)
16:00  Operator B starts shift
18:00  Vehicle B 1234 XYZ checks out → Operator B collects payment
```

**Attribution Model: Transaction-Level Shift Tracking**

| Record | Shift Attribution | Operator Attribution |
|--------|-------------------|----------------------|
| Session | Check-in shift (Shift A) | Check-in operator (Operator A) |
| Transaction | Payment shift (Shift B) | Payment collector (Operator B) |

**Data Model:**

```sql
sessions
  shift_id       UUID  -- check-in shift
  operator_id    UUID  -- check-in operator

transactions
  shift_id       UUID  -- payment collection shift
  operator_id    UUID  -- payment collector
```

**Why this design:**
- **Session** = parking event → attributed to whoever registered the vehicle (accountability for correct plate/vehicle type)
- **Transaction** = payment → attributed to whoever collected the money (accountability for cash handling)
- **Cash reconciliation** is based on `transactions.shift_id`, ensuring correct handover amounts

**Shift Summary Example:**

| Shift | Operator | Vehicles Checked In | Payments Collected | Cash to Hand Over |
|-------|----------|---------------------|--------------------|--------------------|
| Shift A | Operator A | 1 | 0 | Rp 0 |
| Shift B | Operator B | 0 | 1 | Rp 15,000 |

**Reports Implications:**
- "Sessions by operator" report shows who checked in each vehicle
- "Revenue by operator" report shows who collected each payment
- Cash discrepancy calculation uses `transactions.shift_id` for accuracy

---

## 17.10 Impact on Reports

| Report | Shift Impact |
|--------|-------------|
| Daily Revenue Summary | Can now be filtered by shift |
| Operator Activity Log | Activities grouped by shift |
| Cash discrepancy | New report: shifts with non-zero discrepancy |

---

## 17.11 Permissions

| Permission | Description |
|-----------|-------------|
| `shifts:start` | Start a shift (operator) |
| `shifts:end` | End a shift (operator) |
| `shifts:view` | View shift records (manager) |
| `shifts:force_close` | Forcibly close an unclosed shift (manager) |
| `shifts:resolve_discrepancy` | Mark a flagged shift as resolved (manager) |

---

## 17.12 Data Model

```sql
shifts
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  operator_id             UUID NOT NULL REFERENCES users(id),
  location_id             UUID NOT NULL REFERENCES locations(id),
  status                  VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                            CHECK (status IN ('OPEN', 'CLOSED', 'FLAGGED', 'RESOLVED', 'FORCE_CLOSED')),
  started_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at                TIMESTAMPTZ,
  expected_cash           NUMERIC(12,2),   -- populated on shift end
  cash_handover_amount    NUMERIC(12,2),   -- entered by operator
  discrepancy             NUMERIC(12,2),   -- cash_handover - expected_cash
  discrepancy_notes       TEXT,            -- manager resolution notes
  force_closed_by         UUID REFERENCES users(id),
  force_closed_reason     TEXT,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shifts_operator    ON shifts (operator_id);
CREATE INDEX idx_shifts_location    ON shifts (location_id, started_at);
CREATE INDEX idx_shifts_status      ON shifts (status);
```

Sessions are linked to shifts:

```sql
-- Add to sessions table:
ALTER TABLE sessions ADD COLUMN shift_id UUID REFERENCES shifts(id);
CREATE INDEX idx_sessions_shift ON sessions (shift_id);
```
