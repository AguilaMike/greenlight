# Include variables from the .env file
include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# Create the new confirm target.
.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

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

## lint: run linters
.PHONY: lint
lint:
	@echo 'Running linters...'
	golangci-lint run

## test: run tests
.PHONY: test
test:
	@echo 'Running tests...'
	go tests ./...

## coverage: run tests and generate coverage report
.PHONY: coverage
coverage:
	@echo 'Running tests and generating coverage report...'
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out && rm coverage.out

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format all .go files and tidy module dependencies
.PHONY: tidy
tidy:
	@echo 'Formatting .go files...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy
	@echo 'Verifying and vendoring module dependencies...'
	go mod verify
	go mod vendor

## tidy/clean: remove all the temporary files
.PHONY: tidy/clean
tidy/clean:
	@echo 'Removing temporary files...'
	go clean -cache -testcache -modcache

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...
