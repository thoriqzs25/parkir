# Milestone 8 — Testing & Deploy

## 1. Goal

Get the entire PARKIR system live on Tencent Cloud Jakarta with a staging environment, automated daily DB backups, and a documented runbook so operators and managers can log in and perform core flows.

## 2. Scope

### In Scope

- Backend integration tests complete and passing
- Dashboard smoke tests (build, type-check, basic page load)
- Desktop app smoke tests (login, check-in, check-out, payment, receipt)
- Tencent Cloud Jakarta VM provisioning for staging and production
- PostgreSQL 15 install + configuration on both VMs
- Daily DB backup to local disk on the VM with 90-day retention
- Dashboard backup status page showing latest backup progress, timestamp, and file size
- Go backend deploy via Docker on both staging and production
- Next.js dashboard deploy via Docker (served by Go backend or reverse-proxied)
- SSL/TLS via Let's Encrypt (certbot) on both environments
- Staging environment for backend + dashboard + desktop staging builds
- Desktop manual download/install for both staging and production builds
- Environment-specific config (dev/staging/prod)
- Production smoke test (manual, by you)
- Deployment runbook (written as deploy.md or equivalent)
- Loki log shipping integration (scaffold + config; actual Loki endpoint to be added later)
- Log rotation on-disk as fallback until Loki is wired

### Out of Scope / Deferred

- Automated CI/CD pipeline (deployment is manual Docker operations for now)
- Centralized monitoring pod integration (deferred; you'll wire it later)
- Electron auto-update pipeline (manual download only)
- Load testing / performance benchmarking
- Real user monitoring (RUM) or synthetic monitoring
- Canary / blue-green deployment (simple stop-and-swap)

## 3. Dependencies

- **All Milestones 0–7 complete** — every feature and report is implemented, reviewed, and merged
- Tencent Cloud Jakarta VM provisioned and accessible via SSH
- Domain/subdomain deferred — will decide later
- Docker installed on staging and production VMs
- Loki endpoint URL + credentials (you'll provide later — placeholder config for now)

## 4. Detailed Tasks

### Backend

- [ ] Review and complete backend integration test suite:
  - Auth: login, refresh, logout, me
  - Users: CRUD, deactivate, RBAC enforcement
  - Roles: CRUD, permissions
  - Locations: CRUD, operator assignment, deactivation
  - Rates: CRUD, effective date overlap
  - Sessions: check-in (duplicate plate warning), check-out (fee calc), void
  - Transactions: cash payment, digital payment, void, calculate-fee
  - Shifts: start, end, force-close, discrepancy calculation
  - Incidents: create, resolve, notes
  - Adjustments: void-transaction, reassign-session
  - Reports: daily-revenue, occupancy, vehicle-breakdown, operator-activity
  - Audit logs: list, filter
  - Alerts: list, acknowledge, resolve, config
- [ ] Add `Dockerfile` for production Go build (multi-stage: build with `go build`, run with `scratch` or `alpine`)
- [ ] Add `Dockerfile` for staging Go build (similar but with `GIN_MODE=release` overridable)
- [ ] Add `docker-compose.prod.yml` for production stack (backend + dashboard behind Caddy or nginx for TLS)
- [ ] Add `docker-compose.staging.yml` for staging stack
- [ ] Environment config via env vars: `DATABASE_URL`, `JWT_SECRET`, `JWT_PUBLIC_KEY`, `JWT_PRIVATE_KEY`, `GIN_MODE`, `CORS_ORIGINS`, `LOKI_URL` (optional), `LOG_LEVEL`
- [ ] Structured JSON logging (already expected from M0) — ensure it outputs to stdout for Docker + Loki
- [ ] Loki log shipping: add a `loki` driver or use Docker logging plugin; alternatively, add a lightweight log shipper sidecar (promtail) to docker-compose
- [ ] Health endpoint `GET /health/ready` (already exists per AGENTS.md) — verify it works with `SELECT 1`
- [ ] Graceful shutdown on `SIGTERM` / `SIGINT`

### Dashboard

- [ ] Review and fix any TypeScript errors in the dashboard (run `npm run typecheck`)
- [ ] Add `Dockerfile` for Next.js dashboard (standalone output mode, `node:22-alpine` runner)
- [ ] Verify all pages load without errors (basic smoke test):
  - Login page
  - Users, roles, locations, rates pages
  - Active sessions, session history, session detail
  - Transactions list
  - Shifts list, shift detail
  - Incidents list, detail
  - Audit log page
  - Reports (daily revenue, occupancy, vehicle breakdown, operator activity)
  - Alert list and config
- [ ] Verify error boundaries catch and display errors (not blank pages)
- [ ] Verify all API calls use the correct `baseUrl` from environment config
- [ ] Verify CORS is configured correctly between dashboard domain and API domain
- [ ] Verify location selector works in production (user assigned to correct locations)
- [ ] Add loading states to any pages that still lack them
- [ ] Ensure `NEXT_PUBLIC_API_URL` is configurable per environment
- [ ] Create backup status page `app/(dashboard)/[locationId]/settings/backups/page.tsx`:
  - Table showing backup filename, file size, timestamp, status (success/failed/running)
  - "Trigger Backup Now" button (manager/owner only)
  - Auto-refresh or manual refresh

### Desktop

- [ ] Review and fix any TypeScript errors (run `npm run typecheck` or equivalent)
- [ ] Verify desktop app builds for Linux (and Windows/macOS if applicable)
- [ ] Staging desktop build: generate a binary/package tagged with staging env config
- [ ] Production desktop build: generate a binary/package tagged with production env config
- [ ] Verify full flow in online mode:
  - Login → select location → start shift → check-in → check-out → payment (cash + digital) → receipt print
- [ ] Verify offline detection works (disable network → app shows offline indicator)
- [ ] Verify printer connection and receipt print
- [ ] Verify incident filing works
- [ ] Manual distribution: desktop binary is built locally and shared directly (e.g., USB, SCP, or link)
- [ ] Staging desktop build: generate and test locally
- [ ] Production desktop build: generate and distribute manually

### DevOps / QA

- [ ] Provision staging VM on Tencent Cloud Jakarta (e.g., 2C4G, 40GB SSD) — allow ports 22 and 443 only
- [ ] Provision production VM on Tencent Cloud Jakarta (e.g., 4C8G, 80GB SSD) — allow ports 22 and 443 only
- [ ] Install Docker + Docker Compose on both VMs
- [ ] Install PostgreSQL 15 on both VMs (or run as a Docker container with persistent volume)
- [ ] PostgreSQL config:
  - Set `max_connections` appropriately (e.g., 100)
  - Enable `log_statement = 'ddl'` for audit
  - Set up `pg_hba.conf` for app connections
- [ ] Create PostgreSQL users and databases for staging and production
- [ ] Run migrations and seed on staging, then verify data
- [ ] Run migrations and seed on production
- [ ] Set up Let's Encrypt SSL via certbot on both VMs (or use a reverse proxy like Caddy with auto-TLS)
- [ ] Configure domain DNS (e.g., `staging-api.parkir.local` → staging VM IP, `api.parkir.local` → production VM IP)
- [ ] Daily DB backup script:
  - Use `pg_dump` to dump to a local directory (e.g., `/var/backups/parkir/`)
  - Filename format: `parkir-YYYYMMDD-HHMMSS.sql.gz`
  - Retention: 90 days (rotate by deleting files older than 90 days)
  - Schedule via cron: `0 2 * * *`
- [ ] Backend endpoint `GET /api/v1/backups` — returns list of backup files (filename, size, timestamp, status)
- [ ] Backend runs the backup via a goroutine on schedule (or cron triggers a lightweight API call)
- [ ] Backup restore procedure documented in runbook
- [ ] Log rotation on-disk (fallback): configure `logrotate` for Docker container logs
- [ ] Loki log shipping (placeholder):
  - Add promtail sidecar config to docker-compose
  - Loki endpoint config: placeholder URL (you'll fill in later)
  - Verify promtail scrapes Docker logs correctly
- [ ] Environment config files (`.env.staging`, `.env.production`) with documentation
- [ ] Production smoke test (manual, by you):
  - Visit dashboard, log in as owner
  - Create a location, configure rates
  - Create an operator user
  - Log in via desktop app, start shift, check in a vehicle, check out, process payment, print receipt
  - Verify receipt in session history
  - Verify transaction appears in reports
  - Verify shift shows correct cash discrepancy
  - Verify audit log entries for all actions
- [ ] Write deployment runbook (`deploy.md` or in `docs/`):
  - Prerequisites (SSH access, Docker, Docker Compose, domain placeholder, SSL)
  - Staging deploy steps (copy source → `docker compose build` → stop old container → start new container → verify health)
  - Production deploy steps (same)
  - Desktop build and distribution steps (local build, manual transfer)
  - Database backup restore procedure
  - Rollback procedure (re-tag previous Docker image and restart)
  - Environment variable reference
  - Monitoring checklist (health endpoint, logs, backups)
  - Incident response (what to do if the API crashes, DB is down, desktop can't connect)

## 5. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Deployment model | Docker on VM; docker-compose for stack | Consistent with existing dev setup; simple stop-and-swap rollback |
| SSL | Let's Encrypt certbot (or Caddy with auto-TLS) | Free, automated renewal; user preference |
| DB backups | `pg_dump` to local disk on VM; dashboard shows backup status | User choice — no external storage needed in v1 |
| Backup retention | 90 days, rotate by file age | User needs last 2 months queryable |
| Staging | Separate VM, same deploy pattern as production | User wants staging for backend, dashboard, and desktop |
| Desktop distribution | Manual build from user's machine, distributed directly | User preference — no download hosting needed |
| Logging | Docker JSON driver + promtail sidecar → Loki (placeholder) | User will wire centralized Loki later; promtail config ready |
| Rollback | Re-tag previous Docker image and restart | Simple, no orchestration needed in v1 |
| Test coverage requirement | No minimum % — user performs manual smoke test | User explicitly waived coverage requirement |
| Firewall | Ports 22 (SSH) and 443 (HTTPS) only | User preference; minimal attack surface |
| Docker image delivery | Build directly on the VM via `docker compose build` | User preference — no registry needed |
| Production smoke test | Manual by user | User preference |

## 6. Open Questions / Risks

| Question / Risk | Owner | Due Date |
|-----------------|-------|----------|
| Domain names for staging and production (`*.s.*` / `*.p.*`) — deferred; will decide later | You | Before SSL setup |
| Loki endpoint URL and credentials — placeholder config for now; user will provide later | You | Before Loki goes live |

## 7. Acceptance Criteria

- [ ] Backend integration tests pass (`cd backend && go test ./...`)
- [ ] Dashboard builds and type-checks without errors (`cd dashboard && npm run build`)
- [ ] Desktop app builds for Linux (staging and production versions)
- [ ] Staging VM is provisioned and accessible; staging stack runs with SSL
- [ ] Production VM is provisioned and accessible; production stack runs with SSL
- [ ] `GET /health/ready` returns OK on both environments
- [ ] Daily DB backup runs to local disk; 90-day rotation works; dashboard backup page shows status
- [ ] Logs are forwarded to Loki (placeholder URL configured; promtail running)
- [ ] Manual smoke test passes: full check-in → check-out → payment → receipt flow via desktop, verified in dashboard
- [ ] Deployment runbook is complete and accurate

## 8. Definition of Done

- All three projects (backend, dashboard, desktop) can be built and deployed to staging and production environments.
- Staging environment is running and testable.
- Production environment is running with SSL, daily local backups, backup status page, and log shipping configured.
- Manual smoke test has been performed by you and all core flows pass.
- Deployment runbook is documented and committed to the repository.
- Loki integration is scaffolded (placeholder config) — user will wire the actual endpoint later.
