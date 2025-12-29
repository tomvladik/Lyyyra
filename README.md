# README

## ToDO
 - DB sorting - CS
 - full text search - UI marking
 - load button/progress messages
 - persist select setting
 - UI styles

## Quick Start

### Build & Test

```bash
# Build the application
make build

# Run tests
make test

# Install frontend dependencies
make frontend-install

# Build frontend
make frontend-build
```

### Development

Run in live development mode:

```bash
make wails-dev
# or
wails dev
```

This runs a Vite development server with hot reload. A dev server runs on http://localhost:34115 for calling Go methods from the browser.

### Building for Production

```bash
# Build in devcontainer (recommended for development)
make build

# Build with optimizations
make build-prod

# Or use Wails CLI (may be slow in containers)
wails build -tags "fts5" -s -nopackage
```

**DevContainer Support:** The dev container now includes webkit2gtk libraries and can build the application. After making changes to [.devcontainer/Dockerfile](.devcontainer/Dockerfile), rebuild the container with: `Dev Containers: Rebuild Container`

## Makefile Commands

Run `make help` to see all available commands:
- `make build` - Build the Go backend
- `make test` - Run Go tests
- `make frontend-install` - Install frontend dependencies
- `make frontend-build` - Build frontend for production
- `make wails-dev` - Start Wails development server
- `make wails-build` - Build Wails application for production
- `make clean` - Clean build artifacts

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.
