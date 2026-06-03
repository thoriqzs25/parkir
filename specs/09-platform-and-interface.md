# Chapter 9 — Platform & Interface

## 9.1 Overview

The system has two distinct interfaces serving different users and use cases:

| Interface | Target Users | Platform | Primary Use |
|-----------|-------------|----------|-------------|
| **Operator Desktop App** | Parking operators / attendants | Windows desktop | Check-in, check-out, payment, incidents |
| **Web Dashboard** | Facility managers, admins | Browser (desktop) | Reports, management, configuration, monitoring |

Both interfaces communicate with a shared backend API. The desktop app also has offline capability with local data storage.

---

## 9.2 Operator Desktop App

### 9.2.1 Platform

- **Target OS:** Windows 10 / 11 (primary).
- **Tech:** **Electron** — web-based desktop app. UI is built with the same web tech stack as the web dashboard (React), packaged as a desktop app via Electron.
- **Benefits:** Shared component library with web dashboard, easy auto-update delivery via Electron updater, no separate native codebase.
- **Screen resolution target:** 1024×768 minimum (accounting for older booth hardware).
- **Input:** Keyboard and mouse; touch is optional.
- **Auto-update:** App checks for updates on launch and after reconnect; updates are applied on next restart.

### 9.2.2 Core Screens

#### Login Screen
- Email + password fields.
- On login: user selects their active location (if assigned to multiple).
- Displays system connectivity status (online / offline).

#### Home / Dashboard
- Current location name.
- Live count of active sessions by vehicle type.
- Quick-action buttons: **Check-in**, **Check-out**, **Search Session**, **Report Incident**.
- Connectivity indicator (online / offline mode badge).

#### Check-in Form
- Fields: plate number (text, auto-uppercase), vehicle type (dropdown).
- Submit button: **Check In**.
- Duplicate plate warning (modal with existing session details).
- Confirmation screen after successful check-in (session ID, timestamp).

#### Check-out / Payment Screen
- Search by plate number or session ID.
- Session details displayed: plate, vehicle type, check-in time, elapsed time, calculated fee.
- Payment method selection: **Cash** / **Digital**.
- Cash: amount tendered input → change displayed.
- Digital: gateway/method selector → reference input → confirm.
- **Confirm Payment** button → receipt auto-prints → success screen.

#### Session Search
- Search bar (plate or session ID).
- Filter by state (Active / All).
- Results list: plate, vehicle type, check-in time, state, fee.
- Tap/click a session to view details.

#### Incident Report Form
- Incident type dropdown: STUCK_AT_GATE / PAYMENT_DISPUTE / OPERATOR_ERROR / SYSTEM_DOWNTIME.
- Optional session linkage: search and attach a session.
- Description text area (required).
- Submit button.

#### Settings
- Printer configuration (connection type, port, paper width).
- Active location (change without re-login).
- App version and sync status.

### 9.2.3 Offline Mode UI
- Offline badge visible at all times when not connected.
- All core flows (check-in, check-out, payment, receipt) remain available.
- Sync status indicator shows count of unsynced records.
- On reconnect: auto-sync begins; operator notified of completion or conflicts.

---

## 9.3 Web Dashboard

### 9.3.1 Platform

- **Browser support:** Chrome, Firefox, Edge (latest 2 major versions).
- **Recommended tech:** React SPA or Next.js.
- **Responsive:** Desktop-first; tablet usable; mobile not required.
- **Authentication:** Email + password; JWT token; 8-hour session with refresh.

### 9.3.2 Navigation Structure

```
Web Dashboard
├── Overview (home)
├── Sessions
│   ├── Active Sessions
│   └── Session History
├── Reports
│   ├── Daily Revenue
│   ├── Occupancy Over Time
│   ├── Vehicle Type Breakdown
│   └── Operator Activity
├── Incidents
│   ├── Open Incidents
│   └── Incident History
├── Adjustments
│   ├── Void Transaction
│   └── Reassign Session
├── Administration
│   ├── Locations
│   ├── Rates
│   ├── Users
│   └── Roles & Permissions
└── Observability
    ├── System Health
    ├── Audit Log
    └── Alerts
```

### 9.3.3 Core Pages

#### Overview (Home)
- Location selector (single location or all).
- Live occupancy cards per vehicle type (current count / capacity).
- Today's revenue summary (total, cash, digital).
- Open incident count with quick-link.
- System health status widget.
- Recent activity feed (last 10 sessions).

#### Active Sessions
- Real-time list of all `ACTIVE` sessions for selected location(s).
- Columns: plate, vehicle type, check-in time, elapsed duration, operator.
- Sortable and filterable.
- Click to view session detail.

#### Session History
- All sessions with full filter support: date range, state, vehicle type, operator, location.
- Export to CSV.

#### Reports
- Each report has a standard filter bar: **Location**, **Date Range**, **Vehicle Type**.
- Charts and tables; exportable to CSV or PDF.
- See Chapter 10 for report specs.

#### Incidents
- Open incidents listed with priority (time since filing).
- Manager can view details, add notes, and mark as resolved.
- History tab shows all closed incidents.

#### Adjustments
- **Void Transaction:** search by session/receipt ID → confirm void with reason → manager PIN authorization.
- **Reassign Session:** search session → select target operator → confirm with reason → manager PIN.

#### Administration
- Locations: create/edit locations, configure rates per vehicle type.
- Users: create, edit, deactivate users; assign roles and locations.
- Roles: create/edit roles; assign permissions per module.

#### Observability
- System health dashboard with component status.
- Audit log with full filter and search.
- Alerts list with threshold configuration.

---

## 9.4 Shared UX Principles

- **Confirmations for destructive actions** — Void, deactivate, and delete operations always require a confirmation step with reason input.
- **Optimistic UI** — Check-in and check-out flows show instant feedback; errors are surfaced clearly.
- **Role-aware UI** — Menu items and actions not permitted by the user's role are hidden (not just disabled).
- **Timezone handling** — All timestamps stored in UTC; displayed in the configured local timezone per location.
- **Accessibility** — Keyboard navigability for all primary operator flows; sufficient color contrast (WCAG AA).
