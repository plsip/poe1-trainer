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
func (r *Repository) CreateRun(ctx context.Context, guideID int, characterName string) (*RunSession, error) {
	run := &RunSession{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO runs (guide_id, character_name)
		VALUES ($1, $2)
		RETURNING id, guide_id, character_name, started_at, finished_at, is_active`,
		guideID, characterName,
	).Scan(&run.ID, &run.GuideID, &run.CharacterName, &run.StartedAt, &run.FinishedAt, &run.IsActive)
	if err != nil {
		return nil, fmt.Errorf("run: insert: %w", err)
	}
	return run, nil
}

// GetRun fetches a single run by ID.
func (r *Repository) GetRun(ctx context.Context, runID int) (*RunSession, error) {
	run := &RunSession{}
	err := r.db.QueryRow(ctx, `
		SELECT id, guide_id, character_name, started_at, finished_at, is_active
		FROM runs WHERE id = $1`, runID,
	).Scan(&run.ID, &run.GuideID, &run.CharacterName, &run.StartedAt, &run.FinishedAt, &run.IsActive)
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
		SELECT id, guide_id, character_name, started_at, finished_at, is_active
		FROM runs WHERE guide_id = $1 ORDER BY started_at DESC`, guideID)
	if err != nil {
		return nil, fmt.Errorf("run: list: %w", err)
	}
	defer rows.Close()
	runs := []RunSession{}
	for rows.Next() {
		var run RunSession
		if err := rows.Scan(&run.ID, &run.GuideID, &run.CharacterName,
			&run.StartedAt, &run.FinishedAt, &run.IsActive); err != nil {
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

// GetRanking returns the top N runs for a guide ordered by total time (fastest first).
func (r *Repository) GetRanking(ctx context.Context, guideID, limit int) ([]RankingEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.character_name, r.started_at,
		       EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000 AS total_ms
		FROM runs r
		WHERE r.guide_id = $1 AND r.finished_at IS NOT NULL
		ORDER BY total_ms ASC
		LIMIT $2`,
		guideID, limit)
	if err != nil {
		return nil, fmt.Errorf("run: ranking: %w", err)
	}
	defer rows.Close()
	entries := []RankingEntry{}
	for rows.Next() {
		var e RankingEntry
		if err := rows.Scan(&e.RunID, &e.CharacterName, &e.StartedAt, &e.TotalMs); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// RankingEntry is a row in the local ranking table.
type RankingEntry struct {
	RunID         int       `json:"run_id"`
	CharacterName string    `json:"character_name"`
	StartedAt     time.Time `json:"started_at"`
	TotalMs       float64   `json:"total_ms"`
}
