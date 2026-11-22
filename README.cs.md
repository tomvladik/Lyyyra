# Lyyyra

> <img src="https://flagcdn.com/gb.svg" width="20" alt="United Kingdom"> [English](README.md) · <img src="https://flagcdn.com/cz.svg" width="20" alt="Czechia"> **Čeština**

Lyyyra je desktopová aplikace pro zpěvníky. Stahuje podklady z [evangelickyzpevnik.cz](https://www.evangelickyzpevnik.cz) a nabízí pohodlné vyhledávání, třídění i práci s notovými podklady.

## Hlavní funkce

- **Aktuální databáze písní** – aplikace automaticky stáhne texty, metadata i PDF s notami.
- **Vyhledávání a filtrování** – pište libovolný výraz (název, autor, text) a Lyyyra průběžně zužuje výběr.
- **Řazení** – přepínejte mezi číselným pořadím, názvy a autory hudby či textu.
- **Výběr písní** – u každé skladby přidejte noty do „Připravených not" a stáhněte je jako jedno PDF.
- **Offline režim** – po stažení zůstane databáze uložená lokálně.

## Jak začít

1. Stáhněte (nebo zkompilujte) aplikaci dle návodu níže.
2. Spusťte aplikaci a v horní části klikněte na tlačítko **„Stáhnout data z internetu"**.
3. Po dokončení importu můžete okamžitě vyhledávat, filtrovat a tisknout.
4. Ikona 📋 přidá píseň do pravého panelu „Připravené noty", kde lze stáhnout společné PDF.

> [!WARNING]
> **Licenční upozornění:**
>
> Materiály stažené z evangelickyzpevnik.cz slouží pouze pro osobní potřebu. Pro veřejné použití je nutné zajistit licenci u držitelů práv.

## Ukázka aplikace

![Lyyyra demo](docs/images/demo.gif)

## Stažení hotových binárek

### Z vydaných verzí (doporučeno)

Hotové binárky pro Windows a Linux najdete v [sekci Releases](https://github.com/tomvladik/Lyyyra/releases):

1. Přejděte na [GitHub Releases](https://github.com/tomvladik/Lyyyra/releases).
2. Vyberte poslední verzi (tag `v*`).
3. Stáhněte odpovídající archiv:
   - `Lyyyra-windows-amd64-*.zip` pro Windows
   - `Lyyyra-linux-amd64-*.tar.gz` pro Linux

**Pro Windows uživatele:** Binárka není digitálně podepsaná, proto Windows může zobrazit varování:

- Po stažení extrahujte `Lyyyra.exe` z archivu.
- Při prvním spuštění klikněte na **„Další informace"** (More info) v okně Windows SmartScreen.
- Poté vyberte **„Přesto spustit"** (Run anyway).
- Alternativně: klikněte pravým tlačítkem na `Lyyyra.exe` → Vlastnosti → zaškrtněte „Odblokovat" (Unblock) → OK.

### Z GitHub Actions (nejnovější buildy)

Pro nejnovější neveřejné buildy z větve `main`:

1. Přejděte na [GitHub Actions](https://github.com/tomvladik/Lyyyra/actions).
2. Otevřete poslední úspěšný běh workflow **Build and Package**.
3. V sekci **Artifacts** stáhněte binárku pro svůj systém.

### Ruční build

Manuální build je stále možný z příkazové řádky podle instrukcí v části Poznámky pro vývojáře níže.

## Často kladené dotazy

**Musím být online?**
Ano při prvním stažení dat. Poté může aplikace fungovat offline.

**Kde najdu hotové PDF?**
Každý song otevřete ikonou 🎵. Více písní lze seřadit do „Připravených not" a získat jedno PDF klikem na „Zobrazit připravené noty".

**Co dělat, když se stahování zasekne?**
Zkontrolujte připojení k internetu a klikněte znovu na „Stáhnout data z internetu".

**Jak přepínat třídění?**
V InfoBoxu je rozbalovací nabídka „Řadit podle". Volba se uloží a příště se použije automaticky.

---

# Poznámky pro vývojáře

## Rychlý start

```bash
# Instalace závislostí
make frontend-install

# Sestavení všeho
make build

# Spuštění testů backendu + frontendu
make test-all
```

### Kontinuální integrace a vydávání

- Automatická sestavení běží v [.github/workflows/build-release.yml](.github/workflows/build-release.yml).
- Workflow spouští testy Go/Vitest, provádí sestavení pro více platforem (Windows přes mingw-w64, Linux nativně) a nahrává komprimované artefakty.
- **Vydání:** Push tagu (`v*`) automaticky spustí workflow; binárky se objeví v [Releases](https://github.com/tomvladik/Lyyyra/releases).
- **Nejnovější buildy:** Ruční spuštění nebo push do větve `main` nahrává artefakty do [Actions](https://github.com/tomvladik/Lyyyra/actions).

## Vývoj

Repozitář obsahuje dev-kontejnér pro VS Code (`.devcontainer/devcontainer.json` a `.devcontainer/Dockerfile`). S běžícím Dockerem na hostu otevřete projekt ve VS Code a zvolte **Dev Containers: Reopen in Container**; repozitář je připojen na `/workspaces/Lyyyra`. Image obsahuje Go 1.26, Node 22, Wails CLI, WebKitGTK 4.1, MinGW, Xvfb a nástroje/lintery používané Make cíli; pojmenované diskové svazky uchovávají cache a historii bashu mezi sestaveními kontejnéru.

- `make wails-dev` / `wails dev` – Wails + Vite vývojový server (hot reload na http://localhost:34115)
- Devcontainer cílí na WebKitGTK 4.1 (`webkit2_41`). V případě potřeby přepište pomocí `WEBKIT_TAG=webkit2_40 make wails-dev`.
- Uvnitř headless devcontaineru není GUI relace, proto `make wails-dev` automaticky použije `xvfb-run`, když je `$DISPLAY` prázdné. Nainstalujte jej pomocí `sudo apt-get update && sudo apt-get install -y xvfb`, pokud příkaz chybí, nebo spusťte ručně `xvfb-run -a wails dev -tags "dev webkit2_41"`.

### Historie bashu v dev-kontejnéru

Devcontainer ukládá historii bashu na pojmenovaný diskový svazek `lyyyra-bash-history` (viz `.devcontainer/devcontainer.json`), připojený na `/home/vscode/.bash_history_data`. `HISTFILE` ukazuje na `/home/vscode/.bash_history_data/.bash_history`; `histappend` a `history -a` / `history -n` v `PROMPT_COMMAND` zajistí, že příkazy ze všech terminálů se okamžitě zapisují a jsou viditelné v ostatních terminálech (`ignoreboth` ignoruje duplicity a příkazy začínající mezerou). Limity jsou 10 000 příkazů v paměti a 20 000 v souboru.

Protože jde o pojmenovaný svazek, historie přetrvává i po rebuildu či znovuvytvoření kontejneru a není součástí repozitáře (workspace je samostatný bind mount). Na hostu spusťte `docker volume inspect lyyyra-bash-history` a podívejte se na `Mountpoint`; u Docker Desktopu je tato cesta uvnitř Docker VM (nyní `/data/docker/volumes/lyyyra-bash-history/_data`), takže soubor je tam `/data/docker/volumes/lyyyra-bash-history/_data/.bash_history`. Pro čtení nebo zálohu bez procházení VM spusťte na hostu (nikoli uvnitř kontejneru):

```bash
docker run --rm -v lyyyra-bash-history:/history alpine cat /history/.bash_history
```

## Make cíle

Spusťte `make help` pro rychlý přehled všech dostupných cílů.

**Sestavení**

| Cíl | Popis |
|---|---|
| `make build` | Sestaví Go backend (dev tagy) |
| `make build-prod` | Sestaví produkční binárku s optimalizacemi `-ldflags="-s -w"` |
| `make wails-build` | Sestaví celou Wails aplikaci pro produkci |
| `make wails-build-windows` | Zkříženě sestaví Wails aplikaci pro Windows (nejdřív sestaví frontend) |
| `make wails-build-windows-skip-frontend` | Zkřížené sestavení pro Windows bez přestavby frontendu |
| `make wails-install` | Nainstaluje Wails CLI |

**Vývoj**

| Cíl | Popis |
|---|---|
| `make wails-dev` | Spustí Wails + Vite vývojový server (hot reload) |
| `make frontend-install` | Nainstaluje npm závislosti frontendu |
| `make frontend-build` | Sestaví frontend pro produkci |
| `make frontend-dev` | Spustí pouze vývojový server frontendu |

**Testování**

| Cíl | Popis |
|---|---|
| `make test` | Spustí Go testy |
| `make test-verbose` | Spustí Go testy s úplným výstupem |
| `make test-coverage` | Spustí Go testy a vygeneruje `coverage.html` |
| `make frontend-test` | Spustí Vitest testy frontendu (bez watch režimu) |
| `make frontend-test-watch` | Spustí frontendové testy ve watch režimu |
| `make frontend-test-coverage` | Spustí frontendové testy s přehledem pokrytí |
| `make test-all` | Spustí všechny testy (Go + frontend) |
| `make test-all-coverage` | Spustí všechny testy s pokrytím (Go + frontend) |

**Kontrola kvality**

| Cíl | Popis |
|---|---|
| `make fmt` | Formátuje Go kód (`go fmt ./internal/...`) |
| `make lint` | Kontroluje Go kód pomocí golangci-lint a frontend pomocí eslint |

**Údržba**

| Cíl | Popis |
|---|---|
| `make clean` | Odstraní artefakty sestavení Go a frontendu |
| `make clean-data` | Smaže lokální data aplikace (`~/Lyyyra`) – při dalším spuštění začne čistě |
| `make install-tools` | Nainstaluje Go nástroje (gotestsum, golangci-lint), plugin pro pokrytí a act |
| `make ci-test` | Otestuje GitHub Actions workflow lokálně pomocí act |

## Testování a nástroje

- Konfigurace Vitest je v `frontend/vitest.config.ts` se `src/test/setup.ts`.
- Spusťte `make test-coverage` a `make frontend-test-coverage` pro vygenerování přehledů pokrytí.
- Spusťte `npx tsc --noEmit` v adresáři `frontend/`, abyste ověřili, že React kód kompiluje.

## Sestavování vydání

```bash
make build                        # Dev sestavení (pouze Go backend)
make build-prod                   # Optimalizované sestavení Go backendu (-s -w)
make wails-build                  # Kompletní produkční sestavení Wails (aktuální platforma)
make wails-build-windows          # Zkřížené sestavení pro Windows (devcontainer / mingw-w64)
```
