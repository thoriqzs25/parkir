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
| POST | `/locations/:id/assign-operator` | `locations:create` | Assign operator |
| POST | `/locations/:id/remove-operator` | `locations:create` | Remove operator |
| GET | `/locations/:id/rates` | `rates:view` | List rates for location |
| POST | `/locations/:id/rates` | `rates:create` | Create rate for location |
| PATCH | `/rates/:id` | `rates:edit` | Update rate |

Responses use the envelope format `{ data, error, meta }`.

## Documentation

- `PLAN.md` — Overall implementation plan
- `MILESTONE_PLANNING.md` — Milestone planning workflow
- `plans/milestone-0.md` — Foundation milestone plan
- `plans/milestone-1.md` — Backend Core Entities milestone plan
- `specs/` — Full product specifications
