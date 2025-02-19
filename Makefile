# --------------------------------------------------------------------------------------------------------------#
# Load environment variables
# --------------------------------------------------------------------------------------------------------------#

include .envrc

# --------------------------------------------------------------------------------------------------------------#
# HELPERS
# --------------------------------------------------------------------------------------------------------------#

## help: Print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## confirm: Ask for confirmation before running a command
.PHONY: confirm 
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

# --------------------------------------------------------------------------------------------------------------#
# Development
# --------------------------------------------------------------------------------------------------------------#

## run/api: Run the cmd/api application
.PHONY: run/shorten
run/shorten:
	go run ./shorten/cmd/ -db-dsn=${URL_SHORTENER_DSN}

.PHONY: run/redirect
run/redirect:
	go run ./redirect/cmd/ -db-dsn=${URL_SHORTENER_DSN}

.PHONY: run/gateway
run/gateway:
	go run ./shortener-gateway/cmd/ -cors-trusted-origin="http://localhost:5173"


# --------------------------------------------------------------------------------------------------------------#
# Database Operations
# --------------------------------------------------------------------------------------------------------------#

## db/psql: Connect to the PostgreSQL database using psql
.PHONY: db/psql
db/psql:
	psql ${URL_SHORTENER_DSN}

# --- Database Migrations ---

# Default values (can be overridden when running make)
MIGRATION_DIR ?= ./migration  # Default migration directory

## db/migration/new: Create a new set of database migrations (requires name)
.PHONY: db/migration/new
db/migration/new:
ifndef name
	$(error "Usage: make db/migration/new name=your_migration_name")
endif
	@echo "Creating migration file for '${name}' in $(MIGRATION_DIR)"
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(name)

## db/migration/up: Apply all 'up' database migrations (default: ./migration)
.PHONY: db/migration/up
db/migration/up: confirm
	@echo "Running up migrations in $(MIGRATION_DIR)..."
	migrate -path $(MIGRATION_DIR) -database ${URL_SHORTENER_DSN} up

## db/migration/down: Rollback the last migration
.PHONY: db/migration/down
db/migration/down: confirm
	@echo "Rolling back last migration in $(MIGRATION_DIR)..."
	migrate -path $(MIGRATION_DIR) -database ${URL_SHORTENER_DSN} down 1
