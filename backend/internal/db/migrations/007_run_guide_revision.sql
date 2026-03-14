BEGIN;

ALTER TABLE runs
	ADD COLUMN IF NOT EXISTS guide_revision INTEGER;

UPDATE runs r
SET guide_revision = COALESCE(
	(
		SELECT gs.revision
		FROM run_splits rs
		JOIN guide_steps gs ON gs.id = rs.step_id
		WHERE rs.run_id = r.id AND gs.guide_id = r.guide_id
		ORDER BY gs.revision DESC
		LIMIT 1
	),
	(
		SELECT gs.revision
		FROM run_checkpoints rc
		JOIN guide_steps gs ON gs.id = rc.step_id
		WHERE rc.run_id = r.id AND gs.guide_id = r.guide_id
		ORDER BY gs.revision DESC
		LIMIT 1
	),
	g.current_revision
)
FROM guides g
WHERE g.id = r.guide_id
	AND (r.guide_revision IS NULL OR r.guide_revision = 0);

ALTER TABLE runs
	ALTER COLUMN guide_revision SET NOT NULL;

ALTER TABLE runs
	ALTER COLUMN guide_revision SET DEFAULT 1;

COMMIT;