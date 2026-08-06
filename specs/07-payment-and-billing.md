# Chapter 7 — Payment & Billing

## 7.1 Overview

Payment is collected at the end of a parking session (pay-on-exit model). The system calculates the fee based on duration, vehicle type, and the active rate for the location. **Only tap-to-go payment methods are supported** (e-money, Flazz). Payment is processed by a third-party vendor; PARKIR receives a completion signal and records the transaction.

Every completed payment produces an immutable transaction record with config version tracking for audit purposes.

---

## 7.2 Rate Models

The system supports the following rate models. For multi-day stays, a **recurring 24-hour block model** is used.

| Model | Description |
|-------|-------------|
| **First-Hour Rate** | A separate, distinct rate applied to the first hour of each 24-hour block |
| **Subsequent Hourly Rate** | Rate applied per hour (or fraction) after the first hour within each block |
| **Daily Flat Rate** | A maximum cap applied per 24-hour block when the total would otherwise exceed it |
| **Pay-on-Exit** | Fee is always calculated at check-out; never pre-charged |

### How They Work Together (Multi-Day Model)

For multi-day stays, the fee is calculated in 24-hour blocks:

```
duration_hours = CEIL((check_out_at - check_in_at) / 3600)
duration_hours = MAX(duration_hours, 1)  // minimum 1 hour

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

### Example

Config: first_hour_rate = Rp 5,000 | subsequent_hourly_rate = Rp 3,000 | daily_flat_rate = Rp 50,000

| Duration | Calculation | Fee |
|----------|-------------|-----|
| 30 min → 1h | first_hour only | Rp 5,000 |
| 2h 10min → 3h | 5,000 + (2 × 3,000) | Rp 11,000 |
| 10h | 5,000 + (9 × 3,000) = 32,000 → capped | Rp 50,000 |
| 25h (multi-day) | Block 1: 50,000 + Block 2: 5,000 | Rp 55,000 |
| 48h (2 full days) | Block 1: 50,000 + Block 2: 50,000 | Rp 100,000 |

> If `first_hour_rate` and `subsequent_hourly_rate` are configured to the same value, the system behaves as a simple flat hourly rate within each 24-hour block.

---

## 7.3 Rate Configuration

### Configuration Structure

Rates are configured per location and vehicle type:

```json
{
  "location_id": "uuid",
  "vehicle_type": "CAR",
  "version": "rate_v42",
  "first_hour_rate": 500000,  // Rp 5,000 (in cents)
  "subsequent_hourly_rate": 300000,  // Rp 3,000
  "daily_flat_rate": 5000000,  // Rp 50,000
  "effective_from": "2026-01-01",
  "effective_until": null
}
```

### Configuration Versioning

- Every rate config has a `version` field (e.g., "rate_v1", "rate_v2", ...).
- Version auto-increments on create/update.
- Transactions record `rate_config_version` (which version was used).
- Ensures audit trail even if rates change later.

### Configuration Distribution

**Cloud Backend (Source of Truth):**
- Stores all rate configs (versioned).
- Serves configs to server room apps.

**Server Room App (Local Cache):**
- Polls cloud every 1 minute for config updates.
- Stores configs in local SQLite database.
- Uses local configs for fee calculation (offline resilience).
- If rates change while offline, uses old rates until reconnected.

**Fee Calculation:**
- Server room app calculates fee (not cloud).
- Uses local rate config (polled from cloud).
- Records `rate_config_version` in transaction (audit trail).

---

## 7.4 Payment Methods

### Supported Methods (v2)

| Method | Description | Processing |
|--------|-------------|------------|
| **EMONEY** | E-money card (tap-to-go) | Payment vendor terminal |
| **FLAZZ** | Flazz card (tap-to-go) | Payment vendor terminal |
| **CASH** | Cash payment (staff taps with own e-money) | Staff handles via offline SOP |
| **OFFLINE_SOP** | Offline SOP (staff manual handling) | Staff handles via offline SOP |

### Payment Vendor Integration

**Vendor Responsibility:**
- Payment terminal hardware (e-money/Flazz card reader).
- Card reading and validation.
- Payment processing and authorization.
- Signal completion to PARKIR.

**PARKIR Responsibility:**
- Display fee amount on driver-facing monitor.
- Wait for payment completion signal from vendor.
- Record transaction on payment success.
- Open gate on payment success.
- Handle payment failure (display error, keep gate closed).

**Integration (TBD):**
- Communication protocol: TBD (serial, USB, network?)
- Signal format: TBD (HTTP callback, webhook, shared file?)
- Error handling: TBD (insufficient balance, card error, timeout?)

---

## 7.5 Payment Flow

### Exit Flow (Automated)

1. Driver scans QR ticket at exit gate.
2. Gate app reads QR data → sends to server room app.
3. Server room app:
   - Looks up session in local DB.
   - Calculates fee using local rate config.
   - Returns: `{fee_amount, check_in_time, duration, vehicle_type}`.
4. Gate app displays on driver-facing monitor:
   - Fee amount (large, bold).
   - Check-in time.
   - Duration.
   - Instruction: "Please tap your e-money card".
5. Driver taps e-money/Flazz card.
6. Payment vendor processes payment.
7. Payment vendor signals success to gate app.
8. Gate app → Server room app: `POST /gate/session/{id}/close`.
9. Server room app:
   - Creates transaction record (amount, payment_method, payment_reference).
   - Records `rate_config_version` (audit trail).
   - Updates session state to `CLOSED`.
   - Returns: `{transaction_id}`.
10. Gate app opens gate.
11. Vehicle exits.

### Payment Failure Flow

1. Driver taps card → payment vendor processes.
2. Payment vendor signals failure (insufficient balance, card error).
3. Gate app displays on monitor:
   - "Payment failed" (red X icon).
   - "Insufficient balance, please topup".
4. Gate stays closed.
5. If driver asks for help:
   - Driver presses alert button.
   - Staff walks to gate.
   - Staff handles via offline SOP (e.g., staff taps with own e-money, driver pays cash to staff).
   - Staff records transaction manually (payment_method: CASH or OFFLINE_SOP).

### Offline Payment (Offline SOP)

If payment vendor system is down:
1. Gate app displays: "Payment system offline".
2. Driver presses alert button.
3. Staff handles via offline SOP:
   - Manual gate opening (if authorized).
   - Paper receipt (if needed).
   - Record transaction later (when system restored).
   - Payment method: OFFLINE_SOP.

---

## 7.6 Transaction Record

### Transaction Data

```json
{
  "transaction_id": "uuid",
  "session_id": "uuid",
  "location_id": "uuid",
  "gate_id": "GATE-EXIT-01",
  "amount": 1100000,  // Rp 11,000 (in cents)
  "payment_method": "EMONEY",
  "payment_reference": "vendor-ref-12345",
  "config_version": "rate_v42",
  "shift_number": 15,
  "created_at": "2026-08-06T10:30:00Z",
  "synced_at": "2026-08-06T10:31:00Z",
  "voided": false,
  "voided_at": null,
  "voided_by": null,
  "void_reason": null
}
```

### Transaction Voiding

- Transactions can be voided by a manager (requires `payments:void` permission).
- Void requires manager PIN verification (6-digit PIN).
- Voided transactions are marked: `voided = true`, `voided_at`, `voided_by`, `void_reason`.
- Voided transactions are excluded from revenue reports.
- Audit log records void action.

---

## 7.7 Receipt Printing

### Entry Ticket (Automatic)

Printed automatically on entry:
- QR code (session_id, location_id, timestamp, vehicle_type).
- Timestamp (check-in time).
- Vehicle type.
- Warning text: "Kunci kendaraan anda dengan rapat. Jangan tinggalkan karcis parkir di dalam kendaraan Anda".

### Exit Receipt (Optional)

Printed on demand (driver presses receipt button):
- Check-in time.
- Check-out time.
- Duration.
- Fee amount.
- Vehicle type.
- Shift number.
- Transaction ID.
- Payment method.

---

## 7.8 Sync Behavior

### Transaction Sync

**Server Room App → Cloud Backend:**
- Transactions synced to cloud every 1 minute (batch sync).
- Sync includes: all transactions (including voided).
- Cloud stores transactions for reporting and dashboard.
- Duplicate detection: cloud checks `transaction_id`, skips if exists.

**Offline Mode:**
- If internet is down, transactions stored in local DB only.
- Sync queue tracks unsynced transactions.
- When internet restored, sync queue processed (exponential backoff on failure).

---

## 7.9 Fee Calculation Examples

### Example 1: Short Stay (1 hour)

- Check-in: 08:00, Check-out: 08:45 (45 min → 1 hour)
- Rate: first_hour = Rp 5,000, subsequent = Rp 3,000, daily_cap = Rp 50,000
- Fee: Rp 5,000

### Example 2: Medium Stay (3 hours)

- Check-in: 08:00, Check-out: 11:10 (3h 10min → 4 hours)
- Rate: first_hour = Rp 5,000, subsequent = Rp 3,000, daily_cap = Rp 50,000
- Fee: 5,000 + (3 × 3,000) = Rp 14,000

### Example 3: Long Stay (10 hours, hits daily cap)

- Check-in: 08:00, Check-out: 18:00 (10 hours)
- Rate: first_hour = Rp 5,000, subsequent = Rp 3,000, daily_cap = Rp 50,000
- Fee: min(5,000 + (9 × 3,000), 50,000) = min(32,000, 50,000) = Rp 32,000

### Example 4: Multi-Day Stay (25 hours)

- Check-in: 08:00 Day 1, Check-out: 09:00 Day 2 (25 hours)
- Rate: first_hour = Rp 5,000, subsequent = Rp 3,000, daily_cap = Rp 50,000
- Fee: Block 1 (24h): 50,000 + Block 2 (1h): 5,000 = Rp 55,000

### Example 5: Full 2-Day Stay (48 hours)

- Check-in: 08:00 Day 1, Check-out: 08:00 Day 3 (48 hours)
- Rate: first_hour = Rp 5,000, subsequent = Rp 3,000, daily_cap = Rp 50,000
- Fee: Block 1 (24h): 50,000 + Block 2 (24h): 50,000 = Rp 100,000

---

## 7.10 Design Decisions

**Why no cash payments?**
- Fully automated gates (no operator to handle cash).
- Faster throughput (tap-to-go is faster than cash).
- Reduces security risk (no cash on-site).
- Aligns with AMB's current system (cashless).

**Why server room app calculates fee (not cloud)?**
- Offline resilience: gates work without internet.
- Lower latency: no round-trip to cloud for fee calculation.
- Local rate config polled every 1 min (fresh enough).

**Why config version tracking?**
- Audit trail: know which rate was used for each transaction.
- If rates change, old transactions still reference old rates.
- Prevents retroactive fee recalculation.
- Supports offline operation (use last known config).

**Why multi-day block model?**
- Fair pricing for long-term parkers.
- Daily cap prevents excessive fees.
- Simple to understand and implement.
- Matches AMB's current pricing model.

**Why payment vendor integration is TBD?**
- AMB will provide payment vendor.
- PARKIR needs to integrate with vendor's terminal.
- Protocol and signal format not yet known.
- Will be finalized during implementation.

---

*End of Chapter 7 — Payment & Billing (v2)*
