.PHONY: help proto sqlc migrate-up migrate-down db-start db-stop test build run clean

DB_URL=postgres://healthapp_user:verysecretpassword@localhost:5432/healthapp_db?sslmode=disable
MIGRATIONS_DIR=./db/migrations

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

proto: ## Generate connect-rpc Go code and OpenAPI spec from proto files
	@echo "Generating Protobuf/Connect code and OpenAPI spec..."
	buf generate api/proto

sqlc: ## Generate Go code from SQL queries using sqlc
	@echo "Generating sqlc code..."
	sqlc generate

# migrate-create: NAME?=new_migration ## Create new migration files (e.g., make migrate-create NAME=add_indexes)
# 	@echo "Creating migration: $(NAME)"
# 	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

migrate-up: ## Apply all database migrations
	@echo "Applying migrations..."
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) up

migrate-down: ## Revert the last database migration
	@echo "Reverting last migration..."
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down 1

db-start: ## Start PostgreSQL container via Docker Compose
	@echo "Starting PostgreSQL container..."
	docker compose up -d postgres

db-stop: ## Stop PostgreSQL container
	@echo "Stopping PostgreSQL container..."
	docker compose down

test: ## Run Go tests with colored output
	@echo "Running tests..."
	richgo test -v -race ./...

build: ## Build the server binary
	@echo "Building server..."
	mkdir -p bin
	go build -o bin/server ./cmd/server

run: build ## Build and run the server
	@echo "Running server..."
	./bin/server serve

seed: build ## Seed the database with mock data
	@echo "Seeding database with mock data..."
	./bin/server seed

clean: ## Clean generated files and build artifacts
	@echo "Cleaning generated files and build artifacts..."
	rm -rf bin
	rm -rf internal/infrastructure/persistence/postgres/db/*
	rm -rf internal/infrastructure/rpc/gen/*
	rm -rf third_party/openapi/*
	# Keep .gitkeep files if any
	find internal/infrastructure/persistence/postgres/db/ -type f ! -name '.gitkeep' -delete
	find internal/infrastructure/rpc/gen/ -type f ! -name '.gitkeep' -delete
	find third_party/openapi/ -type f ! -name '.gitkeep' -delete


# Combine setup steps
setup-tools:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/kyoh86/richgo@latest

generate-all: proto sqlc ## Generate all code (protobuf, connect, sqlc)

init-db: db-start ## Initialize DB: Start container & run migrations
	@echo "Waiting for DB to be ready..."
	@until docker compose exec postgres pg_isready -U healthapp_user -d healthapp_db -q; do \
		sleep 1; \
	done
	$(MAKE) migrate-up
