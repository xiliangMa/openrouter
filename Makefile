.PHONY: help start stop restart build test lint clean migrate-up migrate-down docker-up docker-down

# Default target
help:
	@echo "MassRouter SaaS Platform - Make Commands"
	@echo ""
	@echo "Development:"
	@echo "  make start          - Start all services (docker-compose up)"
	@echo "  make stop           - Stop all services (docker-compose down)"
	@echo "  make restart        - Restart all services"
	@echo "  make logs           - View logs from all services"
	@echo ""
	@echo "Backend:"
	@echo "  make backend-build  - Build backend Go application"
	@echo "  make backend-run    - Run backend locally (requires local DB)"
	@echo "  make backend-test   - Run backend tests"
	@echo "  make backend-lint   - Run golangci-lint on backend"
	@echo ""
	@echo "Frontend:"
	@echo "  make admin-dev      - Start admin frontend in dev mode"
	@echo "  make portal-dev     - Start portal frontend in dev mode"
	@echo "  make admin-build    - Build admin frontend"
	@echo "  make portal-build   - Build portal frontend"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make db-reset       - Reset database (warning: destructive)"
	@echo ""
	@echo "Deployment:"
	@echo "  make docker-build   - Build all Docker images"
	@echo "  make docker-push    - Push Docker images to registry"
	@echo ""

# Development
start:
	docker-compose up -d

stop:
	docker-compose down

restart: stop start

logs:
	docker-compose logs -f

# Backend
backend-build:
	cd backend && go build -o bin/server ./cmd/server

backend-run:
	cd backend && go run ./cmd/server

backend-test:
	cd backend && go test ./... -v

backend-lint:
	cd backend && golangci-lint run

# Frontend
admin-dev:
	cd admin && npm run dev

portal-dev:
	cd portal && npm run dev

admin-build:
	cd admin && npm run build

portal-build:
	cd portal && npm run build

# Database
migrate-up:
	cd backend && go run ./cmd/migrate up

migrate-down:
	cd backend && go run ./cmd/migrate down

db-reset:
	@echo "WARNING: This will delete all data in the database!"
	@read -p "Are you sure? (y/N): " confirm && [ $${confirm:-N} = y ]
	docker-compose down -v
	docker-compose up -d postgres redis
	sleep 5
	cd backend && go run ./cmd/migrate up

# Docker
docker-build:
	docker-compose build

docker-push:
	@echo "Not implemented: Set DOCKER_REGISTRY and image tags"

# Setup
setup: 
	@echo "Setting up development environment..."
	cp .env.example .env
	@echo "Please edit .env file with your configuration"
	@echo "Run 'make start' to start services"
	@echo "Run 'make migrate-up' to apply database migrations"

# Code quality
lint: backend-lint
	@echo "Running frontend linting..."
	cd admin && npm run lint || true
	cd portal && npm run lint || true

# Clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf backend/bin
	rm -rf admin/.next admin/out admin/node_modules
	rm -rf portal/.next portal/out portal/node_modules
