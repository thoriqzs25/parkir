# Chapter 6 — Parking Sessions

## 6.1 Overview

A **parking session** is the core operational unit of the system. It represents a single vehicle's stay — from the moment it enters (ticket dispensed) to the moment it exits (payment completed). Every billing, reporting, and audit action is anchored to a session.

In v2, sessions are created and managed automatically by the system (no operators). The gate app handles hardware interaction, while the server room app handles business logic.

---

## 6.2 Session Lifecycle

```
         ┌──────────────┐
         │   ACTIVE     │  Vehicle entered; ticket dispensed; gate opened
         └──────┬───────┘
                │  Driver scans QR ticket at exit gate
                ▼
    ┌──────────────────────┐
    │  PENDING_PAYMENT     │  Fee calculated; displayed on monitor; awaiting tap-to-go payment
    └──────────┬───────────┘
               │  Payment confirmed (e-money/Flazz tap)
               ▼
    ┌──────────────────────┐
    │      CLOSED          │  Gate opened; session complete
    └──────────────────────┘

   At any point (ACTIVE or PENDING_PAYMENT):
               │  Hardware failure / exception / manager void
               ▼
    ┌──────────────────────┐
    │      VOIDED          │  Session cancelled; excluded from revenue
    └──────────────────────┘
```

---

## 6.3 Session States

| State | Description | Transitions To |
|-------|-------------|---------------|
| `ACTIVE` | Vehicle entered; ticket dispensed; gate opened | `PENDING_PAYMENT`, `VOIDED` |
| `PENDING_PAYMENT` | QR scanned; fee calculated; awaiting payment | `CLOSED`, `VOIDED` |
| `CLOSED` | Payment confirmed; gate opened | `VOIDED` (via manager void only) |
| `VOIDED` | Session cancelled; terminal state | — |

### Notes on State Transitions
- A session in `ACTIVE` remains until driver exits or exception occurs.
- A session in `PENDING_PAYMENT` transitions to `CLOSED` on successful payment, or `VOIDED` on exception/void.
- A `CLOSED` session can only be voided by a manager with the `payments:void` permission (requires manager PIN).
- `VOIDED` is a terminal state — it cannot be re-opened.

---

## 6.4 Entry Flow (Automated)

**Actor:** System (Gate App + Server Room App)

1. Vehicle enters loop sensor → Gate app detects vehicle presence.
2. Driver presses ticket button.
3. Gate app verifies vehicle is still in loop sensor.
   - If vehicle not detected: ignore button press (prevents wasted tickets).
   - If vehicle detected: continue.
4. Gate app → Server Room App: `POST /gate/session/create`
5. Server Room App:
   - Generates `session_id` (UUID).
   - Determines current shift (from shift config).
   - Assigns `shift_number` (increments continuously).
   - Creates session in local DB with state `ACTIVE`.
   - Adds session to sync queue (for cloud sync).
6. Server Room App → Gate App: returns `{session_id, location_id, timestamp, vehicle_type, shift_number}`.
7. Gate app:
   - Generates QR code encoding: `session_id`, `location_id`, `timestamp`, `vehicle_type`.
   - Prints ticket via thermal printer (QR code + timestamp + vehicle type + warning text: "Kunci kendaraan anda dengan rapat. Jangan tinggalkan karcis parkir di dalam kendaraan Anda").
   - Opens gate motor.
8. Vehicle enters → loop sensor clears → gate closes.

**Validation:**
- Vehicle must be detected in loop sensor (prevents wasted tickets).
- Gate must be configured with vehicle type (set via dashboard).
- Printer must be operational (if jammed/empty: gate stays closed, alert triggered).

**QR Code Format:**
```json
{
  "session_id": "uuid-v4",
  "location_id": "uuid-v4",
  "timestamp": "2026-08-06T10:30:00Z",
  "vehicle_type": "motorcycle"
}
```

---

## 6.5 Exit Flow (Automated)

**Actor:** System (Gate App + Server Room App)

1. Driver scans QR ticket at fixed-mount scanner (USB HID).
2. Gate app reads QR data → parses `session_id`.
3. Gate app → Server Room App: `POST /gate/session/{session_id}/calculate-fee`
4. Server Room App:
   - Looks up session in local DB (must be in `ACTIVE` state).
   - Calculates fee using local rate config (polled from cloud every 1 min).
   - Assigns checkout `shift_number`.
   - Updates session: `state = PENDING_PAYMENT`, `fee_amount`, `check_out_at`, `shift_number`.
   - Returns: `{fee_amount, check_in_time, duration, vehicle_type, shift_number}`.
5. Gate app displays on driver-facing monitor:
   - Fee amount (large, bold).
   - Check-in time.
   - Duration (e.g., "2 hours 15 minutes").
   - Vehicle type.
   - Instruction: "Please tap your e-money card".
6. Driver taps e-money/Flazz card → payment terminal processes.
7. Payment vendor signals success → Gate app.
8. Gate app → Server Room App: `POST /gate/session/{session_id}/close`
9. Server Room App:
   - Creates transaction record (amount, payment_method, payment_reference).
   - Updates session: `state = CLOSED`.
   - Adds transaction to sync queue.
   - Returns: `{transaction_id}`.
10. Gate app opens gate motor.
11. Vehicle exits → gate closes.
12. (Optional) Driver presses receipt button → Gate app prints receipt (check-in/out time, fee, vehicle type, shift number).

**Validation:**
- QR code must be readable (if unreadable: gate stays closed, alert triggered).
- Session must exist in local DB and be in `ACTIVE` state.
- Payment must succeed (if failed: gate stays closed, display "Insufficient balance, please topup").
- If exception occurs: driver presses alert button → staff handles via offline SOP.

---

## 6.6 Fee Calculation

Fee is calculated by the **Server Room App** at the moment QR is scanned (check-out initiated).

**Inputs:**
- `check_in_at` (from session)
- `check_out_at` (current time)
- `vehicle_type` (from session)
- `location_id` (from session)
- Local rate config (polled from cloud, versioned)

**Calculation Logic:**
```
duration_hours = CEIL((check_out_at - check_in_at) / 3600 seconds)
duration_hours = MAX(duration_hours, 1)  // minimum 1 hour

// Recurring 24-hour block model
total_fee = 0
remaining_hours = duration_hours

WHILE remaining_hours > 0:
    block_hours = MIN(remaining_hours, 24)
    block_fee = first_hour_rate
    IF block_hours > 1:
        block_fee += (block_hours - 1) * subsequent_hourly_rate
    block_fee = MIN(block_fee, daily_flat_rate)  // cap at daily rate
    total_fee += block_fee
    remaining_hours -= block_hours

RETURN total_fee
```

**Example:**
- Check-in: 08:00, Check-out: 11:30 (3.5 hours → 4 hours)
- Rate: first_hour = 5000, subsequent = 3000, daily_cap = 50000
- Fee: 5000 + (3 * 3000) = 14000

**Config Version Tracking:**
- Transaction records `rate_config_version` (which rate version was used).
- Ensures audit trail even if rates change later.

---

## 6.7 Shift Assignment

Shift numbers are assigned by the **Server Room App** based on check-in/check-out time.

**Default Shift Config (3 shifts/day):**
- Shift 1: Morning (e.g., 06:00 - 14:00)
- Shift 2: Afternoon (e.g., 14:00 - 22:00)
- Shift 3: Evening (e.g., 22:00 - 06:00)

**Shift Number Increment:**
- Shift numbers increment continuously across days.
- Example: Day 1: shift_1, shift_2, shift_3 → Day 2: shift_4, shift_5, shift_6 → ...

**Assignment Logic:**
```
current_time = now()
shift_config = lookup_shift_config(location_id, current_time)
shift_number = calculate_continuous_shift_number(shift_config, current_date)
```

**Storage:**
- Session records `shift_number` (assigned at check-in).
- Transaction records `shift_number` (assigned at check-out, may differ from session if overnight).

---

## 6.8 Session Data Model

```
Session {
  session_id: UUID (primary key)
  location_id: UUID (foreign key → locations)
  gate_id: string (entry gate hardware ID)
  vehicle_type: string (car, motorcycle, truck)
  check_in_at: timestamp
  check_out_at: timestamp (nullable)
  fee_amount: integer (nullable, set at check-out)
  rate_config_version: string (nullable, set at check-out)
  shift_config_version: string (nullable, set at check-out)
  shift_number: integer (assigned at check-in)
  state: enum (ACTIVE, PENDING_PAYMENT, CLOSED, VOIDED)
  qr_data: string (QR code content for audit)
  created_at: timestamp
  updated_at: timestamp
  synced_at: timestamp (nullable, when synced to cloud)
}
```

**Notes:**
- No `operator_id` (no operators in v2).
- `gate_id` is hardware ID (from gate app), not a database entity.
- `qr_data` stores what was encoded in QR code (for audit/validation).
- `synced_at` tracks when session was synced to cloud backend.

---

## 6.9 Sync Behavior

**Server Room App → Cloud Backend:**
- Sessions are synced to cloud every 1 minute (batch sync).
- Sync includes: all sessions (ACTIVE, PENDING_PAYMENT, CLOSED, VOIDED).
- Cloud backend stores sessions for reporting and dashboard.
- Duplicate detection: cloud checks `session_id`, skips if exists.

**Offline Mode:**
- If internet is down, sessions stored in local DB only.
- Sync queue tracks unsynced sessions.
- When internet restored, sync queue processed (exponential backoff on failure).

---

## 6.10 Exception Handling

**Hardware Failures:**
- Printer jammed/empty: gate stays closed, alert triggered, staff refills/fixes.
- QR scanner unreadable: gate stays closed, alert triggered, staff runs offline SOP.
- Gate motor failure: gate stays closed, alert triggered, staff handles.
- Payment terminal failure: gate stays closed, alert triggered, staff runs offline SOP.

**Payment Failures:**
- Insufficient balance: gate stays closed, display "Insufficient balance, please topup".
- Card error: gate stays closed, display error message.
- Driver asks for help: staff taps with their own e-money (driver pays cash to staff).

**Alert Flow:**
- Driver presses alert button (physical) → Gate app → Server Room App.
- Server Room App plays audio alert: "Gate {gate_id} needs assistance".
- Staff walks to gate, handles exception (offline SOP if needed).

---

## 6.11 Design Decisions

**Why no plate number tracking?**
- Simplifies entry flow (no LPR/ANPR camera needed).
- Reduces hardware cost and complexity.
- QR ticket is sufficient for session tracking.
- Can be added later if needed (future enhancement).

**Why shift_number in session and transaction?**
- Audit trail: know which shift session was created in.
- Reporting: aggregate by shift for reconciliation.
- Shift numbers increment continuously (not reset daily).

**Why config version tracking?**
- Audit trail: know which rate/shift config was used.
- If rates change, old transactions still reference old rates.
- Prevents retroactive fee recalculation.

**Why server room app calculates fee (not cloud)?**
- Offline resilience: gates work without internet.
- Lower latency: no round-trip to cloud for fee calculation.
- Local rate config polled every 1 min (fresh enough).

---

*End of Chapter 6 — Parking Sessions (v2)*
