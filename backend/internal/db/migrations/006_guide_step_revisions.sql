BEGIN;

ALTER TABLE guides
	ADD COLUMN IF NOT EXISTS current_revision INTEGER NOT NULL DEFAULT 1;

ALTER TABLE guide_steps
	ADD COLUMN IF NOT EXISTS revision INTEGER NOT NULL DEFAULT 1;

ALTER TABLE guide_steps
	DROP CONSTRAINT IF EXISTS guide_steps_guide_id_step_number_key;

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'guide_steps_guide_revision_step_number_key'
	) THEN
		ALTER TABLE guide_steps
			ADD CONSTRAINT guide_steps_guide_revision_step_number_key
			UNIQUE (guide_id, revision, step_number);
	END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_guide_steps_guide_revision
	ON guide_steps(guide_id, revision, sort_order);

COMMIT;