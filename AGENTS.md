# PARKIR — Agent Quick Reference

Project context for LLM agents. Read this first when starting a new session.

## What is PARKIR?

Multi-location parking management system.

- **Backend:** Go 1.22+ (Gin, pgx raw SQL, JWT RS256)
- **Dashboard:** Next.js 14 App Router + TypeScript + Tailwind
- **Desktop:** Electron + React + TypeScript
- **Database:** PostgreSQL 15

## Project layout

```
PARKIR/
├── backend/               # Go API
│   ├── cmd/api/           # Main HTTP server entrypoint
│   ├── cmd/migrate/       # DB migrations tool
│   ├── cmd/seed/          # Seed default roles + owner user
│   ├── internal/domain/   # Domain handlers (auth, users, locations, sessions, ...)
│   ├── internal/store/    # Data access layer (pgx)
│   ├── internal/db/       # Pool + migrate helpers
│   └── migrations/        # SQL migration files
├── dashboard/             # Next.js web dashboard
├── desktop/               # Electron desktop app
├── bruno/PARKIR/          # Bruno API collection + LOCAL environment
├── docker-compose.yml     # Dev stack: postgres + backend + dashboard
└── Makefile               # Common commands
```

## Common commands

```bash
# Start full dev stack (Postgres 5432, backend 8080, dashboard 3000)
make dev

# Run backend locally with hot reload (air), connecting to local postgresql@18
cd backend && air -c .air.toml
# (Set DATABASE_URL, MIGRATIONS_PATH, JWT_PRIVATE_KEY_PATH, JWT_PUBLIC_KEY_PATH as needed)

# Stop the dev stack
make stop

# Run build for backend / dashboard / desktop
make build                 # all three
make build-backend
make build-dashboard
make build-desktop

# Run backend outside Docker
make backend-run

# Run dashboard outside Docker
make dashboard-run

# Run desktop app
make desktop-run

# DB migrations + seed (require DATABASE_URL or PARKIR_TEST_DATABASE_URL)
make migrate-up
make migrate-down
make seed

# Tests (integration tests need a running PostgreSQL)
make test-backend
```

## Authentication

- Auth is split between public and protected `/api/v1/auth` routes:
  - `POST /api/v1/auth/login` — public
  - `POST /api/v1/auth/logout` — public
  - `POST /api/v1/auth/refresh` — public
  - `GET  /api/v1/auth/me` — requires valid JWT cookie
- Dashboard uses the `access_token` httpOnly cookie.
- Desktop uses `Authorization: Bearer <token>` header.

Default seed user:
- Email: `owner@parkir.local`
- Password: `owner123`
- PIN: `123456`

## API conventions

- Base path: `/api/v1`
- All responses use envelope format: `{ data, error, meta }`.
- List endpoints return: `{ data: { items, meta } }`.
- Timezone: dashboard timestamps are rendered in `Asia/Jakarta` (WIB, UTC+7).

## Recent fixes & gotchas

1. **`make build-desktop` failed** due to a typo in `desktop/src/renderer/screens/History.tsx` (extra `"`). Fixed.
2. **`make dev` failed** because `backend/Dockerfile.dev` used `golang:1.23-alpine` but `air@latest` requires Go 1.25+. Upgraded to `golang:1.25-alpine`.
3. **Dashboard container failed** because `node_modules` volume was empty. Updated `docker-compose.yml` dashboard command to `npm install && npm run dev`.
4. **Auth routes were inconsistent** (login at `/auth/login`, me at `/api/v1/auth/me`). All auth routes are now under `/api/v1/auth/`.
5. **Added `/health/ready`** endpoint that runs `SELECT 1` to verify DB connectivity.

## Local Postgres conflict

If you have a local Postgres already running on `localhost:5432`, host-side `make migrate-up` / `make seed` / `make test-backend` may connect to the wrong instance. Either stop the local Postgres or use the Docker container hostname (`postgres`) when running inside Docker.

## Bruno collection

Location: `bruno/PARKIR/`
- Environment: `LOCAL` (`baseUrl: http://localhost:8080`, `apiV1: {{baseUrl}}/api/v1`)
- Request files use lowercase-kebab naming (e.g. `create-user.bru`).
- Run `Auth / Login` first; Bruno’s cookie jar stores the JWT for other requests.

## Key files for common tasks

| Task | File(s) |
|------|---------|
| Add/modify API routes | `backend/internal/domain/<domain>/handler.go` + `backend/cmd/api/main.go` |
| Add/modify DB tables | `backend/migrations/` + `backend/internal/store/` |
| Authentication logic | `backend/internal/auth/`, `backend/internal/middleware/auth.go` |
| Permissions | `backend/internal/permissions/permissions.go` |
| Add health/observability | `backend/internal/domain/health/handler.go` |
| Dashboard pages | `dashboard/src/app/` |
| Desktop screens | `desktop/src/renderer/screens/` |

## Test commands

```bash
# Backend integration tests (need DATABASE_URL or PARKIR_TEST_DATABASE_URL pointing to Postgres)
make test-backend

# Equivalent:
cd backend && go test ./...
```
