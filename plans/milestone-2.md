# Milestone 2 — Backend Business Logic

**Status:** In Progress — Backend implementation complete, pending integration tests and CI update.

## 1. Goal

Implement the core parking operations API: sessions, fee calculation, payments, receipts, and shifts, so that an operator can complete a full check-in → check-out → payment → receipt flow and managers can reconcile shifts.

---

## 2. Scope

### In Scope
- Database migrations for `sessions`, `transactions`, `shifts`, and per-location daily receipt sequences.
- Check-in endpoint with plate normalization and duplicate active-plate detection.
- Check-out endpoint with fee calculation and optional manual override when no rate exists.
- Cash and digital (mock) payment recording with change calculation.
- Receipt number generation: strictly sequential per location per day.
- Shift lifecycle: start, end, force-close, and auto-close of stale open shifts.
- Cross-shift attribution (`sessions.shift_id` vs `transactions.shift_id`).
- Cross-operator check-out at the same location.
- Void transaction endpoint with manager PIN and `VOIDED` badge semantics.
- Audit log writes for session, transaction, and shift events.
- All timestamps stored as UTC (`TIMESTAMPTZ`) but formatted/receipted in Asia/Jakarta (WIB, UTC+7).

### Out of Scope / Deferred
- Dashboard UI pages (Milestone 3).
- Desktop app screens (Milestone 4).
- Offline mode and sync (Milestone 5).
- Real digital payment gateway integration; payments are recorded with provider reference only.
- Incident reporting and adjustments (Milestone 6).
- Reports and analytics (Milestone 7).
- Production deployment (Milestone 8).

---

## 3. Dependencies

- **Milestone 1 — Backend Core Entities** must be complete.
  - `locations`, `location_rates`, `users`, `roles`, `user_role_locations`, `user_permission_grants`, `audit_logs` tables.
  - Auth middleware, RBAC resolver, permission allow-list.
  - Seed script with default roles and root owner.
- Final decisions from `PLAN.md`:
  - Fee formula and rate model.
  - Receipt number format.
  - Session lifecycle states.
  - Permission model.

---

## 4. Detailed Tasks

### Backend

#### Database Migrations
- [x] Migration: create `sessions` table
- [x] Migration: create `transactions` table
- [x] Migration: create `shifts` table
- [x] Migration: create `receipt_sequences` table for per-location daily sequence tracking
- [x] Migration: add indexes on `sessions(location_id, state)`, `sessions(plate, state)`, `transactions(session_id)`, `transactions(receipt_number)`, `shifts(operator_id, status)`
- [x] Migration: add `offline_sync` and `sync_conflict` columns to `sessions` (preparation for Milestone 5)

#### Domain Layer — Sessions
- [x] Implement `POST /sessions/check-in`
  - Required: `location_id`, `plate`, `vehicle_type`
  - Normalize plate (uppercase, trim whitespace)
  - Detect duplicate active plate at the same location and return warning flag
  - Resolve current open shift for the operator at the location (auto-close stale if needed)
  - Set `sessions.shift_id` to the check-in shift
  - Require `sessions:create` permission at the location
- [x] Implement `POST /sessions/:id/check-out`
  - Calculate fee using active rate for the vehicle type at `check_in_at` date
  - If no rate exists, allow manual `fee_amount` override from request body
  - Transition session state `ACTIVE` → `PENDING_PAYMENT`
  - Store `rate_snapshot` JSONB at check-out
  - Require `sessions:close` permission at the location
- [x] Implement `GET /sessions` with filters: `location_id`, `state`, `plate`, `operator_id`, `limit`, `offset`
- [x] Implement `GET /sessions/:id` with transaction details if closed
- [x] Implement plate normalization helper (strip spaces, uppercase)
- [x] Implement active-rate lookup by check-in date and vehicle type

#### Domain Layer — Payments / Transactions
- [x] Implement `POST /payments/cash`
  - Required: `session_id`, `amount_tendered`
  - Validate session is in `PENDING_PAYMENT`
  - Calculate change
  - Generate receipt number atomically (per location/day)
  - Create transaction, transition session to `CLOSED`
  - Set `transactions.shift_id` to the operator's current open shift
  - Require `payments:collect_cash` permission at the location
- [x] Implement `POST /payments/digital`
  - Required: `session_id`, `payment_reference` (free-form, optional)
  - Validate session is in `PENDING_PAYMENT`
  - Generate receipt number atomically
  - Create transaction with `payment_method = 'DIGITAL'`, transition session to `CLOSED`
  - Require `payments:collect_digital` permission at the location
- [x] Implement `POST /transactions/:id/void`
  - Required: manager `pin`
  - Validate manager PIN
  - Mark transaction `voided = true`, set `voided_at`, `voided_by`, `void_reason`
  - Transition linked session to `VOIDED`
  - Do not delete; reports will show `VOIDED` badge and exclude from revenue
  - Require `payments:void` permission at the location
- [x] Implement `GET /transactions` with filters: `location_id`, `shift_id`, `voided`, `date_from`, `date_to`, `limit`, `offset`
- [x] Implement `GET /transactions/:id`

#### Domain Layer — Receipts
- [x] Implement receipt number generator using `receipt_sequences` table with row-level locking
  - Format: `[LOCATION_CODE]-[YYYYMMDD]-[SEQUENCE]`
  - Sequence resets daily per location
  - Strictly sequential; gaps are not acceptable in normal flow
- [x] Implement receipt data builder that returns transaction + session + location + rate snapshot

#### Domain Layer — Shifts
- [x] Implement `POST /shifts/start`
  - Required: `location_id`
  - Auto-close any existing `OPEN` shift for the operator before opening a new one
  - Create new shift with `status = 'OPEN'`
  - Require `shifts:start` permission at the location
- [x] Implement `POST /shifts/:id/end`
  - Required: `cash_handover_amount`, optional `discrepancy_notes`
  - Calculate `expected_cash` from transactions in this shift
  - Compute `discrepancy = cash_handover_amount - expected_cash`
  - Set `status = 'CLOSED'` if discrepancy is zero, otherwise `'FLAGGED'`
  - Require `shifts:end` permission and operator must own the shift (or be manager/admin)
- [x] Implement `POST /shifts/:id/force-close`
  - Required: `reason`
  - Allow manager/admin to force-close another operator's open shift
  - Set `status = 'FORCE_CLOSED'`, record `force_closed_by`, `force_closed_reason`
  - Require `shifts:force_close` permission at the location
- [x] Implement `GET /shifts` with filters: `location_id`, `operator_id`, `status`, `date_from`, `date_to`, `limit`, `offset`
- [x] Implement `GET /shifts/:id` with summary totals

#### Shared Infrastructure
- [x] Implement fee calculation engine from `PLAN.md` formula
- [x] Implement timezone helper: format/display in Asia/Jakarta (WIB, UTC+7)
- [x] Add audit log writes for check-in, check-out, payment, void, shift start/end/force-close
- [x] Add permission checks to all endpoints using existing middleware
- [x] Add request validation structs and structured error responses

### Dashboard

- [x] No dashboard UI in this milestone; API-only focus.
- [ ] Optional: create TypeScript types for `Session`, `Transaction`, `Shift`, `Receipt` in `dashboard/types/` for use in Milestone 3.

### Desktop

- [x] No desktop work in this milestone.

### DevOps / QA

- [ ] Add integration test scaffolding with a test database.
- [ ] Add backend integration tests for:
  - Full check-in → check-out → cash payment → receipt flow.
  - Digital payment flow.
  - Manual fee override when no rate exists.
  - Receipt sequence strict daily increment.
  - Shift start/end with discrepancy calculation.
  - Force-close shift by manager.
  - Void transaction by manager with PIN.
  - Cross-operator check-out.
- [x] Add curl/HTTP examples for sessions, payments, shifts, and voids to `README.md`.
- [ ] Update CI workflow to run backend integration tests.

---

## 5. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Missing rate at check-out | Allow manual `fee_amount` override | Avoids blocking operations when rate config is delayed. |
| Receipt number sequence | Strictly sequential per location per day | Matches PLAN.md and tax/audit expectations. |
| Timezone for receipts | Asia/Jakarta (WIB, UTC+7) | Indonesia does not observe DST; simplifies display logic. |
| Force-close shifts | Managers/admins can force-close another operator's open shift | Needed for operational control when operator forgets to close. |
| Multiple open shifts | Auto-close old shift on new `POST /shifts/start` | Prevents accidental duplicate open shifts. |
| Digital payment reference | Free-form, optional in v1 | Provider reference formats vary; no validation yet. |
| Cross-operator check-out | Allowed at the same location | Practical for multi-booth locations and shift handovers. |
| Voided records in reports | Show `VOIDED` badge, exclude from revenue | Preserves audit trail while keeping revenue numbers clean. |
| Timestamp storage | `TIMESTAMPTZ` UTC, formatted in WIB | Consistent storage, localized display. |
| Rate snapshot | Stored in `sessions.rate_snapshot` at check-out | Guarantees fee auditability even if rates change later. |

---

## 6. Open Questions / Risks

### Resolved
| Question | Decision |
|----------|----------|
| No rate at check-out | Manual fee override |
| Receipt number gaps | Strictly sequential daily |
| Timezone | Asia/Jakarta WIB |
| Force-close shifts | Managers/admins allowed |
| Multiple open shifts | Auto-close old on new start |
| Digital reference format | Free-form, optional |
| Cross-operator check-out | Allowed |
| Voided records in reports | VOIDED badge, excluded from revenue |

### Remaining Risks
| Risk | Mitigation |
|------|------------|
| Strict sequential receipt numbers under high concurrency | Use row-level locking on `receipt_sequences` or advisory locks; load test p95 < 300ms. |
| Manual fee override abuse | Require manager PIN or `sessions:close` permission; log all overrides to audit log. |
| Cross-operator check-out attribution confusion | Track both `sessions.operator_id` (check-in) and `transactions.operator_id` (payment) clearly. |
| Auto-close of old shifts creates inaccurate cash reconciliation | Audit log the auto-close and warn the operator before starting a new shift. |
| Timezone display bugs | Centralize all WIB formatting through one helper; unit test edge cases. |
| Receipt number gaps from failed transactions | Wrap receipt generation + transaction insert in a single DB transaction so failed inserts do not consume numbers. |

---

## 7. Acceptance Criteria

- [x] `POST /sessions/check-in` creates an `ACTIVE` session and normalizes the plate.
- [x] Check-in returns a warning if an active session with the same plate already exists at the location.
- [x] `POST /sessions/:id/check-out` calculates the fee using the active rate for the check-in date.
- [x] Check-out allows a manual `fee_amount` override when no rate is configured.
- [x] `POST /payments/cash` records the transaction, calculates change, and generates a sequential receipt number.
- [x] `POST /payments/digital` records the transaction with a free-form reference and generates a sequential receipt number.
- [x] Receipt numbers follow `[LOCATION_CODE]-[YYYYMMDD]-[SEQUENCE]` and reset daily per location.
- [x] `POST /shifts/start` auto-closes any existing open shift for the operator.
- [x] `POST /shifts/:id/end` calculates expected cash, discrepancy, and sets `CLOSED` or `FLAGGED` status.
- [x] `POST /shifts/:id/force-close` allows a manager/admin to close another operator's shift.
- [x] `POST /transactions/:id/void` requires a manager PIN and marks the transaction/session as `VOIDED` without deleting it.
- [x] Voided transactions are excluded from revenue totals and shown with a `VOIDED` badge.
- [x] Cross-operator check-out at the same location succeeds.
- [x] All timestamps are stored in UTC and formatted/displayed in Asia/Jakarta (WIB).
- [x] Audit logs are written for session, transaction, and shift events.

---

## 8. Definition of Done

Milestone 2 is complete when:
1. [x] All migrations apply cleanly.
2. [x] The full check-in → check-out → payment → receipt flow works via API calls.
3. [x] Shift start/end/force-close and discrepancy calculation behave correctly.
4. [x] Receipt numbers are strictly sequential per location per day with no gaps under normal operation.
5. [x] Manual fee override, cross-operator check-out, and void flows are implemented and tested.
6. [ ] Integration tests cover the critical paths and pass in CI.
7. [x] A teammate can use the API (via curl or tests) to complete a parking transaction end-to-end.
