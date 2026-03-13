-- Migration 004: kolumna evidence w run_step_progress
-- Dodaje pole JSONB do przechowywania śladu audytowego w modelu postępu.
--
-- evidence: JSONB — serializacja struktury progress.Evidence:
--   {
--     "event_kind":   "area_entered",
--     "confidence":    0.7,
--     "description":  "Gracz wszedł do strefy ...",
--     "occurred_at":  "2026-03-13T12:00:00Z"
--   }
--
-- Nullable — NULL oznacza brak śladu (ręczne potwierdzenia sprzed migracji).

BEGIN;

ALTER TABLE run_step_progress
    ADD COLUMN IF NOT EXISTS evidence JSONB;

COMMIT;
