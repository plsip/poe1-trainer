-- Migration 002: domain v2 — rozszerzony model domenowy
-- Dodaje encje: builds, build_versions, guide_step_conditions,
--               gems, gem_availability_rules, gem_upgrade_rules, gear_hint_rules,
--               run_characters, run_step_progress, character_snapshots,
--               manual_checks, local_rankings
-- Modyfikuje:   guides      (+build_version_id)
--               guide_steps (+completion_mode, +section)
--               runs        (+league, +status, +notes)
--
-- Tabela run_checkpoints z migracji 001 zostaje zachowana dla kompatybilności.
-- run_step_progress zastępuje ją semantycznie jako główne źródło prawdy.

BEGIN;

-- ─── builds ──────────────────────────────────────────────────────────────────
-- Archetyp buildu (np. "Storm Burst Totemy").
-- Źródło prawdy: ręcznie importowane z pliku MD lub UI.

CREATE TABLE IF NOT EXISTS builds (
    id          SERIAL      PRIMARY KEY,
    slug        TEXT        NOT NULL UNIQUE,      -- "stormburst_totems"
    name        TEXT        NOT NULL,             -- "Storm Burst Totemy"
    class       TEXT        NOT NULL DEFAULT '',  -- "Witch", "Shadow", …
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── build_versions ──────────────────────────────────────────────────────────
-- Konkretna wersja teorii buildu (np. "3.25 — Settlers").
-- Jeden build może mieć wiele wersji w miarę zmian meta lub patchów.

CREATE TABLE IF NOT EXISTS build_versions (
    id          SERIAL      PRIMARY KEY,
    build_id    INTEGER     NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    version     TEXT        NOT NULL,             -- "1", "2", "3.25"
    patch_tag   TEXT        NOT NULL DEFAULT '',  -- "3.25", "settlers"
    notes       TEXT        NOT NULL DEFAULT '',
    is_current  BOOLEAN     NOT NULL DEFAULT FALSE,
    released_at DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (build_id, version)
);

CREATE INDEX IF NOT EXISTS idx_build_versions_build_id ON build_versions(build_id);

-- ─── alter guides ────────────────────────────────────────────────────────────
-- Dodaje opcjonalne powiązanie z build_version.
-- Nullable w MVP; backfill przez: UPDATE guides SET build_version_id = ...
-- Po backfillu można dodać NOT NULL CONSTRAINT w kolejnej migracji.

ALTER TABLE guides
    ADD COLUMN IF NOT EXISTS build_version_id INTEGER
        REFERENCES build_versions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_guides_build_version_id ON guides(build_version_id);

-- ─── alter guide_steps ───────────────────────────────────────────────────────
-- completion_mode: tryb weryfikacji ukończenia kroku.
--   manual        — gracz klika "Potwierdź" w UI
--   logtail       — inferencja z Client.txt (bez potwierdzenia)
--   logtail_ask   — inferencja z Client.txt + wymaga potwierdzenia
--   ggg_api       — inferencja z GGG API (bez potwierdzenia)
--   ggg_api_ask   — inferencja z GGG API + wymaga potwierdzenia
-- section: nazwa sekcji/rozdziału (np. "Akt 1 — Pustkowia")

ALTER TABLE guide_steps
    ADD COLUMN IF NOT EXISTS completion_mode TEXT NOT NULL DEFAULT 'manual';

ALTER TABLE guide_steps
    ADD COLUMN IF NOT EXISTS section TEXT NOT NULL DEFAULT '';

-- ─── guide_step_conditions ───────────────────────────────────────────────────
-- Reguły inferencji przypisane do kroków.
-- Wiele warunków per krok, wykonywane w kolejności `priority` (rosnąco).
-- Pierwszy spełniony warunek wywołuje zmianę statusu kroku.
--
-- condition_type / wymagany payload:
--   logtail_area   — {"area": "The Twilight Strand"}
--   logtail_event  — {"pattern": "^.*You have entered.*$"}
--   ggg_level      — {"min_level": 12}
--   ggg_quest      — {"quest_id": "a1q3", "state": "Finished"}
--   manual_confirm — {} (jawne wymuszenie potwierdzenia)

CREATE TABLE IF NOT EXISTS guide_step_conditions (
    id             SERIAL  PRIMARY KEY,
    step_id        INTEGER NOT NULL REFERENCES guide_steps(id) ON DELETE CASCADE,
    condition_type TEXT    NOT NULL,
    payload        JSONB   NOT NULL DEFAULT '{}',
    priority       INTEGER NOT NULL DEFAULT 0,
    notes          TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_step_conditions_step_id ON guide_step_conditions(step_id);

-- ─── gems ────────────────────────────────────────────────────────────────────
-- Słownik gemów — źródło prawdy: nazwy z gry (angielskie).
-- Zbiór statyczny importowany przy starcie lub z pliku MD.

CREATE TABLE IF NOT EXISTS gems (
    id         SERIAL  PRIMARY KEY,
    name       TEXT    NOT NULL UNIQUE,       -- "Stormburst", "Added Lightning Damage Support"
    color      TEXT    NOT NULL DEFAULT '',   -- red | green | blue | white
    is_skill   BOOLEAN NOT NULL DEFAULT TRUE,
    is_support BOOLEAN NOT NULL DEFAULT FALSE,
    wiki_url   TEXT    NOT NULL DEFAULT '',
    notes      TEXT    NOT NULL DEFAULT ''
);

-- ─── gem_availability_rules ──────────────────────────────────────────────────
-- Kiedy i gdzie gracz MOŻE PO RAZ PIERWSZY zdobyć dany gem w ramach poradnika.
-- Źródło prawdy: poradnik (ręcznie wpisane lub zparsowane z MD).
-- Derived: nic — wszystkie kolumny są danymi wejściowymi.

CREATE TABLE IF NOT EXISTS gem_availability_rules (
    id                  SERIAL  PRIMARY KEY,
    guide_id            INTEGER NOT NULL REFERENCES guides(id) ON DELETE CASCADE,
    gem_id              INTEGER NOT NULL REFERENCES gems(id)   ON DELETE CASCADE,
    act_first_available INTEGER NOT NULL,
    vendor_name         TEXT    NOT NULL DEFAULT '',  -- "Nessa", "Lilly Roth"
    vendor_act          INTEGER,
    acquisition_type    TEXT    NOT NULL DEFAULT 'vendor', -- vendor | drop | quest
    quest_name          TEXT    NOT NULL DEFAULT '',
    notes               TEXT    NOT NULL DEFAULT '',
    UNIQUE (guide_id, gem_id)
);

CREATE INDEX IF NOT EXISTS idx_gem_avail_guide_id ON gem_availability_rules(guide_id);
CREATE INDEX IF NOT EXISTS idx_gem_avail_gem_id   ON gem_availability_rules(gem_id);

-- ─── gem_upgrade_rules ───────────────────────────────────────────────────────
-- Co należy zrobić z gemem na konkretnym kroku poradnika.
-- Generuje alert w silniku rekomendacji przy osiągnięciu danego kroku.

CREATE TABLE IF NOT EXISTS gem_upgrade_rules (
    id           SERIAL  PRIMARY KEY,
    guide_id     INTEGER NOT NULL REFERENCES guides(id)      ON DELETE CASCADE,
    step_id      INTEGER NOT NULL REFERENCES guide_steps(id) ON DELETE CASCADE,
    gem_id       INTEGER NOT NULL REFERENCES gems(id)        ON DELETE CASCADE,
    action_type  TEXT    NOT NULL, -- socket | link | quality | vaal | swap | buy
    target_value TEXT    NOT NULL DEFAULT '', -- "4L", "B-B-R-R", "20/20", "Level 20"
    priority     INTEGER NOT NULL DEFAULT 0,
    notes        TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_gem_upgrade_guide_id ON gem_upgrade_rules(guide_id);
CREATE INDEX IF NOT EXISTS idx_gem_upgrade_step_id  ON gem_upgrade_rules(step_id);

-- ─── gear_hint_rules ─────────────────────────────────────────────────────────
-- Sugestie dotyczące ekwipunku powiązane z krokiem lub całym poradnikiem.
-- step_id = NULL oznacza hint globalny (np. "kupuj flaki z odporą").

CREATE TABLE IF NOT EXISTS gear_hint_rules (
    id          SERIAL  PRIMARY KEY,
    guide_id    INTEGER NOT NULL REFERENCES guides(id)      ON DELETE CASCADE,
    step_id     INTEGER          REFERENCES guide_steps(id) ON DELETE CASCADE,
    slot        TEXT    NOT NULL DEFAULT '', -- helmet | chest | gloves | boots |
                                             -- weapon | offhand | ring | amulet |
                                             -- belt | flask | any
    description TEXT    NOT NULL,
    min_life    INTEGER,                     -- sugerowane minimum HP w tym miejscu
    min_res     INTEGER,                     -- sugerowana suma odporności
    priority    TEXT    NOT NULL DEFAULT 'medium', -- high | medium | low
    notes       TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_gear_hint_guide_id ON gear_hint_rules(guide_id);
CREATE INDEX IF NOT EXISTS idx_gear_hint_step_id  ON gear_hint_rules(step_id);

-- ─── alter runs ──────────────────────────────────────────────────────────────
-- league: liga w której postać jest rozgrywana ("Settlers", "Standard", …)
-- status: aktywny stan runu
--   active    — run w toku
--   finished  — campain ukończona
--   abandoned — run porzucony przed ukończeniem
-- notes: dowolne komentarze gracza po run

ALTER TABLE runs ADD COLUMN IF NOT EXISTS league TEXT NOT NULL DEFAULT '';
ALTER TABLE runs ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE runs ADD COLUMN IF NOT EXISTS notes  TEXT NOT NULL DEFAULT '';

-- ─── run_characters ──────────────────────────────────────────────────────────
-- Szczegóły postaci powiązanej z runem.
-- Uzupełniane manualnie przy starcie runu lub automatycznie z GGG API.
-- level_current: aktualizowany po każdym snapshote (cached — źródłem jest GGG API).

CREATE TABLE IF NOT EXISTS run_characters (
    id              SERIAL      PRIMARY KEY,
    run_id          INTEGER     NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    character_name  TEXT        NOT NULL,
    character_class TEXT        NOT NULL DEFAULT '',
    league          TEXT        NOT NULL DEFAULT '',
    level_at_start  INTEGER     NOT NULL DEFAULT 1,
    level_current   INTEGER     NOT NULL DEFAULT 1,  -- cached z ostatniego snapshotu
    last_seen_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (run_id)   -- MVP: jeden run = jedna postać
);

CREATE INDEX IF NOT EXISTS idx_run_characters_run_id ON run_characters(run_id);

-- ─── run_step_progress ───────────────────────────────────────────────────────
-- Główna tabela postępu kroków w danym runie.
-- Zastępuje semantycznie run_checkpoints (zachowane z 001 dla kompatybilności).
--
-- status:
--   pending           — krok jeszcze nie osiągnięty
--   in_progress       — gracz widzi krok jako aktywny / jest w tej strefie
--   needs_confirmation — inferencja wykryła ukończenie, czeka na potwierdzenie
--   completed         — krok ukończony (confirmed_at IS NOT NULL lub auto)
--   skipped           — gracz jawnie pominął krok
--
-- Źródło prawdy: completed_at + confirmed_by.
-- status = 'completed' może być re-derivowany z tych pól.

CREATE TABLE IF NOT EXISTS run_step_progress (
    id            SERIAL      PRIMARY KEY,
    run_id        INTEGER     NOT NULL REFERENCES runs(id)        ON DELETE CASCADE,
    step_id       INTEGER     NOT NULL REFERENCES guide_steps(id) ON DELETE CASCADE,
    status        TEXT        NOT NULL DEFAULT 'pending',
    completed_at  TIMESTAMPTZ,
    confirmed_by  TEXT        NOT NULL DEFAULT 'manual', -- manual | logtail | ggg
    confirmed_at  TIMESTAMPTZ,
    notes         TEXT        NOT NULL DEFAULT '',
    UNIQUE (run_id, step_id)
);

CREATE INDEX IF NOT EXISTS idx_rsp_run_id         ON run_step_progress(run_id);
CREATE INDEX IF NOT EXISTS idx_rsp_step_id        ON run_step_progress(step_id);
CREATE INDEX IF NOT EXISTS idx_rsp_run_status     ON run_step_progress(run_id, status);

-- ─── character_snapshots ─────────────────────────────────────────────────────
-- Migawka stanu postaci w danym momencie runu.
-- Kolumny skalarne (level, life_max, res_*) = źródło prawdy dla alertów i rankingu.
-- equipped_items, skills, raw_response = cache z GGG API (derived, do odczytu UI).
-- Snapshoty są immutable — nowy stan = nowy wiersz.

CREATE TABLE IF NOT EXISTS character_snapshots (
    id             BIGSERIAL   PRIMARY KEY,
    run_id         INTEGER     NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    captured_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source         TEXT        NOT NULL DEFAULT 'manual', -- manual | ggg
    level          INTEGER     NOT NULL DEFAULT 1,
    life_max       INTEGER,
    mana_max       INTEGER,
    res_fire       INTEGER,
    res_cold       INTEGER,
    res_lightning  INTEGER,
    res_chaos      INTEGER,
    equipped_items JSONB       NOT NULL DEFAULT '{}',  -- cache: pełny ekwipunek z GGG
    skills         JSONB       NOT NULL DEFAULT '{}',  -- cache: socketed gems z GGG
    raw_response   JSONB       NOT NULL DEFAULT '{}'   -- cache: pełna odpowiedź GGG API
);

CREATE INDEX IF NOT EXISTS idx_snapshots_run_id      ON character_snapshots(run_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_captured_at ON character_snapshots(run_id, captured_at DESC);

-- ─── manual_checks ───────────────────────────────────────────────────────────
-- Pytania wymagające ręcznej odpowiedzi gracza, powiązane z krokiem lub runem.
-- Generowane przez silnik rekomendacji lub importera poradnika.
-- step_id = NULL: check globalny (np. "Czy masz flakę life?")
--
-- check_type: gear | gem | level | resist | flask | quest | free_form
-- is_confirmed = FALSE + confirmed_at IS NULL = oczekuje odpowiedzi

CREATE TABLE IF NOT EXISTS manual_checks (
    id             SERIAL      PRIMARY KEY,
    run_id         INTEGER     NOT NULL REFERENCES runs(id)        ON DELETE CASCADE,
    step_id        INTEGER              REFERENCES guide_steps(id) ON DELETE CASCADE,
    check_type     TEXT        NOT NULL,
    prompt         TEXT        NOT NULL,
    is_confirmed   BOOLEAN     NOT NULL DEFAULT FALSE,
    response_value TEXT        NOT NULL DEFAULT '', -- "yes" | "no" | wartość wpisana
    confirmed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_manual_checks_run_id              ON manual_checks(run_id);
CREATE INDEX IF NOT EXISTS idx_manual_checks_step_id             ON manual_checks(step_id);
CREATE INDEX IF NOT EXISTS idx_manual_checks_run_unconfirmed     ON manual_checks(run_id)
    WHERE is_confirmed = FALSE;

-- ─── local_rankings ──────────────────────────────────────────────────────────
-- Pre-obliczona tabela wyników dla danego poradnika.
-- DERIVED — wszystkie kolumny obliczane z run_splits + runs.
-- Aktualizowane po każdym zakończeniu runu (status = 'finished').
-- rank przeliczany dla wszystkich runów danego guide po każdej zmianie.
--
-- total_ms: derived z runs.finished_at - runs.started_at
-- act_splits: derived z run_splits (agregacja per akt)
-- rank: derived — pozycja wśród wszystkich finished runów tego guide

CREATE TABLE IF NOT EXISTS local_rankings (
    id          SERIAL      PRIMARY KEY,
    guide_id    INTEGER     NOT NULL REFERENCES guides(id) ON DELETE CASCADE,
    run_id      INTEGER     NOT NULL REFERENCES runs(id)   ON DELETE CASCADE,
    total_ms    BIGINT      NOT NULL,
    act_splits  JSONB       NOT NULL DEFAULT '{}', -- {"1": 120000, "2": 345000, …}
    rank        INTEGER     NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (run_id)
);

CREATE INDEX IF NOT EXISTS idx_rankings_guide_id  ON local_rankings(guide_id);
CREATE INDEX IF NOT EXISTS idx_rankings_guide_rank ON local_rankings(guide_id, rank);

COMMIT;
