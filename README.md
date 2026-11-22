# Lyyyra

> <img src="https://flagcdn.com/gb.svg" width="20" alt="United Kingdom"> **English** · <img src="https://flagcdn.com/cz.svg" width="20" alt="Czechia"> [Čeština](README.cs.md)

Lyyyra is a desktop application for songbooks. It downloads sheet music from [evangelickyzpevnik.cz](https://www.evangelickyzpevnik.cz) and offers convenient searching, sorting, and working with sheet music PDFs.

## Features

- **Up-to-date song database** – lyrics, metadata, and sheet music PDFs are downloaded automatically.
- **Search and filtering** – type any term (title, author, lyrics) and Lyyyra narrows the results as you type.
- **Sorting** – switch between song number, title, music author, and lyric author.
- **Song selection** – add sheet music to "Prepared sheets" for any song and download them as a single PDF.
- **Offline mode** – after the initial download the database is stored locally.

## Getting Started

1. Download (or compile) the application using the instructions below.
2. Launch Lyyyra and click **"Stáhnout data z internetu"** (Download data from the internet) at the top.
3. Once the import finishes, you can immediately search, filter, and print.
4. The 📋 icon adds a song to the **"Připravené noty"** (Prepared sheets) panel on the right, where you can download a combined PDF.

> [!WARNING]
> **License notice:**
>
> Materials downloaded from evangelickyzpevnik.cz are for personal use only. Public use requires a license from the rights holders.

## App Preview

![Lyyyra demo](docs/images/demo.gif)

## Download Prebuilt Binaries

### From Releases (recommended)

Prebuilt binaries for Windows and Linux are available in the [Releases section](https://github.com/tomvladik/Lyyyra/releases):

1. Go to [GitHub Releases](https://github.com/tomvladik/Lyyyra/releases).
2. Pick the latest version (tag `v*`).
3. Download the matching archive:
   - `Lyyyra-windows-amd64-*.zip` for Windows
   - `Lyyyra-linux-amd64-*.tar.gz` for Linux

**For Windows users:** The binary is not digitally signed, so Windows may show a warning:

- After downloading, extract `Lyyyra.exe` from the archive.
- On first launch, click **"More info"** in the Windows SmartScreen dialog.
- Then choose **"Run anyway"**.
- Alternatively: right-click `Lyyyra.exe` → Properties → check "Unblock" → OK.

### From GitHub Actions (latest builds)

For the newest unpublished builds from the `main` branch:

1. Go to [GitHub Actions](https://github.com/tomvladik/Lyyyra/actions).
2. Open the latest successful **Build and Package** workflow run.
3. In the **Artifacts** section, download the binary for your OS.

### Manual build

You can still build manually from the command line — see the Developer Notes below.

## Frequently Asked Questions

**Do I need to be online?**
Yes, for the initial data download. After that the app works offline.

**Where do I find the finished PDFs?**
Open a song with the 🎵 icon. Multiple songs can be added to "Prepared sheets" and exported as one PDF by clicking "Zobrazit připravené noty" (Show prepared sheets).

**What if the download gets stuck?**
Check your internet connection and click "Stáhnout data z internetu" again.

**How do I change the sorting?**
Use the "Řadit podle" (Sort by) dropdown in the InfoBox. The choice is saved and applied automatically next time.

---

# Developer Notes

## Quick Start

```bash
# Install deps
make frontend-install

# Build everything
make build

# Run backend + frontend tests
make test-all
```

### Continuous Integration & Releases

- Automated builds live in [.github/workflows/build-release.yml](.github/workflows/build-release.yml).
- The workflow runs Go/Vitest tests, performs cross-platform builds (Windows via mingw-w64, Linux native), and uploads compressed artifacts.
- **Releases**: Tag pushes (`v*`) automatically trigger the workflow; binaries appear in [Releases](https://github.com/tomvladik/Lyyyra/releases).
- **Latest builds**: Manual dispatches or `main` branch pushes upload artifacts to [Actions](https://github.com/tomvladik/Lyyyra/actions).

## Development

The repo includes a VS Code devcontainer (`.devcontainer/devcontainer.json` and `.devcontainer/Dockerfile`). With Docker running on the host, open the project in VS Code and choose **Dev Containers: Reopen in Container**; the repository is mounted at `/workspaces/Lyyyra`. The image provides Go 1.26, Node 22, Wails CLI, WebKitGTK 4.1, MinGW, Xvfb, and the linters/tools used by the Make targets; named volumes preserve caches and bash history across rebuilds, and the host Docker socket is mounted for Docker-based tooling.

- `make wails-dev` / `wails dev` – Wails + Vite dev server (hot reload on http://localhost:34115)
- The devcontainer targets WebKitGTK 4.1 (`webkit2_41`). Override via `WEBKIT_TAG=webkit2_40 make wails-dev` if needed.
- Inside a headless devcontainer there is no GUI session, so `make wails-dev` automatically falls back to `xvfb-run` when `$DISPLAY` is empty. Install it via `sudo apt-get update && sudo apt-get install -y xvfb` if the command is missing, or run `xvfb-run -a wails dev -tags "dev webkit2_41"` manually.

### Bash history in the devcontainer

The devcontainer stores bash history on the named Docker volume `lyyyra-bash-history` (see `.devcontainer/devcontainer.json`), mounted at `/home/vscode/.bash_history_data`. `HISTFILE` points to `/home/vscode/.bash_history_data/.bash_history`; `histappend` and `history -a` / `history -n` in `PROMPT_COMMAND` make commands from all terminals append immediately and appear in other terminals (`ignoreboth` skips duplicates and commands starting with a space). The limits are 10,000 commands in memory and 20,000 in the file.

Because it is a named volume, the history survives rebuilding/recreating the container and is not part of the repository (the workspace is a separate bind mount). On the host run `docker volume inspect lyyyra-bash-history` to see its `Mountpoint`; with Docker Desktop this path is inside the Docker VM (currently `/data/docker/volumes/lyyyra-bash-history/_data`), so the file there is `/data/docker/volumes/lyyyra-bash-history/_data/.bash_history`. To read or back it up without going into the VM, run this on the host (not inside the container):

```bash
docker run --rm -v lyyyra-bash-history:/history alpine cat /history/.bash_history
```

## Make Targets

Run `make help` for a quick overview of all available targets.

**Build**

| Target | Description |
|---|---|
| `make build` | Build the Go backend (dev tags) |
| `make build-prod` | Build production binary with `-ldflags="-s -w"` optimizations |
| `make wails-build` | Build full Wails application for production |
| `make wails-build-windows` | Cross-compile Wails app for Windows (builds frontend first) |
| `make wails-build-windows-skip-frontend` | Cross-compile for Windows, skipping frontend rebuild |
| `make wails-install` | Install Wails CLI |

**Development**

| Target | Description |
|---|---|
| `make wails-dev` | Start Wails + Vite dev server (hot reload) |
| `make frontend-install` | Install frontend npm dependencies |
| `make frontend-build` | Build frontend for production |
| `make frontend-dev` | Start frontend development server only |

**Testing**

| Target | Description |
|---|---|
| `make test` | Run Go tests |
| `make test-verbose` | Run Go tests with full output |
| `make test-coverage` | Run Go tests and generate `coverage.html` |
| `make frontend-test` | Run frontend Vitest tests (non-watch) |
| `make frontend-test-watch` | Run frontend tests in watch mode |
| `make frontend-test-coverage` | Run frontend tests with coverage report |
| `make test-all` | Run all tests (Go + frontend) |
| `make test-all-coverage` | Run all tests with coverage (Go + frontend) |

**Code Quality**

| Target | Description |
|---|---|
| `make fmt` | Format Go code (`go fmt ./internal/...`) |
| `make lint` | Lint Go code with golangci-lint and frontend with eslint |

**Maintenance**

| Target | Description |
|---|---|
| `make clean` | Remove Go and frontend build artifacts |
| `make clean-data` | Delete local app data (`~/Lyyyra`) – next run starts fresh |
| `make install-tools` | Install Go tools (gotestsum, golangci-lint), coverage plugin, and act |
| `make ci-test` | Test the GitHub Actions workflow locally using act |

## Testing & Tooling

- Vitest config lives in `frontend/vitest.config.ts` with `src/test/setup.ts`.
- Run `make test-coverage` and `make frontend-test-coverage` to generate coverage reports.
- Run `npx tsc --noEmit` inside `frontend/` to ensure the React code compiles.

## Building Releases

```bash
make build                        # Dev build (Go backend only)
make build-prod                   # Optimized Go backend build (-s -w)
make wails-build                  # Full Wails production build (current platform)
make wails-build-windows          # Cross-compile for Windows (devcontainer / mingw-w64)
```
