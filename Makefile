# Show help for all tasks
.PHONY: help
help:
ifeq ($(OS),Windows_NT)
	@echo "Available tasks:"
	@powershell -NoProfile -Command "$$re='^[a-zA-Z_-]+:.*?## '; Get-Content Makefile | Where-Object { $$_ -match $$re } | ForEach-Object { if ($$_ -match '^(?<name>[a-zA-Z_-]+):.*?## (?<desc>.*)') { $$name=$$Matches['name']; $$desc=$$Matches['desc']; $$pad=20; $$cyan=[char]27 + '[36m'; $$reset=[char]27 + '[0m'; $$fmt='  {0,-' + $$pad + '} {1}'; Write-Output ($$fmt -f ($$cyan + $$name + $$reset), $$desc) } }"
else
	@echo "Available tasks:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
endif

# Build tags
# Devcontainer ships with WebKitGTK 4.1 libraries, so default to that.
WEBKIT_TAG ?= webkit2_41
GO_DEV_TAGS ?= dev $(WEBKIT_TAG)
GO_PROD_TAGS ?= $(WEBKIT_TAG)

# Go backend tasks
.PHONY: build test fmt lint

build: ## Build the Go backend
	@echo "Building Go backend..."
	go build -v -tags "$(GO_DEV_TAGS)" -o build/bin/lyyyra .

build-prod: ## Build production binary with optimizations
	@echo "Building production binary..."
	go build -v -tags "$(GO_PROD_TAGS)" -ldflags="-s -w" -o build/bin/lyyyra .

test: ## Run Go tests
	@echo "Running Go tests..."
	gotestsum --format testname -- -tags "$(GO_PROD_TAGS)" ./internal/...

test-verbose: ## Run Go tests with full output
	@echo "Running Go tests (verbose)..."
	gotestsum --format standard-verbose -- -tags "$(GO_PROD_TAGS)" ./internal/...
fmt: ## Format Go code
	go fmt ./internal/...

lint: ## Lint Go code (requires golangci-lint)
	@echo "Linting Go code in ./internal/..."
	golangci-lint run ./internal/... -v
	@echo "Linting TypeScript in frontend/src..."
	cd frontend && npm run lint

# Node.js/Frontend tasks
.PHONY: frontend-install frontend-build frontend-dev frontend-test

frontend-install: ## Install frontend dependencies
	cd frontend && npm install

frontend-build: ## Build frontend for production
	cd frontend && npm run build

frontend-dev: ## Start frontend development server
	cd frontend && npm run dev

frontend-test: ## Run frontend tests (non-watch)
	cd frontend && npm run test:run

frontend-test-watch: ## Run frontend tests in watch mode
	cd frontend && npm test


clean: ## Clean Go and frontend build artifacts
	go clean
	rm -rf frontend/dist 2>/dev/null || powershell -Command "Remove-Item -Recurse -Force -ErrorAction Ignore 'frontend/dist'" 2>/dev/null || true

clean-data: ## Delete local app data (database, songs, status, logs)
	@echo "Deleting app data in ~/Lyyyra..."
	rm -rf ~/Lyyyra 2>/dev/null || powershell -Command "Remove-Item -Recurse -Force -ErrorAction Ignore '$$HOME/Lyyyra'" 2>/dev/null || true
	@echo "App data deleted. Next run will start fresh."

test-all: test frontend-test ## Run all tests (Go + frontend)

test-coverage: ## Run Go tests with coverage
	@echo "Running Go tests with coverage..."
	go test -tags "$(GO_PROD_TAGS)" -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

frontend-test-coverage: ## Run frontend tests with coverage
	@echo "Running frontend tests with coverage..."
	cd frontend && npm run test:coverage -- --run

test-all-coverage: test-coverage frontend-test-coverage ## Run all tests with coverage

install-tools: ## Install Go test, lint tools, and act for CI testing
	go install gotest.tools/gotestsum@v1.11.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
	@echo "Installing frontend coverage tool..."
	cd frontend && npm install -D @vitest/coverage-v8
	@echo "Installing act for local CI testing..."
	@curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Wails tasks
.PHONY: wails-dev wails-build wails-build-windows wails-install

wails-dev: ## Start Wails development server (with devtools enabled)
ifeq ($(OS),Windows_NT)
	wails dev -tags "$(GO_DEV_TAGS)"
else
	@if [ -z "$$DISPLAY" ] && command -v xvfb-run >/dev/null 2>&1; then \
		echo "[wails-dev] DISPLAY not set, starting via xvfb-run"; \
		xvfb-run -a wails dev -tags "$(GO_DEV_TAGS)" -devtools; \
	else \
		wails dev -tags "$(GO_DEV_TAGS)"; \
	fi
endif

wails-build: ## Build Wails application for production
	wails build -tags "$(GO_PROD_TAGS)"-devtools

wails-build-windows-skip-frontend: ## Build Windows app (skip frontend rebuild, devcontainer cross-compile)
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ wails build -platform windows/amd64 -tags "$(GO_PROD_TAGS)" -devtools -skipbindings -s

wails-build-windows: frontend-build wails-build-windows-skip-frontend ## Build Wails for Windows (cross-compile in devcontainer)



wails-install: ## Install Wails CLI
	go install github.com/wailsapp/wails/v2/cmd/wails@latest

# CI/CD testing
.PHONY: ci-test

ci-test: ## Test GitHub Actions workflow locally (requires act and sudo access)
	@if ! command -v act >/dev/null 2>&1; then \
		echo "Installing act locally..."; \
		curl -fsSL https://raw.githubusercontent.com/nektos/act/master/install.sh | bash; \
	fi
	@echo "Running GitHub Actions workflow locally..."
	sudo -E $(shell command -v act) workflow_dispatch -j test --artifact-server-path /tmp/artifacts