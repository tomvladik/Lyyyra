# Lyyyra

Lyyyra je desktopová aplikace pro zpěvníky. Stahuje podklady z [evangelickyzpevnik.cz](https://www.evangelickyzpevnik.cz) a nabízí pohodlné vyhledávání, třídění i práci s notovými podklady.

## Hlavní funkce

- **Aktuální databáze písní** – aplikace automaticky stáhne texty, metadata i PDF s notami.
- **Vyhledávání a filtrování** – pište libovolný výraz (název, autor, text) a Lyyyra průběžně zužuje výběr.
- **Řazení** – přepínejte mezi číselným pořadím, názvy a autory hudby či textu.
- **Výběr písní** – u každé skladby přidejte noty do „Připravených not“ a stáhněte je jako jedno PDF.
- **Offline režim** – po stažení zůstane databáze uložená lokálně.

## Jak začít

1. Stáhněte (nebo zkompilujte) aplikaci dle návodu níže.
2. Spusťte Lyyru a v horní části klikněte na tlačítko **„Stáhnout data z internetu“**.
3. Po dokončení importu můžete okamžitě vyhledávat, filtrovat a tisknout.
4. Ikona 📋 přidá píseň do pravého panelu „Připravené noty“, kde lze stáhnout společné PDF.

> [!WARNING]
> **Licenční upozornění:**
> 
> Materiály stažené z evangelickyzpevnik.cz slouží pouze pro osobní potřebu. Pro veřejné použití je nutné zajistit licenci u držitelů práv.

## Ukázka aplikace

![Lyyyra demo](docs/images/demo.gif)

## Stažení hotových binárek

### Z vydaných verzí (doporučeno)

Hotové binárky pro Windows a Linux najdete v [sekci Releases](https://github.com/tomvladik/Lyyyra/releases):

1. Přejděte na [GitHub Releases](https://github.com/tomvladik/Lyyyra/releases)
2. Vyberte poslední verzi (tag `v*`)
3. Stáhněte odpovídající archiv:
   - `Lyyyra-windows-amd64-*.zip` pro Windows
   - `Lyyyra-linux-amd64-*.tar.gz` pro Linux

**Pro Windows uživatele:** Binárka není digitálně podepsaná, proto Windows může zobrazit varování. Postup:
- Po stažení extrahujte `Lyyyra.exe` z archivu
- Při prvním spuštění klikněte na **„Další informace"** (More info) v okně Windows SmartScreen
- Poté vyberte **„Přesto spustit"** (Run anyway)
- Alternativně: klikněte pravým tlačítkem na `Lyyyra.exe` → Vlastnosti → zaškrtněte „Odblokovat" (Unblock) → OK

### Z GitHub Actions (nejnovější buildy)

Pro nejnovější neveřejné buildy z větve `main`:

1. Přejděte na [GitHub Actions](https://github.com/tomvladik/Lyyyra/actions)
2. Otevřete poslední úspěšný běh workflow **Build and Package**
3. V sekci **Artifacts** stáhněte binárku pro svůj systém

### Ručnní build

Manuální build je stále možný z příkazové řádky podle instrukcí v části Developer Notes níže.

## Často kladené dotazy

**Musím být online?**
Ano při prvním stažení dat. Poté může aplikace fungovat offline.

**Kde najdu hotové PDF?**
Každý song otevřete ikonou 🎵. Více písní lze seřadit do „Připravených not“ a získat jedno PDF klikem na „Zobrazit připravené noty“.

**Co dělat, když se stahování zasekne?**
Zkontrolujte připojení k internetu a klikněte znovu na „Stáhnout data z internetu“.

**Jak přepínat třídění?**
V InfoBoxu je rozbalovací nabídka „Řadit podle“. Volba se uloží a příště se použije automaticky.

---

# Developer Notes (English)

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

- `make wails-dev` / `wails dev` – Wails + Vite dev server (hot reload on http://localhost:34115)
- The devcontainer targets WebKitGTK 4.1 (`webkit2_41`). Override via `WEBKIT_TAG=webkit2_40 make wails-dev` if needed.
- Inside a headless devcontainer there is no GUI session, so `make wails-dev` automatically falls back to `xvfb-run` when `$DISPLAY` is empty. Install it via `sudo apt-get update && sudo apt-get install -y xvfb` if the command is missing, or run `xvfb-run -a wails dev -tags "dev webkit2_41"` manually.

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
- See [TESTING.md](TESTING.md) for detailed coverage notes.
- Run `npx tsc --noEmit` inside `frontend/` to ensure the React code compiles.

## Building Releases

```bash
make build                        # Dev build (Go backend only)
make build-prod                   # Optimized Go backend build (-s -w)
make wails-build                  # Full Wails production build (current platform)
make wails-build-windows          # Cross-compile for Windows (devcontainer / mingw-w64)
```
