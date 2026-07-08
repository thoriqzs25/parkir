# Milestone 6 — Incidents, Adjustments & Observability

## 1. Goal

Build exception handling, corrections, audit, and monitoring so that managers can resolve operational incidents, perform adjustments with PIN verification, query the immutable audit log, and monitor system health and alerts — all from the dashboard, with email notification for alerts.

## 2. Scope

### In Scope
- `incidents` and `incident_notes` DB tables (new migration)
- `alerts` and `alert_configs` DB tables (new migration)
- Incident reporting from the desktop (operator) and dashboard (manager)
- Incident list/detail + note thread + resolve workflow (dashboard)
- Adjustments: void transaction + manager PIN confirmation
- Adjustments: reassign session to another operator/shift + manager PIN confirmation
- When resolving an incident, option to auto-trigger an adjustment (void/reassign)
- Audit log list/detail page in dashboard with action, actor, entity, date filters
- System health dashboard page (API status, DB connectivity, uptime, last check)
- Alert engine: `LONG_SESSION`, `UNPAID_EXIT`, `HIGH_VOID_RATE`, `SYNC_FAILURE`, `GATEWAY_FAILURE`
- Alert list + acknowledge/resolve workflow in dashboard
- Alert config management page (admin/owner only)
- Email push notification for triggered alerts (SMTP, configured at system level)
- Health polling every 1 minute for dashboard health page
- CSV export for audit logs

### Out of Scope / Deferred
- SMS/WhatsApp notifications for alerts (email only in v1)
- Real-time WebSocket push for alerts (polling-based refresh on page load)
- Automated alert-to-incident conversion
- Self-healing actions (restart, failover)
- Historical alert replay
- Advanced alert aggregation or deduplication
- Automated email alert digest or scheduled reports

## 3. Dependencies
- **Milestone 2 — Backend Business Logic** (provides sessions, transactions, shifts that incidents/alerts reference)
- **Milestone 3 — Web Dashboard Foundation** (provides layout, navigation, auth, table components)
- **Milestone 4 — Desktop App Online Mode** (desktop incident filing builds on its UI patterns)
- **Milestone 5 — Offline Mode & Sync** (desktop uses sync infrastructure for sending incident records)
- PIN verification logic already exists in `backend/internal/auth/auth.go` and `backend/internal/domain/transactions/handler.go`
- `audit_logs` DB table already exists (migration 000002) — no new migration needed
- Permissions for all domains already defined in `backend/internal/permissions/permissions.go`
- Email sending capability (SMTP configuration) must be added to backend config

## 4. Detailed Tasks

### Backend

- [ ] Create migration `000005_incidents_alerts.up.sql`:
  - `incidents` table (id, location_id, type, state, session_id, reported_by, reported_at, description, resolved_by, resolved_at, resolution_notes, offline_sync, created_at, updated_at)
  - `incident_notes` table (id, incident_id, author_id, note, created_at)
  - `alerts` table (id, code, location_id, state, entity_type, entity_id, triggered_at, acknowledged_by, acknowledged_at, resolved_by, resolved_at, resolution_notes, metadata, created_at)
  - `alert_configs` table (id, location_id nullable, code, enabled, threshold JSONB, updated_by, updated_at, unique per location_id+code)
- [ ] Create `internal/domain/incidents/` with handler, routes, store:
  - `GET /api/v1/incidents` (list with filters: location, type, state, date range)
  - `POST /api/v1/incidents` (file incident, operator or manager)
  - `GET /api/v1/incidents/:id` (detail with notes)
  - `PATCH /api/v1/incidents/:id` (update state + resolution notes, trigger optional adjustment)
  - `POST /api/v1/incidents/:id/notes` (add note to thread)
- [ ] When resolving an incident, offer optional linked adjustment: void associated transaction or reassign session
- [ ] Create `internal/domain/adjustments/` with handler, routes (no dedicated table needed):
  - `POST /api/v1/adjustments/void-transaction` (requires manager PIN, marks transaction voided, writes audit log)
  - `POST /api/v1/adjustments/reassign-session` (requires manager PIN, updates session.operator_id + shift_id, writes audit log)
- [ ] Reuse existing `validateManagerPIN()` pattern for both adjustment endpoints
- [ ] Create `internal/domain/auditlogs/` with handler, routes:
  - `GET /api/v1/audit-logs` (list with filters: action, actor, entity_type, entity_id, location, date range)
  - `GET /api/v1/audit-logs/export` (CSV download)
- [ ] Create `internal/domain/alerts/` with handler, routes, store:
  - `GET /api/v1/alerts` (list with filters: location, state, code)
  - `POST /api/v1/alerts/:id/acknowledge`
  - `POST /api/v1/alerts/:id/resolve`
  - `GET /api/v1/alert-configs` (list per location)
  - `PATCH /api/v1/alert-configs/:id` (update enabled, threshold; admin/owner only)
- [ ] Implement alert engine:
  - Check `LONG_SESSION`: session active > configured threshold (e.g., 24h)
  - Check `UNPAID_EXIT`: session in `PENDING_PAYMENT` > threshold
  - Check `HIGH_VOID_RATE`: void count per shift > threshold
  - Check `SYNC_FAILURE`: offline sync batch had failures
  - Check `GATEWAY_FAILURE`: health check for payment/db fails
  - Run checks on a periodic goroutine (every 5 minutes) or event-triggered
- [ ] Add SMTP email config to backend config (`internal/config/`)
- [ ] Implement email notification sender (simple SMTP client, HTML template per alert code)
- [ ] On alert trigger, send email to configured recipients (system-level email list)
- [ ] Extend `/health/ready` to include component status (DB, uptime, last alert check)
- [ ] Extend existing `logAudit()` for all new mutation endpoints
- [ ] Write integration tests:
  - File incident → resolve with adjustment → verify transaction voided
  - Void transaction with correct PIN → success
  - Void transaction with wrong PIN → reject
  - Reassign session → verify attribution preserved
  - Audit log query with filters
  - Alert lifecycle: trigger → acknowledge → resolve
  - Alert config update (admin vs non-admin)
  - CSV export returns valid CSV

### Dashboard

- [ ] Add **Incidents** nav item and page group
- [ ] Incidents list page with filters (location, type, state, date range)
- [ ] Incidents detail page with note thread, state timeline, linked session/transaction
- [ ] Incident resolve form: resolution notes + optional adjustment checkbox (void/reassign)
- [ ] Add **Adjustments** section (or combine with incidents):
  - Void transaction modal/page: enter transaction ID + manager PIN
  - Reassign session modal/page: search session + new operator/shift + manager PIN
- [ ] Add **Audit Logs** nav item and page:
  - Audit log list with filters (action, actor, entity type, date range)
  - Audit log CSV export button
- [ ] Add **Health** nav item and page:
  - Component cards: API status (up/down), DB connectivity, Uptime (seconds), Last health check timestamp
  - Auto-refresh every 60 seconds via client-side polling
- [ ] Add **Alerts** nav item:
  - Alerts list page with filters (location, state, code)
  - Alert detail page with acknowledge/resolve actions
  - Alert config management page (admin/owner only): enable/disable per location, set thresholds
- [ ] Add alert notification badge in nav bar (count of active/triggered alerts)
- [ ] Wire all pages to use canonical API client in `dashboard/src/lib/api.ts`

### Desktop

- [ ] Incident report screen (existing nav entry from Milestone 4) — already exists as a screen placeholder
- [ ] Ensure incident filing sends to `POST /api/v1/incidents` endpoint
- [ ] Desktop offline incident queue (from Milestone 5) sends to incident endpoint on reconnect
- [ ] Minor: show alert/health status indicator (if API responds, show green dot)

### DevOps / QA

- [ ] SMTP configuration documented in `.env.example` and README
- [ ] Integration test for full incident → resolve → adjustment flow
- [ ] Integration test for alert trigger → email notification (mock SMTP)
- [ ] Test: manager PIN rejection on wrong PIN (no lockout, just retry)
- [ ] Test: HTML email rendering for at least one alert code
- [ ] Verify dashboard health page auto-refreshes at 60s interval
- [ ] Verify CSV export downloads correctly
- [ ] Seed data: add some alert configs for the default location

## 5. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Incident resolution | Any manager at the location can resolve | Simpler workflow; no need for assignment tracking |
| Incident → Adjustment link | Optional checkbox during resolution | Gives managers flexibility to resolve without auto-changing data |
| Alert notifications | Email + in-app badge | Email for async notification, badge for immediate awareness in dashboard |
| Manager PIN failure | Reject + retry (no lockout) | Consistent with user preference; avoids support escalations for forgotten PINs |
| Session reassignment | Preserve original operator/shift in historical reports | `transactions` record is immutable; reassignment adds a reassignment audit trail without rewriting history |
| Health polling interval | Every 60 seconds | Balances freshness with server load; matches user preference |
| Audit log export | CSV | User explicitly requested CSV export in v1 |
| Alert config access | Admin/owners only | Sensitive thresholds should not be changeable by location managers |
| Alert engine scheduling | Periodic goroutine every 5 min + event-triggered on relevant mutations | Keeps implementation simple while still catching issues promptly |
| Email implementation | Go `net/smtp` with HTML templates | No external dependency; easy to configure; sufficient for v1 |

## 6. Open Questions / Risks

| Question / Risk | Owner | Due Date |
|-----------------|-------|----------|
| What SMTP server/credentials will be used for email alerts (Tencent Cloud SES, self-hosted SMTP, or third-party)? | Product/Ops | Before alert implementation |
| Who are the default email recipients for alerts? System-level list or per-location? | Product | Before alert implementation |
| Should `HIGH_VOID_RATE` threshold be absolute count per shift, or percentage of total transactions? | Backend + Product | Before alert engine implementation |
| How do we handle SMTP transient failures — retry, queue, or skip? | Backend | Before email sender implementation |
| Should the desktop app show any alert/notification from the server (e.g., banner for "system maintenance")? | Desktop | Optional — can defer |
| Audit log table already exists but has no index on location_id + timestamp for efficient querying — may need a migration for composite index | Backend | Before audit log handler implementation |
| The health endpoint currently returns minimal info — extending it with component status may need a new response shape | Backend | Before health dashboard implementation |

## 7. Acceptance Criteria

- [ ] Operator can file an incident from the desktop app (when online)
- [ ] Manager can view, filter, and resolve incidents from the dashboard
- [ ] Resolving an incident can optionally void the linked transaction or reassign the session
- [ ] Adjustments (void transaction, reassign session) require correct manager PIN
- [ ] Wrong PIN is rejected with a retry message (no lockout)
- [ ] Reassigning a session preserves original operator/shift attribution in reports
- [ ] Audit log page lists all mutations with filtering and CSV export
- [ ] Health dashboard shows component status and auto-refreshes every 60 seconds
- [ ] Alerts are triggered for LONG_SESSION, UNPAID_EXIT, HIGH_VOID_RATE, SYNC_FAILURE, GATEWAY_FAILURE
- [ ] Triggered alerts appear in the dashboard alert list with a badge count
- [ ] Alert configs can be enabled/disabled and thresholds adjusted by admin/owners
- [ ] Triggered alerts send an email to configured recipients
- [ ] All new API endpoints return consistent envelope format `{ data, error, meta }`
- [ ] Integration tests pass for incidents, adjustments, audit logs, and alerts
- [ ] Dashboard builds and type-checks without errors

## 8. Definition of Done

- Backend: incidents, adjustments, audit logs, and alert CRUD + alert engine + email notification are implemented and integration-tested.
- Dashboard: incidents, adjustments, audit logs, health, and alerts pages are implemented, type-check, and connect to real backend endpoints.
- Desktop: incident filing is wired to the backend endpoint (online mode).
- All code is reviewed and merged to `main`.
- `PLAN.md` is updated if any decisions diverged from this plan.