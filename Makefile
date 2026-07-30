.PHONY: help backend-run backend-build backend-test frontend-install frontend-dev frontend-build \
        db-setup docker-up docker-down docker-logs fmt lint

help:
	@echo "Available targets:"
	@echo "  make backend-run       Run the Go API locally (go run ./cmd/api)"
	@echo "  make backend-build     Build the Go API binary into backend/bin/"
	@echo "  make backend-test      Run Go tests"
	@echo "  make frontend-install  Install frontend dependencies"
	@echo "  make frontend-dev      Run the Vite dev server"
	@echo "  make frontend-build    Build the production frontend bundle"
	@echo "  make db-setup          Apply schema.sql to a local MySQL instance"
	@echo "  make docker-up         Build and start mysql + backend + frontend via Docker Compose"
	@echo "  make docker-down       Stop and remove Docker Compose containers"
	@echo "  make docker-logs       Tail logs from all Docker Compose services"
	@echo "  make fmt               Format Go and frontend source"
	@echo "  make lint              Lint frontend source"

# ---------------------------------------------------------------------
# Backend (Go)
# ---------------------------------------------------------------------
backend-run:
	cd backend && go run ./cmd/api

backend-build:
	cd backend && mkdir -p bin && go build -o bin/task-manager-api ./cmd/api

backend-test:
	cd backend && go test ./...

# ---------------------------------------------------------------------
# Frontend (React + Vite)
# ---------------------------------------------------------------------
frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

# ---------------------------------------------------------------------
# Database
# ---------------------------------------------------------------------
db-setup:
	mysql -u root -p < backend/schema.sql

# ---------------------------------------------------------------------
# Docker Compose (mysql + backend + frontend)
# ---------------------------------------------------------------------
docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# ---------------------------------------------------------------------
# Formatting / linting
# ---------------------------------------------------------------------
fmt:
	cd backend && go fmt ./...
	cd frontend && npm run format --if-present

lint:
	cd frontend && npm run lint
