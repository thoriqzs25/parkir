# Milestone 1 — Backend Core Entities

**Status:** Completed

**Completed on:** 2026-06-25

## 1. Goal

Build the complete data layer and authentication/RBAC foundation for PARKIR, including users, roles, locations, rates, and permission resolution, exposed through a secured REST API.

---

## 2. Scope

### In Scope
- Database migrations for all core entity tables
- Repository layer with raw SQL/pgx
- Password hashing (bcrypt) and PIN hashing (bcrypt)
- JWT access token authentication with 8-hour expiry
- RBAC permission resolution (role-location + independent grants)
- Auth middleware that validates JWT, loads user, and resolves permissions
- CRUD API endpoints for users, roles, locations, and rates
- Seed script for default roles and root owner
- Request/response envelope format
- Request validation and structured error responses
- Password policy: minimum length only

### Out of Scope / Deferred
- Parking sessions, payments, receipts, shifts (Milestone 2)
- Web dashboard UI pages (Milestone 3)
- Desktop app (Milestone 4)
- Offline mode (Milestone 5)
- Incidents, adjustments, observability (Milestone 6)
- Reports (Milestone 7)
- Production deployment (Milestone 8)
- Refresh token rotation/blacklisting
- Password reset / self-service flows
- Email verification
- Rate model Option B / time-of-day pricing

---

## 3. Dependencies

- **Milestone 0 — Foundation** must be complete.
- Final decisions already captured in `PLAN.md`.

---

## 4. Detailed Tasks

### Backend

#### Database Migrations
- [x] Migration: create `locations` table
- [x] Migration: create `location_rates` table
- [x] Migration: extend `roles` table (already exists from M0; verify constraints)
- [x] Migration: extend `users` table (already exists from M0; add indexes)
- [x] Migration: create `user_role_locations` table
- [x] Migration: create `user_permission_grants` table
- [x] Migration: add trigger to prevent overlapping rate effective dates per location/vehicle type

#### Domain Layer — Permissions
- [x] Define hardcoded allow-list of valid permission strings
- [x] Implement `permissions.GetEffectivePermissions(ctx, userID, locationID)` using raw SQL
- [x] Implement helper functions: `HasPermission(permissions, permission)`, `HasAnyPermission(permissions, ...permissions)`

#### Domain Layer — Auth
- [x] Implement password hashing with bcrypt
- [x] Implement PIN hashing with bcrypt
- [x] Implement `POST /auth/login` — validate email/password, issue JWT
- [x] Implement `POST /auth/logout` — clear auth cookie
- [x] Implement `POST /auth/refresh` — issue new JWT from valid existing token
- [x] Implement auth middleware: extract JWT, validate, load user, resolve permissions

#### Domain Layer — Users
- [x] Implement `GET /users` with pagination
- [x] Implement `POST /users` — create user with role and locations
- [x] Implement `GET /users/:id`
- [x] Implement `PATCH /users/:id` — update name, email, role, locations, status
- [x] Implement `POST /users/:id/deactivate`
- [x] Implement `POST /users/:id/reset-password`
- [x] Implement `POST /users/:id/reset-pin`

#### Domain Layer — Roles
- [x] Implement `GET /roles`
- [x] Implement `POST /roles` — create role with validated permissions
- [x] Implement `GET /roles/:id`
- [x] Implement `PATCH /roles/:id` — update name/permissions
- [x] Enforce that only owners can grant `finance:*` permissions

#### Domain Layer — Locations
- [x] Implement `GET /locations`
- [x] Implement `POST /locations` — create location with capacity
- [x] Implement `GET /locations/:id`
- [x] Implement `PATCH /locations/:id` — update name, address, city, capacity, status
- [x] Implement `POST /locations/:id/deactivate`
- [x] Implement `POST /locations/:id/assign-operator` and `POST /locations/:id/remove-operator`

#### Domain Layer — Rates
- [x] Implement `GET /locations/:id/rates`
- [x] Implement `POST /locations/:id/rates` — create rate for vehicle type
- [x] Implement `PATCH /rates/:id` — update rate
- [x] Implement rate validation: reject overlapping effective dates for same location/vehicle type
- [x] Implement rate lookup by check-in date

#### Shared Infrastructure
- [x] Implement request/response envelope: `{ data: ..., error: { code, message, details }, meta: {...} }`
- [x] Implement centralized error handling middleware
- [x] Implement request validation helpers
- [x] Add audit log writes for user/role/location/rate mutations

### Dashboard

- [x] No dashboard pages in this milestone; API-only focus.

### Desktop

- [x] No desktop work in this milestone.

### DevOps / QA

- [x] Add curl examples for auth and CRUD endpoints to README
- [ ] Add integration test scaffolding (test database setup) — deferred to Milestone 2
- [ ] Update CI to run backend build and basic integration tests — deferred to Milestone 2
- [x] Update `Makefile` with `test-backend` and `migrate-create`

---

## 5. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Password policy | Minimum length only (e.g., 8 chars) | Simpler UX; complexity rules often frustrate users. |
| First owner account | Seed script | Zero-touch onboarding; `make seed` creates root owner. |
| Permission validation | Hardcoded allow-list | Prevents typos and ensures RBAC consistency. |
| Location deactivation | Kick out active operators immediately | Simpler mental model; prevents operations at inactive locations. |
| Rate retroactivity | Only new check-ins | Matches spec: rate is determined at check-in date; rate snapshot preserves historical accuracy. |
| Overlapping rates | Reject overlaps | Avoids ambiguity; clean data model. |
| API response format | Envelope format | Consistent error handling across frontend clients. |
| JWT permissions | Look up per request | Token stays small; permissions can be revoked immediately. |
| JWT expiry | 8 hours | Matches typical operator shift length. |
| JWT signing | RS256 | More secure than HS256; supports key rotation. |
| Token transport | httpOnly cookie for dashboard auth | XSS protection; desktop uses secure storage + cookie. |
| Role deletion | Soft delete (`deleted_at`) | Preserves audit trail and historical references. |
| Location code updates | Disabled after creation | Code is used in receipts and reports; changing it breaks references. |
| Audit log writes | Synchronous | Simpler and ensures audit trail is always current. |

---

## 6. Open Questions / Risks

### Resolved
| Question | Decision |
|----------|----------|
| JWT signing algorithm | RS256 |
| Dashboard auth transport | httpOnly cookies |
| Role deletion | Soft delete via `deleted_at` |
| `location.code` changes | Disabled after creation |
| Audit log writes | Synchronous |

### Remaining Risks
| Risk | Mitigation |
|------|------------|
| RS256 key generation complexity | Provide Makefile target and dev fallback keys |
| httpOnly cookies complicate cross-origin local dev | Configure CORS + credentials properly in docker-compose |
| Soft-deleted roles still referenced by users | Show role as archived in UI; prevent assignment to new users |
| Synchronous audit logs may slow down write-heavy endpoints | Keep audit metadata small; optimize later if needed |

---

## 7. Acceptance Criteria

- [x] `POST /auth/login` returns a valid JWT for correct credentials.
- [x] `POST /auth/login` returns a clear error for invalid credentials.
- [x] Auth middleware rejects requests without a valid JWT.
- [x] Auth middleware attaches user and resolved permissions to the request context.
- [x] `POST /roles` rejects unknown permissions (hardcoded allow-list).
- [x] `POST /roles` rejects `finance:*` permissions unless the actor is an owner.
- [x] `POST /users` creates a user with role and assigned locations.
- [x] `GET /users/:id` returns user details including role and locations.
- [x] `POST /locations` creates a location with capacity.
- [x] `POST /locations/:id/deactivate` sets location status to INACTIVE.
- [x] `POST /locations/:id/rates` rejects overlapping effective dates.
- [x] `GET /locations/:id/rates` returns rates ordered by effective date.
- [x] API responses use envelope format.
- [x] Seed script creates default roles and root owner.

---

## 8. Definition of Done

Milestone 1 is complete when:
1. [x] All migrations apply cleanly.
2. [x] The auth system issues and validates JWTs.
3. [x] RBAC resolves effective permissions correctly for role-location assignments and independent grants.
4. [x] CRUD endpoints for users, roles, locations, and rates are implemented and manually tested.
5. [x] API responses follow the envelope format.
6. [x] Seed script produces a usable root owner account.
7. [x] A teammate can use the API (via curl or Postman) to create users, roles, locations, and rates.
