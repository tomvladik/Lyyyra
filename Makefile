# Show help for all tasks
.PHONY: help
help:
	@echo "Available tasks:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build tags
# Devcontainer ships with WebKitGTK 4.1 libraries, so default to that.
WEBKIT_TAG ?= webkit2_41
GO_DEV_TAGS ?= dev fts5 $(WEBKIT_TAG)
GO_PROD_TAGS ?= fts5 $(WEBKIT_TAG)

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
	gotestsum --format testname -- -tags "$(GO_PROD_TAGS)" .

test-verbose: ## Run Go tests with full output
	@echo "Running Go tests (verbose)..."
	gotestsum --format standard-verbose -- -tags "$(GO_PROD_TAGS)" .

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

clean: ## Clean Go and frontend build artifacts
	go clean
	rm -rf frontend/dist 2>/dev/null || powershell -Command "Remove-Item -Recurse -Force -ErrorAction Ignore 'frontend/dist'" 2>/dev/null || true

clean-data: ## Delete local app data (database, songs, status, logs)
	@echo "Deleting app data in ~/Lyyyra..."
	rm -rf ~/Lyyyra 2>/dev/null || powershell -Command "Remove-Item -Recurse -Force -ErrorAction Ignore '$$HOME/Lyyyra'" 2>/dev/null || true
	@echo "App data deleted. Next run will start fresh."

test-all: test frontend-test ## Run all tests (Go + frontend)
	cd frontend && npm test || true

install-tools: ## Install Go test and lint tools
	go install gotest.tools/gotestsum@v1.11.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2

# Wails tasks
.PHONY: wails-dev wails-build wails-build-windows wails-install
wails-dev: ## Start Wails development server
	wails dev -tags "$(GO_DEV_TAGS)"

wails-build: ## Build Wails application for production
	wails build -tags "$(GO_PROD_TAGS)"

wails-build-windows: ## Build Wails for Windows (cross-compile in devcontainer)
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ wails build -platform windows/amd64 -tags "$(GO_PROD_TAGS)"

wails-build-windows-skip-frontend: ## Build Windows app (skip frontend rebuild, devcontainer cross-compile)
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ wails build -platform windows/amd64 -tags "$(GO_PROD_TAGS)" -skipbindings -s

wails-build-windows-fast: frontend-build wails-build-windows-skip-frontend ## Fast Windows build in devcontainer (pre-build frontend)

wails-build-native-windows: ## Build Wails for Windows (native build on Windows)
	wails build -platform windows/amd64 -tags "$(GO_PROD_TAGS)"

wails-build-native-windows-skip-frontend: ## Build Windows app native (skip frontend rebuild)
	wails build -platform windows/amd64 -tags "$(GO_PROD_TAGS)" -skipbindings -s

wails-install: ## Install Wails CLI
	go install github.com/wailsapp/wails/v2/cmd/wails@latest