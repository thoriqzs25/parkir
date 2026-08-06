# Chapter 4 — Locations

## 4.1 Overview

A **Location** represents a single physical parking facility. In v2, each location has:
- 2-10 automated gates (entry/exit).
- 1 server room app (mini PC in server room).
- 2-3 server room staff + 1 leader.

The system manages 20+ locations for AMB (first tenant), grouped by city. All locations are visible and manageable from the central web dashboard.

---

## 4.2 Location Attributes

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | System-generated unique identifier |
| `name` | String | Human-readable name (e.g., "AMB Parking Central") |
| `code` | String | Short unique identifier (e.g., `AMB-CENTRAL`) |
| `address` | String | Full street address |
| `city` | String | City (used for grouping in dashboard) |
| `status` | Enum | `ACTIVE` or `INACTIVE` |
| `created_at` | Timestamp | When the location was added |
| `updated_at` | Timestamp | Last modified |

---

## 4.3 Location Components

Each location consists of:

### Gates
- 2-10 automated gates per location.
- Each gate is either ENTRY or EXIT.
- Each gate configured with vehicle type (CAR, MOTO, TRUCK, or ALL).
- Gate hardware: barrier, thermal printer (entry), QR scanner + payment terminal (exit), sensors.
- Gate app runs on mini PC at each gate.

### Server Room App
- 1 server room app per location.
- Runs on mini PC in server room.
- Local SQLite database (configs, sessions, transactions).
- Manages all gates at location.
- Syncs to cloud backend every 1 minute.

### Staff
- 2-3 server room staff per location.
- 1 leader per location.
- Monitor gates, handle exceptions, run offline SOP.

---

## 4.4 Location Hierarchy

```
AMB (Tenant)
├── Jakarta (City)
│   ├── Location A (2 entry, 2 exit gates)
│   ├── Location B (1 entry, 1 exit gate)
│   └── Location C (3 entry, 2 exit gates)
├── Surabaya (City)
│   ├── Location D (2 entry, 2 exit gates)
│   └── Location E (1 entry, 1 exit gate)
└── Bandung (City)
    └── Location F (2 entry, 2 exit gates)
```

**Dashboard Filtering:**
- Locations grouped by city.
- Filter by city (dropdown).
- View all locations (owner/admin only).

---

## 4.5 Location Lifecycle

### Creation
1. Admin creates location via dashboard (name, code, address, city).
2. System creates default shift config (3 shifts/day).
3. System creates default rate config (per vehicle type).
4. Location status: `ACTIVE`.

### Configuration
1. Admin configures rates (per vehicle type, versioned).
2. Admin configures shifts (time windows, versioned).
3. Admin installs server room app (USB).
4. Admin installs gate apps (USB).
5. Gate apps discover via mDNS, register as `UNREGISTERED`.
6. Admin configures gates via dashboard (vehicle type, gate type).
7. Gates become `OPERATIONAL`.

### Deactivation
1. Admin deactivates location via dashboard.
2. Location status: `INACTIVE`.
3. Gates stop accepting new sessions.
4. Existing sessions can still close.
5. Location hidden from dashboard (unless filter: all).

---

## 4.6 Location Management (Dashboard)

### Location List View

| Column | Description |
|--------|-------------|
| Name | Location name |
| Code | Short code |
| City | City (grouped) |
| Status | ACTIVE / INACTIVE |
| Gates | Count of gates (online/total) |
| Revenue Today | Total revenue today |
| Actions | Edit, Deactivate |

### Location Detail View

- Location info (name, code, address, city).
- Gate list (status, vehicle type, last seen).
- Rate configs (versioned, editable).
- Shift configs (versioned, editable).
- Reports (revenue, occupancy, transactions).
- Audit log (location-specific).

---

## 4.7 Design Decisions

**Why no capacity tracking?**
- Automated system doesn't need to track slots.
- No operators to manage capacity.
- Can be added later if needed (sensors).

**Why group by city?**
- AMB has 20+ locations across multiple cities.
- Easier to manage and report by region.
- Matches AMB's organizational structure.

**Why 1 server room app per location?**
- Single source of truth for location.
- Manages all gates at location.
- Simplifies sync and monitoring.

**Why 2-10 gates per location?**
- Flexible (small to large locations).
- Each gate independent (stateless).
- Server room app manages all gates.

---

*End of Chapter 4 — Locations (v2)*
