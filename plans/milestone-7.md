# Milestone 7 — Reports & Polish

## 1. Goal

Deliver four data-driven reports (daily revenue, occupancy over time, vehicle-type breakdown, operator activity) with CSV/PDF export and charts, plus a UI/UX polish pass across the dashboard.

## 2. Scope

### In Scope

- Daily revenue summary report: cards (total revenue, transaction count, average fee) + bar chart + table with filters (location, date range up to 90 days)
- Occupancy over time report: line chart showing active sessions by hour/day + heatmap view
- Per-vehicle-type breakdown report: table + pie/bar chart split by CAR/MOTO/TRUCK
- Operator activity report: sessions handled, revenue collected, shift hours per operator (managers, admins, owners can view)
- CSV export for all four reports
- PDF export for all four reports (browser print-to-PDF via `window.print()` or a server-side PDF library)
- Voided transactions shown in revenue reports with a toggleable flag column
- UI/UX consistency pass: uniform spacing, button styles, loading states, error boundaries
- YYYY-MM-DD date format throughout reports

### Out of Scope / Deferred

- Scheduled/digest report emails (deferred to post-v1)
- Cached/pre-aggregated reports (always query live)
- Report date ranges beyond 90 days
- Drill-down from report charts to individual transactions
- Real-time dashboard (reports are refresh-on-load)
- Role-based row-level filtering on reports (all visible data for the location is shown)
- Multi-location aggregate reports (reports are per-location only)
- Report comparison (week-over-week, month-over-month)

## 3. Dependencies

- **Milestone 2 — Backend Business Logic** provides the sessions, transactions, and shift data that reports aggregate
- **Milestone 3 — Web Dashboard Foundation** provides layout, navigation, table components, permission helpers
- **Milestone 6 — Incidents, Adjustments & Observability** is complete (audit/report permissions rely on `finance:*` and `reports:*` permission model)
- Charting library decision (Recharts or Chart.js) must be made before implementation
- `reports:view_revenue`, `reports:view_occupancy`, `reports:view_operators`, `finance:export` permissions already defined

## 4. Detailed Tasks

### Backend

- [ ] Create `internal/domain/reports/` with handler, routes, store:
  - `GET /api/v1/reports/daily-revenue` — query params: `location_id`, `date_from`, `date_to` (max 90 days), `include_voided` (bool)
  - `GET /api/v1/reports/occupancy` — query params: `location_id`, `date_from`, `date_to`, `granularity` (hour/day)
  - `GET /api/v1/reports/vehicle-breakdown` — query params: `location_id`, `date_from`, `date_to`
  - `GET /api/v1/reports/operator-activity` — query params: `location_id`, `date_from`, `date_to`, `operator_id` (optional)
- [ ] Daily revenue store query:
  - Aggregate `SUM(fee_amount)`, `COUNT(*)`, `AVG(fee_amount)` grouped by day
  - If `include_voided=true`, include voided transactions in a separate column
- [ ] Occupancy store query:
  - Count active sessions per time bucket using `check_in_at` / `check_out_at`
  - Support hourly and daily granularity
- [ ] Vehicle breakdown store query:
  - Group transactions by `vehicle_type`, sum fee_amount and count
- [ ] Operator activity store query:
  - Group by operator_id, sum transactions, count sessions, sum shift hours
  - Join with users table for operator names
- [ ] CSV export endpoint per report (or a generic `?format=csv` query param)
- [ ] PDF export endpoint per report (generate HTML → convert to PDF using a Go PDF library, or return HTML for browser print)
- [ ] Add `reports:*` permission checks to all report endpoints
- [ ] Write integration tests for each report query

### Dashboard

- [ ] Add **Reports** nav item to sidebar
- [ ] Create `app/(dashboard)/[locationId]/reports/` page group:
  - `reports/page.tsx` — reports hub with cards linking to each report
  - `reports/daily-revenue/page.tsx` — date range picker (max 90 days), toggle for voided transactions, bar chart + summary cards + table + CSV/PDF export buttons
  - `reports/occupancy/page.tsx` — date range picker, granularity toggle (hour/day), line chart + heatmap + CSV/PDF export
  - `reports/vehicle-breakdown/page.tsx` — date range picker, pie/bar chart + table + CSV/PDF export
  - `reports/operator-activity/page.tsx` — date range picker, optional operator filter, table + bar chart + CSV/PDF export
- [ ] Install and configure a charting library (Recharts recommended for React)
- [ ] Build reusable chart components: `BarChart`, `LineChart`, `PieChart`, `Heatmap`
- [ ] Build reusable `DateRangePicker` component (max 90-day spread, YYYY-MM-DD format)
- [ ] Build reusable `ReportExport` component with CSV and PDF download buttons
- [ ] UI/UX consistency pass:
  - Add loading skeletons to all list pages
  - Add error boundaries to all page groups
  - Fix any spacing, button sizing, or color inconsistencies
  - Ensure all action buttons show loading state
  - Add toast confirmation on all mutation actions
- [ ] TypeScript type-check passes for all new pages

### Desktop

- [ ] Minimal: Add a read-only reports section or a link to the dashboard reports URL (browser-based)
- [ ] No new desktop screens for reports (deferred to post-v1 if needed)

### DevOps / QA

- [ ] Integration test for each report query with known seed data
- [ ] End-to-end smoke test: set date range → view report → export CSV → verify content
- [ ] PDF output verification for at least one report
- [ ] Verify 90-day max enforcement on backend
- [ ] Verify voided toggle works correctly in revenue report
- [ ] Verify operator activity only accessible to managers/admins/owners

## 5. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Maximum date range | 90 days | User preference; prevents unbounded queries |
| Report caching | None (live queries) | User preference; data is always fresh |
| Charting library | Recharts | React-native, well-maintained, supports bar/line/pie/heatmap |
| PDF export | Browser `window.print()` with print CSS | Simplest approach; no server-side PDF dependency in v1 |
| CSV export | Server-side generation (`?format=csv`) | Consistent with audit log CSV export pattern |
| Voided transactions in revenue | Toggleable column | User specified "add flag" — show as separate column with show/hide toggle |
| Report granularity | Hourly and daily for occupancy | User can choose level of detail |
| Date format | YYYY-MM-DD | User preference; consistent across all reports |

## 6. Open Questions / Risks

| Question / Risk | Owner | Due Date |
|-----------------|-------|----------|
| Should occupancy heatmap show X=hour, Y=day or vice versa? | Resolved: X=hour (0–23), Y=date | — |
| What is the exact definition of "occupancy" — active sessions at a point in time, or total check-ins per bucket? | Resolved: total check-ins per time bucket | — |
| Browser print-to-PDF may produce inconsistent results across browsers — acceptable? | Frontend | Before PDF feature sign-off |
| Chart responsiveness on smaller screens — should reports be scrollable or responsive? | Frontend | Before chart implementation |
| Operator activity: should it include only closed sessions or all sessions? | Product | Before operator activity query |
| Large report datasets (>90 days of data for busy locations) — ~260k sessions max with aggregation; aggregated reports produce <100 rows, occupancy hourly ~2,160 rows — no special optimization needed | Resolved: acceptable with default limit param | — |

## 7. Acceptance Criteria

- [ ] Daily revenue report shows total revenue, transaction count, average fee per day with bar chart
- [ ] Daily revenue report includes a toggle to show/hide voided transactions in a separate column
- [ ] Occupancy report shows active sessions per hour/day with line chart and heatmap
- [ ] Vehicle breakdown report shows count and revenue per vehicle type with chart and table
- [ ] Operator activity report shows sessions handled, revenue collected, and hours worked per operator
- [ ] All reports respect the 90-day maximum date range (backend enforces)
- [ ] CSV export downloads a valid CSV file for each report
- [ ] PDF export (browser print) produces a readable print layout for each report
- [ ] Reports pages show loading skeletons while data loads
- [ ] Error boundaries catch and display errors gracefully on all report pages
- [ ] Only users with `reports:view_revenue` can access revenue reports
- [ ] Only users with `reports:view_occupancy` can access occupancy reports
- [ ] Only users with `reports:view_operators` can access operator activity reports
- [ ] Backend integration tests pass for all four report queries
- [ ] Dashboard builds and type-checks without errors

## 8. Definition of Done

- Backend: all four report endpoints with CSV export and 90-day enforcement are implemented and integration-tested.
- Dashboard: all four report pages with charts, filters, export buttons, loading states, and error boundaries are implemented and type-check.
- UI/UX consistency pass is complete across the dashboard.
- All code is reviewed and merged to `main`.
- `PLAN.md` is updated if any decisions diverged from this plan.