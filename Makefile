.PHONY: help db-up db-down db-logs db-shell migrate-up migrate-down migrate-version migration-create ingest test-llm reset dev

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
	@go run ./cmd/migrate/main.go up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@go run ./cmd/migrate/main.go down

migrate-version: ## Show current migration version
	@go run ./cmd/migrate/main.go version

migrate-force: ## Force migration to specific version (use with caution)
	@read -p "Enter version number: " version; \
	go run ./cmd/migrate/main.go force $$version

migration-create: ## Create new migration (usage: make migration-create name=add_users_table)
	@if [ -z "$(name)" ]; then \
	   echo "Error: name is required. Usage: make migration-create name=add_users_table"; \
	   exit 1; \
	fi
	@echo "Creating migration: $(name)"
	@migrate create -ext sql -dir db/migrations -seq $(name)
	@echo "Migration files created in db/migrations/"

# Application commands
ingest: ## Run data ingestion (fetch and load from connectors)
	@echo "Running data ingestion..."
	@go run ./cmd/purepulse/main.go

test-llm: ## Test LLM end-to-end pipeline
	@echo "Testing LLM pipeline..."
	@. ./.env && go run ./cmd/test-llm/main.go

# Development helpers
db-reset: db-clean db-up migrate-up ## Reset database (WARNING: deletes all data and re-runs migrations)
	@echo "Database has been reset!"

reset: migrate-down migrate-up ## Rollback and re-apply migrations (keeps data)
	@echo "Migrations reset!"

dev: migrate-up ingest test-llm ## Full development workflow (migrate → ingest → test)
	@echo "Development workflow complete!"

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

helptest-weekly-metrics: ## Test weekly metrics aggregation
	@echo "Testing weekly metrics aggregation..."
	@. ./.env && go run ./cmd/test-weekly-metrics/main.go

refresh-views: ## Refresh materialized views
	@echo "Refreshing materialized views..."
	@. ./.env && psql $(DATABASE_URL) -c "REFRESH MATERIALIZED VIEW CONCURRENTLY daily_user_activity;"
	@echo "✅ Views refreshed"

check-views: ## Check materialized view data
	@. ./.env && psql $(DATABASE_URL) << EOF
	SELECT 'daily_user_activity' as view_name, COUNT(*) as rows FROM daily_user_activity
	UNION ALL
	SELECT 'weekly_team_activity', COUNT(*) FROM weekly_team_activity
	UNION ALL
	SELECT 'user_correlation_summary', COUNT(*) FROM user_correlation_summary;
	EOF


test-metrics-refresh: ## Just refresh views
	@. ./.env && psql $(DATABASE_URL) << EOF
	REFRESH MATERIALIZED VIEW CONCURRENTLY daily_user_activity;
	REFRESH MATERIALIZED VIEW CONCURRENTLY weekly_team_activity;
	REFRESH MATERIALIZED VIEW CONCURRENTLY user_correlation_summary;
	EOF
	@echo "Views refreshed"


test-metrics-full: ## Full diagnostic for weekly metrics
	@echo "Running full weekly metrics diagnostic..."
	@. ./.env && go run ./cmd/test-weekly-metrics/main.go

test-metrics-psql: ## Check view structure via psql
	@. ./.env && psql $$DATABASE_URL -c "\d daily_user_activity"
	@. ./.env && psql $$DATABASE_URL -c "\d weekly_team_activity"
	@. ./.env && psql $$DATABASE_URL -c "\d user_correlation_summary"

test-metrics-columns: ## Show exact columns in views
	@. ./.env && psql $$DATABASE_URL -c "SELECT table_name, column_name, data_type FROM information_schema.columns WHERE table_name IN ('daily_user_activity', 'weekly_team_activity', 'user_correlation_summary') ORDER BY table_name, ordinal_position;"

test-team-weekly: ## Test team weekly aggregation and Claude analysis
	@echo "Testing team weekly aggregation..."
	@. ./.env && go run ./cmd/test-team-weekly/main.go


test-user-weekly: ## Test all user weekly report generation
	@echo "Testing user weekly reports..."
	@. ./.env && go run ./cmd/test-user-weekly/main.go

# GraphQL testing commands
graphql-server: ## Start GraphQL server
	@echo "Starting GraphQL server..."
	@. ./.env && go run ./pkg/graphql/server.go

graphql-generate: graphql-deps  ## Generate GraphQL code (gqlgen)
	@echo "Generating GraphQL code..."
	@cd pkg/graphql && go run github.com/99designs/gqlgen generate

# Weekly report generation pipeline
generate-reports: refresh-views test-user-weekly test-team-weekly ## Generate all weekly reports (user + team)
	@echo "✅ All reports generated!"


# Complete end-to-end testing pipeline
full-stack: ## Complete setup: reset DB → ingest → generate reports → start GraphQL
	@echo "🚀 Starting full stack setup..."
	@echo "Step 1/8: Resetting database..."
	@make db-reset
	@echo "Step 2/8: Ingesting data from connectors..."
	@make ingest
	@echo "Step 3/8: Refreshing materialized views..."
	@make refresh-views
	@echo "Step 4/8: Generating user weekly reports..."
	@make test-user-weekly
	@echo "Step 5/8: Generating team weekly reports..."
	@make test-team-weekly
	@echo "Step 6/8: Generating GraphQL code..."
	@make graphql-generate
	@echo "Step 7/8: Verifying data..."
	@psql $(DATABASE_URL) -c "SELECT COUNT(*) as user_reports FROM weekly_reports;"
	@psql $(DATABASE_URL) -c "SELECT COUNT(*) as team_reports FROM team_weekly_reports;"
	@echo "Step 8/8: Starting GraphQL server..."
	@echo "✅ Setup complete! GraphQL playground at http://localhost:8080/"
	@make graphql-server

# Alternative: run server in background and open browser
full-stack-demo: ## Same as full-stack but opens browser automatically
	@make db-reset ingest refresh-views test-user-weekly test-team-weekly graphql-generate
	@echo "✅ All data generated! Starting server and opening browser..."
	@open http://localhost:8080/ 2>/dev/null || xdg-open http://localhost:8080/ 2>/dev/null || echo "Open http://localhost:8080/ in your browser"
	@make graphql-server

# GraphQL dependencies
graphql-deps: ## Install/update gqlgen dependencies
	@echo "Installing gqlgen dependencies..."
	@go get github.com/99designs/gqlgen/codegen@v0.17.81
	@go get github.com/99designs/gqlgen@v0.17.81
	@echo "✅ GraphQL dependencies installed"