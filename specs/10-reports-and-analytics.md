# Chapter 10 — Reports & Analytics

## 10.1 Overview

The reporting module provides managers and admins with data-driven insight into revenue performance, occupancy trends, vehicle composition, and operator behavior. All reports are available through the web dashboard.

### Common Filter Bar (All Reports)

| Filter | Options |
|--------|---------|
| Location | Single location or all accessible locations |
| Date Range | Custom range picker; presets: Today, Yesterday, This Week, This Month, Last Month |
| Vehicle Type | All, CAR, MOTO, TRUCK |

Reports only show data for locations the viewing user has `reports:*` permission for.

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
| Total Revenue | Sum of `fee_amount` for selected period |
| Cash Revenue | Sum where `payment_method = CASH` |
| Digital Revenue | Sum where `payment_method = DIGITAL` |
| Total Sessions | Count of closed non-voided sessions |
| Average Fee | Total Revenue / Total Sessions |

#### Daily Revenue Bar Chart
- X-axis: dates in selected range
- Y-axis: revenue (currency)
- Bars stacked by payment method (cash / digital)

#### Day-by-Day Table
| Column | Description |
|--------|-------------|
| Date | |
| Total Sessions | |
| Total Revenue | |
| Cash | |
| Digital | |
| vs. Previous Period | % change vs same date range in prior period |

#### Previous Period Comparison
- If date range = "This Month", comparison = "Last Month"
- If custom range, comparison = same duration shifted back by range length
- Displayed as: `▲ +12.4%` or `▼ -3.1%`

### Export
- CSV: day-by-day table
- PDF: full report with charts (optional, v2)

---

## 10.3 Report 2 — Occupancy Over Time

### Purpose
Understand parking demand patterns — identify peak hours, low periods, and trends across days.

### Data Source
Derived from `sessions`: for each hour, count sessions where `check_in_at <= hour_start AND (check_out_at IS NULL OR check_out_at >= hour_start)`.

### Display Components

#### Occupancy Heatmap (primary view)
- Rows: hours of the day (00:00 – 23:00)
- Columns: dates in selected range
- Cell value: occupancy count or % of capacity
- Color scale: light (low) → dark (high)

#### Occupancy Line Chart (secondary view)
- X-axis: time
- Y-axis: occupancy count or %
- One line per vehicle type (optional toggle)
- Toggleable: absolute count vs percentage of capacity

#### Peak Hours Summary Table
| Hour | Avg Occupancy | Peak Day | Peak Count |
|------|--------------|----------|-----------|
| 08:00 | 72% | Mon | 89% |
| ... | | | |

### Notes
- If capacity is not configured for a vehicle type, occupancy % is shown as `N/A` and raw count is shown instead.
- For "All Locations" view, occupancy is aggregated (sum of counts; % uses sum of capacities).

---

## 10.4 Report 3 — Per-Vehicle-Type Breakdown

### Purpose
Compare the contribution of each vehicle type to total sessions and revenue.

### Data Source
`transactions` grouped by `vehicle_type`, where `voided = false`.

### Display Components

#### Summary Cards
| Metric | Per Vehicle Type |
|--------|----------------|
| Session Count | |
| Revenue | |
| Revenue Share (%) | |
| Average Fee | |
| Average Duration | |

#### Stacked Bar Chart
- X-axis: dates in selected range
- Y-axis: session count or revenue
- Bars stacked by vehicle type (CAR / MOTO / TRUCK)
- Toggle between "Sessions" and "Revenue" view

#### Breakdown Table
| Vehicle Type | Sessions | Revenue | Avg Fee | Avg Duration | Revenue % |
|-------------|----------|---------|---------|-------------|-----------|
| CAR | | | | | |
| MOTO | | | | | |
| TRUCK | | | | | |
| **Total** | | | | | |

#### Trend Chart (secondary)
- Line chart showing session count per vehicle type over the selected date range.
- Useful for spotting shifts in vehicle mix over time.

---

## 10.5 Report 4 — Operator Activity Log

### Purpose
Track what each operator did during their work period — for performance review, accountability, and anomaly investigation.

### Additional Filters (beyond common bar)
| Filter | Options |
|--------|---------|
| Operator | Dropdown of all operators at selected location |
| Action Type | Check-in, Check-out, Payment, Incident Filed, Adjustment |

### Display Components

#### Operator Summary Table
| Operator | Check-ins | Check-outs | Cash Collected | Digital Collected | Incidents Filed | Voids / Adjustments |
|----------|-----------|-----------|---------------|-----------------|----------------|-------------------|

Click a row to drill into that operator's detailed log.

#### Detailed Activity Timeline
Chronological list of all actions by the selected operator in the period:

| Timestamp | Action | Details |
|-----------|--------|---------|
| 2025-03-15 08:02 | CHECK_IN | B 1234 XYZ (CAR) |
| 2025-03-15 09:14 | CHECK_OUT | B 1234 XYZ — Rp 15,000 (CASH) |
| 2025-03-15 09:45 | INCIDENT_FILED | PAYMENT_DISPUTE — Session #abc123 |
| 2025-03-15 11:30 | CHECK_IN | D 5678 ABC (TRUCK) |

#### Performance Metrics Cards (per operator)
| Metric | Value |
|--------|-------|
| Total Check-ins | |
| Total Check-outs | |
| Total Revenue Collected | |
| Avg Sessions per Hour | |
| Incidents Filed | |
| Voids Involved In | |

---

## 10.6 Data Freshness & Caching

| Report | Refresh Strategy |
|--------|----------------|
| Daily Revenue Summary | Near real-time (< 1 min lag acceptable) |
| Occupancy Over Time | Historical data cached; current hour live |
| Vehicle Type Breakdown | Near real-time |
| Operator Activity Log | Real-time (direct DB query) |

- Reports for the current day query live data.
- Reports for completed past days may use pre-aggregated snapshots for performance.
- All report queries must complete within 3 seconds for a 30-day date range.
