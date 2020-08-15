.DEFAULT_GOAL:=help

ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
MIGRATIONS_DIR := $(ROOT_DIR)/script/db/migrations
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)
TARGET_MAX_CHAR_NUM := 15

POSTGRES_USER := app
POSTGRES_PASSWORD := app121
POSTGRES_HOST := localhost
POSTGRES_PORT := 5432
POSTGRES_DB := crius
POSTGRESQL_URL := "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"

CRIUS_PORT := 3000

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

.PHONY: debug
## Print internal Makefile variables
debug:
	@echo ROOT_DIR=$(ROOT_DIR)
	@echo MIGRATIONS_DIR=$(MIGRATIONS_DIR)
	@echo POSTGRES_USER=$(POSTGRES_USER)
	@echo POSTGRES_PASSWORD=$(POSTGRES_PASSWORD)
	@echo POSTGRES_HOST=$(POSTGRES_HOST)
	@echo POSTGRES_PORT=$(POSTGRES_PORT)
	@echo POSTGRES_DB=$(POSTGRES_DB)
	@echo POSTGRESQL_URL=$(POSTGRESQL_URL)
	@echo CRIUS_PORT=$(CRIUS_PORT)

.PHONY: tidy
## Tidy up go code
tidy:
	go mod tidy
	go fmt ./...

.PHONY: service/build
## Build the Crius HTTP server
service/build:
	go build ./...

.PHONY: service/run
## Run the Crius HTTP server. Will run with gin (https://github.com/codegangsta/gin) if available
service/run:
	MIGRATIONS_DIR=$(MIGRATIONS_DIR) APP_DB_USERNAME=$(POSTGRES_USER) APP_DB_PASSWORD=$(POSTGRES_PASSWORD) APP_DB_NAME=$(POSTGRES_DB) PORT=$(CRIUS_PORT) go run $(ROOT_DIR)/internal/cmd/main/main.go

.PHONY: service/test
## Runt the tests
service/test:
	MIGRATIONS_DIR=$(MIGRATIONS_DIR) APP_DB_USERNAME=$(POSTGRES_USER) APP_DB_PASSWORD=$(POSTGRES_PASSWORD) APP_DB_NAME=$(POSTGRES_DB) go test -v ./...

.PHONY: db/await
# Internal target, waits for the DB to come up
db/await:
	@db_up=0; \
	while [ $${db_up} -eq 0 ]; do \
		echo 'Waiting for DB to come up ...'; \
		db_up=`docker exec -t criusdb /bin/bash -c 'psql $(POSTGRESQL_URL) -c "SELECT 1"' > /dev/null 2>&1 && echo 1 || echo 0`; \
		if [ $$db_up -eq 1 ]; then echo 'DB is up'; else sleep 1; fi \
	done

.PHONY: db/run
## Start the database
db/run:
	@docker rm -f criusdb > /dev/null 2>&1 || true
	docker run --name criusdb -e POSTGRES_USER=$(POSTGRES_USER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -e POSTGRES_DB=$(POSTGRES_DB) -p $(POSTGRES_PORT):5432 -d postgres:13
	@$(MAKE) db/await

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
