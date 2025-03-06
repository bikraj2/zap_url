# --------------------------------------------------------------------------------------------------------------#
# Load environment variables
# --------------------------------------------------------------------------------------------------------------#

include .envrc

# --------------------------------------------------------------------------------------------------------------#
# HELPER
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
	go run ./gateway/cmd/ -cors-trusted-origin="http://localhost:5173"


SERVICES =  gateway kgs shorten redirect
# Default target (build all services)
build/docker: $(SERVICES)
	docker build -t postgres -f Dockerfile.postgres .
# Rule to build each service
$(SERVICES):
	docker build -t $@ -f $@/Dockerfile .
.PHONY: build/docker $(SERVICES) 

.PHONY: run/docker
run/docker:
	- docker run -d --restart unless-stopped --name shorten --network url_network -p 127.0.0.1:8081:8081 shorten
	- docker run -d --restart unless-stopped --name redirect --network url_network -p 127.0.0.1:8082:8082 redirect
	- docker run -d --restart unless-stopped --name kgs --network url_network -p 127.0.0.1:8080:8080 kgs
	- docker run -d --restart unless-stopped --name  gateway --network url_network -p  127.0.0.1:8084:8084 gateway
	- docker run -d --restart unless-stopped --name redis-rebloom --network url_network -p 127.0.0.1:6379:6379 goodform/rebloom:latest 
	- docker run -d \
		--network url_network \
		-p 127.0.0.1:8500:8500 \
		-p 127.0.0.1:8600:8600/udp \
		--name=dev-consul \
		hashicorp/consul:latest \
		agent -server -ui \
		-node=server-1 \
		-bootstrap-expect=1 \
		-client=0.0.0.0
	- docker run -d --restart unless-stopped \
		--name pg-container \
		--network url_network \
		-p 127.0.0.1:5432:5432 \
		-v pg_data:/var/lib/postgresql/data \
		postgres
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

