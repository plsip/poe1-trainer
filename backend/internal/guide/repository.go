package guide

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides persistence for guides.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Save persists a parsed Guide (upsert by slug).
// Returns the guide with IDs filled in.
func (r *Repository) Save(ctx context.Context, g *Guide) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("guide: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var id int
	err = tx.QueryRow(ctx, `
		INSERT INTO guides (slug, title, build_name, version)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (slug) DO UPDATE
			SET title = EXCLUDED.title,
			    build_name = EXCLUDED.build_name,
			    version = EXCLUDED.version
		RETURNING id`,
		g.Slug, g.Title, g.BuildName, g.Version,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("guide: upsert guide: %w", err)
	}
	g.ID = id

	// Delete existing steps for re-import.
	if _, err := tx.Exec(ctx, `DELETE FROM guide_steps WHERE guide_id = $1`, id); err != nil {
		return fmt.Errorf("guide: delete old steps: %w", err)
	}

	for i := range g.Steps {
		step := &g.Steps[i]
		step.GuideID = id
		var stepID int
		err = tx.QueryRow(ctx, `
			INSERT INTO guide_steps
				(guide_id, step_number, act, section, title, description, area,
				 quest_name, step_type, completion_mode,
				 is_checkpoint, requires_manual, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			RETURNING id`,
			id, step.StepNumber, step.Act, step.Section,
			step.Title, step.Description, step.Area,
			step.QuestName, string(step.StepType), string(step.CompletionMode),
			step.IsCheckpoint, step.RequiresManual, step.SortOrder,
		).Scan(&stepID)
		if err != nil {
			return fmt.Errorf("guide: insert step %d: %w", step.StepNumber, err)
		}
		step.ID = stepID

		for j := range step.GemRequirements {
			gem := &step.GemRequirements[j]
			gem.StepID = stepID
			var gemID int
			err = tx.QueryRow(ctx, `
				INSERT INTO guide_gem_requirements (step_id, gem_name, color, note)
				VALUES ($1,$2,$3,$4) RETURNING id`,
				stepID, gem.GemName, gem.Color, gem.Note,
			).Scan(&gemID)
			if err != nil {
				return fmt.Errorf("guide: insert gem req for step %d: %w", step.StepNumber, err)
			}
			gem.ID = gemID
		}

		for k := range step.Conditions {
			cond := &step.Conditions[k]
			cond.StepID = stepID
			payloadJSON, err := json.Marshal(cond.Payload)
			if err != nil {
				return fmt.Errorf("guide: marshal condition payload for step %d: %w", step.StepNumber, err)
			}
			var condID int
			err = tx.QueryRow(ctx, `
				INSERT INTO guide_step_conditions
					(step_id, condition_type, payload, priority, notes)
				VALUES ($1,$2,$3,$4,$5) RETURNING id`,
				stepID, string(cond.ConditionType), payloadJSON,
				cond.Priority, cond.Notes,
			).Scan(&condID)
			if err != nil {
				return fmt.Errorf("guide: insert condition for step %d: %w", step.StepNumber, err)
			}
			cond.ID = condID
		}
	}

	return tx.Commit(ctx)
}

// GetBySlug fetches a guide with all its steps and gem requirements.
func (r *Repository) GetBySlug(ctx context.Context, slug string) (*Guide, error) {
	g := &Guide{}
	err := r.db.QueryRow(ctx, `
		SELECT id, slug, title, build_name, version, created_at
		FROM guides WHERE slug = $1`, slug,
	).Scan(&g.ID, &g.Slug, &g.Title, &g.BuildName, &g.Version, &g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("guide: get by slug %q: %w", slug, err)
	}
	steps, err := r.loadSteps(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	g.Steps = steps
	return g, nil
}

// GetByID fetches a guide by its primary key.
func (r *Repository) GetByID(ctx context.Context, id int) (*Guide, error) {
	g := &Guide{}
	err := r.db.QueryRow(ctx, `
		SELECT id, slug, title, build_name, version, created_at
		FROM guides WHERE id = $1`, id,
	).Scan(&g.ID, &g.Slug, &g.Title, &g.BuildName, &g.Version, &g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("guide: get by id %d: %w", id, err)
	}
	steps, err := r.loadSteps(ctx, id)
	if err != nil {
		return nil, err
	}
	g.Steps = steps
	return g, nil
}

// List returns all guides without their steps.
func (r *Repository) List(ctx context.Context) ([]Guide, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, title, build_name, version, created_at FROM guides ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("guide: list: %w", err)
	}
	defer rows.Close()
	guides := []Guide{}
	for rows.Next() {
		var g Guide
		if err := rows.Scan(&g.ID, &g.Slug, &g.Title, &g.BuildName, &g.Version, &g.CreatedAt); err != nil {
			return nil, err
		}
		guides = append(guides, g)
	}
	return guides, rows.Err()
}

func (r *Repository) loadSteps(ctx context.Context, guideID int) ([]Step, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, guide_id, step_number, act, section, title, description, area,
		       quest_name, step_type, completion_mode,
		       is_checkpoint, requires_manual, sort_order
		FROM guide_steps WHERE guide_id = $1 ORDER BY sort_order`, guideID)
	if err != nil {
		return nil, fmt.Errorf("guide: load steps: %w", err)
	}
	defer rows.Close()

	var steps []Step
	var stepIDs []int
	for rows.Next() {
		var s Step
		var stepType, completionMode string
		if err := rows.Scan(
			&s.ID, &s.GuideID, &s.StepNumber, &s.Act, &s.Section,
			&s.Title, &s.Description, &s.Area,
			&s.QuestName, &stepType, &completionMode,
			&s.IsCheckpoint, &s.RequiresManual, &s.SortOrder,
		); err != nil {
			return nil, err
		}
		s.StepType = StepType(stepType)
		s.CompletionMode = CompletionMode(completionMode)
		steps = append(steps, s)
		stepIDs = append(stepIDs, s.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load gem requirements in bulk.
	if len(stepIDs) == 0 {
		return steps, nil
	}
	gemRows, err := r.db.Query(ctx, `
		SELECT id, step_id, gem_name, color, note
		FROM guide_gem_requirements
		WHERE step_id = ANY($1)`, stepIDs)
	if err != nil {
		return nil, fmt.Errorf("guide: load gems: %w", err)
	}
	defer gemRows.Close()

	gemsByStep := map[int][]GemRequirement{}
	for gemRows.Next() {
		var gem GemRequirement
		if err := gemRows.Scan(&gem.ID, &gem.StepID, &gem.GemName, &gem.Color, &gem.Note); err != nil {
			return nil, err
		}
		gemsByStep[gem.StepID] = append(gemsByStep[gem.StepID], gem)
	}
	if err := gemRows.Err(); err != nil {
		return nil, err
	}

	for i := range steps {
		steps[i].GemRequirements = gemsByStep[steps[i].ID]
	}

	// Load step conditions in bulk.
	condRows, err := r.db.Query(ctx, `
		SELECT id, step_id, condition_type, payload, priority, notes
		FROM guide_step_conditions
		WHERE step_id = ANY($1)
		ORDER BY step_id, priority`, stepIDs)
	if err != nil {
		return nil, fmt.Errorf("guide: load conditions: %w", err)
	}
	defer condRows.Close()

	condsByStep := map[int][]StepCondition{}
	for condRows.Next() {
		var c StepCondition
		var condType string
		var payloadRaw []byte
		if err := condRows.Scan(&c.ID, &c.StepID, &condType, &payloadRaw, &c.Priority, &c.Notes); err != nil {
			return nil, err
		}
		c.ConditionType = ConditionType(condType)
		if err := json.Unmarshal(payloadRaw, &c.Payload); err != nil {
			c.Payload = map[string]string{}
		}
		condsByStep[c.StepID] = append(condsByStep[c.StepID], c)
	}
	if err := condRows.Err(); err != nil {
		return nil, err
	}

	for i := range steps {
		steps[i].Conditions = condsByStep[steps[i].ID]
	}
	return steps, nil
}

// GetStepByID returns a single step.
func (r *Repository) GetStepByID(ctx context.Context, stepID int) (*Step, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, guide_id, step_number, act, section, title, description, area,
		       quest_name, step_type, completion_mode,
		       is_checkpoint, requires_manual, sort_order
		FROM guide_steps WHERE id = $1`, stepID)
	var s Step
	var stepType, completionMode string
	if err := row.Scan(
		&s.ID, &s.GuideID, &s.StepNumber, &s.Act, &s.Section,
		&s.Title, &s.Description, &s.Area,
		&s.QuestName, &stepType, &completionMode,
		&s.IsCheckpoint, &s.RequiresManual, &s.SortOrder,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("guide: step %d not found", stepID)
		}
		return nil, fmt.Errorf("guide: get step: %w", err)
	}
	s.StepType = StepType(stepType)
	s.CompletionMode = CompletionMode(completionMode)
	return &s, nil
}
