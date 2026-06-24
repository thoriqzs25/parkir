# Milestone Planning Workflow

## How to Use This

Before writing code for any milestone, create a detailed milestone plan.

**To trigger a milestone planning session, ask OpenCode:**

> "Create a detailed plan for Milestone X."

OpenCode will then **grill you with questions** specific to that milestone. Once you answer, it will produce a detailed milestone plan in a dedicated file (e.g., `plans/milestone-0.md`).

---

## Why This Matters

Milestones in PARKIR are complex. They touch multiple layers (backend, dashboard, desktop), have unclear decisions hidden in the specs, and often require tradeoffs between speed and correctness. The grilling step surfaces assumptions before code is written.

---

## Milestone Plan Template

Every milestone plan will be written using this structure:

```markdown
# Milestone X — [Name]

## 1. Goal (one sentence)
What is the single outcome this milestone must deliver?

## 2. Scope
### In Scope
- ...

### Out of Scope / Deferred
- ...

## 3. Dependencies
- Requires Milestone Y to be complete?
- Needs decisions from product/business?
- External integrations or hardware?

## 4. Detailed Tasks
### Backend
- [ ] Task 1
- [ ] Task 2

### Dashboard
- [ ] Task 1
- [ ] Task 2

### Desktop
- [ ] Task 1
- [ ] Task 2

### DevOps / QA
- [ ] Task 1

## 5. Technical Decisions
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Decision 1 | ... | ... |

## 6. Open Questions / Risks
| Question | Owner | Due Date |
|----------|-------|----------|
| ... | ... | ... |

## 7. Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2

## 8. Definition of Done
What must be true for this milestone to be considered complete?
```

---

## Milestone-Specific Grill Questions

### Milestone 0 — Foundation
1. Do you want the backend, dashboard, and desktop app in a single monorepo or separate repos?
2. Should `docker-compose.yml` include hot-reload for Go and Next.js in development?
3. Do you want database migrations managed by a Go migration tool (golang-migrate), or plain SQL files run manually?
4. Should the backend serve the Next.js dashboard as static files, or should they run as separate services in production?
5. What Go version? What Node.js version?
6. Do you want structured logging (JSON) from day one, or plain text logs?
7. Should tests run in CI from day one, or is CI just lint/build for now?
8. Do you want a seed script with default roles/users, or create them manually?

### Milestone 1 — Backend Core Entities
1. Should passwords require complexity rules (length, symbols), or just minimum length?
2. How do you want to handle the first owner/admin account — seed data, CLI command, or manual DB insert?
3. Should role permissions be validated against a hardcoded allow-list, or free-form strings?
4. Do you want location deactivation to immediately kick out active operators, or let them finish their shift?
5. Should rate changes apply retroactively to existing active sessions, or only to new check-ins?
6. How do you want to handle overlapping rate effective dates for the same location/vehicle type?
7. Do you want API response envelopes (`{ data: ..., error: ... }`) or raw resources?
8. Should JWT tokens include user permissions, or look them up per request?

### Milestone 2 — Backend Business Logic
1. What should happen if no rate is configured at check-out — block check-out or allow manual override?
2. Should receipt numbers be strictly sequential, or is gap tolerance acceptable?
3. How should we handle daylight saving time / timezone for receipt timestamps?
4. Should a manager be able to force-close another operator's active shift?
5. What happens if an operator tries to start a new shift while another is still open?
6. Should digital payments require a reference code minimum length or format?
7. Do we allow check-out of a session checked in by another operator?
8. Should voided sessions be physically deleted from reports, or shown with a VOIDED badge?

### Milestone 3 — Web Dashboard Foundation
1. What should the dashboard landing page show — location selector first, or default to the user's first location?
2. Do you want inline editing for rates, or a separate edit form?
3. Should session lists auto-refresh, or require manual refresh?
4. Do you want pagination, infinite scroll, or simple limit/offset tables?
5. Should the dashboard be mobile-responsive, or desktop-only?
6. Do you want toast notifications for actions, or inline success messages?
7. Should forms validate on blur, on submit, or both?
8. What timezone should the dashboard display timestamps in — user timezone or location timezone?

### Milestone 4 — Desktop App Online Mode
1. Should the desktop app support multiple active locations per login, or lock to one location until switch?
2. Should the operator be forced to start a shift before any check-in, or can they browse first?
3. Do you want barcode scanner support for plate entry, or manual keyboard only?
4. Should the check-out screen auto-search as the operator types, or require explicit submit?
5. How should the app behave if the printer fails mid-print — retry, skip, or block session close?
6. Do you want a "quick reprint" button on the success screen, or only from session history?
7. Should the desktop app remember the logged-in operator between restarts, or require login every time?
8. What screen resolution is the minimum target — 1024x768, or something else?

### Milestone 5 — Offline Mode & Sync
1. What is the maximum time an operator should be able to work offline — hours, days?
2. Should the app warn the operator when the rate cache is stale, or silently use last known rates?
3. How should we handle a sync where the operator's local clock is wrong?
4. Should offline receipts be reprinted automatically after sync with the official receipt number?
5. What happens if a vehicle checks in offline and the same plate is checked in online before sync?
6. Do you want sync to happen automatically on reconnect, or require operator confirmation?
7. Should incidents filed offline be auto-synced, or require manager approval?
8. How much local storage space should the app reserve/cap for offline data?

### Milestone 6 — Incidents, Adjustments & Observability
1. Who can resolve incidents — any manager at the location, or only assigned managers?
2. Should incident resolution trigger an automatic adjustment (void/reassign), or be separate?
3. Do you want email/SMS/WhatsApp alerts in v1 if an alert fires, or only in-app badges?
4. Should the manager PIN expire after failed attempts (lockout), or just reject?
5. Should reassigning a session update historical reports retroactively, or preserve original attribution?
6. Do you want real-time health polling, or refresh-on-load?
7. Should audit logs be exportable to CSV in v1?
8. Who can configure alert thresholds — admins only, or managers too?

### Milestone 7 — Reports & Polish
1. What is the maximum date range for reports — 30 days, 90 days, unlimited?
2. Should reports be cached/pre-aggregated for past dates, or always query live?
3. Do you want charts in v1, or tables only?
4. Should reports be printable as PDF, or CSV export only?
5. Who can see operator activity reports — managers at their locations, or only admins/owners?
6. Should revenue reports include voided transactions in a separate column, or hide them entirely?
7. Do you want scheduled report emails in v1?
8. What date format should reports use — DD/MM/YYYY, YYYY-MM-DD, or localized?

### Milestone 8 — Testing & Deploy
1. What is the rollback strategy if a deployment breaks?
2. Do you want a staging environment, or deploy straight to production?
3. How should database backups be scheduled — daily, hourly?
4. Should logs be shipped to an external service (e.g., Loki, CloudWatch), or kept on disk?
5. Do you want SSL/TLS handled by the VM (Let's Encrypt) or by Tencent Cloud load balancer?
6. Should the Electron app auto-update on launch, or prompt the user?
7. What is the minimum acceptable test coverage before launch?
8. Who performs the final production smoke test — you, the team, or end users?

---

## Output Location

Milestone plans will be saved to:

```
plans/milestone-0.md
plans/milestone-1.md
plans/milestone-2.md
...
```

If the `plans/` directory does not exist, OpenCode will create it.

---

## Example Trigger Phrases

- "Create a detailed plan for Milestone 0."
- "Plan Milestone 4 for me."
- "I want to start Milestone 2 — give me the detailed plan and ask me the hard questions."
- "Break down Milestone 5 into tasks and grill me first."
