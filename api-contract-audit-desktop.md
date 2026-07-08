# API Contract Audit: Desktop ↔ Backend

**Date:** 2026-07-01
**Scope:** All 9 desktop screens + AuthContext + SyncManager ↔ 17 backend endpoints

---

## Legend

| Icon | Meaning |
|------|---------|
| ✅ | Aligned — contract matches |
| ⚠️ | Warning — minor type/status mismatch, runtime may still work |
| ❌ | Misaligned — broken contract, will cause runtime errors |

---

## Architecture

- **Base URL:** `http://localhost:8080/api/v1` (hardcoded in `lib/api.ts:3`)
- **Auth:** `Authorization: Bearer <token>` header (JWT stored in Electron main process memory via IPC)
- **Envelope:** `request()` function reads `{ data: T, error?, meta? }` and returns `envelope.data`
- **Offline-first:** Check-in, check-out, and payment have complete offline code paths using `localStorage`. Data is synced via `POST /sync/batch` when connectivity is restored.

---

## 1. Auth

**File:** `contexts/AuthContext.tsx`
**Screens:** `Login.tsx`, `Layout.tsx` (logout)

| # | Method | Route | Desktop Expectation | Backend Reality | Status | Notes |
|---|--------|-------|---------------------|-----------------|--------|-------|
| 1 | POST | `/auth/login` | `{ user: User; token: string }` | `LoginResponse { user, token }` via `response.Created()` → `{ data: { user, token } }` | ✅ | Request: `{ email, password }` ✅. Token stored via IPC `set-token`. |
| 2 | POST | `/auth/logout` | `void` | `{ data: { message: "logged out" } }` via `response.OK()` | ✅ | Return value ignored. Token cleared via IPC `clear-token`. |
| 3 | GET | `/auth/me` | `MeResponse { user: User; permissions: string[] }` | `MeResponse { user, permissions }` via `response.OK()` → `{ data: { user, permissions } }` | ✅ | Called on startup (session restore) and after login. |

---

## 2. Locations

**File:** `contexts/AuthContext.tsx`

| # | Method | Route | Desktop Expectation | Backend Reality | Status | Notes |
|---|--------|-------|---------------------|-----------------|--------|-------|
| 4 | GET | `/locations` | `{ items: Location[]; meta?: { total: number } }` | `response.OK(c, gin.H{"items": ..., "meta": ...})` → `{ data: { items: [...], meta: {...} } }` | ✅ | Backend paginates with `Limit: total` (no real pagination). Desktop uses `items[]`. |

---

## 3. Shifts

**Files:** `contexts/AuthContext.tsx`, `screens/LocationSelect.tsx`, `screens/Dashboard.tsx`

| # | Method | Route | Desktop Expectation | Backend Reality | Status | Notes |
|---|--------|-------|---------------------|-----------------|--------|-------|
| 5 | GET | `/shifts/me/open` | `Shift` or `null` (on 404) | `response.OK(c, shift)` → `{ data: shift }` | ✅ | Desktop catches 404 and returns `null`. |
| 6 | POST | `/shifts/start` | `Shift` | `response.Created(c, shift)` → `{ data: shift }` | ✅ | Request: `{ location_id }` ✅ |
| 7 | POST | `/shifts/{id}/end` | `Shift` | `response.OK(c, shift)` → `{ data: shift }` | ✅ | Request: `{ cash_handover_amount, discrepancy_notes? }` ✅ |

---

## 4. Sessions (Check-In / Check-Out)

**Files:** `screens/CheckIn.tsx`, `screens/CheckOut.tsx`, `screens/History.tsx`

| # | Method | Route | Desktop Expectation | Backend Reality | Status | Notes |
|---|--------|-------|---------------------|-----------------|--------|-------|
| 8 | POST | `/sessions/check-in` | `CheckInResponse { session: Session; duplicate_plate_warning: boolean }` | `SessionResponse { session, duplicate_plate_warning }` via `response.Created()` → `{ data: { session, duplicate_plate_warning } }` | ✅ | Request: `{ location_id, plate, city_code, vehicle_type }` ✅ |
| 9 | POST | `/sessions/{id}/check-out` | `Session` | `response.OK(c, session)` → `{ data: session }` | ✅ | Request: `{ fee_amount? }` (or empty `{}`) ✅ |
| 10 | GET | `/sessions` | `{ items: Session[]; meta?: { total: number } }` | `response.OK(c, gin.H{"items": ..., "meta": ...})` → `{ data: { items: [...], meta: {...} } }` | ✅ | Query: `location_id`, `state`, `plate`, `operator_id`, `limit`, `offset` ✅. Comma-separated `state` (e.g. `"CLOSED,VOIDED"`) handled by backend store (`store/sessions.go:215` splits on comma). |

---

## 5. Payments

**File:** `screens/Payment.tsx`

| # | Method | Route | Desktop Expectation | Backend Reality | Status | Notes |
|---|--------|-------|---------------------|-----------------|--------|-------|
| 11 | POST | `/payments/cash` | `Transaction` | `response.Created(c, tx)` → `{ data: tx }` | ✅ | Request: `{ session_id, amount_tendered }` ✅ |
| 12 | POST | `/payments/digital` | `Transaction` | `response.Created(c, tx)` → `{ data: tx }` | ✅ | Request: `{ session_id, payment_reference? }` ✅ |

---

## 6. Transactions

**File:** `screens/Success.tsx`

| # | Method | Route | Desktop Expectation | Backend Reality | Status | Notes |
|---|--------|-------|---------------------|-----------------|--------|-------|
| 13 | GET | `/transactions/{id}` | `Transaction` | `response.OK(c, tx)` → `{ data: tx }` | ✅ | Called from Success screen to fetch full receipt data. |

---

## 7. Sync (Offline Batch)

**File:** `lib/sync.ts` (called from `App.tsx` SyncManager)

| # | Method | Route | Desktop Expectation | Backend Reality | Status | Notes |
|---|--------|-------|---------------------|-----------------|--------|-------|
| 14 | POST | `/sync/batch` | `BatchSyncResponse { results: SyncResult[] }` | `BatchSyncResponse { results: [SyncResult] }` via `response.OK()` → `{ data: { results: [...] } }` | ✅ | |

### Batch payload shapes:

**`check_in` item:**
| Field | Desktop sends | Backend `OfflineSessionData` | Match |
|-------|--------------|------------------------------|-------|
| `id` | `string` (uuid) | `ID string (required,uuid)` | ✅ |
| `location_id` | `string` | `LocationID string (required,uuid)` | ✅ |
| `operator_id` | `string` | `OperatorID string (required,uuid)` | ✅ |
| `shift_id` | `string` | `ShiftID string (required,uuid)` | ✅ |
| `plate` | `string` | `Plate string (required)` | ✅ |
| `city_code` | `string` | `CityCode string` | ✅ |
| `vehicle_type` | `"CAR"\|"MOTO"\|"TRUCK"` | `VehicleType string (required,oneof=...)` | ✅ |
| `check_in_at` | `string` (ISO) | `CheckInAt time.Time (required)` | ✅ |

**`check_out` item:**
| Field | Desktop sends | Backend `SyncItem` | Match |
|-------|--------------|---------------------|-------|
| `session_id` | `string` | `SessionID string` | ✅ |
| `check_out_at` | `string` (ISO) | `CheckOutAt time.Time` | ✅ |
| `fee_amount` | `number` | `FeeAmount *float64` | ✅ |
| `rate_snapshot` | `object?` | `RateSnapshot map[string]interface{}` | ✅ |

**`payment` item:**
| Field | Desktop sends | Backend `SyncItem` | Match |
|-------|--------------|---------------------|-------|
| `transaction_id` | `string` | `TransactionID string` | ✅ |
| `session_id` | `string` | `SessionID string` | ✅ |
| `shift_id` | `string` | `ShiftID string` | ✅ |
| `operator_id` | `string` | `OperatorID string` | ✅ |
| `location_id` | `string` | `LocationID string` | ✅ |
| `duration_hours` | `number` | `DurationHours int` | ⚠️ Desktop sends `number`, backend expects `int`. Go will truncate to int via JSON decode. |
| `rate_first_hour` | `number` | `RateFirstHour float64` | ✅ |
| `rate_subsequent_hourly` | `number` | `RateSubsequentHourly float64` | ✅ |
| `rate_daily` | `number` | `RateDaily float64` | ✅ |
| `fee_amount` | `number` | `FeeAmount float64` | ✅ |
| `payment_method` | `string` | `PaymentMethod string` | ✅ |
| `amount_tendered` | `number?` | `AmountTendered *float64` | ✅ |
| `change_amount` | `number?` | `ChangeAmount *float64` | ✅ |
| `payment_reference` | `string?` | `PaymentReference *string` | ✅ |

---

## 8. Incidents

**File:** `screens/IncidentReport.tsx`

| # | Method | Route | Desktop Expectation | Backend Reality | Status | Notes |
|---|--------|-------|---------------------|-----------------|--------|-------|
| 15 | POST | `/incidents` | Response ignored | `response.Created(c, inc)` → `{ data: inc }` | ✅ | Request: `{ location_id, type, description }` ✅. Desktop uses bare `request()` import, ignores response. |

---

## 9. Rates

**File:** `contexts/AuthContext.tsx` (rate cache on location change)

| # | Method | Route | Desktop Expectation | Backend Reality | Status | Notes |
|---|--------|-------|---------------------|-----------------|--------|-------|
| 16 | GET | `/locations/{id}/rates` | `Rate[]` | `response.OK(c, rates)` → `{ data: [...rates] }` | ✅ | 24-hour TTL cache in `localStorage`. |

---

## Type Alignment: Desktop ↔ Backend Store

### `User`
Desktop `types/index.ts:13` ↔ Backend `store/user.go:12` — fully aligned. ✅

### `Location`
Desktop `types/index.ts:1` ↔ Backend `store/location.go:12` — fully aligned. ✅

### `Shift`
Desktop `types/index.ts:30` ↔ Backend `store/shifts.go:14` — fully aligned. ✅

### `Session`
Desktop `types/index.ts:47` ↔ Backend `store/sessions.go:15` — fully aligned. ✅

### `Transaction`
Desktop `types/index.ts:66` ↔ Backend `store/transactions.go:12` — fully aligned. ✅

### `Rate`
| Field | Desktop `types/index.ts:94` | Backend `store/rate.go:12` | Match |
|-------|----------------------------|---------------------------|-------|
| `id` | `string` | `string` | ✅ |
| `location_id` | `string` | `string` | ✅ |
| `vehicle_type` | `"CAR"\|"MOTO"\|"TRUCK"` | `string` | ⚠️ No TS enum enforcement |
| `first_hour_rate` | `number` | `float64` | ✅ |
| `subsequent_hourly_rate` | `number` | `float64` | ✅ |
| `daily_flat_rate` | `number` | `float64` | ✅ |
| `effective_from` | `string` | `time.Time` | ✅ |
| `effective_until` | `string \| null` | `*time.Time,omitempty` | ✅ |
| `created_by` | `string?` | `*string,omitempty` | ✅ |
| `created_at` | `string` | `time.Time` | ✅ |
| `updated_at` | `string` | **Not in backend struct** | ⚠️ Backend `store.Rate` has no `updated_at` field. Field is never accessed by any screen — harmless. |

---

## Summary

### ✅ Fully Aligned (16/17 endpoints)

All core flows — login, session restore, check-in, check-out, payment, sync, history, incidents, rates — have matching contracts between desktop and backend.

### ⚠️ Warnings

| # | Detail | Impact |
|---|--------|--------|
| 1 | Desktop `Rate.updated_at` field has no backend counterpart | None — field is never read by any screen |
| 2 | Batch sync `duration_hours` sent as `number`, backend expects `int` | Go `json.Unmarshal` accepts numeric JSON values for any int/float type — works fine |

### 🔍 Interesting Observations

1. **Desktop uses flat `/incidents` path** — but the dashboard's `api.ts` uses `POST /api/v1/incidents`. Both point to the same endpoint thanks to the base URL. ✅
2. **`listTransactions`** is exported by `api.ts` but never called from any screen — dead code.
3. **`/health` and `/health/components`** are not used by the desktop app — only the dashboard uses them.
4. **No auto-refresh mechanism** — unlike the dashboard (which has `apiRequest` with 401 → refresh retry), the desktop calls `logout()` on any auth error in `me()`. The token is stored in the main process memory and never refreshed.
