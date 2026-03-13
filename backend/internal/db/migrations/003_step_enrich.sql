-- Migration 003: wzbogacenie kroków poradnika
-- Dodaje kolumny step_type i quest_name do guide_steps.
-- step_type — klasyfikacja heurystyczna kroku (boss_kill, navigation, gem_acquire, …)
-- quest_name — nazwa questa wyekstrahowana z tekstu kroku (może być pusta)

BEGIN;

ALTER TABLE guide_steps
    ADD COLUMN IF NOT EXISTS step_type  TEXT NOT NULL DEFAULT 'general';

ALTER TABLE guide_steps
    ADD COLUMN IF NOT EXISTS quest_name TEXT NOT NULL DEFAULT '';

COMMIT;
