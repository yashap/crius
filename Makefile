.DEFAULT_GOAL:=help

ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
MIGRATIONS_DIR := $(ROOT_DIR)/scripts/db/migrations
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)
TARGET_MAX_CHAR_NUM = 15

POSTGRES_USER = app
POSTGRES_PASSWORD = app121
POSTGRES_HOST = localhost
POSTGRES_PORT = 5432
POSTGRES_DB = crius
POSTGRESQL_URL = "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"

GIN_INSTALLED := $(shell type -p gin >/dev/null 2>&1 && echo 1 || echo 0)

.PHONY: help
## Show this help
help:
	@printf 'Usage:\n  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}\n\n'
	@printf 'Targets:\n'
	@awk '/^[a-zA-Z\-\_\/0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-$(TARGET_MAX_CHAR_NUM)s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.PHONY: format
## Format all go code
fmt:
	go fmt ./...

.PHONY: service/build
## Build the Crius HTTP server
service/build:
	go build ./...

.PHONY: service/run
## Run the Crius HTTP server. Will run with gin (https://github.com/codegangsta/gin) if available
service/run:
	MIGRATIONS_DIR=$(MIGRATIONS_DIR) APP_DB_USERNAME=$(POSTGRES_USER) APP_DB_PASSWORD=$(POSTGRES_PASSWORD) APP_DB_NAME=$(POSTGRES_DB) PORT=3000 go run internal/cmd/main/main.go

.PHONY: service/test
## Runt the tests
service/test:
	MIGRATIONS_DIR=$(MIGRATIONS_DIR) APP_DB_USERNAME=$(POSTGRES_USER) APP_DB_PASSWORD=$(POSTGRES_PASSWORD) APP_DB_NAME=$(POSTGRES_DB) go test -v ./...

.PHONY: db/run
## Start the database
db/run:
	@docker rm -f criusdb > /dev/null 2>&1 || true
	docker run --name criusdb -e POSTGRES_USER=$(POSTGRES_USER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -e POSTGRES_DB=$(POSTGRES_DB) -p $(POSTGRES_PORT):5432 -d postgres:13
	# TODO: better way to check if postgres is up
	sleep 5

.PHONY: migrate/up
## Run all DB migrations
migrate/up:
	migrate -database $(POSTGRESQL_URL) -path $(MIGRATIONS_DIR) up

.PHONY: migrate/down
## Reverse all DB migrations
migrate/down:
	migrate -database $(POSTGRESQL_URL) -path $(MIGRATIONS_DIR) down

.PHONY: migrate/new
## Create a new DB migration script. Must set MIGRATE_FILENAME, e.g. `MIGRATE_FILENAME=create_products_table make migrate/new`
migrate/new:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(MIGRATE_FILENAME)
