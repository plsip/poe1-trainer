# poe1-trainer

PoE1 Trainer is a web application designed to help players learn Path of Exile 1 through deliberate campaign progression and later character advancement.

The system should suggest decisions, highlight important moments, and explain why recommendations matter, but it should not replace the player's own thinking.

## Product Goal

- help the player understand what to do next and why,
- teach patterns used by experienced players,
- track run progress against the guide,
- highlight when to update setup, gems, gear, and flasks,
- stay extensible toward maps, atlas, and endgame features.

## Technical Assumptions

- frontend: React + TypeScript
- backend: Go
- database: PostgreSQL
- local environment: Docker Compose

## Product Principles

- the trainer suggests, but does not hand-hold without explanation,
- the backend is the source of truth for business logic and integrations,
- the markdown guide is the source of build content,
- Client.txt parsing and other external integrations are treated as unreliable inputs,
- when something cannot be detected safely, the system requires manual confirmation.

## MVP

- import one markdown guide for one build,
- model campaign steps and checkpoints,
- manual or semi-automatic run progress tracking,
- suggestions for gems, gear, flasks, and key decisions,
- a simple view of the current stage and next recommendations.

## Out Of Scope For MVP

- full endgame support,
- advanced trade purchase recommendations,
- broad multi-build support,
- agentic decision workflows.

## Starter Documents

- [docs/product/vision.md](docs/product/vision.md)
- [docs/architecture/overview.md](docs/architecture/overview.md)
- [docs/decisions/0001-trainer-not-agent.md](docs/decisions/0001-trainer-not-agent.md)
- [docs/next-steps.md](docs/next-steps.md)

## Uruchamianie lokalnie

### Wymagania

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose v2
- Go 1.24+ (tylko jeśli chcesz uruchamiać backend poza Dockerem)
- Node.js 22+ (tylko jeśli chcesz uruchamiać frontend poza Dockerem)

### Konfiguracja środowiska

```bash
cp .env.example .env
# Edytuj .env jeśli chcesz zmienić domyślne wartości.
```

### Tryb deweloperski (hot-reload)

Uruchamia bazę danych, backend z `air` (hot-reload Go) i frontend z Vite (HMR):

```bash
docker compose -f docker-compose.dev.yml up --build
```

Aplikacja dostępna pod adresem:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

### Zasilanie danymi (seed)

Po pierwszym uruchomieniu baza jest pusta. Zaimportuj przykładowy przewodnik jedną z dwóch metod:

**Opcja A — przez Docker (zalecane):**

```bash
docker compose -f docker-compose.dev.yml --profile seed up seed
```

**Opcja B — przez skrypt hosta (wymaga Go i działającej bazy):**

```bash
bash scripts/seed-dev.sh
```

Po imporcie przewodnik Storm Burst Totemy jest widoczny w aplikacji pod `/guides`.

### Pełny stack produkcyjny

```bash
docker compose up --build
```

Frontend serwowany przez nginx na porcie 3000, backend na 8080.

### Tylko baza danych (praca ręczna)

```bash
docker compose -f docker-compose.dev.yml up -d db
```

Następnie uruchom backend i frontend lokalnie zgodnie z [docs/next-steps.md](docs/next-steps.md).

### Konfiguracja ścieżki do Client.txt

Obserwator `Client.txt` jest opcjonalny — aplikacja działa bez niego.
Kiedy zdecydujesz się go włączyć:

1. Ustal ścieżkę do `Client.txt` na swoim hoście:
   - Windows/Steam: `C:\Program Files (x86)\Steam\steamapps\common\Path of Exile\logs\Client.txt`

2. W pliku `.env` ustaw:
   ```
   LOG_PATH=/mnt/poe-logs/Client.txt
   ```

3. W `docker-compose.dev.yml` odkomentuj wolumen `poe-logs` w sekcji `backend.volumes`:
   ```yaml
   - ${POE_LOG_DIR}:/mnt/poe-logs:ro
   ```
   Ustaw `POE_LOG_DIR` w `.env` na katalog zawierający `Client.txt`:
   ```
   POE_LOG_DIR=C:/Program Files (x86)/Steam/steamapps/common/Path of Exile/logs
   ```

> **Uwaga bezpieczeństwa:** wolumen montowany jest zawsze jako read-only (`:ro`).
> Backend nigdy nie zapisuje do katalogu logów gry.
> Na Windows z Docker Desktop upewnij się, że dysk `C:` jest udostępniony w ustawieniach Docker Desktop → Resources → File Sharing.

### Bez integracji z GGG OAuth

Aplikacja działa w pełni bez konfiguracji klienta OAuth GGG.
Pola `GGG_CLIENT_ID` i `GGG_CLIENT_SECRET` w `.env` można pozostawić puste.
Integracja z GGG API jest zaplanowana na fazę 2 i nie jest wymagana do nauki rozgrywki.

## Working Prompts

The `prompts/` directory contains the working prompt pack for continuing development with Copilot in the new repository.

Recommended starting order:

1. `prompts/00_shared_context.md`
2. `prompts/01_architecture_and_plan.md`
3. `prompts/02_domain_model_and_database.md`
4. `prompts/03_markdown_guide_importer.md`
5. `prompts/08_backend_http_api.md`
6. `prompts/09_frontend_react_ui.md`
