# Milestone 3 — Web Dashboard Foundation

**Status:** Completed — backend endpoint added and all dashboard pages implemented; builds successfully.

## 1. Goal

Build the manager-facing Next.js dashboard so that authenticated managers can configure the system (users, roles, locations, rates) and view operational data (active sessions, session history, transactions, shifts) through a desktop-optimized web UI.

---

## 2. Scope

### In Scope
- Login page with email/password and httpOnly cookie handling.
- Authenticated layout with sidebar navigation, header, and location selector.
- Default-location routing: after login, redirect to the user's first assigned location context.
- Users management page: list, create, edit, deactivate, reset password/PIN, assign locations.
- Roles & permissions page: list, create, edit, soft-delete roles with permission allow-list validation.
- Locations management page: list, create, edit, deactivate, assign operators.
- Rates configuration page: list rates per location, create rate, edit rate in a separate dialog/form.
- Active sessions page: list active sessions with manual refresh.
- Session history page: list closed/voided sessions with filters (plate, state, date range).
- Transactions list page: list transactions with void badge and filters.
- Shifts list and detail page: list shifts, view shift details and discrepancy.
- Toast notifications for create/update/delete actions.
- Form validation on blur using a form library (React Hook Form + Zod).
- Client-side permission helpers to show/hide actions based on resolved permissions.
- Timestamps displayed in Asia/Jakarta (WIB, UTC+7).
- API client with cookie-based auth and envelope response handling.
- Dashboard TypeScript types for all M2/M3 entities.

### Out of Scope / Deferred
- Reports and charts (Milestone 7).
- Audit log viewer (Milestone 6).
- Incidents and adjustments UI (Milestone 6).
- Sync conflict resolution UI (Milestone 5).
- Mobile-responsive layout (deferred until after v1).
- Real-time WebSocket updates.
- Dashboard user preferences or profile page.
- Production deployment specifics (Milestone 8).

---

## 3. Dependencies

- **Milestone 1 — Backend Core Entities** must be complete.
  - Auth endpoints (`/auth/login`, `/auth/refresh`, `/auth/logout`).
  - Users, roles, locations, rates CRUD endpoints.
  - RBAC middleware and permission resolver.
- **Milestone 2 — Backend Business Logic** must be complete.
  - Sessions, transactions, shifts endpoints.
  - Fee calculation and receipt generation.
- Final decisions from `PLAN.md`:
  - Next.js App Router + Tailwind CSS.
  - httpOnly cookie auth.
  - WIB timezone for display.
  - API envelope format.

---

## 4. Detailed Tasks

### Backend

- [x] Verify all dashboard-required endpoints return the envelope format consistently.
- [x] Add `GET /auth/me` endpoint to return current user (id, name, email, role, locations, permissions).
- [x] Ensure CORS config allows dashboard origin with credentials in development and production.
- [x] Add query param `include=transaction` to `GET /sessions/:id`.

### Dashboard

#### Project Setup & Shared Infrastructure
- [x] Install dependencies: `react-hook-form`, `zod`, `@hookform/resolvers`, `sonner` (toast), `date-fns` or `dayjs` (WIB formatting).
- [x] Create `lib/api.ts` with fetch wrapper that:
  - Sends cookies (`credentials: 'include'`).
  - Parses `{ data, error, meta }` envelope.
  - Refreshes token on 401 and retries once.
  - Throws typed API errors.
- [x] Create `lib/permissions.ts` with `hasPermission(permissions, permission)` and `hasAnyPermission` helpers.
- [x] Create `lib/time.ts` with WIB formatting helpers.
- [x] Create `types/` files: `auth.ts`, `user.ts`, `role.ts`, `location.ts`, `rate.ts`, `session.ts`, `transaction.ts`, `shift.ts`, `api.ts`.
- [x] Create shared UI components in `components/ui/`: `Button`, `Input`, `Select`, `Dialog`, `Table`, `Badge`, `Toast`, `Skeleton`, `EmptyState`, `Pagination/LoadMore`.
- [x] Create `components/layout/`:
  - `AuthGuard` — redirects to `/login` if not authenticated.
  - `DashboardLayout` — sidebar, header, location switcher, user menu.
  - `LocationProvider` — React context for current location, defaulting to first assigned location.

#### Auth
- [x] Build `/login` page with email/password form.
- [x] Call `POST /auth/login` and redirect to dashboard on success.
- [x] Handle login errors (invalid credentials, server error).
- [x] Add logout button that calls `POST /auth/logout` and clears client state.
- [x] Create `useAuth` hook that fetches `/auth/me` on app mount and provides user + permissions.

#### Navigation & Routing
- [x] Define dashboard route groups under `app/(dashboard)/[locationId]/`.
- [x] Sidebar links:
  - Dashboard home (active sessions)
  - Sessions (history)
  - Transactions
  - Shifts
  - Locations
  - Rates
  - Users
  - Roles
- [x] Highlight active route.
- [x] Location selector in header; switching updates the `[locationId]` route.
- [x] Redirect `/` to `/[defaultLocationId]/sessions/active`.

#### Users Page
- [x] `app/(dashboard)/[locationId]/users/page.tsx`
- [x] List users in a table with name, email, role, status, locations.
- [x] Load more pagination.
- [x] Create user dialog with name, email, password, role, location assignment.
- [x] Edit user dialog (name, email, role, locations, status).
- [x] Deactivate user action with confirmation.
- [x] Reset password/PIN actions (manager/admin only).
- [x] Hide create/edit actions if user lacks `users:create`/`users:edit` permissions.

#### Roles Page
- [x] `app/(dashboard)/[locationId]/roles/page.tsx`
- [x] List roles with permissions displayed as badges.
- [x] Create role dialog with name and permission multi-select using the allow-list.
- [x] Edit role dialog.
- [x] Soft-delete role with confirmation.
- [x] Enforce that only owners can grant `finance:*` permissions.

#### Locations Page
- [x] `app/(dashboard)/[locationId]/locations/page.tsx`
- [x] List locations with name, code, city, status, capacity summary.
- [x] Create location form.
- [x] Edit location form (name, address, city, capacity, status).
- [x] Deactivate location action.
- [x] Assign/remove operators (requires `locations:assign_operators`).

#### Rates Page
- [x] `app/(dashboard)/[locationId]/rates/page.tsx`
- [x] List rates per vehicle type with effective dates.
- [x] Create rate dialog.
- [x] Edit rate in a separate dialog/form (not inline).
- [x] Display rate overlap errors from backend as toast/inline error.

#### Sessions Pages
- [x] `app/(dashboard)/[locationId]/sessions/active/page.tsx`
  - List `ACTIVE` and `PENDING_PAYMENT` sessions.
  - Manual refresh button.
  - Filter by plate.
  - Link to session detail.
- [x] `app/(dashboard)/[locationId]/sessions/history/page.tsx`
  - List `CLOSED` and `VOIDED` sessions.
  - Filters: plate, state, date range.
  - Load more pagination.
- [x] `app/(dashboard)/[locationId]/sessions/[id]/page.tsx`
  - Session detail with linked transaction if closed.

#### Transactions Page
- [x] `app/(dashboard)/[locationId]/transactions/page.tsx`
- [x] List transactions with receipt number, plate, amount, method, void badge.
- [x] Filter by voided status, date range.
- [x] Load more pagination.
- [x] Link to session detail.

#### Shifts Page
- [x] `app/(dashboard)/[locationId]/shifts/page.tsx`
- [x] List shifts with operator, status, started/ended, discrepancy.
- [x] Filter by status, operator, date range.
- [x] `app/(dashboard)/[locationId]/shifts/[id]/page.tsx`
  - Shift detail with cash summary and transaction list.

### Desktop

- [x] No desktop work in this milestone.

### DevOps / QA

- [x] Add dashboard build to CI.
- [x] Add basic type-check step for dashboard in CI.
- [ ] Add at least one happy-path dashboard smoke test (login → view sessions).
- [x] Update `README.md` with dashboard dev commands and page overview.

---

## 5. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Landing page | Default to user's first assigned location | Faster navigation; managers usually manage one primary location. |
| Rate editing | Separate edit dialog/form | Safer UX; avoids accidental edits and complex inline validation. |
| Session refresh | Manual refresh button | Predictable data state; avoids distracting auto-polling. |
| Pagination | Limit/offset tables with Load more | Simple to implement and aligns with backend pagination. |
| Mobile support | Desktop-only for v1 | Managers primarily use desktop; mobile responsiveness deferred. |
| Action feedback | Toast notifications | Non-intrusive, consistent feedback across pages. |
| Form validation | On blur | Immediate feedback without waiting for submit. |
| Timezone display | Asia/Jakarta (WIB, UTC+7) | Matches receipt timestamps and business operations; no DST. |
| Auth state | `useAuth` hook fetching `/auth/me` | Keeps client state in sync with server-side session. |
| API client | Fetch with credentials + envelope parser | Native fetch, no extra dependency, supports httpOnly cookies. |
| Location context | React Context + route param | Avoids prop drilling; URL is shareable. |

---

## 6. Open Questions / Risks

### Resolved
| Question | Decision |
|----------|----------|
| Landing page | Default to first assigned location |
| Rate editing | Separate edit form/dialog |
| Session refresh | Manual refresh |
| Pagination | Load more |
| Mobile | Desktop-only for v1 |
| Notifications | Toast |
| Validation | On blur |
| Timezone | WIB |

### Remaining Risks
| Risk | Mitigation |
|------|------------|
| httpOnly cookies + CORS complexity in local dev | Use `FRONTEND_URL` env and `credentials: 'include'`; test early. |
| Permission rendering drift between backend and frontend | Share permission allow-list or keep frontend helper in sync manually. |
| Large list performance with load-more | Keep default limit reasonable (20–50) and add backend search filters. |
| Redirect logic breaks if user has no assigned location | Fallback to a "Select location" page if `user.locations` is empty. |
| Form library adds bundle size | Only install required packages; tree-shake unused components. |

---

## 7. Acceptance Criteria

- [x] `/login` accepts email/password and redirects to the dashboard on success.
- [x] Authenticated users see a sidebar, header, and location selector.
- [x] After login, the dashboard defaults to the user's first assigned location.
- [x] Managers can create, edit, and deactivate users.
- [x] Managers can create, edit, and soft-delete roles.
- [x] Managers can create, edit, and deactivate locations.
- [x] Managers can create and edit rates in a separate form/dialog.
- [x] Active sessions page lists sessions and supports manual refresh.
- [x] Session history page supports filters and load-more pagination.
- [x] Transactions page shows voided transactions with a badge.
- [x] Shifts page lists shifts.
- [x] Shifts detail page with discrepancy and transaction list.
- [x] Actions show toast notifications on success/error.
- [x] Forms validate on blur.
- [x] Timestamps are displayed in WIB.
- [x] UI actions are hidden when the user lacks the required permission.
- [x] Dashboard builds and type-checks successfully.

---

## 8. Definition of Done

Milestone 3 is complete when:
1. [x] The dashboard login flow works end-to-end with the backend.
2. [x] Managers can configure users, roles, locations, and rates through the dashboard.
3. [x] Managers can view active sessions, session history, transactions, and shifts.
4. [x] Permission-based UI hiding is implemented for all privileged actions.
5. [x] Toast feedback, blur validation, and WIB timestamps are consistently applied.
6. [x] The dashboard builds successfully and passes type checks.
7. [x] CI runs backend tests, dashboard build, and dashboard type-check.
8. [x] A teammate can run `make dashboard-run` and complete the login → view sessions flow.
