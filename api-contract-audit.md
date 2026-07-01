# API Contract Audit: Dashboard ↔ Backend

**Date:** 2026-07-01
**Scope:** All 27 dashboard pages ↔ 69 backend endpoints

---

## Legend

| Icon | Meaning |
|------|---------|
| ✅ | Aligned — contract matches |
| ⚠️ | Warning — minor type/status mismatch, runtime may still work |
| ❌ | Misaligned — broken contract, will cause runtime errors |
| 🔍 | Needs verification — unclear from code, requires runtime check |

---

## Envelope Structure

**Backend standard** (`response.go:18`): `{ data, error, meta }`

**Dashboard API client** (`lib/api.ts:60`): `return body?.data as T` — unwraps outer `data` field.

### Paginated list contract

Backend **should** return:
```json
{
  "data": { "items": [...], "meta": { "limit": 20, "offset": 0, "total": 100 } }
}
```

Dashboard expects `PaginatedItems<T>` = `{ items: T[]; meta: ApiMeta }`.

---

## 1. Health `/health` + `/health/components`

**Page:** `app/(dashboard)/[locationId]/health/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 1 | GET | `/health` | `{ status: string }` via envelope unwrap | **FIXED** — now uses `response.OK()` → `{ data: { status: "ok" } }` | ✅ | Fixed 2026-07-01: replaced raw `c.JSON()` with `response.OK()`. |
| 2 | GET | `/health/components` | `HealthComponents` via envelope unwrap | **FIXED** — now uses `response.OK()` → `{ data: { status, components, last_check } }` | ✅ | Fixed 2026-07-01: replaced raw `c.JSON()` with `response.OK()`. |

---

## 2. Auth

**Page:** `app/login/page.tsx`
**Hooks:** `hooks/useAuth.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 3 | POST | `/api/v1/auth/login` | `{ user: User; token: string }` | `LoginResponse { user, token }` via `response.Created()` | ✅ | Envelope unwrap gives `{ user, token }`. |
| 4 | POST | `/api/v1/auth/logout` | `{ message: string }` | `{ message: "logged out" }` via `response.OK()` | ✅ | |
| 5 | POST | `/api/v1/auth/refresh` | `{ token: string }` | `{ token: string }` via `response.OK()` | ✅ | |
| 6 | GET | `/api/v1/auth/me` | `MeResponse { user: User; permissions: string[] }` | `MeResponse { user, permissions }` via `response.OK()` | ✅ | |

**Login request body:** Dashboard `LoginInput` = `{ email, password }` ↔ Backend `LoginRequest` = `{ email, password }`. ✅

---

## 3. Users

**Page:** `app/(dashboard)/[locationId]/users/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 7 | GET | `/api/v1/users` | `PaginatedItems<User>` → `.items[]` | **FIXED** — now uses `gin.H{"items": users, "meta": ...}` → `{ data: { items: [...], meta: {...} } }` | ✅ | Fixed 2026-07-01: changed from double-wrapped `response.Response{}` to `gin.H{}` pattern. |
| 8 | GET | `/api/v1/users/:id` | `User` | `response.OK(c, user)` → `{ data: user }` | ✅ | Not actively called by any page. |
| 9 | POST | `/api/v1/users` | `User` | `response.Created(c, user)` → `{ data: user }` | ✅ | Request body: `{ name, email, password, role_id, location_ids? }` ✅ |
| 10 | PATCH | `/api/v1/users/:id` | `User` | `response.OK(c, user)` → `{ data: user }` | ✅ | Request body: `{ name?, email?, role_id?, location_ids?, status? }` ✅ |
| 11 | POST | `/api/v1/users/:id/deactivate` | `void` | `response.NoContent(c)` → 204 + no body | ⚠️ | Works at runtime but `request()` catches JSON parse error silently. Consider returning `204` explicitly. |
| 12 | POST | `/api/v1/users/:id/reset-password` | `void` | `response.OK(c, gin.H{"message": "..."})` → `{ data: { message: "..." } }` | ⚠️ | Typed as `void` but response has body. Caller ignores return value, so runtime is fine. |
| 13 | POST | `/api/v1/users/:id/reset-pin` | `void` | `response.OK(c, gin.H{"message": "..."})` → `{ data: { message: "..." } }` | ⚠️ | Same as reset-password. |

**Request body alignment:**
| Field | Dashboard `CreateUserInput` | Backend `CreateUserRequest` | Match |
|-------|---------------------------|---------------------------|-------|
| `name` | `string` | `string` (required) | ✅ |
| `email` | `string` | `string` (required,email) | ✅ |
| `password` | `string` | `string` (required,min=8) | ✅ |
| `role_id` | `string` | `string` (required,uuid) | ✅ |
| `location_ids` | `string[]` optional | `[]string` optional | ✅ |

| Field | Dashboard `UpdateUserInput` | Backend `UpdateUserRequest` | Match |
|-------|---------------------------|---------------------------|-------|
| `name` | `string?` | `*string` | ✅ |
| `email` | `string?` | `*string` | ✅ |
| `role_id` | `string?` | `*string` | ✅ |
| `location_ids` | `string[]?` | `[]string` | ✅ |
| `status` | `"ACTIVE"\|"DEACTIVATED"?` | `*string` | ✅ |

**Dashboard `User` type** ↔ **Backend `store.User` JSON:**
| Field | Dashboard | Backend | Match |
|-------|-----------|---------|-------|
| `id` | `string` | `string` | ✅ |
| `name` | `string` | `string` | ✅ |
| `email` | `string` | `string` | ✅ |
| `role_id` | `string` | `string` | ✅ |
| `role_name` | `string?` | `string,omitempty` | ✅ |
| `status` | `"ACTIVE"\|"DEACTIVATED"` | `string` | ⚠️ Backend uses plain string, no enum validation |
| `location_ids` | `string[]?` | `[]string,omitempty` | ✅ |
| `created_at` | `string` | `time.Time` → RFC3339 | ✅ |
| `updated_at` | `string` | `time.Time` → RFC3339 | ✅ |

---

## 4. Roles

**Page:** `app/(dashboard)/[locationId]/roles/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 14 | GET | `/api/v1/roles` | `PaginatedItems<Role>` → `.items[]` | **FIXED** — now returns `{ data: { items: [...], meta: {...} } }` | ✅ | Fixed 2026-07-01: wrapped flat array in `{ items, meta }` envelope. |
| 15 | GET | `/api/v1/roles/:id` | `Role` | `response.OK(c, role)` → `{ data: role }` | ✅ | Not actively called by any page. |
| 16 | POST | `/api/v1/roles` | `Role` | `response.Created(c, role)` → `{ data: role }` | ✅ | Request body: `{ name, permissions }` ✅ |
| 17 | PATCH | `/api/v1/roles/:id` | `Role` | `response.OK(c, role)` → `{ data: role }` | ✅ | Request body: `{ name?, permissions? }` ✅ |
| 18 | DELETE | `/api/v1/roles/:id` | `void` | `response.NoContent(c)` → 204 | ⚠️ | Same JSON-parse-on-empty-body caveat as users/deactivate. |

**Dashboard `Role` type** ↔ **Backend `store.Role` JSON:**
| Field | Dashboard | Backend | Match |
|-------|-----------|---------|-------|
| `id` | `string` | `string` | ✅ |
| `name` | `string` | `string` | ✅ |
| `permissions` | `string[]` | `[]string` | ✅ |
| `created_at` | `string` | `time.Time` → RFC3339 | ✅ |
| `updated_at` | `string` | `time.Time` → RFC3339 | ✅ |
| `deleted_at` | (not in type) | `*time.Time,omitempty` | ⚠️ Dashboard ignores soft-delete field (OK) |

---

## 5. Locations

**Pages:** `app/(dashboard)/[locationId]/locations/page.tsx`, `(dashboard)/select/page.tsx`, `components/layout/LocationProvider.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 19 | GET | `/api/v1/locations` | `PaginatedItems<Location>` → `.items[]` | `response.OK(c, gin.H{"items": locs, "meta": ...})` → `{ data: { items: [...], meta: {...} } }` | ✅ | Correct `items` key. |
| 20 | GET | `/api/v1/locations/:id` | `Location` | `response.OK(c, loc)` → `{ data: loc }` | ✅ | |
| 21 | POST | `/api/v1/locations` | `Location` | `response.Created(c, loc)` → `{ data: loc }` | ✅ | Request: `{ name, code, address?, city?, capacity? }` ✅ |
| 22 | PATCH | `/api/v1/locations/:id` | `Location` | `response.OK(c, loc)` → `{ data: loc }` | ✅ | Request: `{ name?, address?, city?, status?, capacity? }` ✅ |
| 23 | POST | `/api/v1/locations/:id/deactivate` | `Location` | `response.OK(c, loc)` → `{ data: loc }` | ✅ | |
| 24 | POST | `/api/v1/locations/:id/assign-operator` | `void` | `response.NoContent(c)` → 204 | ⚠️ | Same JSON-parse caveat. Request: `{ user_id }` ✅ |
| 25 | POST | `/api/v1/locations/:id/remove-operator` | `void` | `response.NoContent(c)` → 204 | ⚠️ | Same. Request: `{ user_id }` ✅ |

**Note:** Backend locations List handler (line 95-98) uses `Limit: total` instead of a true `limit` query param. `ListLocations()` doesn't accept pagination params.

**Dashboard `Location` type** ↔ **Backend `store.Location` JSON:**
| Field | Dashboard | Backend | Match |
|-------|-----------|---------|-------|
| `id` | `string` | `string` | ✅ |
| `name` | `string` | `string` | ✅ |
| `code` | `string` | `string` | ✅ |
| `address` | `string?` | `string,omitempty` | ✅ |
| `city` | `string?` | `string,omitempty` | ✅ |
| `status` | `"ACTIVE"\|"INACTIVE"` | `string` | ⚠️ Backend plain string, no enum |
| `capacity` | `Record<string, number>?` | `map[string]interface{},omitempty` | ⚠️ Backend uses `interface{}` values, not `number` |
| `created_at` | `string` | `time.Time` → RFC3339 | ✅ |
| `updated_at` | `string` | `time.Time` → RFC3339 | ✅ |

---

## 6. Rates

**Page:** `app/(dashboard)/[locationId]/rates/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 26 | GET | `/api/v1/locations/:locationId/rates` | `Rate[]` | `response.OK(c, rates)` → `{ data: [...rates] }` | ✅ | |
| 27 | POST | `/api/v1/locations/:locationId/rates` | `Rate` | `response.Created(c, rate)` → `{ data: rate }` | ✅ | Request: `{ vehicle_type, first_hour_rate, subsequent_hourly_rate, daily_flat_rate, effective_from, effective_until? }` ✅ |
| 28 | PATCH | `/api/v1/rates/:rateId` | `Rate` | `response.OK(c, rate)` → `{ data: rate }` | ✅ | Request: `{ first_hour_rate?, subsequent_hourly_rate?, daily_flat_rate?, effective_until? }` ✅ |

**Dashboard `Rate` type** ↔ **Backend `store.Rate` JSON:**
| Field | Dashboard | Backend | Match |
|-------|-----------|---------|-------|
| `id` | `string` | `string` | ✅ |
| `location_id` | `string` | `string` | ✅ |
| `vehicle_type` | `"CAR"\|"MOTO"\|"TRUCK"` | `string` | ⚠️ Backend has `oneof` validation but no TS enum |
| `first_hour_rate` | `number` | `float64` | ✅ |
| `subsequent_hourly_rate` | `number` | `float64` | ✅ |
| `daily_flat_rate` | `number` | `float64` | ✅ |
| `effective_from` | `string` | `time.Time` → RFC3339 | ✅ |
| `effective_until` | `string?` | `*time.Time,omitempty` | ✅ |
| `created_by` | `string?` | `*string,omitempty` | ✅ |
| `created_at` | `string` | `time.Time` → RFC3339 | ✅ |

---

## 7. Sessions

**Pages:** `app/(dashboard)/[locationId]/sessions/active/page.tsx`, `sessions/history/page.tsx`, `sessions/[id]/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 29 | GET | `/api/v1/sessions` | `PaginatedItems<Session>` → `.items[]` | `response.OK(c, gin.H{"items": ..., "meta": ...})` → `{ data: { items: [...], meta: {...} } }` | ✅ | Query params: `location_id`, `state`, `plate`, `operator_id`, `limit`, `offset` ✅ |
| 30 | GET | `/api/v1/sessions/:id?include=transaction` | `{ session: Session; transaction?: Transaction }` | `response.OK(c, gin.H{"session": ..., "transaction": tx})` → `{ data: { session, transaction } }` | ✅ | Backend only returns transaction if `include=transaction` query param present. |

---

## 8. Transactions

**Page:** `app/(dashboard)/[locationId]/transactions/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 31 | GET | `/api/v1/transactions` | `PaginatedItems<Transaction>` → `.items[]` | `response.OK(c, gin.H{"items": ..., "meta": ...})` → `{ data: { items: [...], meta: {...} } }` | ✅ | Query: `location_id`, `shift_id`, `voided`, `date_from`, `date_to`, `limit`, `offset` |
| 32 | GET | `/api/v1/transactions/:id` | `Transaction` | `response.OK(c, tx)` → `{ data: tx }` | ✅ | Not actively called by any page. |
| 33 | POST | `/api/v1/transactions/:id/void` | `Transaction` | `response.OK(c, tx)` → `{ data: tx }` | ✅ | Request: `{ manager_pin, void_reason }` ✅. Not actively called by any page. |

---

## 9. Shifts

**Pages:** `app/(dashboard)/[locationId]/shifts/page.tsx`, `shifts/[id]/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 34 | GET | `/api/v1/shifts` | `PaginatedItems<Shift>` → `.items[]` | `response.OK(c, gin.H{"items": ..., "meta": ...})` → `{ data: { items: [...], meta: {...} } }` | ✅ | Query: `location_id`, `operator_id`, `status`, `date_from`, `date_to` ✅ |
| 35 | GET | `/api/v1/shifts/:id?include=transactions` | `{ shift: Shift; transactions?: Transaction[]; summary?: { transaction_count, expected_cash } }` | `response.OK(c, gin.H{"shift": ..., "transactions": ..., "summary": {...}})` → `{ data: { shift, transactions, summary } }` | ✅ | Backend returns `summary.transaction_count` (int) and `summary.expected_cash` (float64) ✅ |

---

## 10. Sync Conflicts

**Page:** `app/(dashboard)/[locationId]/sync-conflicts/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 36 | GET | `/api/v1/sync/conflicts` | `PaginatedItems<Session>` → `.items[]` | `response.OK(c, gin.H{"items": ..., "meta": ...})` → `{ data: { items: [...], meta: {...} } }` | ✅ | Query: `location_id`, `limit`, `offset` |
| 37 | POST | `/api/v1/sync/conflicts/:id/resolve` | `Session` | `response.OK(c, session)` → `{ data: session }` | ✅ | Request: `{ action: "VOID_OFFLINE"\|"IGNORE", void_reason? }` ✅ |

---

## 11. Incidents

**Pages:** `app/(dashboard)/[locationId]/incidents/page.tsx`, `incidents/[id]/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 38 | GET | `/api/v1/incidents` | `PaginatedItems<Incident>` → `.items[]` | `response.OK(c, gin.H{"items": ..., "meta": ...})` → `{ data: { items: [...], meta: {...} } }` | ✅ | Query: `location_id`, `type`, `state`, `date_from`, `date_to` |
| 39 | GET | `/api/v1/incidents/:id` | `Incident` | `response.OK(c, inc)` → `{ data: inc }` | ✅ | |
| 40 | POST | `/api/v1/incidents` | `Incident` | `response.Created(c, inc)` → `{ data: inc }` | ✅ | Request: `{ location_id, type, session_id?, description }` ✅. Not actively called by any page. |
| 41 | PATCH | `/api/v1/incidents/:id/resolve` | `Incident` | `response.OK(c, inc)` → `{ data: inc }` | ✅ | Request: `{ resolution_notes, adjustment_action?, adjustment_entity_id?, manager_pin? }` ✅ |
| 42 | GET | `/api/v1/incidents/:id/notes` | `IncidentNote[]` | `response.OK(c, notes)` → `{ data: [...notes] }` | ✅ | |
| 43 | POST | `/api/v1/incidents/:id/notes` | `IncidentNote` | `response.Created(c, note)` → `{ data: note }` | ✅ | Request: `{ note }` ✅ |

---

## 12. Adjustments

**Defined in** `lib/api.ts:356-368` but **not actively called by any page** (available for incident resolution flow).

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 44 | POST | `/api/v1/adjustments/void-transaction` | `Transaction` | `response.OK(c, tx)` → `{ data: tx }` | ✅ | Request: `{ transaction_id, reason, manager_pin }` ✅ |
| 45 | POST | `/api/v1/adjustments/reassign-session` | `Session` | `response.OK(c, session)` → `{ data: session }` | ✅ | Request: `{ session_id, new_operator_id, new_shift_id, manager_pin }` ✅ |

---

## 13. Audit Logs

**Page:** `app/(dashboard)/[locationId]/audit-logs/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 46 | GET | `/api/v1/audit-logs` | `PaginatedItems<AuditLog>` → `.items[]` | `response.OK(c, gin.H{"items": ..., "meta": ...})` → `{ data: { items: [...], meta: {...} } }` | ✅ | Query: `action`, `actor_id`, `entity_type`, `entity_id`, `location_id`, `date_from`, `date_to` |
| 47 | GET | `/api/v1/audit-logs/export` | Direct URL to CSV file | CSV file with `Content-Type: text/csv` | ✅ | Dashboard opens URL in new tab. Backend streams CSV directly. |

---

## 14. Alerts

**Page:** `app/(dashboard)/[locationId]/alerts/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 48 | GET | `/api/v1/alerts` | `PaginatedItems<Alert>` → `.items[]` | **FIXED** — now uses `gin.H{"items": ..., "meta": gin.H{...}}` instead of `response.Response{Data: ...}` | ✅ | Fixed 2026-07-01: meta was wrapped in `response.Response{Data: ...}` causing `meta.data.total` instead of `meta.total`. Page breaks pagination. |
| 49 | GET | `/api/v1/alerts/:id` | `Alert` | `response.OK(c, alert)` → `{ data: alert }` | ✅ | Not actively called. |
| 50 | POST | `/api/v1/alerts/:id/acknowledge` | `Alert` | `response.OK(c, alert)` → `{ data: alert }` | ✅ | Request body (optional): `{ resolution_notes? }`. Dashboard sends empty body. ✅ |
| 51 | POST | `/api/v1/alerts/:id/resolve` | `Alert` | `response.OK(c, alert)` → `{ data: alert }` | ✅ | Request: `{ resolution_notes }` ✅ |

---

## 15. Alert Configs

**Page:** `app/(dashboard)/[locationId]/alerts/config/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 52 | GET | `/api/v1/alert-configs?location_id=:id` | `AlertConfig[]` | `response.OK(c, configs)` → `{ data: [...configs] }` | ✅ | |
| 53 | PATCH | `/api/v1/alert-configs/:id` | `AlertConfig` | `response.OK(c, config)` → `{ data: config }` | ✅ | Request: `{ enabled?, threshold? }` ✅ |

---

## 16. Reports

**Pages:** `app/(dashboard)/[locationId]/reports/{daily-revenue,occupancy,vehicle-breakdown,operator-activity}/page.tsx`

All reports use the same pattern.

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 54 | GET | `/api/v1/reports/daily-revenue` | `DailyRevenueRow[]` | `response.OK(c, rows)` → `{ data: [...rows] }` | ✅ | Query: `location_id`, `date_from`, `date_to`, `include_voided` |
| 55 | GET | `/api/v1/reports/daily-revenue?format=csv` | Direct CSV URL | CSV response | ✅ | |
| 56 | GET | `/api/v1/reports/occupancy` | `OccupancyRow[]` | `response.OK(c, rows)` → `{ data: [...rows] }` | ✅ | Query: `location_id`, `granularity`, `date_from`, `date_to` |
| 57 | GET | `/api/v1/reports/occupancy?format=csv` | Direct CSV URL | CSV response | ✅ | |
| 58 | GET | `/api/v1/reports/vehicle-breakdown` | `VehicleBreakdownRow[]` | `response.OK(c, rows)` → `{ data: [...rows] }` | ✅ | Query: `location_id`, `date_from`, `date_to` |
| 59 | GET | `/api/v1/reports/vehicle-breakdown?format=csv` | Direct CSV URL | CSV response | ✅ | |
| 60 | GET | `/api/v1/reports/operator-activity` | `OperatorActivityRow[]` | `response.OK(c, rows)` → `{ data: [...rows] }` | ✅ | Query: `location_id`, `operator_id`, `date_from`, `date_to` |
| 61 | GET | `/api/v1/reports/operator-activity?format=csv` | Direct CSV URL | CSV response | ✅ | |

**Dashboard report types** all match backend `store/report.go` fields exactly. ✅

---

## 17. Backups

**Page:** `app/(dashboard)/[locationId]/backups/page.tsx`

| # | Method | Route | Dashboard Expectation | Backend Reality | Status | Notes |
|---|--------|-------|----------------------|-----------------|--------|-------|
| 62 | GET | `/api/v1/backups` | `BackupListResponse` = `{ items, status, last_run_at?, last_status? }` | `response.OK(c, resp)` → `{ data: { items, status, last_run_at?, last_status? } }` | ✅ | |
| 63 | POST | `/api/v1/backups/run` | `BackupListResponse` | `response.Created(c, resp)` → `{ data: { items, status, last_run_at?, last_status? } }` | ✅ | |

---

## Unused Backend Endpoints (not consumed by dashboard)

These endpoints exist in the backend but are **never called** from dashboard pages (they serve the desktop app or are unused):

| # | Method | Route | Used By | Notes |
|---|--------|-------|---------|-------|
| — | POST | `/api/v1/sessions/check-in` | Desktop app | Check-in flow |
| — | POST | `/api/v1/sessions/:id/check-out` | Desktop app | Check-out flow |
| — | POST | `/api/v1/payments/cash` | Desktop app | Cash payment |
| — | POST | `/api/v1/payments/digital` | Desktop app | Digital payment |
| — | GET | `/api/v1/shifts/me/open` | Desktop app | Open shift check |
| — | POST | `/api/v1/shifts/start` | Desktop app | Start shift |
| — | POST | `/api/v1/shifts/:id/end` | Desktop app | End shift |
| — | POST | `/api/v1/shifts/:id/force-close` | Desktop app | Force close |
| — | POST | `/api/v1/sync/batch` | Desktop app | Offline sync |
| — | GET | `/health/ready` | — | Health check endpoint not used by dashboard |

---

## Summary

### ❌ Broken (will cause runtime errors)

*All previously broken endpoints have been fixed. See fixes below.*

| # | Endpoint | Issue | Status |
|---|----------|-------|--------|
| 1 | `GET /api/v1/users` | Double-wrapped envelope → `items` field absent | ✅ **Fixed** — uses `gin.H{"items": ...}` |
| 2 | `GET /api/v1/roles` | Flat array instead of `{ items, meta }` | ✅ **Fixed** — wrapped in `{ items, meta }` |
| 3 | `GET /health` | Raw `c.JSON()` → no envelope | ✅ **Fixed** — uses `response.OK()` |
| 4 | `GET /health/components` | Raw `c.JSON()` → no envelope | ✅ **Fixed** — uses `response.OK()` |
| 5 | `GET /api/v1/alerts` | `meta` wrapped in `response.Response{Data: ...}` → `res.meta.total` is `undefined`. Pagination broken. | ✅ **Fixed** — uses `gin.H{...}` directly |

### ⚠️ Warnings (minor type/status mismatches)

| # | Endpoint | Issue |
|---|----------|-------|
| 1 | `POST .../deactivate` (user) | Backend returns 204 No Content. Dashboard typed `void`. Runtime works but JSON parse skips silently. |
| 2 | `POST .../reset-password` | Typed as `void` but backend sends `{ message: "..." }`. Caller ignores return, so OK at runtime. |
| 3 | `POST .../reset-pin` | Same as reset-password. |
| 4 | `DELETE /api/v1/roles/:id` | Backend returns 204. Same JSON caveat. |
| 5 | `POST .../assign-operator`, `remove-operator` | Backend returns 204. Same JSON caveat. |
| 6 | Backend store `Location.Capacity` uses `map[string]interface{}` | Dashboard types it as `Record<string, number>`. Numeric values stored as `float64` by Go JSON decoder will work, but non-numeric values would break. |

### 🔍 Needs Verification

*All previously 🔍 endpoints have been verified and are correct. Nothing outstanding.*

### ✅ Fully Aligned

- Auth (login, logout, refresh, me)
- Users CRUD (create, update, get by id)
- Roles CRUD (create, update, get by id, delete — minus list pagination)
- Locations (all 7 endpoints)
- Rates (all 3 endpoints)
- Sessions list + detail
- Shifts list + detail
- Sync conflicts resolve
- Incidents (detail, notes, resolve — minus list pagination check)
- Adjustments
- Alert configs
- Alert acknowledge + resolve
- All reports (JSON + CSV)
- Backups

---

## Backend Inconsistencies Found

1. ~~**Inconsistent list response format** — Some handlers use `gin.H{"items": ..., "meta": ...}` (correct) and others use `response.Response{Data: ..., Meta: ...}` (double-wrap). Affected: **users** handler.~~ ✅ **Fixed 2026-07-01**

2. ~~**Health endpoints bypass envelope** — `/health`, `/health/components` use raw `c.JSON()` instead of `response.OK()`.~~ ✅ **Fixed 2026-07-01**

3. ~~**Alerts meta wrapped in Response struct** — `meta` was nested inside `response.Response{Data: ...}`, making `meta.total` inaccessible.~~ ✅ **Fixed 2026-07-01**

4. **No standard `deleted_at` handling** — `store.Role` includes `deleted_at` but dashboard type doesn't account for it. Not urgent since soft-deleted roles are likely filtered at the store layer.
