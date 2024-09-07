include .env

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# Create the new confirm target.
.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api

## migrate/new name=$1: create a new database migration
.PHONY: migrate/new
migrate/new:
	@echo 'Creating new migration...'
	migrate create -seq -ext .sql -dir ./scripts/migrations $(name)

## migrate/up: apply all up database migrations
.PHONY: migrate/up
migrate/up:
	@echo 'Running up migrations...'
	migrate -path ./scripts/migrations -database ${DB_DSN} up

## migrate/down: apply all down database migrations
.PHONY: migrate/down
migrate/down: confirm
	@echo 'Running up migrations...'
	migrate -path ./scripts/migrations -database ${DB_DSN} down
