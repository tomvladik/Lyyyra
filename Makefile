# Show help for all tasks
.PHONY: help
help:
	@echo "Available tasks:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Go backend tasks

build: ## Build the Go backend
	go build -v ./...

test: ## Run Go tests
	gotestsum --format testname -- -tags "fts5" ./...

fmt: ## Format Go code
	go fmt ./...

lint: ## Lint Go code (requires golangci-lint)
	golangci-lint run

# Node.js/Frontend tasks

frontend-install: ## Install frontend dependencies
	cd frontend && npm install

frontend-build: ## Build frontend for production
	cd frontend && npm run build

frontend-dev: ## Start frontend development server
	cd frontend && npm run dev

# Clean up build artifacts

clean: ## Clean Go and frontend build artifacts
	go clean
	rm -rf frontend/dist

test-all: test ## Run all tests (Go + frontend, if available)
	cd frontend && npm test || true

install-tools: ## Install Go test and lint tools
	go install gotest.tools/gotestsum@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2

# Wails tasks
wails-dev: ## Start Wails development server
	wails dev

wails-build: ## Build Wails application for production
	wails build

wails-build-windows: ## Build Wails application for Windows
	wails build -platform windows/amd64

wails-install: ## Install Wails CLI
	go install github.com/wailsapp/wails/v2/cmd/wails@latest