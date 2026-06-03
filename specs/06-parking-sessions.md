# Chapter 6 — Parking Sessions

## 6.1 Overview

A **parking session** is the core operational unit of the system. It represents a single vehicle's stay — from the moment it enters to the moment it exits and payment is confirmed. Every billing, reporting, and audit action is anchored to a session.

---

## 6.2 Session Lifecycle

```
         ┌──────────┐
         │ CHECK-IN │  Operator enters plate + vehicle type
         └────┬─────┘
              │
              ▼
        ┌───────────┐
        │  ACTIVE   │  Vehicle is parked; session clock running
        └────┬──────┘
             │  Operator initiates check-out
             ▼
    ┌──────────────────┐
    │ PENDING_PAYMENT  │  Fee calculated; awaiting payment
    └────────┬─────────┘
             │  Payment confirmed
             ▼
    ┌──────────────────┐
    │     CLOSED       │  Receipt printed; session complete
    └──────────────────┘

   At any point (ACTIVE or PENDING_PAYMENT):
             │  Manager authorizes void
             ▼
    ┌──────────────────┐
    │     VOIDED       │  Session cancelled; excluded from revenue
    └──────────────────┘
```

---

## 6.3 Session States

| State | Description | Transitions To |
|-------|-------------|---------------|
| `ACTIVE` | Vehicle is parked; check-in recorded | `PENDING_PAYMENT`, `VOIDED` |
| `PENDING_PAYMENT` | Check-out initiated; fee calculated; awaiting payment | `CLOSED`, `ACTIVE` (if cancelled back), `VOIDED` |
| `CLOSED` | Payment confirmed; receipt printed | `VOIDED` (via manual adjustment only) |
| `VOIDED` | Session cancelled; terminal state | — |

### Notes on State Transitions
- A session in `PENDING_PAYMENT` can be reverted to `ACTIVE` if the operator cancels the check-out (e.g. driver decides to stay).
- A `CLOSED` session can only be voided by a manager with the `adjustments:void_transaction` permission.
- `VOIDED` is a terminal state — it cannot be re-opened.

---

## 6.4 Check-in Flow

**Actor:** Operator (Desktop App)

1. Operator opens the check-in form.
2. Operator enters:
   - Plate number (text input, auto-uppercased)
   - Vehicle type (dropdown: CAR / MOTO / TRUCK)
3. System checks for duplicate active plate at the same location.
   - If found: show warning with existing session details. Operator confirms or cancels.
4. On confirm: system creates session with state `ACTIVE`, records `check_in_at` (server time), `operator_id`, `location_id`.
5. Confirmation shown to operator with session ID and check-in timestamp.

**Validation:**
- Plate number: required, non-blank, max 20 chars.
- Vehicle type: required, must be one of CAR / MOTO / TRUCK.
- Operator must be assigned to the active location.

---

## 6.5 Check-out Flow

**Actor:** Operator (Desktop App)

1. Operator searches for the session by plate number or session ID.
2. System displays session details: plate, vehicle type, check-in time, elapsed duration.
3. System calculates the fee (see Chapter 7 for billing logic).
4. Operator confirms check-out — session moves to `PENDING_PAYMENT`.
5. Operator selects payment method (Cash / Digital) and records payment.
6. System confirms payment, moves session to `CLOSED`, records `check_out_at`.
7. Receipt is sent to thermal printer automatically.

**Validation:**
- Session must be in `ACTIVE` state to initiate check-out.
- Payment amount must be >= calculated fee (for cash: change is displayed).
- Payment method is required.

---

## 6.6 Fee Calculation

Fee is calculated at the moment check-out is initiated.

```
duration_hours = CEIL((check_out_at - check_in_at) in hours)
if duration_hours == 0: duration_hours = 1   -- minimum 1 hour

if duration_hours == 1:
    raw_fee = first_hour_rate
else:
    raw_fee = first_hour_rate + (duration_hours - 1) × subsequent_hourly_rate

fee = MIN(raw_fee, daily_flat_rate)
```

- The applicable rate is fetched from `location_rates` where `vehicle_type` matches and `effective_from <= check_in_date`.
- If no rate is configured for the vehicle type at that location, the system blocks check-out and alerts the operator to contact a manager.

### Example
Config: first_hour = Rp 5,000 | subsequent /hr = Rp 3,000 | daily cap = Rp 30,000

| Scenario | Duration | Calculation | Fee |
|----------|----------|-------------|-----|
| Short stay (CAR) | 45min → 1h | first_hour only | Rp 5,000 |
| Medium stay (CAR) | 2h 15min → 3h | 5,000 + (2 × 3,000) | Rp 11,000 |
| All day (CAR) | 10h | 5,000 + (9 × 3,000) = 32,000 → capped | Rp 30,000 |
| Overnight (MOTO) | 14h | 2,000 + (13 × 1,000) = 15,000 → capped | Rp 15,000 |

---

## 6.7 Session Search

Operators and managers can search sessions using:

| Filter | Notes |
|--------|-------|
| Plate number | Partial or full match |
| Session ID | Exact match |
| State | ACTIVE / PENDING_PAYMENT / CLOSED / VOIDED |
| Location | Scoped to user's accessible locations |
| Date range | By `check_in_at` |
| Vehicle type | CAR / MOTO / TRUCK |

---

## 6.8 Offline Sessions

When the operator desktop app is in offline mode:
- Sessions are created locally with a temporary ID and flagged `offline_sync = true`.
- Check-out and payment can also be completed offline using locally cached rates.
- On reconnect, all unsynced sessions are pushed to the backend in `check_in_at` order.
- If a sync conflict occurs (e.g. duplicate plate), the session is flagged `SYNC_CONFLICT` and presented to the manager for review.

See Chapter 11 for full offline mode behavior.

---

## 6.9 Data Model

```
sessions
  id                UUID, primary key
  location_id       UUID, FK → locations.id, not null
  operator_id       UUID, FK → users.id, not null
  plate             VARCHAR(10), not null  -- normalized: A-1234-BCD
  city_code         VARCHAR(4), not null   -- extracted from plate prefix
  vehicle_type      ENUM('CAR', 'MOTO', 'TRUCK'), not null
  state             ENUM('ACTIVE', 'PENDING_PAYMENT', 'CLOSED', 'VOIDED'), not null
  check_in_at       TIMESTAMP WITH TIME ZONE, not null
  check_out_at      TIMESTAMP WITH TIME ZONE, nullable
  fee_amount        NUMERIC(12,2), nullable  -- populated at check-out
  rate_snapshot     JSONB, nullable          -- snapshot of rate at time of billing
  offline_sync      BOOLEAN, default false
  sync_conflict     BOOLEAN, default false
  created_at        TIMESTAMP
  updated_at        TIMESTAMP
```

### `rate_snapshot` Structure
```json
{
  "first_hour_rate": 5000,
  "subsequent_hourly_rate": 3000,
  "daily_flat_rate": 30000,
  "vehicle_type": "CAR",
  "effective_from": "2025-01-01"
}
```
Storing a snapshot at billing time ensures historical accuracy even if rates change later.
