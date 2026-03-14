package run

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides persistence for runs.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateRun inserts a new active run.
// When autoStart=true the timer_started_at is left NULL; it will be set on the
// first logtail area event. When autoStart=false the timer starts immediately.
func (r *Repository) CreateRun(ctx context.Context, guideID int, characterName, league string, autoStart bool) (*RunSession, error) {
	run := &RunSession{}
	var timerSQL string
	if autoStart {
		// timer_started_at stays NULL — set later by HandleAreaEvent
		timerSQL = `INSERT INTO runs (guide_id, guide_revision, character_name, league)
		VALUES ($1, (SELECT current_revision FROM guides WHERE id = $1), $2, $3)
		RETURNING id, guide_id, guide_revision, character_name, league, status, notes, started_at, finished_at,
		          is_active, timer_started_at, paused_at, total_paused_ms`
	} else {
		// timer starts immediately
		timerSQL = `INSERT INTO runs (guide_id, guide_revision, character_name, league, timer_started_at)
		VALUES ($1, (SELECT current_revision FROM guides WHERE id = $1), $2, $3, NOW())
		RETURNING id, guide_id, guide_revision, character_name, league, status, notes, started_at, finished_at,
		          is_active, timer_started_at, paused_at, total_paused_ms`
	}
	err := r.db.QueryRow(ctx, timerSQL, guideID, characterName, league).
		Scan(&run.ID, &run.GuideID, &run.GuideRevision, &run.CharacterName, &run.League, &run.Status, &run.Notes,
			&run.StartedAt, &run.FinishedAt, &run.IsActive,
			&run.TimerStartedAt, &run.PausedAt, &run.TotalPausedMs)
	if err != nil {
		return nil, fmt.Errorf("run: insert: %w", err)
	}
	return run, nil
}

// GetActiveRun returns the most recently started active run, or nil when no run is active.
// Used by the logtail event dispatcher to route area/level events.
func (r *Repository) GetActiveRun(ctx context.Context) (*RunSession, error) {
	run := &RunSession{}
	err := r.db.QueryRow(ctx, `
		SELECT id, guide_id, guide_revision, character_name, league, status, notes, started_at, finished_at,
		       is_active, timer_started_at, paused_at, total_paused_ms
		FROM runs WHERE is_active = TRUE ORDER BY started_at DESC LIMIT 1`,
	).Scan(&run.ID, &run.GuideID, &run.GuideRevision, &run.CharacterName, &run.League, &run.Status, &run.Notes,
		&run.StartedAt, &run.FinishedAt, &run.IsActive,
		&run.TimerStartedAt, &run.PausedAt, &run.TotalPausedMs)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("run: get active: %w", err)
	}
	return run, nil
}

// GetRun fetches a single run by ID.
func (r *Repository) GetRun(ctx context.Context, runID int) (*RunSession, error) {
	run := &RunSession{}
	err := r.db.QueryRow(ctx, `
		SELECT id, guide_id, guide_revision, character_name, league, status, notes, started_at, finished_at,
		       is_active, timer_started_at, paused_at, total_paused_ms
		FROM runs WHERE id = $1`, runID,
	).Scan(&run.ID, &run.GuideID, &run.GuideRevision, &run.CharacterName, &run.League, &run.Status, &run.Notes,
		&run.StartedAt, &run.FinishedAt, &run.IsActive,
		&run.TimerStartedAt, &run.PausedAt, &run.TotalPausedMs)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("run: %d not found", runID)
		}
		return nil, fmt.Errorf("run: get %d: %w", runID, err)
	}
	return run, nil
}

// ListRuns returns all runs for a guide, most recent first.
func (r *Repository) ListRuns(ctx context.Context, guideID int) ([]RunSession, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, guide_id, guide_revision, character_name, league, status, notes, started_at, finished_at,
		       is_active, timer_started_at, paused_at, total_paused_ms
		FROM runs WHERE guide_id = $1 ORDER BY started_at DESC`, guideID)
	if err != nil {
		return nil, fmt.Errorf("run: list: %w", err)
	}
	defer rows.Close()
	runs := []RunSession{}
	for rows.Next() {
		var run RunSession
		if err := rows.Scan(&run.ID, &run.GuideID, &run.GuideRevision, &run.CharacterName, &run.League, &run.Status, &run.Notes,
			&run.StartedAt, &run.FinishedAt, &run.IsActive,
			&run.TimerStartedAt, &run.PausedAt, &run.TotalPausedMs); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

// ConfirmStep records that a step was completed.
func (r *Repository) ConfirmStep(ctx context.Context, runID, stepID int, by ConfirmedBy) (*Checkpoint, error) {
	cp := &Checkpoint{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO run_checkpoints (run_id, step_id, confirmed_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (run_id, step_id) DO UPDATE SET confirmed_at = NOW(), confirmed_by = EXCLUDED.confirmed_by
		RETURNING id, run_id, step_id, confirmed_at, confirmed_by`,
		runID, stepID, string(by),
	).Scan(&cp.ID, &cp.RunID, &cp.StepID, &cp.ConfirmedAt, &cp.ConfirmedBy)
	if err != nil {
		return nil, fmt.Errorf("run: confirm step: %w", err)
	}
	return cp, nil
}

// ListCheckpoints returns all confirmed checkpoints for a run.
func (r *Repository) ListCheckpoints(ctx context.Context, runID int) ([]Checkpoint, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, run_id, step_id, confirmed_at, confirmed_by
		FROM run_checkpoints WHERE run_id = $1 ORDER BY confirmed_at`, runID)
	if err != nil {
		return nil, fmt.Errorf("run: list checkpoints: %w", err)
	}
	defer rows.Close()
	cps := []Checkpoint{}
	for rows.Next() {
		var cp Checkpoint
		if err := rows.Scan(&cp.ID, &cp.RunID, &cp.StepID, &cp.ConfirmedAt, &cp.ConfirmedBy); err != nil {
			return nil, err
		}
		cps = append(cps, cp)
	}
	return cps, rows.Err()
}

// RecordEvent appends an event to the run's event log.
func (r *Repository) RecordEvent(ctx context.Context, runID int, eventType EventType, payload map[string]string) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("run: marshal event payload: %w", err)
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO run_events (run_id, event_type, payload)
		VALUES ($1, $2, $3)`,
		runID, string(eventType), data,
	)
	return err
}

// FinishRun marks a run as finished at the current time.
func (r *Repository) FinishRun(ctx context.Context, runID int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE runs SET is_active = FALSE, finished_at = $1 WHERE id = $2`,
		time.Now(), runID)
	return err
}

// RecordSplit records a timing split for a step.
func (r *Repository) RecordSplit(ctx context.Context, runID, stepID int, splitMs int64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO run_splits (run_id, step_id, split_ms)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`,
		runID, stepID, splitMs)
	return err
}

// GetDetailedRanking returns up to limit entries from local_rankings for a guide,
// ordered by rank ascending. Returns an empty slice if no runs have been computed yet.
func (r *Repository) GetDetailedRanking(ctx context.Context, guideID, limit int) ([]DetailedRankingEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT lr.rank, lr.run_id, ru.character_name, ru.started_at, lr.total_ms, lr.act_splits
		FROM local_rankings lr
		JOIN runs ru ON ru.id = lr.run_id
		WHERE lr.guide_id = $1
		ORDER BY lr.rank ASC
		LIMIT $2`, guideID, limit)
	if err != nil {
		return nil, fmt.Errorf("run: detailed ranking: %w", err)
	}
	defer rows.Close()
	entries := []DetailedRankingEntry{}
	for rows.Next() {
		var e DetailedRankingEntry
		var actJSON []byte
		if err := rows.Scan(&e.Rank, &e.RunID, &e.CharacterName, &e.StartedAt, &e.TotalMs, &actJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(actJSON, &e.ActSplits); err != nil {
			e.ActSplits = map[string]int{}
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(entries) > 0 {
		entries[0].IsPB = true
	}
	return entries, nil
}

// GetRankingStats returns aggregate timing statistics (count, PB, median, last run)
// for all finished runs of a guide.
func (r *Repository) GetRankingStats(ctx context.Context, guideID int) (*RankingStats, error) {
	rows, err := r.db.Query(ctx, `
		SELECT lr.run_id, lr.total_ms, lr.rank, ru.finished_at
		FROM local_rankings lr
		JOIN runs ru ON ru.id = lr.run_id
		WHERE lr.guide_id = $1
		ORDER BY lr.rank ASC`, guideID)
	if err != nil {
		return nil, fmt.Errorf("run: ranking stats: %w", err)
	}
	defer rows.Close()

	type row struct {
		runID      int
		totalMs    int64
		rank       int
		finishedAt *time.Time
	}
	var all []row
	for rows.Next() {
		var ro row
		if err := rows.Scan(&ro.runID, &ro.totalMs, &ro.rank, &ro.finishedAt); err != nil {
			return nil, err
		}
		all = append(all, ro)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	stats := &RankingStats{Count: len(all)}
	if len(all) == 0 {
		return stats, nil
	}

	// PB = rank 1 (already sorted)
	pbMs := all[0].totalMs
	stats.PBMs = &pbMs

	// Median
	medianMs := all[len(all)/2].totalMs
	stats.MedianMs = &medianMs

	// Last run by finished_at
	var lastRunID int
	var lastMs int64
	var latestFinish *time.Time
	for _, ro := range all {
		if ro.finishedAt != nil && (latestFinish == nil || ro.finishedAt.After(*latestFinish)) {
			latestFinish = ro.finishedAt
			lastRunID = ro.runID
			lastMs = ro.totalMs
		}
	}
	if latestFinish != nil {
		stats.LastRunMs = &lastMs
		stats.LastRunID = &lastRunID
	}
	return stats, nil
}

// ComputeAndUpsertRanking derives act_splits + total_ms for the given finished run,
// upserts the local_rankings row, then recomputes ranks for the whole guide.
func (r *Repository) ComputeAndUpsertRanking(ctx context.Context, runID, guideID int) error {
	// total_ms: finished_at - COALESCE(timer_started_at, started_at) - total_paused_ms
	var totalMs int64
	err := r.db.QueryRow(ctx, `
		SELECT GREATEST(0,
		    EXTRACT(EPOCH FROM (finished_at - COALESCE(timer_started_at, started_at))) * 1000
		    - total_paused_ms)
		FROM runs WHERE id = $1 AND finished_at IS NOT NULL`, runID).Scan(&totalMs)
	if err != nil {
		return fmt.Errorf("run: ranking compute total_ms: %w", err)
	}

	// act_splits: for each act, take the last (maximum) split_ms in that act
	actRows, err := r.db.Query(ctx, `
		SELECT gs.act, MAX(rs.split_ms)
		FROM run_splits rs
		JOIN guide_steps gs ON gs.id = rs.step_id
		WHERE rs.run_id = $1
		GROUP BY gs.act
		ORDER BY gs.act`, runID)
	if err != nil {
		return fmt.Errorf("run: ranking compute act_splits: %w", err)
	}
	defer actRows.Close()
	actSplits := map[string]int64{}
	for actRows.Next() {
		var act int
		var ms int64
		if err := actRows.Scan(&act, &ms); err != nil {
			return err
		}
		actSplits[fmt.Sprint(act)] = ms
	}
	if err := actRows.Err(); err != nil {
		return err
	}

	actJSON, err := json.Marshal(actSplits)
	if err != nil {
		return fmt.Errorf("run: ranking marshal act_splits: %w", err)
	}

	// Upsert local_rankings
	_, err = r.db.Exec(ctx, `
		INSERT INTO local_rankings (guide_id, run_id, total_ms, act_splits, rank, computed_at)
		VALUES ($1, $2, $3, $4, 0, NOW())
		ON CONFLICT (run_id)
		DO UPDATE SET total_ms = EXCLUDED.total_ms, act_splits = EXCLUDED.act_splits,
		              computed_at = NOW()`,
		guideID, runID, totalMs, actJSON)
	if err != nil {
		return fmt.Errorf("run: ranking upsert: %w", err)
	}

	// Recompute ranks for all runs of this guide
	_, err = r.db.Exec(ctx, `
		UPDATE local_rankings lr
		SET rank = sub.new_rank
		FROM (
		    SELECT id, ROW_NUMBER() OVER (ORDER BY total_ms ASC) AS new_rank
		    FROM local_rankings
		    WHERE guide_id = $1
		) sub
		WHERE lr.id = sub.id`, guideID)
	if err != nil {
		return fmt.Errorf("run: ranking rerank: %w", err)
	}
	return nil
}

// GetPBSplitsForGuide returns the run_id and splits of the rank-1 run for a guide.
// Returns (0, nil, nil) when no ranking entry exists yet.
func (r *Repository) GetPBSplitsForGuide(ctx context.Context, guideID int) (int, []Split, error) {
	var pbRunID int
	err := r.db.QueryRow(ctx,
		`SELECT run_id FROM local_rankings WHERE guide_id = $1 AND rank = 1`, guideID,
	).Scan(&pbRunID)
	if err == pgx.ErrNoRows {
		return 0, nil, nil
	}
	if err != nil {
		return 0, nil, fmt.Errorf("run: get pb run: %w", err)
	}
	splits, err := r.ListSplits(ctx, pbRunID)
	return pbRunID, splits, err
}

// GetPrevRunSplits returns the run_id and splits of the most recently finished
// run for guideID, excluding excludeRunID.
// Returns (0, nil, nil) when no such run exists.
func (r *Repository) GetPrevRunSplits(ctx context.Context, guideID, excludeRunID int) (int, []Split, error) {
	var prevRunID int
	err := r.db.QueryRow(ctx, `
		SELECT r.id
		FROM runs r
		JOIN local_rankings lr ON lr.run_id = r.id
		WHERE r.guide_id = $1 AND r.id != $2 AND r.finished_at IS NOT NULL
		ORDER BY r.finished_at DESC
		LIMIT 1`, guideID, excludeRunID,
	).Scan(&prevRunID)
	if err == pgx.ErrNoRows {
		return 0, nil, nil
	}
	if err != nil {
		return 0, nil, fmt.Errorf("run: get prev run: %w", err)
	}
	splits, err := r.ListSplits(ctx, prevRunID)
	return prevRunID, splits, err
}

// SetTimerStartedAt sets timer_started_at on a run (used for auto-start from logtail).
func (r *Repository) SetTimerStartedAt(ctx context.Context, runID int, t time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE runs SET timer_started_at = $1 WHERE id = $2 AND timer_started_at IS NULL`,
		t, runID)
	return err
}

// PauseRun records the current time as paused_at (idempotent if already paused).
func (r *Repository) PauseRun(ctx context.Context, runID int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE runs SET paused_at = NOW() WHERE id = $1 AND is_active = TRUE AND paused_at IS NULL`,
		runID)
	return err
}

// ResumeRun clears paused_at and accumulates the paused duration into total_paused_ms.
func (r *Repository) ResumeRun(ctx context.Context, runID int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE runs
		SET total_paused_ms = total_paused_ms +
		    GREATEST(0, EXTRACT(EPOCH FROM (NOW() - paused_at)) * 1000)::BIGINT,
		    paused_at = NULL
		WHERE id = $1 AND is_active = TRUE AND paused_at IS NOT NULL`, runID)
	return err
}

// AbandonRun marks a run as abandoned.
func (r *Repository) AbandonRun(ctx context.Context, runID int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE runs SET is_active = FALSE, status = 'abandoned', finished_at = $1 WHERE id = $2`,
		time.Now(), runID)
	return err
}

// SkipStep upserts a run_step_progress row with status=skipped.
func (r *Repository) SkipStep(ctx context.Context, runID, stepID int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO run_step_progress (run_id, step_id, status, confirmed_by)
		VALUES ($1, $2, 'skipped', 'manual')
		ON CONFLICT (run_id, step_id)
			DO UPDATE SET status = 'skipped', confirmed_at = NOW()`,
		runID, stepID)
	return err
}

// UndoStep removes the checkpoint and step_progress record for a step in a run.
func (r *Repository) UndoStep(ctx context.Context, runID, stepID int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("run: begin undo tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `
		DELETE FROM run_checkpoints WHERE run_id = $1 AND step_id = $2`, runID, stepID); err != nil {
		return fmt.Errorf("run: undo checkpoint: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM run_step_progress WHERE run_id = $1 AND step_id = $2`, runID, stepID); err != nil {
		return fmt.Errorf("run: undo step_progress: %w", err)
	}
	return tx.Commit(ctx)
}

// UpsertCharacter inserts or updates the character record for a run.
func (r *Repository) UpsertCharacter(ctx context.Context, c *Character) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO run_characters
			(run_id, character_name, character_class, league, level_at_start, level_current, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (run_id)
			DO UPDATE SET
				character_name  = EXCLUDED.character_name,
				character_class = EXCLUDED.character_class,
				league          = EXCLUDED.league,
				level_at_start  = EXCLUDED.level_at_start,
				level_current   = EXCLUDED.level_current,
				last_seen_at    = EXCLUDED.last_seen_at
		RETURNING id, created_at`,
		c.RunID, c.CharacterName, c.CharacterClass, c.League,
		c.LevelAtStart, c.LevelCurrent, c.LastSeenAt,
	).Scan(&c.ID, &c.CreatedAt)
}

// GetCharacter returns the character record for a run.
func (r *Repository) GetCharacter(ctx context.Context, runID int) (*Character, error) {
	c := &Character{}
	err := r.db.QueryRow(ctx, `
		SELECT id, run_id, character_name, character_class, league,
		       level_at_start, level_current, last_seen_at, created_at
		FROM run_characters WHERE run_id = $1`, runID,
	).Scan(&c.ID, &c.RunID, &c.CharacterName, &c.CharacterClass, &c.League,
		&c.LevelAtStart, &c.LevelCurrent, &c.LastSeenAt, &c.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("run: character for run %d not found", runID)
		}
		return nil, fmt.Errorf("run: get character: %w", err)
	}
	return c, nil
}

// ListSnapshots returns character snapshots for a run ordered by captured_at descending.
func (r *Repository) ListSnapshots(ctx context.Context, runID int) ([]CharacterSnapshot, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, run_id, captured_at, source, level,
		       life_max, mana_max, res_fire, res_cold, res_lightning, res_chaos
		FROM character_snapshots WHERE run_id = $1 ORDER BY captured_at DESC`, runID)
	if err != nil {
		return nil, fmt.Errorf("run: list snapshots: %w", err)
	}
	defer rows.Close()
	snaps := []CharacterSnapshot{}
	for rows.Next() {
		var s CharacterSnapshot
		var src string
		if err := rows.Scan(&s.ID, &s.RunID, &s.CapturedAt, &src, &s.Level,
			&s.LifeMax, &s.ManaMax, &s.ResFire, &s.ResCold, &s.ResLightning, &s.ResChaos); err != nil {
			return nil, err
		}
		s.Source = SnapshotSource(src)
		snaps = append(snaps, s)
	}
	return snaps, rows.Err()
}

// CreateSnapshot inserts a new character snapshot (scalar fields only; JSONB caches default to {}).
func (r *Repository) CreateSnapshot(ctx context.Context, s *CharacterSnapshot) error {
	var src string
	if err := r.db.QueryRow(ctx, `
		INSERT INTO character_snapshots
			(run_id, source, level, life_max, mana_max, res_fire, res_cold, res_lightning, res_chaos)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, captured_at, source`,
		s.RunID, string(s.Source), s.Level,
		s.LifeMax, s.ManaMax, s.ResFire, s.ResCold, s.ResLightning, s.ResChaos,
	).Scan(&s.ID, &s.CapturedAt, &src); err != nil {
		return fmt.Errorf("run: create snapshot: %w", err)
	}
	s.Source = SnapshotSource(src)
	return nil
}

// CreateGGGSnapshot inserts a GGG-sourced snapshot including JSONB cache fields
// (equipped_items, skills, raw_response).
func (r *Repository) CreateGGGSnapshot(ctx context.Context, s *CharacterSnapshot) error {
	equippedJSON, err := json.Marshal(s.EquippedItems)
	if err != nil {
		equippedJSON = []byte("{}")
	}
	skillsJSON, err := json.Marshal(s.Skills)
	if err != nil {
		skillsJSON = []byte("{}")
	}
	rawJSON, err := json.Marshal(s.RawResponse)
	if err != nil {
		rawJSON = []byte("{}")
	}
	var src string
	if err := r.db.QueryRow(ctx, `
		INSERT INTO character_snapshots
			(run_id, source, level, life_max, mana_max, res_fire, res_cold, res_lightning, res_chaos,
			 equipped_items, skills, raw_response)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, captured_at, source`,
		s.RunID, string(SnapshotGGG), s.Level,
		s.LifeMax, s.ManaMax, s.ResFire, s.ResCold, s.ResLightning, s.ResChaos,
		equippedJSON, skillsJSON, rawJSON,
	).Scan(&s.ID, &s.CapturedAt, &src); err != nil {
		return fmt.Errorf("run: create ggg snapshot: %w", err)
	}
	s.Source = SnapshotSource(src)
	return nil
}

// ListEvents returns log events for a run, most recent first, up to limit rows.
func (r *Repository) ListEvents(ctx context.Context, runID, limit int) ([]Event, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, run_id, event_type, payload, occurred_at
		FROM run_events WHERE run_id = $1 ORDER BY occurred_at DESC LIMIT $2`, runID, limit)
	if err != nil {
		return nil, fmt.Errorf("run: list events: %w", err)
	}
	defer rows.Close()
	events := []Event{}
	for rows.Next() {
		var e Event
		var evType string
		var payloadRaw []byte
		if err := rows.Scan(&e.ID, &e.RunID, &evType, &payloadRaw, &e.OccurredAt); err != nil {
			return nil, err
		}
		e.EventType = EventType(evType)
		if err := json.Unmarshal(payloadRaw, &e.Payload); err != nil {
			e.Payload = map[string]string{}
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// ListSplits returns all splits for a run ordered by split_ms.
func (r *Repository) ListSplits(ctx context.Context, runID int) ([]Split, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, run_id, step_id, split_ms, recorded_at
		FROM run_splits WHERE run_id = $1 ORDER BY split_ms`, runID)
	if err != nil {
		return nil, fmt.Errorf("run: list splits: %w", err)
	}
	defer rows.Close()
	splits := []Split{}
	for rows.Next() {
		var s Split
		if err := rows.Scan(&s.ID, &s.RunID, &s.StepID, &s.SplitMs, &s.RecordedAt); err != nil {
			return nil, err
		}
		splits = append(splits, s)
	}
	return splits, rows.Err()
}

// ListPendingChecks returns all unanswered manual checks for a run.
func (r *Repository) ListPendingChecks(ctx context.Context, runID int) ([]ManualCheck, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, run_id, step_id, check_type, prompt, is_confirmed, response_value, confirmed_at, created_at
		FROM manual_checks WHERE run_id = $1 AND is_confirmed = FALSE ORDER BY created_at`, runID)
	if err != nil {
		return nil, fmt.Errorf("run: list pending checks: %w", err)
	}
	defer rows.Close()
	checks := []ManualCheck{}
	for rows.Next() {
		var mc ManualCheck
		var checkType string
		if err := rows.Scan(&mc.ID, &mc.RunID, &mc.StepID, &checkType, &mc.Prompt,
			&mc.IsConfirmed, &mc.ResponseValue, &mc.ConfirmedAt, &mc.CreatedAt); err != nil {
			return nil, err
		}
		mc.CheckType = CheckType(checkType)
		checks = append(checks, mc)
	}
	return checks, rows.Err()
}

// AnswerCheck marks a manual check as confirmed with the given response value.
func (r *Repository) AnswerCheck(ctx context.Context, checkID int, responseValue string) (*ManualCheck, error) {
	mc := &ManualCheck{}
	var checkType string
	err := r.db.QueryRow(ctx, `
		UPDATE manual_checks
		SET is_confirmed = TRUE, response_value = $2, confirmed_at = NOW()
		WHERE id = $1
		RETURNING id, run_id, step_id, check_type, prompt, is_confirmed, response_value, confirmed_at, created_at`,
		checkID, responseValue,
	).Scan(&mc.ID, &mc.RunID, &mc.StepID, &checkType, &mc.Prompt,
		&mc.IsConfirmed, &mc.ResponseValue, &mc.ConfirmedAt, &mc.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("run: check %d not found", checkID)
		}
		return nil, fmt.Errorf("run: answer check: %w", err)
	}
	mc.CheckType = CheckType(checkType)
	return mc, nil
}

// CreateManualCheck inserts a new unanswered manual check for a run.
func (r *Repository) CreateManualCheck(ctx context.Context, mc *ManualCheck) error {
	var checkType string
	return r.db.QueryRow(ctx, `
		INSERT INTO manual_checks (run_id, step_id, check_type, prompt)
		VALUES ($1, $2, $3, $4)
		RETURNING id, check_type, created_at`,
		mc.RunID, mc.StepID, string(mc.CheckType), mc.Prompt,
	).Scan(&mc.ID, &checkType, &mc.CreatedAt)
}
