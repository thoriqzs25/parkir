# Chapter 4 — Locations

## 4.1 Overview

A **Location** represents a single physical parking facility. The system
manages multiple locations under one administrative instance. Each location is
independently configured with its own rates, capacity, and assigned operators,
but all locations are visible and manageable from the central web dashboard.

---

## 4.2 Location Attributes

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | System-generated unique identifier |
| `name` | String | Human-readable name (e.g. "Grand Mall Parking") |
| `code` | String | Short unique identifier used in receipts and reports (e.g. `GMP-01`) |
| `address` | String | Full street address |
| `city` | String | City |
| `status` | Enum | `ACTIVE` or `INACTIVE` |
| `capacity` | JSON | Total slot count per vehicle type (see 4.3) |
| `created_at` | Timestamp | When the location was added |
| `updated_at` | Timestamp | Last modified |

---

## 4.3 Capacity Configuration

Each location defines a capacity per vehicle type. This is used for occupancy percentage calculations in the dashboard.

```json
{
  "CAR": 100,
  "MOTO": 50,
  "TRUCK": 20
}
```

- Capacity values are informational — the system does not hard-block check-ins when capacity is reached in v1.
- Occupancy percentage = `(active sessions of type / capacity of type) × 100`
- If capacity is not set for a vehicle type, occupancy % is shown as `N/A`.

---

## 4.4 Rate Configuration per Location

Each location has its own rate table per vehicle type and rate model.

| Field | Type | Description |
|-------|------|-------------|
| `location_id` | UUID | FK to location |
| `vehicle_type` | Enum | CAR, MOTO, TRUCK |
| `first_hour_rate` | Decimal | Fee for the first hour (or fraction) |
| `subsequent_hourly_rate` | Decimal | Fee per hour after the first hour |
| `daily_flat_rate` | Decimal | Maximum daily cap; applies when total exceeds this |
| `effective_from` | Date | When this rate becomes active |
| `effective_until` | Date | Optional end date (null = indefinite) |

### Rate Calculation Logic
1. Calculate duration in hours (round up to next full hour; minimum 1 hour).
2. Apply `first_hour_rate` for the first hour.
3. Apply `subsequent_hourly_rate` × (duration_hours - 1) for remaining hours.
4. Sum both; if result > `daily_flat_rate`, apply `daily_flat_rate` instead.
5. The applicable rate is the one where `effective_from <= check_in_date <= effective_until`.

### Example Rate Table

| Vehicle Type | First Hour | Subsequent /hr | Daily Flat Rate |
|-------------|-----------|---------------|----------------|
| CAR | Rp 5,000 | Rp 3,000 | Rp 30,000 |
| MOTO | Rp 2,000 | Rp 1,000 | Rp 15,000 |
| TRUCK | Rp 15,000 | Rp 10,000 | Rp 100,000 |

---

## 4.5 Operator Assignment

- Operators are assigned to one or more locations via the `user_locations` table (see Chapter 3).
- An operator can only check in vehicles at their assigned location(s).
- If an operator is assigned to multiple locations, they select the active location at login or session start.

---

## 4.6 Location Management

### Owner / Admin Capabilities

Only **Owners** and **System Administrators** can:
- Create a new location
- Edit location name, address, code, capacity
- Activate or deactivate a location
- Manage rate configurations per vehicle type
- Assign or remove operators from a location

### Manager Capabilities

**Facility Managers** have supervisory access only:
- View location details and assigned operators
- View occupancy and session activity
- Authorize manual gate open (via incident workflow)
- Authorize void transactions (sign-off with PIN)
- Report incidents and operational issues
- View reports for their assigned location(s)

Managers **cannot**:
- Create, edit, or deactivate locations
- Modify rate configurations
- Assign operators to locations

### Deactivating a Location

When a location is deactivated (by Owner/Admin):
- Prevents new sessions from being opened at that location.
- Existing `ACTIVE` sessions at the location remain open and must be manually closed.
- Historical data (sessions, transactions, reports) is preserved.
- Operators assigned to that location can no longer select it as their active location.

---

## 4.7 Dashboard Aggregation

The web dashboard supports two viewing modes:

| Mode | Description |
|------|-------------|
| **Single Location** | Shows data for one selected location |
| **All Locations** | Aggregates data across all locations the user has access to |

Metrics available per location or aggregated:
- Current active sessions (by vehicle type)
- Occupancy percentage
- Today's revenue
- Open incidents count

---

## 4.8 Data Model

```
locations
  id                UUID, primary key
  name              VARCHAR(150), not null
  code              VARCHAR(20), unique, not null
  address           TEXT
  city              VARCHAR(100)
  status            ENUM('ACTIVE', 'INACTIVE'), default ACTIVE
  capacity          JSONB  -- { "CAR": 100, "MOTO": 50, "TRUCK": 20 }
  created_at        TIMESTAMP
  updated_at        TIMESTAMP

location_rates
  id                UUID, primary key
  location_id       UUID, FK → locations.id
  vehicle_type      ENUM('CAR', 'MOTO', 'TRUCK'), not null
  hourly_rate       NUMERIC(12, 2), not null
  daily_flat_rate   NUMERIC(12, 2), not null
  effective_from    DATE, not null
  effective_until   DATE, nullable
  created_by        UUID, FK → users.id
  created_at        TIMESTAMP
```
