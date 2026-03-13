-- Migration 001: initial schema
-- Tworzy wszystkie tabele potrzebne do MVP

BEGIN;

-- ─── guides ─────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS guides (
    id          SERIAL PRIMARY KEY,
    slug        TEXT        NOT NULL UNIQUE,       -- np. "stormburst_campaign_v1"
    title       TEXT        NOT NULL,
    build_name  TEXT        NOT NULL,
    version     TEXT        NOT NULL DEFAULT '1',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS guide_steps (
    id              SERIAL PRIMARY KEY,
    guide_id        INTEGER     NOT NULL REFERENCES guides(id) ON DELETE CASCADE,
    step_number     INTEGER     NOT NULL,
    act             INTEGER     NOT NULL,           -- 1–10
    title           TEXT        NOT NULL,
    description     TEXT        NOT NULL DEFAULT '',
    area            TEXT        NOT NULL DEFAULT '',
    is_checkpoint   BOOLEAN     NOT NULL DEFAULT FALSE,
    requires_manual BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order      INTEGER     NOT NULL,
    UNIQUE (guide_id, step_number)
);
CREATE INDEX IF NOT EXISTS idx_guide_steps_guide_id ON guide_steps(guide_id);

CREATE TABLE IF NOT EXISTS guide_gem_requirements (
    id          SERIAL PRIMARY KEY,
    step_id     INTEGER NOT NULL REFERENCES guide_steps(id) ON DELETE CASCADE,
    gem_name    TEXT    NOT NULL,
    color       TEXT    NOT NULL DEFAULT '',        -- red / green / blue
    note        TEXT    NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_gem_reqs_step_id ON guide_gem_requirements(step_id);

-- ─── runs ───────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS runs (
    id              SERIAL PRIMARY KEY,
    guide_id        INTEGER     NOT NULL REFERENCES guides(id),
    character_name  TEXT        NOT NULL DEFAULT '',
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_runs_guide_id ON runs(guide_id);

CREATE TABLE IF NOT EXISTS run_checkpoints (
    id              SERIAL PRIMARY KEY,
    run_id          INTEGER     NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    step_id         INTEGER     NOT NULL REFERENCES guide_steps(id),
    confirmed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_by    TEXT        NOT NULL DEFAULT 'manual', -- manual | logtail | ggg
    UNIQUE (run_id, step_id)
);
CREATE INDEX IF NOT EXISTS idx_run_checkpoints_run_id ON run_checkpoints(run_id);

CREATE TABLE IF NOT EXISTS run_events (
    id          BIGSERIAL   PRIMARY KEY,
    run_id      INTEGER     NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    event_type  TEXT        NOT NULL,               -- area_entered | step_confirmed | hint
    payload     JSONB       NOT NULL DEFAULT '{}',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_run_events_run_id ON run_events(run_id);

-- ─── splits / ranking ───────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS run_splits (
    id          SERIAL PRIMARY KEY,
    run_id      INTEGER NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    step_id     INTEGER NOT NULL REFERENCES guide_steps(id),
    split_ms    BIGINT  NOT NULL,                   -- ms od startu runu
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_run_splits_run_id ON run_splits(run_id);

COMMIT;
