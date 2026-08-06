# Chapter 10 — Reports & Analytics

## 10.1 Overview

The reporting module provides AMB admins and managers with data-driven insight into revenue performance, occupancy trends, vehicle composition, and gate health. All reports are available through the web dashboard.

Reports are grouped by city (for AMB's 20+ locations).

### Common Filter Bar (All Reports)

| Filter | Options |
|--------|---------|
| City | Single city or all cities |
| Location | Single location or all accessible locations |
| Date Range | Custom range picker; presets: Today, Yesterday, This Week, This Month, Last Month |
| Vehicle Type | All, CAR, MOTO, TRUCK |

Reports only show data for locations the viewing user has permission for.

---

## 10.2 Report 1 — Daily Revenue Summary

### Purpose
Track revenue collected per day, understand payment method mix, and compare against prior periods.

### Data Source
`transactions` table where `voided = false`, grouped by `check_out_at::date`.

### Display Components

#### Summary Cards (top of page)
| Metric | Description |
|--------|-------------|
| Total Revenue | Sum of `amount` for selected period |
| E-Money Revenue | Sum where `payment_method = EMONEY` |
| Flazz Revenue | Sum where `payment_method = FLAZZ` |
| Total Sessions | Count of closed non-voided sessions |
| Average Fee | Total Revenue / Total Sessions |

#### Daily Revenue Bar Chart
- X-axis: dates in selected range
- Y-axis: revenue (currency)
- Bars stacked by payment method (EMONEY / FLAZZ)

#### Day-by-Day Table
| Column | Description |
|--------|-------------|
| Date | |
| Total Sessions | |
| Total Revenue | |
| E-Money | |
| Flazz | |
| Average Fee | |

### Export
- CSV export with all columns.
- Includes filters applied.

---

## 10.3 Report 2 — Revenue by Vehicle Type

### Purpose
Understand revenue distribution across vehicle types.

### Data Source
`transactions` joined with `sessions` (for `vehicle_type`), grouped by `vehicle_type`.

### Display Components

#### Pie Chart
- Segments: CAR, MOTO, TRUCK.
- Size: revenue percentage.

#### Vehicle Type Table
| Column | Description |
|--------|-------------|
| Vehicle Type | CAR / MOTO / TRUCK |
| Total Sessions | Count |
| Total Revenue | Sum of `amount` |
| Average Fee | Revenue / Sessions |
| Percentage | Revenue % of total |

### Export
- CSV export.

---

## 10.4 Report 3 — Occupancy Over Time

### Purpose
Understand parking utilization patterns by time of day and day of week.

### Data Source
`transactions` table, grouped by hour-of-day and day-of-week.

### Display Components

#### Heatmap
- X-axis: hour of day (0-23).
- Y-axis: day of week (Mon-Sun).
- Color intensity: average active sessions.

#### Hourly Line Chart
- X-axis: hour of day.
- Y-axis: average active sessions.
- Line per day of week.

### Export
- CSV export (hourly averages).

---

## 10.5 Report 4 — Transaction Report

### Purpose
Detailed transaction list with filtering and search.

### Data Source
`transactions` table with joins to `sessions` and `locations`.

### Display Components

#### Transaction Table
| Column | Description |
|--------|-------------|
| Date | Transaction date |
| Time | Transaction time |
| Location | Location name |
| Gate | Gate ID |
| Vehicle Type | CAR / MOTO / TRUCK |
| Amount | Transaction amount |
| Payment Method | EMONEY / FLAZZ |
| Shift | Shift number |
| Status | OK / VOIDED |

#### Filters
- Date range.
- Location.
- Vehicle type.
- Payment method.
- Status (OK / VOIDED).

#### Search
- Transaction ID.
- Session ID.

### Export
- CSV export with all columns and filters.

---

## 10.6 Report 5 — Gate Health Report

### Purpose
Monitor gate uptime, hardware failures, and maintenance needs.

### Data Source
Gate health checks (from server room app), incidents, audit logs.

### Display Components

#### Gate Uptime Summary
| Column | Description |
|--------|-------------|
| Location | Location name |
| Gate | Gate ID |
| Uptime % | Uptime percentage (last 30 days) |
| Downtime | Total downtime (hours) |
| Incidents | Number of incidents |

#### Gate Status Timeline
- Gantt chart showing gate status over time.
- Colors: ONLINE (green), OFFLINE (red), DEGRADED (yellow).

#### Incident Breakdown
- Pie chart: incident types (printer jam, scanner error, gate motor failure, etc.).

### Export
- CSV export.

---

## 10.7 Report 6 — Shift Summary

### Purpose
Revenue and session counts by shift (for reconciliation).

### Data Source
`transactions` table, grouped by `shift_number`.

### Display Components

#### Shift Table
| Column | Description |
|--------|-------------|
| Shift Number | Continuous shift number |
| Shift Code | Time window (e.g., "06-14") |
| Date | Calendar date |
| Total Sessions | Count |
| Total Revenue | Sum of `amount` |
| Average Fee | Revenue / Sessions |

### Export
- CSV export.

---

## 10.8 Report 7 — Location Comparison

### Purpose
Compare performance across locations (for AMB admins).

### Data Source
`transactions` table, grouped by `location_id`.

### Display Components

#### Location Table
| Column | Description |
|--------|-------------|
| Location | Location name |
| City | City |
| Total Sessions | Count |
| Total Revenue | Sum of `amount` |
| Average Fee | Revenue / Sessions |
| Gate Uptime | Average gate uptime % |

#### Bar Chart
- X-axis: locations.
- Y-axis: revenue.
- Grouped by city.

### Export
- CSV export.

---

## 10.9 Dashboard Home Widgets

### Revenue Today
- Total revenue today (all locations or filtered).
- Compare to yesterday (% change).
- Auto-refresh every 1 minute.

### Active Sessions
- Total active sessions (all locations or filtered).
- Breakdown by vehicle type.
- Auto-refresh every 30 seconds.

### Gate Status
- Total gates: online / offline.
- Alert count (active alerts).
- Auto-refresh every 30 seconds.

---

## 10.10 Design Decisions

**Why group by city?**
- AMB has 20+ locations across multiple cities.
- Easier to manage and report by region.
- Matches AMB's organizational structure.

**Why no operator activity reports?**
- No operators (fully automated).
- Gate activity reports instead.

**Why gate health reports?**
- Critical for automated system.
- Identify maintenance needs.
- Track uptime (SLA compliance).

**Why shift summaries?**
- Reconciliation (shift-based reporting).
- Matches AMB's current system.
- Audit trail.

---

*End of Chapter 10 — Reports & Analytics (v2)*
