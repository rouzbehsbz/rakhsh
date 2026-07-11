ifneq (,$(wildcard .env))
    include .env
    export
endif

SQLC_CONFIG_PATH=./db/postgres/sqlc.yml
POSTGRES_MIGRATION_DIR=./db/postgres/migrations
DATABASE_URL=postgres://$(POSTGRES_USERNAME):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE_NAME)?sslmode=disable

run-core:
	@echo "Running core application ..."
	go run cmd/core/main.go

sqlc-generate:
	@echo "Generating sqlc source code ..."
	@sqlc generate -f $(SQLC_CONFIG_PATH)
	@echo "Done"

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "'name' argument is required"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir $(MIGRATION_DIR) $(name)

migrate-deploy:
	@echo "Deploying database migrations ..."
	@migrate -path $(MIGRATION_DIR) -database $(DATABASE_URL) up