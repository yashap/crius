.DEFAULT_GOAL:=help

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)
TARGET_MAX_CHAR_NUM := 15

ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
POSTGRES_MIGRATIONS_DIR := $(ROOT_DIR)/script/postgresql/migrations
MYSQL_MIGRATIONS_DIR := $(ROOT_DIR)/script/mysql/migrations

POSTGRES_USER := app
POSTGRES_PASSWORD := app121
POSTGRES_HOST := localhost
POSTGRES_PORT := 5432
POSTGRES_DB := crius
POSTGRES_URL := "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"

MYSQL_USER := app
MYSQL_PASSWORD := app121
MYSQL_ROOT_PASSWORD := topsecret
MYSQL_HOST := 127.0.0.1
MYSQL_PORT := 3306
MYSQL_DB := crius
MYSQL_URL := "mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@$(MYSQL_HOST):$(MYSQL_PORT)/$(MYSQL_DB)?multiStatements=true"
MYSQL_MIGRATION_URL := "mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DB)"

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
	@echo POSTGRES_MIGRATIONS_DIR=$(POSTGRES_MIGRATIONS_DIR)
	@echo MYSQL_MIGRATIONS_DIR=$(MYSQL_MIGRATIONS_DIR)
	@echo POSTGRES_USER=$(POSTGRES_USER)
	@echo POSTGRES_PASSWORD=$(POSTGRES_PASSWORD)
	@echo POSTGRES_HOST=$(POSTGRES_HOST)
	@echo POSTGRES_PORT=$(POSTGRES_PORT)
	@echo POSTGRES_DB=$(POSTGRES_DB)
	@echo POSTGRES_URL=$(POSTGRES_URL)
	@echo MYSQL_USER=$(MYSQL_USER)
	@echo MYSQL_PASSWORD=$(MYSQL_PASSWORD)
	@echo MYSQL_HOST=$(MYSQL_HOST)
	@echo MYSQL_PORT=$(MYSQL_PORT)
	@echo MYSQL_DB=$(MYSQL_DB)
	@echo MYSQL_URL=$(MYSQL_URL)
	@echo MYSQL_MIGRATION_URL=$(MYSQL_MIGRATION_URL)
	@echo CRIUS_PORT=$(CRIUS_PORT)

.PHONY: tidy
## Tidy up go code
tidy:
	go mod tidy
	go fmt ./...

.PHONY: generate
## Run code generation
generate:
	sqlboiler --wipe --no-tests --output internal/db/postgresql/dao psql
	sqlboiler --wipe --no-tests --output internal/db/mysql/dao mysql

.PHONY: build-service
## Build the Crius HTTP server
build-service:
	go build ./...

.PHONY: run
## Run the DB (Postgres) and HTTP server (will wipe local DB)
run: pg-run-db run-service

.PHONY: run-service
## Run the Crius HTTP server (against Postgres)
run-service:
	CRIUS_DB_URL=$(POSTGRES_URL) CRIUS_MIGRATIONS_DIR=$(POSTGRES_MIGRATIONS_DIR) PORT=$(CRIUS_PORT) go run $(ROOT_DIR)/internal/cmd/main/main.go

.PHONY: run
## Run the DB (MySQL) and HTTP server (will wipe local DB)
run-mysql: mysql-run-db run-service-mysql

.PHONY: run-service
## Run the Crius HTTP server (against MySQL)
run-service-mysql:
	CRIUS_DB_URL=$(MYSQL_URL) CRIUS_MIGRATIONS_DIR=$(MYSQL_MIGRATIONS_DIR) PORT=$(CRIUS_PORT) go run $(ROOT_DIR)/internal/cmd/main/main.go

.PHONY: test
## Run all unit and integration tests
test:
	go test -v ./...

.PHONY: mysql-run-db
## Start the MySQL DB
mysql-run-db:
	@docker rm -f criusmysql > /dev/null 2>&1 || true
	docker run --name criusmysql -e MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) -e MYSQL_DATABASE=$(MYSQL_DB) -p $(MYSQL_PORT):$(MYSQL_PORT) -d mysql:8 -h 127.0.0.1
	@$(MAKE) mysql-await-db
	@$(MAKE) mysql-create-user

# Internal target, creates MySQL users
mysql-create-user:
	docker cp script/mysql/devusers/create_user.sql criusmysql:/usr/local/bin/create_user.sql
	docker exec -t criusmysql /bin/bash -c 'mysql -u root -p$(MYSQL_ROOT_PASSWORD) -h $(MYSQL_HOST) -P $(MYSQL_PORT) $(MYSQL_DB) < /usr/local/bin/create_user.sql'

.PHONY: mysql-await-db
# Internal target, waits for the MySQL DB to come up
mysql-await-db:
	@db_up=0; \
	while [ $${db_up} -eq 0 ]; do \
		echo 'Waiting for DB to come up ...'; \
		db_up=`docker exec -t criusmysql /bin/bash -c 'mysql -u root -p$(MYSQL_ROOT_PASSWORD) -h $(MYSQL_HOST) -P $(MYSQL_PORT) $(MYSQL_DB) -e "SELECT 1;"' > /dev/null 2>&1 && echo 1 || echo 0`; \
		if [ $$db_up -eq 1 ]; then echo 'DB is up'; else sleep 1; fi \
	done

.PHONY: mysql-migrate-up
## Run all MySQL DB migrations (rarely used, the app does this on startup)
mysql-migrate-up:
	migrate -database $(MYSQL_MIGRATION_URL) -path $(MYSQL_MIGRATIONS_DIR) up

.PHONY: mysql-migrate-down
## Reverse all MySQL DB migrations
mysql-migrate-down:
	migrate -database $(MYSQL_MIGRATION_URL) -path $(MYSQL_MIGRATIONS_DIR) down

.PHONY: mysql-migrate-new
## Create a new MySQL DB migration script. Must set MIGRATE_FILENAME, e.g. `MIGRATE_FILENAME=create_products_table make mysql-migrate-new`
mysql-migrate-new:
	migrate create -ext sql -dir $(MYSQL_MIGRATIONS_DIR) -seq $(MIGRATE_FILENAME)

.PHONY: pg-run-db
## Start the Postgres DB
pg-run-db:
	@docker rm -f criuspostgresql > /dev/null 2>&1 || true
	docker run --name criuspostgresql -e POSTGRES_USER=$(POSTGRES_USER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -e POSTGRES_DB=$(POSTGRES_DB) -p $(POSTGRES_PORT):$(POSTGRES_PORT) -d postgres:13
	@$(MAKE) pg-await-db

.PHONY: pg-await-db
# Internal target, waits for the Postgres DB to come up
pg-await-db:
	@db_up=0; \
	while [ $${db_up} -eq 0 ]; do \
		echo 'Waiting for DB to come up ...'; \
		db_up=`docker exec -t criuspostgresql /bin/bash -c 'psql $(POSTGRES_URL) -c "SELECT 1"' > /dev/null 2>&1 && echo 1 || echo 0`; \
		if [ $$db_up -eq 1 ]; then echo 'DB is up'; else sleep 1; fi \
	done

.PHONY: pg-migrate-up
## Run all Postgres DB migrations (rarely used, the app does this on startup)
pg-migrate-up:
	migrate -database $(POSTGRES_URL) -path $(POSTGRES_MIGRATIONS_DIR) up

.PHONY: pg-migrate-down
## Reverse all Postgres DB migrations
pg-migrate-down:
	migrate -database $(POSTGRES_URL) -path $(POSTGRES_MIGRATIONS_DIR) down

.PHONY: pg-migrate-new
## Create a new Postgres DB migration script. Must set MIGRATE_FILENAME, e.g. `MIGRATE_FILENAME=create_products_table make pg-migrate-new`
pg-migrate-new:
	migrate create -ext sql -dir $(POSTGRES_MIGRATIONS_DIR) -seq $(MIGRATE_FILENAME)
