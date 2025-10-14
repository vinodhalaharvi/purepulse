.PHONY: help db-up db-down db-logs db-shell migrate-up migrate-down migrate-version migration-create

# Load environment variables
include .env
export

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Docker commands
db-up: ## Start PostgreSQL container
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 5
	@docker-compose exec postgres pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB)
	@echo "Database is ready!"

db-down: ## Stop PostgreSQL container
	docker-compose down

db-restart: ## Restart PostgreSQL container
	docker-compose restart postgres

db-logs: ## Show PostgreSQL logs
	docker-compose logs -f postgres

db-shell: ## Open psql shell in container
	docker-compose exec postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

db-clean: ## Remove database container and volumes (WARNING: deletes all data)
	docker-compose down -v
	@echo "All database data has been deleted!"

pgadmin-up: ## Start PgAdmin container
	docker-compose up -d pgadmin
	@echo "PgAdmin available at http://localhost:$(PGADMIN_PORT)"
	@echo "Login: $(PGADMIN_EMAIL) / $(PGADMIN_PASSWORD)"

# Migration commands
migrate-up: ## Run all pending migrations
	@echo "Running migrations..."
	@go run cmd/migrate/main.go up
 migrate-down: ## Rollback last migration @echo "Rolling back last migration..."
	@go run cmd/migrate/main.go down

migrate-version: ## Show current migration version
	@go run cmd/migrate/main.go version

migrate-force: ## Force migration to specific version (use with caution)
	@read -p "Enter version number: " version; \
	go run cmd/migrate/main.go force $$version

migration-create: ## Create new migration (usage: make migration-create name=add_users_table)
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migration-create name=add_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	@migrate create -ext sql -dir db/migrations -seq $(name)
	@echo "Migration files created in db/migrations/"

# Development helpers
db-reset: db-clean db-up migrate-up ## Reset database (WARNING: deletes all data and re-runs migrations)
	@echo "Database has been reset!"

db-seed: ## Run seed data (after migrations)
	@echo "Seeding database..."
	@docker-compose exec postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -f /docker-entrypoint-initdb.d/seed.sql

test-db: ## Run database tests
	@echo "Running database tests..."
	@go test ./db/... -v

# Environment setup
setup: ## Initial setup (copy .env.example, start db, run migrations)
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env file created. Please review and update if needed."; \
	fi
	@make db-up
	@make migrate-up
	@echo "Setup complete!"

.DEFAULT_GOAL := help


