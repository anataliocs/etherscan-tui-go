.PHONY: all lint test test-e2e vulncheck help

all: lint vulncheck test test-e2e ## Run all checks

lint: ## Run linter
	golangci-lint run ./...

vulncheck: ## Run vulnerability check
	govulncheck ./...

test: ## Run unit tests
	go test ./... -v

test-e2e: ## Run E2E tests
	go test ./test/... -v

help: ## Show help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
