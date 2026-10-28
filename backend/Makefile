# Abudemy backend - developer tasks
# Usage: make <target>

GO ?= go
BINARY ?= bin/abudemy-api
PKG := ./...
TEST_DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/abudemy_test?sslmode=disable

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## Tidy go modules
	$(GO) mod tidy

.PHONY: fmt
fmt: ## Format the code
	gofmt -w .

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKG)

.PHONY: build
build: ## Build the API binary
	$(GO) build -trimpath -o $(BINARY) ./cmd/api

.PHONY: run
run: ## Run the API locally (reads .env)
	$(GO) run ./cmd/api

.PHONY: test
test: ## Run unit tests
	$(GO) test $(PKG)

.PHONY: test-integration
test-integration: ## Run integration tests (requires TEST_DATABASE_URL)
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" $(GO) test $(PKG) -run TestIntegration -v

.PHONY: cover
cover: ## Run tests with a coverage report
	$(GO) test $(PKG) -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

.PHONY: docker-up
docker-up: ## Start the local stack (PostgreSQL + API)
	docker compose -f ../docker-compose.yml up --build

.PHONY: docker-down
docker-down: ## Stop the local stack
	docker compose -f ../docker-compose.yml down

.PHONY: docker-testdb
docker-testdb: ## Start a throwaway PostgreSQL for integration tests
	docker run -d --name abudemy-pg-test -e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres \
		-e POSTGRES_DB=abudemy_test -p 5432:5432 postgres:16-alpine
