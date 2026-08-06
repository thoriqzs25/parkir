# Chapter 17 — Shift Management

## 17.1 Overview

The shift system in v2 is **automated and time-based** (not operator-driven). Shift numbers increment continuously across days and are assigned automatically by the server room app based on check-in/check-out time. Shifts are used for reporting, reconciliation, and audit purposes.

**Key difference from v1:** No operators start/end shifts manually. Shifts are predefined time windows (e.g., morning, afternoon, evening) and shift numbers increment continuously.

---

## 17.2 Shift Concepts

### Shift Config

A **shift config** defines the time windows for shifts at a location:

```json
{
  "location_id": "uuid",
  "version": "shift_v15",
  "shift_code": "06-14",
  "shift_number": 1,
  "start_time": "06:00",
  "end_time": "14:00",
  "is_overnight": false
}
```

### Default Shift Config (3 Shifts/Day)

| Shift Code | Start Time | End Time | Is Overnight | Shift Number (Day 1) |
|------------|------------|----------|--------------|---------------------|
| 06-14 | 06:00 | 14:00 | No | 1 |
| 14-22 | 14:00 | 22:00 | No | 2 |
| 22-06 | 22:00 | 06:00 | Yes | 3 |

### Continuous Shift Number Increment

Shift numbers increment **continuously across days** (not reset daily):

```
Day 1: shift_1 (06:00-14:00), shift_2 (14:00-22:00), shift_3 (22:00-06:00)
Day 2: shift_4 (06:00-14:00), shift_5 (14:00-22:00), shift_6 (22:00-06:00)
Day 3: shift_7 (06:00-14:00), shift_8 (14:00-22:00), shift_9 (22:00-06:00)
...
```

**Formula:**
```
shift_number = (day_number - 1) * shifts_per_day + shift_index

Where:
- day_number = days since system start (or epoch)
- shifts_per_day = 3 (default)
- shift_index = 1, 2, or 3 (based on time of day)
```

---

## 17.3 Shift Assignment

### Automatic Assignment

The **server room app** automatically assigns shift numbers based on check-in/check-out time:

**At Check-In (Entry):**
```
current_time = now()
shift_config = lookup_shift_config(location_id, current_time)
shift_number = calculate_continuous_shift_number(shift_config, current_date)
session.shift_number = shift_number
```

**At Check-Out (Exit):**
```
current_time = now()
shift_config = lookup_shift_config(location_id, current_time)
shift_number = calculate_continuous_shift_number(shift_config, current_date)
transaction.shift_number = shift_number
```

**Note:** Session and transaction may have different shift_numbers if the vehicle parks overnight (enters in shift_3, exits in shift_4).

### Shift Lookup Logic

```
function lookup_shift_config(location_id, timestamp):
    configs = get_shift_configs(location_id)
    time = timestamp.time()
    
    for config in configs:
        if config.is_overnight:
            if time >= config.start_time OR time < config.end_time:
                return config
        else:
            if config.start_time <= time < config.end_time:
                return config
    
    return null  // No matching shift (should not happen with proper config)
```

---

## 17.4 Shift Config Management

### Configuration via Dashboard

Admins configure shift configs via the dashboard:

**Create Shift Config:**
- Location (dropdown)
- Shift code (text, e.g., "06-14")
- Start time (time picker)
- End time (time picker)
- Is overnight (checkbox)

**Edit Shift Config:**
- Same fields as create
- Auto-increments version

**Delete Shift Config:**
- Soft delete (mark as inactive)
- Cannot delete if referenced by sessions/transactions

### Config Versioning

- Every shift config has a `version` field (e.g., "shift_v1", "shift_v2", ...).
- Version auto-increments on create/update.
- Sessions and transactions record `shift_config_version` (audit trail).
- Ensures audit trail even if shift configs change later.

### Config Distribution

**Cloud Backend (Source of Truth):**
- Stores all shift configs (versioned).
- Serves configs to server room apps.

**Server Room App (Local Cache):**
- Polls cloud every 1 minute for config updates.
- Stores configs in local SQLite database.
- Uses local configs for shift assignment (offline resilience).

---

## 17.5 Shift Reporting

### Shift Summary Report

Managers can view shift summaries via the dashboard:

**Filters:**
- Location
- Date range
- Shift number (optional)

**Columns:**
| Column | Description |
|--------|-------------|
| Shift Number | Continuous shift number (e.g., 15, 16, 17) |
| Shift Code | Time window (e.g., "06-14") |
| Date | Calendar date |
| Total Sessions | Number of sessions created in this shift |
| Total Transactions | Number of transactions completed in this shift |
| Total Revenue | Sum of transaction amounts |
| Avg Transaction | Average transaction amount |
| Vehicle Breakdown | Count by vehicle type (CAR, MOTO, TRUCK) |

**Example:**
| Shift | Code | Date | Sessions | Transactions | Revenue | Avg |
|-------|------|------|----------|--------------|---------|-----|
| 15 | 06-14 | 2026-08-06 | 120 | 115 | Rp 1,500,000 | Rp 13,043 |
| 16 | 14-22 | 2026-08-06 | 95 | 90 | Rp 1,200,000 | Rp 13,333 |
| 17 | 22-06 | 2026-08-06 | 30 | 28 | Rp 400,000 | Rp 14,286 |

### Export

- Shift summaries exportable to CSV.
- Includes all columns and filters.

---

## 17.6 Overnight Shifts

### Definition

An **overnight shift** crosses midnight (e.g., 22:00 - 06:00).

### Handling

**Shift Assignment:**
```
If current_time >= 22:00:
    shift = evening shift (22:00-06:00) for current date
Else if current_time < 06:00:
    shift = evening shift (22:00-06:00) for previous date
Else:
    shift = morning or afternoon shift for current date
```

**Example:**
- 2026-08-06 23:00 → shift_3 (evening shift of Aug 6)
- 2026-08-07 02:00 → shift_3 (evening shift of Aug 6, continues into Aug 7)
- 2026-08-07 07:00 → shift_4 (morning shift of Aug 7)

**Reporting:**
- Overnight shift spans two calendar dates.
- Report shows shift by shift_number (not by date).
- Sessions/transactions grouped by shift_number.

---

## 17.7 Data Model

### Shift Config Table

```sql
CREATE TABLE shift_configs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_id     UUID NOT NULL REFERENCES locations(id),
  version         VARCHAR(50) NOT NULL,  -- e.g., "shift_v15"
  shift_code      VARCHAR(20) NOT NULL,  -- e.g., "06-14"
  shift_number    INTEGER NOT NULL,  -- display number within day (1, 2, 3)
  start_time      TIME NOT NULL,
  end_time        TIME NOT NULL,
  is_overnight    BOOLEAN NOT NULL DEFAULT false,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Session Table (shift_number field)

```sql
CREATE TABLE sessions (
  ...
  shift_number  INTEGER NOT NULL,  -- assigned at check-in
  shift_config_version  VARCHAR(50),  -- which config version was used
  ...
);
```

### Transaction Table (shift_number field)

```sql
CREATE TABLE transactions (
  ...
  shift_number  INTEGER NOT NULL,  -- assigned at check-out
  ...
);
```

---

## 17.8 Design Decisions

**Why continuous shift numbers (not reset daily)?**
- Unique identifier across time (no ambiguity).
- Easier to query and aggregate.
- Matches AMB's current system.
- Prevents confusion when reporting across multiple days.

**Why automatic assignment (not manual)?**
- No operators in v2 (fully automated).
- Consistent and reliable (no human error).
- Works offline (server room app has local config).
- Reduces complexity (no shift start/end flows).

**Why 3 shifts/day (default)?**
- Matches AMB's current operational model.
- Covers 24 hours (morning, afternoon, evening).
- Configurable per location (can add more shifts if needed).

**Why store shift_number in session and transaction?**
- Audit trail: know which shift session was created in.
- Reporting: aggregate by shift for reconciliation.
- Session and transaction may have different shifts (overnight parking).

**Why shift_config_version tracking?**
- Audit trail: know which shift config was used.
- If configs change, old sessions/transactions still reference old configs.
- Prevents retroactive changes.
- Supports offline operation (use last known config).

**Why server room app assigns shifts (not cloud)?**
- Offline resilience: gates work without internet.
- Lower latency: no round-trip to cloud.
- Local shift config polled every 1 min (fresh enough).

---

## 17.9 Differences from v1

| Aspect | v1 (Manual) | v2 (Automated) |
|--------|-------------|----------------|
| **Shift start** | Operator manually starts shift | Automatic (time-based) |
| **Shift end** | Operator manually ends shift | Automatic (time-based) |
| **Shift number** | Per operator, per day | Continuous across all operators/days |
| **Shift config** | Fixed per location | Configurable, versioned |
| **Cash handover** | Operator records cash handover | N/A (no cash, no operators) |
| **Shift summary** | Per operator | Per shift (time window) |
| **Overnight handling** | Operator ends shift before midnight | Automatic (overnight shift config) |
| **Reporting** | By operator + shift | By shift (time window) |

---

*End of Chapter 17 — Shift Management (v2)*
