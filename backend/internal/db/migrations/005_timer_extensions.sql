-- Migration 005: rozszerzenia timera
-- Dodaje trzy kolumny do tabeli runs obsługujące:
--   timer_started_at — moment rzeczywistego startu timera (może różnić się od started_at
--                      gdy włączony jest tryb auto-startu z logtaila).
--   paused_at        — timestamp WSTRZYMANIA runu (NULL = run działa / nie wstrzymany).
--   total_paused_ms  — skumulowany czas pausy (ms), odejmowany od elapsed_ms.
--
-- Backfill: dla istniejących runów timer_started_at = started_at.

BEGIN;

ALTER TABLE runs
    ADD COLUMN IF NOT EXISTS timer_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS paused_at        TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS total_paused_ms  BIGINT NOT NULL DEFAULT 0;

-- Istniejące runy: timer start = run start (bez opóźnienia).
UPDATE runs SET timer_started_at = started_at WHERE timer_started_at IS NULL;

COMMIT;
