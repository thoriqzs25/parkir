# Chapter 5 — Vehicle Management

## 5.1 Overview

In v2, vehicles are **not identified by plate number** (no LPR/ANPR cameras). Instead, the system uses QR-coded tickets for session tracking. Vehicle type is **configured per gate** (not selected by operator).

There is no persistent vehicle registry. Vehicle type is stored on the session record and used for rate calculation.

---

## 5.2 Supported Vehicle Types

| Type | Code | Description |
|------|------|-------------|
| Car | `CAR` | Standard passenger car, SUV, or van |
| Motorcycle | `MOTO` | Motorcycle or scooter |
| Truck / Heavy Vehicle | `TRUCK` | Truck, pickup, or other heavy vehicle |

- Vehicle type determines which rate is applied at billing.
- Vehicle type is **configured per gate** (via dashboard).
- Gate can be configured for single vehicle type (CAR, MOTO, TRUCK) or ALL.
- Vehicle type cannot be changed after session creation.

---

## 5.3 Gate-Based Vehicle Type

### Configuration

Each gate is configured with a vehicle type:

```json
{
  "gate_id": "GATE-ENTRY-01",
  "gate_type": "ENTRY",
  "vehicle_type": "MOTO"  // CAR, MOTO, TRUCK, or ALL
}
```

**Examples:**
- Entry gate 1: motorcycle only (`MOTO`).
- Entry gate 2: car only (`CAR`).
- Entry gate 3: all vehicle types (`ALL`).

### Driver Flow

1. Driver approaches entry gate.
2. Driver selects gate based on vehicle type (signage).
3. Driver presses button → ticket dispensed.
4. Ticket has vehicle type printed (from gate config).
5. Session created with vehicle type (from gate config).

### Why Gate-Based (Not Operator-Selected)?

- No operators (fully automated).
- Faster throughput (no selection step).
- Reduces errors (driver selects correct gate).
- Matches AMB's current system (dedicated gates per vehicle type).

---

## 5.4 Vehicle Type in Session

### Session Record

```json
{
  "session_id": "uuid",
  "location_id": "uuid",
  "gate_id": "GATE-ENTRY-01",
  "vehicle_type": "MOTO",  // from gate config
  "check_in_at": "2026-08-06T10:30:00Z",
  ...
}
```

### Fee Calculation

Vehicle type determines which rate to use:

```
rate_config = lookup_rate(location_id, vehicle_type, check_in_date)
fee = calculate_fee(duration, rate_config)
```

**Example:**
- Vehicle type: MOTO.
- Rate: first_hour = Rp 2,000, subsequent = Rp 1,000, daily_cap = Rp 15,000.
- Duration: 3 hours.
- Fee: 2,000 + (2 × 1,000) = Rp 4,000.

---

## 5.5 Vehicle Type in Reports

### Revenue by Vehicle Type

Dashboard report: revenue breakdown by vehicle type.

| Vehicle Type | Transactions | Revenue | Avg |
|--------------|--------------|---------|-----|
| CAR | 500 | Rp 10,000,000 | Rp 20,000 |
| MOTO | 800 | Rp 4,000,000 | Rp 5,000 |
| TRUCK | 100 | Rp 3,000,000 | Rp 30,000 |

### Occupancy by Vehicle Type

Dashboard report: active sessions by vehicle type.

| Vehicle Type | Active Sessions |
|--------------|-----------------|
| CAR | 45 |
| MOTO | 120 |
| TRUCK | 8 |

---

## 5.6 Vehicle Type Management (Dashboard)

### Vehicle Type List

Default vehicle types (system-defined):
- CAR
- MOTO
- TRUCK

**Note:** Vehicle types are hardcoded (not user-configurable in v2). Can be extended in v3 if needed.

### Gate Configuration

Admin configures gate vehicle type via dashboard:

1. Navigate to Gates page.
2. Select gate.
3. Edit gate config.
4. Set vehicle type (CAR, MOTO, TRUCK, or ALL).
5. Save → config flows to server room app → gate app.

---

## 5.7 Design Decisions

**Why no plate number tracking?**
- No LPR/ANPR cameras (cost, complexity).
- QR ticket is sufficient for session tracking.
- Faster entry/exit (no plate recognition delay).
- Can be added later if needed (v3).

**Why gate-based vehicle type?**
- No operators (fully automated).
- Faster throughput (no selection step).
- Reduces errors (driver selects correct gate).
- Matches AMB's current system.

**Why hardcoded vehicle types?**
- Simple (3 types cover 99% of cases).
- Rates configured per type.
- Can be extended later if needed.

**Why no vehicle registry?**
- No need (QR ticket tracks session).
- Simpler (no persistent vehicle data).
- Historical lookup via session search.

---

*End of Chapter 5 — Vehicle Management (v2)*
