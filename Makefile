ifneq (,$(wildcard .env))
    include .env
    export
endif

SQLC_CONFIG_PATH=./db/postgres/sqlc.yml
POSTGRES_MIGRATION_DIR=./db/postgres/migrations
POSTGRES_URL=postgres://$(POSTGRES_USERNAME):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE_NAME)?sslmode=disable

run-core:
	@echo "Running core application ..."
	go run cmd/core/main.go

build-core:
	@echo "Building the core project ..."
	go build -o .bin/core cmd/core/main.go
	@echo "Build Completed"

sqlc-generate:
	@echo "Generating sqlc source code ..."
	@sqlc generate -f $(SQLC_CONFIG_PATH)
	@echo "Done"

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "'name' argument is required"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir $(POSTGRES_MIGRATION_DIR) $(name)

migrate-deploy:
	@echo "Deploying database migrations ..."
	@migrate -path $(POSTGRES_MIGRATION_DIR) -database $(POSTGRES_URL) up