# Chapter 5 — Vehicle Management

## 5.1 Overview

Vehicles are not pre-registered in the system. Vehicle identity is captured at the point of check-in and associated with a parking session. The system tracks **plate number** and **vehicle type** as the two mandatory identifiers for every parking event.

There is no persistent vehicle registry — a plate number is simply a string on a session record. Historical lookup is done by searching past sessions.

---

## 5.2 Supported Vehicle Types

| Type | Code | Description |
|------|------|-------------|
| Car | `CAR` | Standard passenger car, SUV, or van |
| Motorcycle | `MOTO` | Motorcycle or scooter |
| Truck / Heavy Vehicle | `TRUCK` | Truck, pickup, or other heavy vehicle |

- Vehicle type determines which rate is applied at billing.
- Vehicle type is selected by the operator at check-in from a fixed dropdown.
- Vehicle type cannot be changed after check-in except via a manual adjustment (see Chapter 12).

---

## 5.2.1 Vehicle Type Classification Options

The following classification approaches were considered. The chosen approach affects operator speed, rate flexibility, and reporting granularity.

### Option A: Simple (Current Choice)
3 types — fast operator selection, simple rate config.

| Type | Code | Covers |
|------|------|--------|
| Car | `CAR` | Passenger car, SUV, van |
| Motorcycle | `MOTO` | Motorcycle, scooter |
| Truck | `TRUCK` | Truck, pickup, heavy vehicle |

### Option B: Size-Based
4-5 types — based on physical space occupied.

| Type | Code | Covers |
|------|------|--------|
| Motorcycle | `MOTO` | Motorcycle, scooter |
| Small | `SMALL` | Compact car, hatchback |
| Medium | `MEDIUM` | Sedan, small SUV |
| Large | `LARGE` | Full-size SUV, MPV, van |
| Extra Large | `XLARGE` | Truck, bus, heavy vehicle |

### Option C: Wheel-Based
3 types — common in Indonesian parking, simple visual check.

| Type | Code | Covers |
|------|------|--------|
| 2-Wheel | `RODA_2` | Motorcycle, scooter |
| 4-Wheel | `RODA_4` | Car, SUV, pickup |
| 6+ Wheel | `RODA_6` | Truck, bus |

### Option D: Expanded Categories
6-8 types — more granular pricing control.

| Type | Code | Covers |
|------|------|--------|
| Motorcycle | `MOTO` | Motorcycle |
| Car | `CAR` | Sedan, hatchback |
| SUV | `SUV` | SUV, crossover |
| MPV | `MPV` | MPV, minivan (Avanza, Innova) |
| Pickup | `PICKUP` | Pickup truck |
| Van | `VAN` | Cargo van, box van |
| Truck | `TRUCK` | Medium/heavy truck |
| Bus | `BUS` | Bus, minibus |

### Option E: Configurable (Admin-Defined)
No hardcoded types — admin creates vehicle types per location.

| Pros | Cons |
|------|------|
| Maximum flexibility | More complex UI and data model |
| Different locations can have different types | Reporting aggregation harder across locations |
| Easy to add new types without code changes | Operator training varies per location |

### Decision Factors

| Factor | Consideration |
|--------|---------------|
| Operator speed | Fewer types = faster check-in (< 2 seconds ideal) |
| Rate flexibility | More types = finer pricing control |
| Indonesian context | Wheel-based (Roda 2/4/6+) is familiar to local operators |
| Future-proofing | Configurable (Option E) allows growth without schema changes |
| MVP simplicity | Option A or C recommended for initial release |

> **Decision:** TBD — awaiting confirmation.

---

## 5.3 Plate Number

### Format
- Free-text input; system normalizes on save.
- **Normalization rules:**
  - Convert to uppercase
  - Replace spaces with dashes
  - Trim leading/trailing whitespace
  - Example: `"f 1327 cbe"` → `"F-1327-CBE"`
- Maximum length: 10 characters (after normalization).
- Standard Indonesian plate format: `[REGION]-[NUMBER]-[SUFFIX]` (e.g., `B-1234-XYZ`, `F-1327-CBE`).

### Validation Rules
- Required — a session cannot be opened without a plate number.
- Must not be blank or whitespace-only.
- Should be unique among currently `ACTIVE` sessions at the same location (warn operator if a duplicate active plate is detected).

### Duplicate Plate Handling
If a plate is already checked in at the same location:
- The operator sees a **warning** with details of the existing active session.
- The operator can choose to proceed (creates a new session) or cancel.
- This handles edge cases like the system missing a previous check-out.

---

## 5.4 Vehicle Data Captured at Check-in

| Field | Source | Notes |
|-------|--------|-------|
| `plate` | Operator input | Normalized: uppercase, dashes (e.g., `BE-1627-PE`) |
| `city_code` | System (auto) | Extracted from plate prefix (e.g., `BE`) |
| `vehicle_type` | Operator selection | CAR / MOTO / TRUCK |
| `check_in_at` | System (auto) | Server timestamp at moment of submission |
| `location_id` | System (auto) | From operator's active location |
| `operator_id` | System (auto) | From authenticated operator session |
| `session_id` | System (auto) | UUID generated on check-in |

### City Code Extraction

The `city_code` is automatically extracted from the plate number prefix (region identifier).

| Input | Normalized Plate | City Code |
|-------|------------------|-----------|
| `BE 1627 PE` | `BE-1627-PE` | `BE` |
| `B 1234 XYZ` | `B-1234-XYZ` | `B` |
| `F 9012 ABC` | `F-9012-ABC` | `F` |
| `DK 5678 CD` | `DK-5678-CD` | `DK` |

**Extraction logic:** Split plate by dash, take first segment.

**Use cases:**
- Analytics: Which regions do most vehicles come from?
- Reports: Revenue breakdown by vehicle origin
- Filtering: Search sessions by city code

---

## 5.5 Vehicle History Lookup

Operators and managers can search historical sessions by plate number.

### Search Behavior
- Returns all sessions (any state) for the given plate, sorted by `check_in_at` descending.
- Results include: session ID, location, vehicle type, check-in time, check-out time, session state, amount paid.
- Available in both the desktop app (for operators) and the web dashboard (for managers).

### Use Cases
- Operator checks if a vehicle is currently parked (duplicate plate detection).
- Manager reviews history of a disputed vehicle.
- Manager traces a plate linked to an incident.

---

## 5.6 Data Model

Vehicle data is embedded in the `sessions` table — there is no separate `vehicles` table in v1.

```
sessions
  id                UUID, primary key
  location_id       UUID, FK → locations.id
  operator_id       UUID, FK → users.id
  plate      VARCHAR(10), not null  -- normalized format: A-1234-BCD
  city_code         VARCHAR(4), not null   -- extracted from plate prefix (e.g., B, BE, DK)
  vehicle_type      ENUM('CAR', 'MOTO', 'TRUCK'), not null
  state             ENUM('ACTIVE', 'PENDING_PAYMENT', 'CLOSED', 'VOIDED')
  check_in_at       TIMESTAMP, not null
  check_out_at      TIMESTAMP, nullable
  offline_sync      BOOLEAN, default false
  created_at        TIMESTAMP
  updated_at        TIMESTAMP
```

### Index Recommendations
```sql
CREATE INDEX idx_sessions_plate ON sessions (plate);
CREATE INDEX idx_sessions_location_state ON sessions (location_id, state);
CREATE INDEX idx_sessions_check_in ON sessions (check_in_at);
```

---

## 5.7 Operator Error Handling

If an operator enters incorrect vehicle data at check-in (wrong plate or wrong vehicle type), this is handled as an **OPERATOR_ERROR** incident combined with a session reassignment or void:

- **Wrong plate number**: File `OPERATOR_ERROR` incident → Manager voids the incorrect session → Operator creates a new correct check-in.
- **Wrong vehicle type**: File `OPERATOR_ERROR` incident → Manager voids the transaction and re-issues with corrected vehicle type (pending manual adjustment flow in Chapter 12).

> Note: In-place editing of plate or vehicle type on an existing session is not supported in v1. Correction always goes through void + re-create.
