# Milestone 0 — Foundation

## 1. Goal

Establish a working local development environment where the Go backend, Next.js dashboard, and Electron desktop app can run together with hot-reload, database migrations, and a shared project structure.

---

## 2. Scope

### In Scope
- Single monorepo structure for backend, dashboard, and desktop
- Docker Compose local dev stack (PostgreSQL + Go backend + Next.js dashboard)
- Go backend scaffold with Gin, pgx, structured JSON logging, and health endpoint
- Next.js App Router dashboard scaffold with Tailwind CSS
- Electron desktop app scaffold with main + renderer processes
- golang-migrate database migrations setup
- Makefile with common dev commands
- CI/CD skeleton (lint + build, no tests yet)
- Seed script for default roles and root owner user
- Initial README with setup instructions

### Out of Scope / Deferred
- Any business logic (auth, sessions, payments, etc.)
- Database tables beyond migrations metadata
- Dashboard pages beyond a placeholder
- Desktop app operational screens
- Production deployment configuration
- Automated test suite (CI runs lint/build only)

---

## 3. Dependencies

- None — this is the first milestone.
- Requires final decisions already captured in `PLAN.md`.

---

## 4. Detailed Tasks

### Backend
- [ ] Initialize Go module at `backend/` with Go 1.23
- [ ] Install dependencies: `gin-gonic/gin`, `jackc/pgx/v5`, `golang-migrate/migrate/v4`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`
- [ ] Create `cmd/api/main.go` with Gin server bootstrap
- [ ] Create `internal/config` for env-based configuration (port, DB URL, JWT secret, log level)
- [ ] Create `internal/db` for PostgreSQL connection pool and migration runner
- [ ] Create `internal/logger` wrapping `slog` with JSON output
- [ ] Create `internal/middleware` for request logging and CORS
- [ ] Implement `GET /health` endpoint returning `{"status":"ok"}`
- [ ] Create `migrations/` directory and initial schema migration files
- [ ] Add `Makefile` commands: `backend-run`, `backend-build`, `backend-lint`, `migrate-up`, `migrate-down`, `seed`
- [ ] Create seed command to insert default roles and root owner user

### Dashboard
- [ ] Initialize Next.js 14+ App Router project at `dashboard/` with TypeScript and Tailwind CSS
- [ ] Set Node.js 20 in `.nvmrc` and `package.json` engines
- [ ] Create basic layout and a placeholder home page
- [ ] Set up API client stub pointing to backend URL
- [ ] Add `Makefile` commands: `dashboard-dev`, `dashboard-build`, `dashboard-lint`

### Desktop
- [ ] Initialize Electron project at `desktop/` with TypeScript and Vite
- [ ] Set up main process entry (`src/main/index.ts`)
- [ ] Set up renderer process entry (`src/renderer/main.tsx`)
- [ ] Create a simple window loading a placeholder UI
- [ ] Add `Makefile` commands: `desktop-dev`, `desktop-build`, `desktop-lint`

### DevOps / QA
- [ ] Create root `docker-compose.yml` with services:
  - `postgres` (PostgreSQL 15+)
  - `backend` (Go with `air` for hot-reload)
  - `dashboard` (Next.js dev server with HMR)
- [ ] Create root `Makefile` with:
  - `make dev` — start Docker Compose
  - `make stop` — stop Docker Compose
  - `make migrate-up` / `make migrate-down`
  - `make seed`
  - `make lint` — lint backend, dashboard, desktop
  - `make build` — build all three
- [ ] Create `.github/workflows/ci.yml` running lint and build on push/PR
- [ ] Add root `.gitignore` covering Go, Node, Electron, and IDE files
- [ ] Write initial `README.md` with setup and run instructions

---

## 5. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Repo structure | Single monorepo | Easier cross-project changes and shared types later. |
| Hot reload | Yes (`air` for Go, Next.js dev server) | Faster iteration during development. |
| Migrations | golang-migrate | Industry standard, versioned, supports up/down. |
| Dashboard serving | Separate services in production | Cleaner separation; CORS handled in dev via compose. |
| Go version | 1.23 | Latest stable; good generics and slog support. |
| Node version | 20 LTS | Stable, widely supported. |
| Logging | Structured JSON via slog | Production-ready and easy to parse. |
| CI scope | Lint + build only | No meaningful tests yet; add tests in later milestones. |
| Seed data | Makefile seed command | Provides default roles and root owner out of the box. |

---

## 6. Open Questions / Risks

| Question | Owner | Due Date |
|----------|-------|----------|
| Which lint tools to enforce — `golangci-lint`, ESLint, Prettier, Biome? | Team | Before coding starts |
| Should backend and dashboard share TypeScript/Go types via code generation or manual copies? | Team | Milestone 1 |
| Should Electron use the same Next.js build as dashboard or a separate renderer build? | Team | Milestone 4 |

---

## 7. Acceptance Criteria

- [ ] `make dev` starts PostgreSQL, Go backend, and Next.js dashboard without errors.
- [ ] `curl http://localhost:8080/health` returns `{"status":"ok"}`.
- [ ] Dashboard loads at `http://localhost:3000` and shows a placeholder page.
- [ ] Desktop app launches and shows a window with placeholder content.
- [ ] `make migrate-up` applies migrations successfully.
- [ ] `make seed` inserts default roles and a root owner into the database.
- [ ] `make lint` passes for all three projects.
- [ ] `make build` produces artifacts for backend, dashboard, and desktop.
- [ ] CI workflow passes on the default branch.

---

## 8. Definition of Done

The milestone is complete when a new developer can clone the repo, run `make dev`, and have a working local environment with a running database, backend API, dashboard, and desktop app scaffold, plus the ability to apply migrations and seed default data.
