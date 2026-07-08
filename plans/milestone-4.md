# Milestone 4 — Desktop App Online Mode

**Status:** Completed — implementation committed to `miles-4` branch.

## 1. Goal

Build a functional Electron desktop operator app that works in online mode, allowing operators to log in, select a location, start a shift, check in vehicles, search active sessions by plate, check out and collect payment (cash or digital), and print receipts.

---

## 2. Scope

### In Scope

- Backend authentication support for desktop via `Authorization: Bearer <token>` header.
- New backend endpoint `GET /auth/me` returning current user, role, assigned locations, and resolved permissions.
- New backend endpoint `GET /shifts/me/open` returning the current user's open shift or 404.
- Electron desktop app structure and routing (login, location select, dashboard, check-in, check-out, payment, success, history, shift).
- Secure in-memory token storage in the main process; expose auth and API helpers through the preload script.
- API client in the renderer that sends the Bearer token, parses envelope responses, and handles common errors.
- Login screen using email/password against `/auth/login`.
- Location selector that lists the operator's assigned locations.
- Shift start/end flow (operator must start a shift before checking in vehicles).
- Check-in screen: vehicle type selector, plate input, city code input, submit, print check-in ticket.
- Active-session search with plate auto-search and explicit check-out action.
- Check-out screen showing fee and duration.
- Payment screen: cash payment with amount tendered + change, and digital payment with reference code.
- Payment success screen with printable receipt and quick reprint button.
- Session history list with filters.
- Network status indicator and basic offline detection warning.
- System print integration: use Electron's `webContents.print()` via preload for receipts/tickets.
- README update for desktop dev commands.

### Out of Scope / Deferred

- Full offline mode and local sync (Milestone 5).
- Incident reporting and adjustments (Milestone 6).
- Manager dashboards, reports, analytics (Milestone 7).
- Auto-updates, packaging, and deployment (Milestone 8).
- Advanced thermal printer integration; v1 uses system print dialog.
- Barcode scanner integration (deferred).
- Mobile/touch-optimized layouts; target is desktop booth terminals.

---

## 3. Dependencies

- **Milestone 1 — Backend Core Entities** must be complete.
  - Auth endpoints and JWT middleware.
  - Users, roles, locations, and rates APIs.
  - Permission allow-list and RBAC resolver.
- **Milestone 2 — Backend Business Logic** must be complete.
  - Sessions, transactions, payments, receipts, and shifts APIs.
  - Fee calculation engine.
- **Milestone 3 — Web Dashboard Foundation** should be present.
  - Provides the envelope format convention and shared API patterns.

---

## 4. Detailed Tasks

### Backend

- [x] **Authentication header support**
  - Update `middleware/auth.go` to read the token from the `Authorization: Bearer <token>` header if the `access_token` cookie is absent.
- [x] **`GET /auth/me` endpoint**
  - Register route and return current user, role, assigned locations, and resolved permissions.
- [x] **`GET /shifts/me/open` endpoint**
  - Register under `/shifts/me/open` using `shifts:view` permission.
  - Return the operator's single open shift (or 404 with code `NO_OPEN_SHIFT`).
- [x] **CORS / cookie clarity**
  - Document that desktop uses Bearer tokens; dashboard continues using httpOnly cookies.
- [x] **List endpoints consistency**
  - Update `GET /locations` to return `{ items, meta }` matching other list endpoints.
  - Update `GET /sessions` state filter to accept comma-separated states (`CLOSED,VOIDED`).

### Desktop

#### Infrastructure

- [x] Set up React Router for screen navigation.
- [x] Create `src/renderer/lib/api.ts` fetch wrapper.
  - Read base URL from an environment variable / build-time constant (default `http://localhost:8080`).
  - Attach `Authorization: Bearer <token>` header.
  - Parse `{ data, error, meta }` envelope and throw typed errors.
  - Handle 401 by redirecting to login.
- [x] Create `src/renderer/types/` files mirroring backend entities.
- [x] Create `src/renderer/stores/authStore.ts` for React Context holding token, user, current location, open shift.
- [x] Expose via `src/preload.ts`:
  - `window.electronAPI.getToken()` / `setToken(token)` / `clearToken()`
  - `window.electronAPI.print(htmlContent, options?)`
  - `window.electronAPI.onOnlineStatusChange(callback)`

#### Screens

- [x] **Login screen**
  - Email/password form.
  - Call `POST /auth/login`, store token in main process, then call `GET /auth/me`.
  - Remember email in local storage; require password on every restart.
  - Error display for invalid credentials / network errors.
- [x] **Location select**
  - Display cards for each assigned location.
  - On select, fetch open shift via `GET /shifts/me/open`.
  - If no open shift, show "Start Shift" prompt.
- [x] **Shift status bar**
  - Always-visible indicator of current location and open shift status.
  - Button to start shift (requires `location_id`).
  - Button to end shift with cash handover input and optional notes.
  - Confirm before ending shift with active sessions warning.
- [x] **Dashboard home / main menu**
  - Large buttons: Check In, Check Out, History.
  - Display current shift summary: transactions, expected cash.
- [x] **Check-in screen**
  - Vehicle type quick-select (CAR / MOTO / TRUCK).
  - Plate input with normalization (uppercase, trim).
  - City code input (default to `B` or last used).
  - Submit calls `POST /sessions/check-in` and handles duplicate plate warning.
  - On success, show ticket with location, plate, vehicle type, check-in time, and print.
- [x] **Check-out / payment flow**
  - Search active sessions by plate with debounced auto-search (≥2 chars).
  - Select session, show duration and fee.
  - Payment method selector: Cash / Digital.
  - Cash: input amount tendered, auto-calculate change.
  - Digital: optional reference input.
  - Submit calls `POST /payments/cash` or `POST /payments/digital`.
  - Show payment success with printable receipt and quick reprint.
- [x] **Session history**
  - List closed/voided sessions for the current location with filters.
  - Load-more pagination.
  - Reprint receipt button for closed sessions.

### DevOps / QA

- [x] Update `desktop/package.json` scripts: `dev`, `build`, and add `lint` placeholder.
- [x] Update root `Makefile` if needed (no changes required; `desktop-run` exists).
- [x] Update `README.md` with desktop setup and dev instructions.
- [x] Add at least one happy-path manual test: login → start shift → check in → check out → cash payment → print receipt.

---

## 5. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Location model | Lock to one location at a time; operator can switch from the header | Matches shift-per-location semantics and simplifies state management. |
| Shift gate | Operator must start a shift before check-in; can browse history before shift | Backend already enforces this; aligns with audit and cash reconciliation. |
| Plate entry | Manual keyboard with uppercase normalization | Barcode scanner deferred; normalization keeps data consistent. |
| Check-out search | Auto-search after 2 characters + explicit check-out confirmation | Faster than submit-only while preventing accidental checkouts. |
| Printer failure | Show error, allow "skip & reprint later" | Do not block a completed payment because of printer hardware. |
| Quick reprint | Available on success screen and from history | Operator convenience; both paths useful. |
| Login persistence | Remember email only; require password on every app restart | Security/convenience balance for shared booth terminals. |
| Minimum resolution | 1024x768 | Common for low-cost booth monitors used by operators. |
| Auth transport | Bearer token stored in main process, sent in `Authorization` header | Simpler than cookie management in Electron; keeps dashboard cookie path unchanged. |
| Receipt printing | System print dialog via `webContents.print()` | No thermal-printer-specific integration yet; works with any printer. |
| Network status | Basic online/offline indicator, no sync | Online mode only in this milestone; sync is M5. |

---

## 6. Open Questions / Risks

### Resolved (proposed defaults)

| Question | Decision |
|----------|----------|
| Multiple locations per login | Lock to one location at a time |
| Shift before check-in | Required |
| Barcode scanner | Manual only in M4 |
| Check-out search | Auto-search + explicit confirmation |
| Printer failure | Skip & reprint later |
| Quick reprint | Success screen and history |
| Login persistence | Remember email, require password |
| Minimum resolution | 1024x768 |

### Remaining Risks

| Risk | Mitigation |
|------|------------|
| Backend still expects cookie token | Add Bearer header support in auth middleware before desktop implementation. |
| Electron print API behaves differently across OS | Test on target OS; fallback to `window.print()` if needed. |
| Active session search performance with many sessions | Use backend plate filter; debounce input; default limit 20. |
| Shared booths with remembered email | Allow manual logout; email-only memory is low risk. |
| CORS issues if backend and desktop run on different origins | Backend already has CORS config; use same origin in production. |

---

## 7. Acceptance Criteria

- [x] `GET /auth/me` returns authenticated user, role, locations, and permissions.
- [x] `GET /shifts/me/open` returns the operator's open shift or a clear `NO_OPEN_SHIFT` response.
- [x] Auth middleware accepts `Authorization: Bearer <token>`.
- [x] Desktop login succeeds with email/password and stores the token securely.
- [x] Operator can select an assigned location and start a shift.
- [x] Operator can check in a vehicle with plate, vehicle type, and city code.
- [x] Searching active sessions by plate returns matching results.
- [x] Operator can check out a session and view the calculated fee.
- [x] Cash payment records the transaction and calculates change.
- [x] Digital payment records a reference code and completes the session.
- [x] Receipt/ticket printing works through the system print dialog.
- [x] Payment success screen offers a quick reprint button.
- [x] History view lists closed sessions for the current location.
- [x] Network status indicator warns when the app goes offline.
- [x] `make desktop-run` launches the app and the happy path works against a running backend.

---

## 8. Definition of Done

Milestone 4 is complete when:
1. The backend supports Bearer token auth and exposes `/auth/me` and `/shifts/me/open`.
2. The Electron desktop app can log in, select a location, and start a shift.
3. Operators can complete check-in → check-out → payment → receipt print end-to-end while online.
4. Common errors (missing shift, duplicate plate, no rate, printer failure) are surfaced clearly.
5. The desktop app builds successfully with `make build-desktop`.
6. `README.md` documents how to run the desktop app in development.
7. All changes are committed on the `miles-4` branch with the message `implement milestone 4`.
