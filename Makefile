# Show help for all tasks
.PHONY: help
help:
	@echo "Available tasks:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Go backend tasks
.PHONY: build test fmt lint

build: ## Build the Go backend
	@echo "Building Go backend..."
	go build -v -tags "dev fts5 webkit2_41" -o build/bin/lyyyra .

build-prod: ## Build production binary with optimizations
	@echo "Building production binary..."
	go build -v -tags "fts5 webkit2_41" -ldflags="-s -w" -o build/bin/lyyyra .

test: ## Run Go tests
	@echo "Running Go tests..."
	gotestsum --format testname -- -tags "fts5" .

test-verbose: ## Run Go tests with full output
	@echo "Running Go tests (verbose)..."
	gotestsum --format standard-verbose -- -tags "fts5" .

fmt: ## Format Go code
	go fmt ./...

lint: ## Lint Go code (requires golangci-lint)
	golangci-lint run

# Node.js/Frontend tasks
.PHONY: frontend-install frontend-build frontend-dev frontend-test

frontend-install: ## Install frontend dependencies
	cd frontend && npm install

frontend-build: ## Build frontend for production
	cd frontend && npm run build

frontend-dev: ## Start frontend development server
	cd frontend && npm run dev

frontend-test: ## Run frontend tests
	cd frontend && npm test -- --run

frontend-test-watch: ## Run frontend tests in watch mode
	cd frontend && npm test

frontend-test-ui: ## Run frontend tests with UI
	cd frontend && npm run test:ui

# Clean up build artifacts
.PHONY: clean test-all install-tools

clean: ## Clean Go and frontend build artifacts
	go clean
	rm -rf frontend/dist

test-all: test frontend-test ## Run all tests (Go + frontend)
	cd frontend && npm test || true

install-tools: ## Install Go test and lint tools
	go install gotest.tools/gotestsum@v1.11.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2

# Wails tasks
.PHONY: wails-dev wails-build wails-build-windows wails-install
wails-dev: ## Start Wails development server
	wails dev

wails-build: ## Build Wails application for production
	wails build

wails-build-windows: ## Build Wails application for Windows
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ wails build -platform windows/amd64 -tags "fts5 webkit2_41"

wails-build-windows-skip-frontend: ## Build Windows app (skip frontend rebuild)
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ wails build -platform windows/amd64 -tags "fts5 webkit2_41" -skipbindings -s

wails-build-windows-fast: frontend-build wails-build-windows-skip-frontend ## Fast Windows build (pre-build frontend)

wails-install: ## Install Wails CLI
	go install github.com/wailsapp/wails/v2/cmd/wails@latest