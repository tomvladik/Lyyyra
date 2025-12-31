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

## Stažení hotových binárek

- Při každém vydání tagu `v*` (nebo ručním spuštění workflow) GitHub Actions spustí sestavení „Build and Package“.
- Po dokončení najdete artefakt `Lyyyra-windows-amd64-<tag>.zip` v sekci **Actions → Build and Package**. Archiv obsahuje `Lyyyra.exe`, připravený ke stažení.
- Manuální build je stále možný z příkazové řádky podle instrukcí v části Developer Notes níže.

> **Licenční upozornění:** Materiály stažené z evangelickyzpevnik.cz slouží pouze pro osobní potřebu. Pro veřejné použití je nutné zajistit licenci u držitelů práv.

## Často kladené dotazy

**Musím být online?**
Ano při prvním stažení dat. Poté může aplikace fungovat offline.

**Kde najdu hotové PDF?**
Každý song otevřete ikonou 🎵. Více písní lze seřadit do „Připravených not“ a získat jedno PDF klikem na „Zobrazit připravené noty“.

**Co dělat, když se stahování zasekne?**
Zkontrolujte připojení k internetu a klikněte znovu na „Stáhnout data z internetu“.

**Jak přepínat třídění?**
V InfoBoxu je rozbalovací nabídka „Řadit podle“. Volba se uloží a příště se použije automaticky.

## Snímky obrazovek

_(Sem můžete doplnit obrázky aplikace, pokud jsou k dispozici.)_

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

### Continuous Integration

- Automated builds live in [.github/workflows/build-release.yml](.github/workflows/build-release.yml).
- The workflow runs Go/Vitest tests, performs a Windows Wails build, zips `Lyyyra.exe`, and exposes it as an artifact on tag pushes (`v*`) or manual dispatches.

## Development

- `make wails-dev` / `wails dev` – Wails + Vite dev server (hot reload on http://localhost:34115)
- The devcontainer targets WebKitGTK 4.1 (`webkit2_41`). Override via `WEBKIT_TAG=webkit2_40 make wails-dev` if needed.
- Inside a headless devcontainer there is no GUI session, so `make wails-dev` automatically falls back to `xvfb-run` when `$DISPLAY` is empty. Install it via `sudo apt-get update && sudo apt-get install -y xvfb` if the command is missing, or run `xvfb-run -a wails dev -tags "dev webkit2_41"` manually.

## Make Targets

- `make build`, `make build-prod`, `make wails-build`
- `make test`, `make frontend-test`, `make test-all`
- `make frontend-test-watch`, `make frontend-test-ui`
- `make frontend-build`, `make frontend-install`
- `make clean`

## Testing & Tooling

- Vitest config lives in `frontend/vitest.config.ts` with `src/test/setup.ts`.
- See [TESTING.md](TESTING.md) for detailed coverage notes.
- Run `npx tsc --noEmit` inside `frontend/` to ensure the React code compiles.

## Building Releases

```bash
make build          # Dev builds
make build-prod     # Optimized builds
wails build -s -nopackage  # Direct Wails build
```
