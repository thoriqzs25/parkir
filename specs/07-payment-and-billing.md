# Chapter 7 — Payment & Billing

## 7.1 Overview

Payment is collected at the end of a parking session (pay-on-exit model). The system calculates the fee based on duration, vehicle type, and the active rate for the location. Both cash and digital payment methods are supported. Every completed payment produces an immutable transaction record.

---

## 7.2 Rate Models

The system supports the following rate models simultaneously. They are not mutually exclusive — the system automatically applies the correct combination.

| Model | Description |
|-------|-------------|
| **First-Hour Rate** | A separate, distinct rate applied to the first hour of parking |
| **Subsequent Hourly Rate** | Rate applied per hour (or fraction) after the first hour |
| **Daily Flat Rate** | A maximum cap applied when the total would otherwise exceed it |
| **Pay-on-Exit** | Fee is always calculated at check-out; never pre-charged |

### How They Work Together

The daily flat rate functions as a **price ceiling**. For any session:

```
duration_hours     = CEIL((check_out_at - check_in_at) / 3600)

if duration_hours == 0:
    duration_hours = 1  -- minimum 1 hour

if duration_hours == 1:
    raw_fee = first_hour_rate
else:
    raw_fee = first_hour_rate + (duration_hours - 1) × subsequent_hourly_rate

final_fee = MIN(raw_fee, daily_flat_rate)
```

### Example

Config: first_hour_rate = Rp 5,000 | subsequent_hourly_rate = Rp 3,000 | daily_flat_rate = Rp 30,000

| Duration | Calculation | Fee |
|----------|-------------|-----|
| 30 min → 1h | first_hour only | Rp 5,000 |
| 2h 10min → 3h | 5,000 + (2 × 3,000) | Rp 11,000 |
| 10h | 5,000 + (9 × 3,000) = 32,000 → capped | Rp 30,000 |

> If `first_hour_rate` and `subsequent_hourly_rate` are configured to the same value, the system behaves as a simple flat hourly rate.

---

## 7.2.1 Rate Configuration Architecture Options

Two architecture options are available for rate configuration:

### Option A: Fixed Structure (Current)

Hardcoded fields: `first_hour_rate`, `subsequent_hourly_rate`, `daily_flat_rate`.

```json
{
  "vehicle_type": "CAR",
  "first_hour_rate": 5000,
  "subsequent_hourly_rate": 3000,
  "daily_flat_rate": 30000
}
```

**Pros:** Simple, covers 80% of use cases (malls, offices, outdoor lots).
**Cons:** Cannot support flat per-entry, progressive discounts, or more than 2 tiers.

### Option B: Unified Rate Steps (Recommended for Flexibility)

Generalized model using rate steps. Supports multiple rate types through configuration.

#### Rate Types

| Type | Description | Use Case |
|------|-------------|----------|
| `tiered_hourly` | Different rates for different duration ranges | Malls, offices |
| `flat_entry` | Single flat fee regardless of duration | Street parking, outdoor lots |
| `progressive` | Rates decrease with longer stays | Airport long-term parking |

#### Structure

```json
{
  "vehicle_type": "CAR",
  "rate_type": "tiered_hourly",
  "steps": [
    { "up_to_minutes": 60, "rate": 5000 },
    { "up_to_minutes": 240, "rate": 3000 },
    { "up_to_minutes": null, "rate": 2000 }
  ],
  "daily_cap": 30000,
  "rounding": "ceiling_hour"
}
```

#### Calculation Logic

```
function calculateFee(duration_minutes, rate_config):
    if rate_config.rate_type == "flat_entry":
        return rate_config.flat_fee

    total_fee = 0
    remaining_minutes = duration_minutes
    previous_threshold = 0

    for step in rate_config.steps:
        threshold = step.up_to_minutes or INFINITY
        minutes_in_step = MIN(remaining_minutes, threshold - previous_threshold)

        if minutes_in_step <= 0:
            break

        if rate_config.rounding == "ceiling_hour":
            hours_in_step = CEIL(minutes_in_step / 60)
        else:
            hours_in_step = minutes_in_step / 60

        total_fee += hours_in_step * step.rate
        remaining_minutes -= minutes_in_step
        previous_threshold = threshold

    if rate_config.daily_cap:
        total_fee = MIN(total_fee, rate_config.daily_cap)

    return total_fee
```

#### Example Configurations

**Tiered Hourly (Mall):**
```json
{
  "rate_type": "tiered_hourly",
  "steps": [
    { "up_to_minutes": 60, "rate": 5000 },
    { "up_to_minutes": null, "rate": 3000 }
  ],
  "daily_cap": 30000
}
```

**Flat Per-Entry (Street Parking):**
```json
{
  "rate_type": "flat_entry",
  "flat_fee": 5000
}
```

**Progressive Discount (Airport Long-Term):**
```json
{
  "rate_type": "tiered_hourly",
  "steps": [
    { "up_to_minutes": 60, "rate": 10000 },
    { "up_to_minutes": 180, "rate": 8000 },
    { "up_to_minutes": 720, "rate": 5000 },
    { "up_to_minutes": null, "rate": 3000 }
  ],
  "daily_cap": 50000
}
```

**Three-Tier (Extended Stay):**
```json
{
  "rate_type": "tiered_hourly",
  "steps": [
    { "up_to_minutes": 60, "rate": 5000 },
    { "up_to_minutes": 240, "rate": 3000 },
    { "up_to_minutes": null, "rate": 2000 }
  ],
  "daily_cap": 30000
}
```

#### Calculation Examples

**Config:** First hour Rp 5,000, hours 2-4 Rp 3,000/hr, hour 5+ Rp 2,000/hr, cap Rp 30,000

| Duration | Calculation | Fee |
|----------|-------------|-----|
| 45 min → 1h | 1 × 5,000 | Rp 5,000 |
| 2h 30min → 3h | 5,000 + (2 × 3,000) | Rp 11,000 |
| 6h | 5,000 + (3 × 3,000) + (2 × 2,000) | Rp 18,000 |
| 12h | 5,000 + (3 × 3,000) + (8 × 2,000) = 30,000 | Rp 30,000 (capped) |

#### Data Model (Option B)

```sql
-- Rate configuration per location and vehicle type
location_rates
  id                  UUID PRIMARY KEY
  location_id         UUID NOT NULL, FK → locations.id
  vehicle_type        VARCHAR(10) NOT NULL
  rate_type           VARCHAR(20) NOT NULL DEFAULT 'tiered_hourly'
                        CHECK (rate_type IN ('tiered_hourly', 'flat_entry'))
  flat_fee            NUMERIC(12,2)           -- for flat_entry type
  daily_cap           NUMERIC(12,2)           -- optional cap
  rounding            VARCHAR(20) DEFAULT 'ceiling_hour'
                        CHECK (rounding IN ('ceiling_hour', 'ceiling_30min', 'exact'))
  effective_from      DATE NOT NULL
  effective_until     DATE
  created_by          UUID, FK → users.id
  created_at          TIMESTAMPTZ DEFAULT now()

-- Rate steps (for tiered_hourly type)
location_rate_steps
  id                  UUID PRIMARY KEY
  rate_id             UUID NOT NULL, FK → location_rates.id
  step_order          INTEGER NOT NULL        -- 1, 2, 3...
  up_to_minutes       INTEGER                 -- NULL = unlimited
  rate                NUMERIC(12,2) NOT NULL  -- rate per hour in this step

  UNIQUE (rate_id, step_order)

-- Indexes
CREATE INDEX idx_rate_steps_rate ON location_rate_steps (rate_id, step_order);
```

#### Migration from Option A to Option B

Existing `first_hour_rate`, `subsequent_hourly_rate`, `daily_flat_rate` maps to:

```json
{
  "rate_type": "tiered_hourly",
  "steps": [
    { "up_to_minutes": 60, "rate": first_hour_rate },
    { "up_to_minutes": null, "rate": subsequent_hourly_rate }
  ],
  "daily_cap": daily_flat_rate
}
```

### Comparison

| Aspect | Option A (Fixed) | Option B (Unified Steps) |
|--------|------------------|--------------------------|
| Flexibility | Low (2 tiers only) | High (unlimited tiers) |
| Flat per-entry | Not supported | Supported |
| Progressive rates | Not supported | Supported |
| Data model | Simple (3 fields) | More complex (steps table) |
| Calculation | Simple | Slightly more complex |
| MVP suitability | Yes | Yes |

> **Decision:** TBD — Option A sufficient for MVP, Option B recommended if multiple rate types needed.

---

## 7.3 Rate Configuration

Rates are configured per **location** and per **vehicle type**. See Chapter 4 for the rate data model.

### Rate Versioning
- Rate records have `effective_from` and `effective_until` dates.
- The rate applied to a session is the one active on the session's `check_in_at` date.
- If rates change mid-day, sessions that started before the change use the old rate.
- The applied rate is snapshotted in `sessions.rate_snapshot` at check-out time for auditability.

### Rate Not Configured
- If no rate is configured for a vehicle type at a location, the check-out flow is blocked.
- The operator sees an error: "Rate not configured for [VEHICLE_TYPE] at this location. Contact your manager."
- The session remains in `ACTIVE` state until a manager configures the rate.

---

## 7.4 Payment Methods

### Cash

| Field | Description |
|-------|-------------|
| `amount_tendered` | Amount given by the driver (operator-entered) |
| `change_amount` | System-calculated: `amount_tendered - fee_amount` |

- Operator enters the amount received from the driver.
- System displays the change to give back.
- If `amount_tendered < fee_amount`, the system shows an error and does not proceed.

### Digital

| Field | Description |
|-------|-------------|
| `payment_reference` | Reference code from the payment gateway or QR scan |
| `gateway` | Gateway used (e.g. QRIS, GoPay, OVO) — configurable list |

- Digital payment confirmation is manual in v1: the operator confirms once they see the payment notification or receipt from the driver.
- Automated gateway callback integration is out of scope for v1.

### Mixed Payment
- Partial cash + partial digital is **out of scope for v1**.
- A single transaction uses one payment method.

---

## 7.5 Payment Flow

**Actor:** Operator (Desktop App)

1. Check-out is initiated (session moves to `PENDING_PAYMENT`).
2. System displays calculated fee and session summary.
3. Operator selects payment method: **Cash** or **Digital**.
4. **If Cash:**
   - Operator enters amount tendered.
   - System shows change to return.
   - Operator confirms collection.
5. **If Digital:**
   - Operator shows QR / payment details to driver.
   - Driver completes payment on their device.
   - Operator confirms payment received.
6. System creates transaction record, session moves to `CLOSED`.
7. Receipt printed automatically.

---

## 7.6 Transaction Record

A transaction record is created for every successfully closed session.

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique transaction ID |
| `session_id` | UUID | FK → sessions.id |
| `location_id` | UUID | Denormalized for query performance |
| `shift_id` | UUID | Shift during which payment was collected (for cash reconciliation) |
| `operator_id` | UUID | Operator who collected payment (may differ from session.operator_id) |
| `vehicle_type` | Enum | CAR / MOTO / TRUCK |
| `plate` | String | Denormalized from session |
| `check_in_at` | Timestamp | Denormalized from session |
| `check_out_at` | Timestamp | Denormalized from session |
| `duration_hours` | Integer | Ceiling of duration in hours |
| `rate_first_hour` | Decimal | First-hour rate applied |
| `rate_subsequent_hourly` | Decimal | Subsequent hourly rate applied |
| `rate_daily` | Decimal | Daily flat rate configured |
| `fee_amount` | Decimal | Final fee charged |
| `payment_method` | Enum | `CASH` / `DIGITAL` |
| `amount_tendered` | Decimal | Null if digital |
| `change_amount` | Decimal | Null if digital |
| `payment_reference` | String | Null if cash; gateway ref if digital |
| `receipt_number` | String | Formatted receipt identifier |
| `voided` | Boolean | True if voided via manual adjustment |
| `voided_at` | Timestamp | Null if not voided |
| `voided_by` | UUID | FK → users.id; null if not voided |
| `void_reason` | Text | Null if not voided |
| `created_at` | Timestamp | — |

### Receipt Number Format
```
[LOCATION_CODE]-[YYYYMMDD]-[SEQUENCE]
Example: GMP01-20250315-00042
```
Sequence resets to 1 each day per location.

---

## 7.7 Voided Transactions

- A transaction can be voided by a manager with `adjustments:void_transaction` permission.
- Voiding sets `transactions.voided = true` and creates an audit log entry.
- Voided transactions are **excluded from all revenue totals** in reports.
- Voided transactions remain visible in the transaction list with a `VOIDED` badge.
- The linked session is also set to `VOIDED` state.

See Chapter 12 for the full void procedure.

---

## 7.8 Data Model

```
transactions
  id                  UUID, primary key
  session_id          UUID, FK → sessions.id, unique
  location_id         UUID, FK → locations.id
  shift_id            UUID, FK → shifts.id, not null  -- payment collection shift
  operator_id         UUID, FK → users.id, not null   -- payment collector (may differ from session.operator_id)
  vehicle_type        ENUM('CAR', 'MOTO', 'TRUCK'), not null
  plate        VARCHAR(10), not null  -- normalized: A-1234-BCD
  check_in_at         TIMESTAMP WITH TIME ZONE, not null
  check_out_at        TIMESTAMP WITH TIME ZONE, not null
  duration_hours      INTEGER, not null
  rate_hourly         NUMERIC(12,2), not null
  rate_daily          NUMERIC(12,2), not null
  fee_amount          NUMERIC(12,2), not null
  payment_method      ENUM('CASH', 'DIGITAL'), not null
  amount_tendered     NUMERIC(12,2), nullable
  change_amount       NUMERIC(12,2), nullable
  payment_reference   VARCHAR(100), nullable
  receipt_number      VARCHAR(50), not null, unique
  voided              BOOLEAN, default false
  voided_at           TIMESTAMP WITH TIME ZONE, nullable
  voided_by           UUID, FK → users.id, nullable
  void_reason         TEXT, nullable
  created_at          TIMESTAMP
```

### Index Recommendations
```sql
CREATE INDEX idx_transactions_location_date ON transactions (location_id, check_out_at);
CREATE INDEX idx_transactions_operator ON transactions (operator_id);
CREATE INDEX idx_transactions_plate ON transactions (plate);
CREATE INDEX idx_transactions_voided ON transactions (voided) WHERE voided = false;
```
