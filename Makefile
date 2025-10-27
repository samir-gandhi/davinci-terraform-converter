# Makefile for DaVinci Terraform Converter

.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the plugin binary
	go build -o davinci-convert .

.PHONY: install
install: build ## Build and install the plugin binary to GOBIN
	go install .

.PHONY: test-all
test-all: ## Run all tests
	go test -tags acceptance ./...


.PHONY: test-verbose
test-verbose: ## Run all tests with verbose output
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -tags acceptance -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: clean
clean: ## Clean build artifacts and test outputs
	rm -f davinci-convert
	rm -f coverage.out coverage.html
	go clean -testcache

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: fmt vet ## Run all linting tools

.PHONY: deps
deps: ## Download dependencies
	go mod download
	go mod tidy

.PHONY: all
all: clean deps lint test-all build ## Run all checks and build

.DEFAULT_GOAL := help
