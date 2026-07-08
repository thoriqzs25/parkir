# PARKIR

Parking Administration System (PAS) — multi-location parking management with operator desktop app and manager web dashboard.

## Stack

- **Backend:** Go 1.22+ + Gin + pgx (raw SQL) + PostgreSQL
- **Dashboard:** Next.js 14+ App Router + TypeScript + Tailwind CSS
- **Desktop:** Electron + React + TypeScript

## Prerequisites

- Go 1.23+
- Node.js 20+
- Docker & Docker Compose

## Quick Start

### 1. Clone and start the dev environment

```bash
git clone <repo-url>
cd PARKIR
make dev
```

This starts:
- PostgreSQL on `localhost:5432`
- Go backend on `http://localhost:8080`
- Next.js dashboard on `http://localhost:3000`

### 2. Run migrations

```bash
make migrate-up
```

### 3. Seed default data

```bash
make seed
```

This creates default roles (`owner`, `admin`, `manager`, `operator`) and a root owner user:
- Email: `owner@parkir.local`
- Password: `owner123`
- PIN: `123456`

### 4. Verify

```bash
curl http://localhost:8080/health
```

Expected response: `{"status":"ok"}`

## Useful Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start Docker Compose dev stack |
| `make stop` | Stop Docker Compose dev stack |
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Roll back one migration |
| `make seed` | Seed default roles and owner user |
| `make build` | Build backend, dashboard, and desktop |
| `make backend-run` | Run backend outside Docker |
| `make dashboard-run` | Run dashboard outside Docker |
| `make desktop-run` | Run desktop app |

## Project Structure

```
PARKIR/
├── backend/         # Go API
├── dashboard/       # Next.js web dashboard
├── desktop/         # Electron desktop app
├── db/              # Database documentation
├── plans/           # Milestone plans
├── specs/           # Product specifications
├── docker-compose.yml
└── Makefile
```

## API Endpoints (Milestone 1)

Base URL: `http://localhost:8080/api/v1`

Authentication uses RS256 JWT in an `access_token` httpOnly cookie.

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | `/auth/login` | — | Login with email/password |
| POST | `/auth/logout` | — | Clear auth cookie |
| POST | `/auth/refresh` | — | Refresh access token |
| GET | `/auth/me` | — | Get current user |
| GET | `/users` | `users:view` | List users |
| GET | `/users/:id` | `users:view` | Get user |
| POST | `/users` | `users:create` | Create user |
| PATCH | `/users/:id` | `users:edit` | Update user |
| POST | `/users/:id/reset-password` | `users:edit` | Reset password |
| POST | `/users/:id/reset-pin` | `users:edit` | Reset PIN |
| POST | `/users/:id/deactivate` | `users:deactivate` | Deactivate user |
| GET | `/roles` | `users:view` | List roles |
| GET | `/roles/:id` | `users:view` | Get role |
| POST | `/roles` | `users:create` | Create role |
| PATCH | `/roles/:id` | `users:create` | Update role |
| DELETE | `/roles/:id` | `users:create` | Soft delete role |
| GET | `/locations` | `locations:view` | List locations |
| GET | `/locations/:id` | `locations:view` | Get location |
| POST | `/locations` | `locations:create` | Create location |
| PATCH | `/locations/:id` | `locations:create` | Update location |
| POST | `/locations/:id/deactivate` | `locations:create` | Deactivate location |
| POST | `/locations/:id/assign-operator` | `locations:assign_operators` | Assign operator |
| POST | `/locations/:id/remove-operator` | `locations:assign_operators` | Remove operator |
| GET | `/locations/:id/rates` | `rates:view` | List rates for location |
| POST | `/locations/:id/rates` | `rates:create` | Create rate for location |
| PATCH | `/rates/:id` | `rates:edit` | Update rate |
| POST | `/sessions/check-in` | `sessions:create` | Check in a vehicle |
| GET | `/sessions` | `sessions:view` | List sessions |
| GET | `/sessions/:id` | `sessions:view` | Get session (use `?include=transaction`) |
| POST | `/sessions/:id/check-out` | `sessions:close` | Check out and calculate fee |
| POST | `/payments/cash` | `payments:collect_cash` | Record cash payment |
| POST | `/payments/digital` | `payments:collect_digital` | Record digital payment |
| GET | `/transactions` | `sessions:view` | List transactions |
| GET | `/transactions/:id` | `sessions:view` | Get transaction |
| POST | `/transactions/:id/void` | `payments:void` | Void transaction (manager PIN) |
| POST | `/shifts/start` | `shifts:start` | Start operator shift |
| GET | `/shifts` | `shifts:view` | List shifts |
| GET | `/shifts/:id` | `shifts:view` | Get shift (use `?include=transactions`)
| POST | `/shifts/:id/end` | `shifts:end` | End shift with cash handover |
| POST | `/shifts/:id/force-close` | `shifts:force_close` | Force-close open shift |

Responses use the envelope format `{ data, error, meta }`. List endpoints return `{ data: { items, meta } }`.

## Dashboard Pages

After logging in, the dashboard defaults to the user's first assigned location. Navigation includes:

- **Active Sessions** — live view of `ACTIVE` and `PENDING_PAYMENT` sessions with manual refresh.
- **Session History** — closed/voided sessions with plate filter and load-more pagination.
- **Session Detail** — full session info and linked transaction (via `?include=transaction`).
- **Transactions** — payment records with void badge and status filter.
- **Shifts** — operator shifts with summary list.
- **Shift Detail** — shift summary, cash summary, and transaction list (via `?include=transactions`).
- **Locations** — create, edit, deactivate, assign/remove operators.
- **Rates** — create and edit rates in a separate dialog/form.
- **Users** — create, edit, deactivate, reset password/PIN, assign locations.
- **Roles** — create, edit, soft-delete with permission allow-list.

## Desktop App

The operator desktop app is an Electron + React + TypeScript client.

### Development

1. Install dependencies:
   ```bash
   cd desktop
   npm install
   ```

2. Run the desktop app in development mode:
   ```bash
   make desktop-run
   # or from the desktop directory:
   npm run dev
   ```

3. The backend must be running at `http://localhost:8080`.

### Operator Flow

1. **Login** with email/password.
2. **Select Location** and start a shift.
3. From the **Main Menu**, choose:
   - **Check In** — register plate, vehicle type, and city code.
   - **Check Out** — search active sessions by plate, then calculate the fee.
   - **Payment** — record cash (with change) or digital payment.
   - **Success** — print receipt with system print dialog and reprint later.
   - **History** — list closed/voided sessions for the current shift.
4. **End Shift** from the dashboard to reconcile cash.

### Authentication

The desktop app authenticates via Bearer token (`Authorization: Bearer <token>`) stored in-memory in the Electron main process. The backend's auth middleware now also accepts this header when the `access_token` cookie is absent, so the same backend serves both dashboard (cookie) and desktop (token) clients.

### Minimum Target Resolution

1024×768.

## Local Development Notes

- Backend runs on `http://localhost:8080`.
- Dashboard runs on `http://localhost:3000`.
- Dashboard uses `NEXT_PUBLIC_API_URL` (default `http://localhost:8080`).
- Auth uses httpOnly `access_token` cookie; dashboard fetch calls use `credentials: 'include'`.
- All dashboard timestamps are displayed in Asia/Jakarta (WIB, UTC+7).

## Documentation

- `PLAN.md` — Overall implementation plan
- `MILESTONE_PLANNING.md` — Milestone planning workflow
- `plans/milestone-0.md` — Foundation milestone plan
- `plans/milestone-1.md` — Backend Core Entities milestone plan
- `plans/milestone-2.md` — Backend Business Logic milestone plan
- `plans/milestone-3.md` — Web Dashboard Foundation milestone plan
- `specs/` — Full product specifications
