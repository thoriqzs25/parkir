# Milestone 5 — Offline Mode & Sync

## 1. Goal

Make the desktop app resilient to connectivity loss so operators can work offline for up to one full shift (24 hours) and automatically sync records when reconnected, with conflicts surfaced to managers for resolution.

## 2. Scope

### In Scope

- Connectivity detection and offline/online state transitions in the desktop app
- Local storage for sessions, transactions, and incidents (up to ~1 day of data)
- Local rate cache with 24h TTL and silent fallback to last known rates
- Offline check-in, check-out, payment, and receipt printing flows
- Sync queue processed in the order operations were captured locally
- Backend batch sync endpoint that accepts offline records
- Sync conflict detection for duplicate active plates
- Dashboard sync conflict resolution UI
- Auto-sync on reconnect
- Automatic reprinting of official receipts after successful sync
- Operator-facing offline/sync status indicator

### Out of Scope / Deferred
- Offline incident filing and sync (incidents table is part of Milestone 6)
- True multi-device real-time collaboration
- Conflict resolution beyond duplicate active plates
- Offline shift start / force-close (shift is assumed started online)
- Offline user/role/rate management
- Emergency offline mode beyond 24 hours
- Local analytics or reports while offline
- Automatic local clock correction

## 3. Dependencies

- **Milestone 4 — Desktop App Online Mode** must be complete or nearly complete, since offline mode builds on the same screens and flows.
- **Milestone 2 — Backend Business Logic** must expose stable endpoints for sessions, payments, receipts, and shifts.
- **Milestone 3 — Web Dashboard Foundation** provides the layout, auth, and table components needed for the conflict UI.
- Stable per-location daily receipt sequence from Milestone 2.

## 4. Detailed Tasks

### Backend

- [ ] Add `offline_sync` and `sync_conflict` handling to existing sessions/transactions logic
- [ ] Create `POST /api/v1/sync/batch` endpoint to receive queued offline records
- [ ] Implement ordered sync processing by `check_in_at`
- [ ] Detect duplicate active plates during sync and mark `sync_conflict = true`
- [ ] Preserve offline record attribution: operator, shift, location, timestamps
- [ ] Assign official receipt numbers to offline transactions during sync
- [ ] Create API to list sync conflicts with filters and pagination
- [ ] Create endpoint for managers to resolve a sync conflict (void/reassign)
- [ ] Add offline incident review endpoint: submitted records remain pending until approved
- [ ] Write audit log entries for sync events and conflict resolutions
- [ ] Add integration tests for sync endpoint, conflict detection, and incident review

### Dashboard

- [ ] Add Sync Conflicts page to navigation
- [ ] Sync conflicts list page with status filters and location selector
- [ ] Sync conflict detail page showing both local/offline and server records
- [ ] Conflict resolution actions: void local record, void server record, or merge
- [ ] Sync status page/indicator showing pending count and last sync time
- [ ] Unresolved-conflict notification badge
- [ ] Unit/component tests for conflict resolution UI

### Desktop

- [ ] Implement connectivity heartbeat polling every 30 seconds
- [ ] Build local storage layer using Electron-safe storage (SQLite via better-sqlite3 or IndexedDB)
- [ ] Cache active rates locally with 24h TTL
- [ ] Queue sessions and transactions locally when offline
- [ ] Generate temporary offline receipt numbers (`[CODE]-OFFLINE-[SEQ]`)
- [ ] Queue offline incidents separately and mark them for manager review
- [ ] Implement sync queue ordered by `check_in_at`
- [ ] Auto-sync on reconnect with retry and operator-facing status
- [ ] Reprint official receipts after successful sync
- [ ] Offline/sync status indicator in the app chrome
- [ ] Enforce ~1-day local storage cap by pruning oldest completed records
- [ ] Smoke tests for offline check-in → payment → reconnect → sync flow

### DevOps / QA

- [ ] End-to-end test: offline → online sync
- [ ] Conflict scenario test: same plate checked in offline then online before sync
- [ ] Auto-reprint after sync test
- [ ] Local storage cap behavior test
- [ ] Dashboard conflict resolution end-to-end test
- [ ] Integration test for batch sync partial-failure handling
- [ ] Verify desktop build still passes with added local storage dependency

## 5. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Max offline duration | Up to one full shift (~24 hours) | Matches the 1-day local storage cap and shift-based operations |
| Stale rate cache behavior | Silent fallback to last known rates | `PLAN.md` already specifies a 24h TTL; avoids blocking operations in low-connectivity sites |
| Local clock drift | Use local timestamps; fix via manual adjustment if needed | Keeps implementation simple; manager can correct bad records later |
| Official receipts after sync | Auto-reprint with the new receipt number | Confirms the correct receipt to the customer automatically |
| Duplicate active plate conflict | Online check-in proceeds; offline record is flagged `sync_conflict` | Server cannot know about un-synced records; conflicts are resolved by managers |
| Sync trigger | Auto-sync on reconnect | Required by `PLAN.md` target of syncing within 60 seconds of reconnect |
| Offline incidents | Pause for manager review before they become active | Operational records need oversight; prevents bad incident data from entering the system |
| Local storage cap | ~1 day of sessions/transactions/incidents | Keeps storage bounded and aligned with the max offline window |

## 6. Open Questions / Risks

| Question / Risk | Owner | Due Date |
|-----------------|-------|----------|
| What is the actual local storage size for one day at the busiest location (120 vehicles/hour)? | Desktop lead | Before storage layer is implemented |
| How should the app behave if the local rate cache is stale and the operator still cannot go online? | Product | Before rate cache implementation |
| Should unresolved sync conflicts block the operator from closing their shift? | Product + Backend lead | Before conflict UI is designed |
| What happens to the printer queue if the printer is unavailable during auto-reprint after sync? | Desktop lead | Before reprint feature is merged |
| How do we handle a partially failed sync batch (some records accepted, some rejected)? | Backend lead | Before sync endpoint is released |
| Electron SQLite dependency may complicate cross-platform builds | DevOps | Before desktop build is finalized |

## 7. Acceptance Criteria

- [ ] Operator can perform check-in, check-out, payment, and receipt print while fully offline
- [ ] Temporary offline receipts use the `[CODE]-OFFLINE-[SEQ]` format
- [ ] On reconnect, all queued records auto-sync within 60 seconds
- [ ] Official receipts are generated after sync and automatically reprinted
- [ ] Duplicate active plates result in `sync_conflict = true` and appear in the dashboard
- [ ] Manager can view and resolve sync conflicts from the dashboard
- [ ] Offline incidents remain pending until a manager reviews and approves them
- [ ] Local storage does not grow unbounded and respects the 1-day cap
- [ ] Backend sync endpoint integration tests pass
- [ ] Dashboard sync conflict page builds and type-checks successfully
- [ ] Desktop offline smoke test passes

## 8. Definition of Done

- The desktop app supports the full offline check-in → check-out → payment → receipt flow.
- Reconnecting the device triggers automatic sync and conflict detection.
- Managers can view and resolve sync conflicts in the dashboard.
- Backend sync paths and conflict scenarios are covered by integration tests.
- Desktop smoke tests pass for offline mode and reconnect flows.
- All code is reviewed and merged to `main`.
- `PLAN.md` is updated if any offline/sync behavior decisions diverged from the original plan.
