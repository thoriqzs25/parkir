# PARKIR v2 — Frontend Timeline & Implementation
**Created:** 2026-08-06

---

## Frontend Components

### 1. AMB Admin Dashboard (Next.js 14)
- Web-based dashboard for AMB admins
- Multi-location management
- Reporting & analytics
- Gate configuration
- User management

### 2. Gate App UI (Electron + React or Tauri + React)
- Driver-facing monitor (exit gate)
- Displays fee, instructions, payment status
- Receipt printing UI
- Error states

### 3. Server Room App UI (Electron + React or Tauri + React)
- Staff monitoring dashboard
- Gate status overview
- Alert notifications (visual + audio)
- Manual refresh controls

---

## Week 1-2: Dashboard Foundation

### Project Setup (2 days)
**Deliverables:**
- [ ] Next.js 14 app (App Router, TypeScript, Tailwind CSS)
- [ ] Project structure (app/, components/, lib/, hooks/, types/)
- [ ] ESLint + Prettier configuration
- [ ] Environment variables setup (.env.local, .env.production)
- [ ] API client setup (fetch wrapper, interceptors, error handling)
- [ ] Authentication context (user state, permissions)

**Project Structure:**
```
dashboard/
├── src/
│   ├── app/
│   │   ├── (auth)/
│   │   │   ├── login/
│   │   │   └── layout.tsx
│   │   ├── (dashboard)/
│   │   │   ├── [locationId]/
│   │   │   │   ├── page.tsx (redirect to /gates)
│   │   │   │   ├── gates/
│   │   │   │   ├── sessions/
│   │   │   │   ├── transactions/
│   │   │   │   ├── reports/
│   │   │   │   └── settings/
│   │   │   ├── locations/
│   │   │   ├── users/
│   │   │   ├── layout.tsx
│   │   │   └── page.tsx (redirect to first location)
│   │   ├── layout.tsx (root)
│   │   └── providers.tsx
│   ├── components/
│   │   ├── ui/ (button, input, modal, table, etc.)
│   │   ├── layout/ (sidebar, header, location-selector)
│   │   ├── gates/ (gate-card, gate-config-modal)
│   │   ├── sessions/ (session-list, session-detail)
│   │   ├── transactions/ (transaction-list, transaction-detail)
│   │   └── reports/ (revenue-chart, occupancy-chart)
│   ├── lib/
│   │   ├── api.ts (API client)
│   │   ├── auth.ts (auth helpers)
│   │   └── utils.ts
│   ├── hooks/
│   │   ├── useAuth.ts
│   │   ├── useLocation.ts
│   │   └── useApi.ts
│   └── types/
│       ├── api.ts
│       ├── gate.ts
│       ├── session.ts
│       └── transaction.ts
├── public/
├── tailwind.config.ts
├── tsconfig.json
└── package.json
```

### Design System (3 days)
**Deliverables:**
- [ ] UI component library (shadcn/ui or custom):
  - Button (primary, secondary, danger, ghost)
  - Input (text, number, select, textarea)
  - Modal (confirm, form, detail)
  - Table (sortable, filterable, paginated)
  - Card (gate card, session card, transaction card)
  - Badge (status: online, offline, busy, unregistered)
  - Alert (info, warning, error, success)
  - Tabs (for reports, settings)
  - Dropdown menu
  - Tooltip
  - Toast notification
- [ ] Design tokens (colors, typography, spacing)
- [ ] Dark mode support (optional)
- [ ] Responsive breakpoints (mobile, tablet, desktop)

### Authentication (2 days)
**Deliverables:**
- [ ] Login page (email + password form)
- [ ] Auth context (user state, permissions, login/logout functions)
- [ ] Protected routes (redirect to /login if not authenticated)
- [ ] JWT cookie handling (httpOnly, secure, sameSite)
- [ ] Auto-refresh token (refresh before expiry)
- [ ] Logout functionality (clear cookie, redirect to /login)

### Layout (3 days)
**Deliverables:**
- [ ] Root layout (html, body, fonts)
- [ ] Auth layout (centered card for login)
- [ ] Dashboard layout:
  - Sidebar (navigation links, collapsible)
  - Header (user menu, location selector, notifications)
  - Main content area
- [ ] Location selector component:
  - Dropdown grouped by city
  - Switch location → update URL (/[locationId]/...)
  - Persist selected location in context
- [ ] Responsive sidebar (hamburger menu on mobile)

---

## Week 3-4: Dashboard - Live Monitoring

### Gate Status Page (3 days)
**Deliverables:**
- [ ] Gate list view (grid of gate cards)
- [ ] Gate card component:
  - Gate name + ID
  - Status badge (online/offline/busy/unregistered)
  - Gate type (entry/exit)
  - Vehicle type
  - Last seen timestamp
  - Health indicator (green/yellow/red)
- [ ] Unregistered gates section:
  - "New gate detected" banner
  - Configure button → opens modal
- [ ] Auto-refresh (every 30s, fetch gate statuses)
- [ ] Filter by status (online, offline, all)
- [ ] Sort by name, status, last seen

### Gate Configuration Modal (2 days)
**Deliverables:**
- [ ] Modal component (overlay, close on escape/backdrop click)
- [ ] Form fields:
  - Gate type (radio: entry/exit)
  - Vehicle type (select: car, motorcycle, truck, all)
  - Location (read-only, from URL)
- [ ] Validation (required fields)
- [ ] Submit → PATCH /api/v1/gates/:id
- [ ] Success → close modal, refresh gate list
- [ ] Error → show error message

### Active Sessions Page (2 days)
**Deliverables:**
- [ ] Session list (table view)
- [ ] Columns: session_id, vehicle_type, check_in_time, duration, gate, status
- [ ] Auto-refresh (every 30s)
- [ ] Filter by vehicle_type, gate
- [ ] Sort by check_in_time (newest first)
- [ ] Click row → session detail page

### Revenue Today Widget (1 day)
**Deliverables:**
- [ ] Dashboard home page (overview)
- [ ] Revenue today card (total amount, transaction count)
- [ ] Compare to yesterday (percentage change)
- [ ] Auto-refresh (every 1 min)

### Alert Counts Widget (1 day)
**Deliverables:**
- [ ] Alert counts card (active alerts, resolved today)
- [ ] Color-coded by severity (critical, warning, info)
- [ ] Click → alert history page

---

## Week 5-6: Dashboard - Configuration Management

### Rate Management Page (3 days)
**Deliverables:**
- [ ] Rate list (table view)
- [ ] Columns: vehicle_type, effective_date, first_hour_rate, subsequent_hourly_rate, daily_flat_rate, version
- [ ] Filter by vehicle_type
- [ ] Create rate button → form modal:
  - Vehicle type (select)
  - Effective date (date picker)
  - First hour rate (number input)
  - Subsequent hourly rate (number input)
  - Daily flat rate (number input)
  - Validation (all required, rates > 0)
  - Submit → POST /api/v1/rates
- [ ] Edit button → edit modal (same form, pre-filled)
  - Submit → PATCH /api/v1/rates/:id
- [ ] Version history (show previous versions, read-only)

### Shift Config Page (2 days)
**Deliverables:**
- [ ] Shift config list (table view)
- [ ] Columns: shift_code, shift_number, start_time, end_time, is_overnight, version
- [ ] Create shift config button → form modal:
  - Shift code (text, e.g., "06-14")
  - Shift number (number)
  - Start time (time picker)
  - End time (time picker)
  - Is overnight (checkbox)
  - Validation
  - Submit → POST /api/v1/locations/:id/shift-configs
- [ ] Edit button → edit modal
  - Submit → PATCH /api/v1/locations/:id/shift-configs/:code

### User Management Page (3 days)
**Deliverables:**
- [ ] User list (table view)
- [ ] Columns: name, email, role, locations, status, created_at
- [ ] Filter by role, status
- [ ] Create user button → form modal:
  - Name (text)
  - Email (email, validation)
  - Password (password, min 8 chars)
  - Role (select: owner, admin, manager, operator)
  - Locations (multi-select)
  - Validation
  - Submit → POST /api/v1/users
- [ ] Edit button → edit modal (name, email, role, locations, status)
  - Submit → PATCH /api/v1/users/:id
- [ ] Deactivate button → confirm modal
  - Submit → POST /api/v1/users/:id/deactivate
- [ ] Reset password button → modal (new password)
  - Submit → POST /api/v1/users/:id/reset-password

### Role/Permission Management Page (2 days)
**Deliverables:**
- [ ] Role list (table view)
- [ ] Columns: name, permissions count, users count
- [ ] Create role button → form modal:
  - Name (text)
  - Permissions (checkboxes, grouped by category)
  - Submit → POST /api/v1/roles
- [ ] Edit button → edit modal
  - Submit → PATCH /api/v1/roles/:id
- [ ] Permission categories:
  - Sessions (view, create, close, void)
  - Payments (view, collect_cash, collect_digital, void)
  - Users (view, create, edit, deactivate)
  - Locations (view, create, edit, deactivate)
  - Rates (view, create, edit)
  - Reports (view_revenue, view_occupancy, view_operators)
  - Observability (view_health, view_audit, view_alerts, manage_alerts)

### Manual Refresh Buttons (1 day)
**Deliverables:**
- [ ] Refresh configs button (header or settings page)
  - Click → POST /api/v1/server-room/refresh-configs
  - Show loading state
  - Success → toast notification
- [ ] Refresh data button (header or dashboard)
  - Click → POST /api/v1/server-room/refresh-data
  - Show loading state
  - Success → toast notification

---

## Week 7-8: Dashboard - Reports

### Daily Revenue Report (3 days)
**Deliverables:**
- [ ] Report page (daily revenue)
- [ ] Date range picker (start date, end date)
- [ ] Revenue chart (line chart, daily totals)
  - Use recharts or chart.js
- [ ] Revenue table (date, total_revenue, transaction_count, avg_transaction)
- [ ] Filter by vehicle_type
- [ ] Export to CSV button
  - Click → GET /api/v1/reports/daily-revenue?format=csv
  - Download file

### Occupancy Report (2 days)
**Deliverables:**
- [ ] Report page (occupancy)
- [ ] Date range picker
- [ ] Occupancy chart (bar chart, by time bucket)
- [ ] Occupancy table (time_bucket, avg_occupancy, peak_occupancy)
- [ ] Filter by vehicle_type
- [ ] Export to CSV

### Transaction Report (2 days)
**Deliverables:**
- [ ] Report page (transactions)
- [ ] Date range picker
- [ ] Transaction table (date, time, location, gate, vehicle_type, amount, payment_method, shift)
- [ ] Filters: vehicle_type, payment_method, shift
- [ ] Sort by date, amount
- [ ] Export to CSV

### Vehicle Breakdown Report (1 day)
**Deliverables:**
- [ ] Report page (vehicle breakdown)
- [ ] Date range picker
- [ ] Pie chart (revenue by vehicle_type)
- [ ] Table (vehicle_type, transaction_count, total_revenue, avg_revenue)
- [ ] Export to CSV

### Operator Activity Report (1 day)
**Deliverables:**
- [ ] Report page (operator activity)
- [ ] Date range picker
- [ ] Table (operator_name, transaction_count, total_revenue, avg_transaction)
- [ ] Sort by transaction_count, total_revenue
- [ ] Export to CSV

---

## Week 9-10: Gate App UI (Driver-Facing Monitor)

### Project Setup (1 day)
**Deliverables:**
- [ ] Electron + React app (or Tauri + React)
- [ ] Project structure (similar to dashboard)
- [ ] API client (communicate with server room app)
- [ ] State management (Zustand or Redux Toolkit)

### Entry Gate UI (2 days)
**Deliverables:**
- [ ] Idle screen:
  - "Press button for ticket" (large text, centered)
  - Animated icon (button press animation)
- [ ] Processing screen:
  - "Dispensing ticket..." (loading spinner)
  - Progress indicator
- [ ] Success screen:
  - "Ticket dispensed" (checkmark icon)
  - "Gate opening..." (animated gate icon)
  - Auto-transition to idle after 3s

### Exit Gate UI (3 days)
**Deliverables:**
- [ ] Idle screen:
  - "Scan your ticket" (large text, centered)
  - Animated QR scanner icon
- [ ] Fee display screen:
  - Fee amount (large, bold)
  - Check-in time
  - Duration (e.g., "2 hours 15 minutes")
  - Vehicle type
  - "Please tap your e-money card" (instruction)
  - Animated card tap icon
- [ ] Payment processing screen:
  - "Processing payment..." (loading spinner)
  - Progress indicator
- [ ] Payment success screen:
  - "Payment successful" (checkmark icon, green)
  - Fee amount
  - "Gate opening..." (animated gate icon)
  - Auto-transition to idle after 3s
- [ ] Payment failed screen:
  - "Payment failed" (X icon, red)
  - "Insufficient balance, please topup"
  - "Press alert button for help" (instruction)
  - Auto-transition to idle after 10s
- [ ] Error screen:
  - "Out of service" (warning icon, yellow)
  - "Please press alert button for help"
- [ ] Receipt screen (optional):
  - Receipt button (bottom corner)
  - On press → print receipt (check-in/out time, fee, vehicle type, shift)

### State Management (1 day)
**Deliverables:**
- [ ] Gate state (idle, processing, success, error)
- [ ] Session data (fee, check_in_time, duration, vehicle_type)
- [ ] Payment state (pending, processing, success, failed)
- [ ] Auto-transition timers (idle after X seconds)

### Hardware Integration (2 days)
**Deliverables:**
- [ ] QR scanner listener (USB HID, read keyboard input)
- [ ] Payment terminal listener (vendor-specific, TBD)
- [ ] Receipt button listener (physical button)
- [ ] Communication with server room app (HTTP client)
- [ ] Printer integration (thermal printer, ESC/POS)

---

## Week 9-10: Server Room App UI (Staff Dashboard)

### Project Setup (1 day)
**Deliverables:**
- [ ] Electron + React app (or Tauri + React)
- [ ] Project structure
- [ ] API client (communicate with local server room backend)
- [ ] State management

### Main Monitoring Screen (3 days)
**Deliverables:**
- [ ] Gate status grid (cards for each gate)
- [ ] Gate card:
  - Gate name + ID
  - Status badge (online/offline/busy)
  - Health indicator (green/yellow/red)
  - Last seen timestamp
  - Click → gate detail modal
- [ ] Auto-refresh (every 15s, fetch gate statuses)
- [ ] Alert banner (top of screen, red background)
  - "Gate {gate_id} needs assistance"
  - Audio alert (play sound, announce gate_id)
  - Dismiss button

### Alert Detail Modal (1 day)
**Deliverables:**
- [ ] Modal with alert details:
  - Gate ID
  - Alert type (hardware failure, payment failure, QR unreadable)
  - Timestamp
  - Action buttons: acknowledge, resolve
- [ ] Acknowledge → mark as acknowledged
- [ ] Resolve → mark as resolved, close modal

### Offline Mode Screen (1 day)
**Deliverables:**
- [ ] Offline banner (top of screen, yellow background)
  - "System offline" (internet disconnected)
  - "Gates still operational"
  - "Transactions queued for sync"
- [ ] Sync status:
  - Queue length (number of unsynced transactions)
  - Last sync attempt (timestamp)
  - Retry button (manual sync)

### Manual Refresh Controls (1 day)
**Deliverables:**
- [ ] Refresh configs button (header)
  - Click → trigger config refresh from cloud
  - Show loading state
  - Success → toast notification
- [ ] Refresh data button (header)
  - Click → trigger data refresh (gate statuses)
  - Show loading state
  - Success → toast notification

### Settings Page (1 day)
**Deliverables:**
- [ ] Server room app settings:
  - Cloud backend URL (configurable)
  - Sync interval (default: 60s)
  - Gate health check interval (default: 15s)
  - Audio alert volume
  - Theme (light/dark)
- [ ] Save → persist to local config file

---

## Frontend Timeline Summary

| Week | Dashboard (Next.js) | Gate App UI (Electron) | Server Room App UI (Electron) |
|------|---------------------|------------------------|-------------------------------|
| **1-2** | Foundation (setup, design system, auth, layout) | — | — |
| **3-4** | Live monitoring (gate status, active sessions, revenue widget, alert widget) | — | — |
| **5-6** | Config management (rates, shifts, users, roles, refresh buttons) | — | — |
| **7-8** | Reports (daily revenue, occupancy, transactions, vehicle breakdown, operator activity) | — | — |
| **9-10** | Polish, bug fixes, accessibility audit | Entry + exit UI (idle, processing, success, error, receipt) | Monitoring screen (gate grid, alerts, offline mode, refresh) |

---

## Frontend Tech Stack

### Dashboard
- **Framework:** Next.js 14 (App Router)
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **UI Components:** shadcn/ui (or custom)
- **State Management:** React Context + hooks (or Zustand)
- **API Client:** fetch (or axios)
- **Charts:** recharts or chart.js
- **Forms:** React Hook Form + Zod validation
- **Testing:** Jest + React Testing Library

### Gate App & Server Room App
- **Framework:** Electron + React (or Tauri + React)
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **UI Components:** shadcn/ui (or custom)
- **State Management:** Zustand
- **API Client:** fetch (communicate with server room backend)
- **Hardware Integration:** Node.js serialport, USB libraries
- **Testing:** Jest + React Testing Library

---

## Frontend Risks & Mitigation

| Risk | Mitigation |
|------|------------|
| Design system inconsistencies | Use component library (shadcn/ui), enforce design tokens |
| Performance (large data tables) | Pagination, virtualization, lazy loading |
| Accessibility issues | WCAG 2.1 AA audit, keyboard navigation, screen reader testing |
| Hardware integration bugs (gate app) | Use mocks for testing, fallback UI for hardware failures |
| State management complexity | Keep state local where possible, use Zustand for shared state |
| Browser compatibility (dashboard) | Test on Chrome, Firefox, Safari, Edge (last 2 versions) |

---

*End of Frontend Timeline*
