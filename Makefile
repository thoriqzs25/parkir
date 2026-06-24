.PHONY: dev stop build lint migrate-up migrate-down seed backend-dashboard desktop

# Development

dev:
	docker compose up --build

stop:
	docker compose down

# Migrations

migrate-up:
	cd backend && go run github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path migrations -database "$(DATABASE_URL)" up

migrate-down:
	cd backend && go run github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path migrations -database "$(DATABASE_URL)" down 1

# Seed

seed:
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed

# Builds

build: build-backend build-dashboard build-desktop

build-backend:
	cd backend && go build -o bin/api ./cmd/api

build-dashboard:
	cd dashboard && npm install && npm run build

build-desktop:
	cd desktop && npm install && npm run build

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
