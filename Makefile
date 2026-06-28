.PHONY: dev stop build lint test-backend migrate-up migrate-down migrate-create seed generate-jwt-keys backend-run dashboard-run desktop-run

# Development

dev:
	docker compose up --build

stop:
	docker compose down

# Migrations

migrate-up:
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate -direction up

migrate-down:
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate -direction down

migrate-create:
	cd backend && go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3 create -ext sql -dir migrations $(name)

# Seed

seed:
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed

# Keys

generate-jwt-keys:
	backend/scripts/generate-jwt-keys.sh

# Builds

build: build-backend build-dashboard build-desktop

build-backend:
	cd backend && go build -o bin/api ./cmd/api

build-dashboard:
	cd dashboard && npm install && npm run build

build-desktop:
	cd desktop && npm install && npm run build

# Tests

test-backend:
	cd backend && go test ./...

# Lint (placeholder until lint tools are configured)

lint:
	@echo "Lint tools not configured yet. Skipping."

# Local non-Docker development helpers

backend-run:
	cd backend && go run ./cmd/api

dashboard-run:
	cd dashboard && npm run dev

desktop-run:
	cd desktop && npm run dev

# Deployment

deploy-prod:
	@echo "=== Production Deploy ==="
	docker compose -f docker-compose.prod.yml build --no-cache backend
	docker compose -f docker-compose.prod.yml build dashboard
	docker compose -f docker-compose.prod.yml up -d
	@echo "=== Done ==="

deploy-staging:
	@echo "=== Staging Deploy ==="
	docker compose -f docker-compose.staging.yml build
	docker compose -f docker-compose.staging.yml up -d
	@echo "=== Done ==="
