package run

import (
	"context"
	"fmt"
	"time"

	"github.com/poe1-trainer/internal/guide"
)

// Service orchestrates run lifecycle and step confirmation.
type Service struct {
	repo      *Repository
	guideRepo *guide.Repository
}

// NewService creates a new Service.
func NewService(repo *Repository, guideRepo *guide.Repository) *Service {
	return &Service{repo: repo, guideRepo: guideRepo}
}

// CreateRun starts a new run for the given guide.
func (s *Service) CreateRun(ctx context.Context, guideID int, characterName, league string) (*RunSession, error) {
	// Verify the guide exists.
	if _, err := s.guideRepo.GetByID(ctx, guideID); err != nil {
		return nil, fmt.Errorf("run: guide %d not found: %w", guideID, err)
	}
	run, err := s.repo.CreateRun(ctx, guideID, characterName, league)
	if err != nil {
		return nil, fmt.Errorf("run: create: %w", err)
	}
	return run, nil
}

// ConfirmStep marks a guide step as completed in the given run.
// Only steps belonging to the guide of the run are accepted.
func (s *Service) ConfirmStep(ctx context.Context, runID, stepID int, by ConfirmedBy) (*Checkpoint, error) {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	if !run.IsActive {
		return nil, fmt.Errorf("run: run %d is not active", runID)
	}

	// Verify step belongs to the run's guide.
	step, err := s.guideRepo.GetStepByID(ctx, stepID)
	if err != nil {
		return nil, fmt.Errorf("run: step %d: %w", stepID, err)
	}
	if step.GuideID != run.GuideID {
		return nil, fmt.Errorf("run: step %d does not belong to guide %d", stepID, run.GuideID)
	}

	cp, err := s.repo.ConfirmStep(ctx, runID, stepID, by)
	if err != nil {
		return nil, err
	}

	// Record event.
	_ = s.repo.RecordEvent(ctx, runID, EventStepConfirmed, map[string]string{
		"step_id": fmt.Sprint(stepID),
		"by":      string(by),
	})
	return cp, nil
}

// GetCurrentState returns the aggregated state for an active run.
func (s *Service) GetCurrentState(ctx context.Context, runID int) (*CurrentState, error) {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}

	checkpoints, err := s.repo.ListCheckpoints(ctx, runID)
	if err != nil {
		return nil, err
	}

	confirmedIDs := make([]int, 0, len(checkpoints))
	for _, cp := range checkpoints {
		confirmedIDs = append(confirmedIDs, cp.StepID)
	}

	// Determine current step: first unconfirmed step in guide order.
	g, err := s.guideRepo.GetByID(ctx, run.GuideID)
	if err != nil {
		return nil, err
	}

	confirmedSet := make(map[int]bool, len(confirmedIDs))
	for _, id := range confirmedIDs {
		confirmedSet[id] = true
	}

	currentStepID := 0
	for _, step := range g.Steps {
		if !confirmedSet[step.ID] {
			currentStepID = step.ID
			break
		}
	}

	var elapsedMs int64
	if run.IsActive {
		elapsedMs = time.Since(run.StartedAt).Milliseconds()
	} else if run.FinishedAt != nil {
		elapsedMs = run.FinishedAt.Sub(run.StartedAt).Milliseconds()
	}

	return &CurrentState{
		Run:              *run,
		CurrentStepID:    currentStepID,
		ConfirmedStepIDs: confirmedIDs,
		ElapsedMs:        elapsedMs,
	}, nil
}

// HandleAreaEvent processes an event emitted by the logtail watcher.
// It does not automatically confirm steps — it only records the event.
// The run service treats logtail signals as informational, not authoritative.
func (s *Service) HandleAreaEvent(ctx context.Context, runID int, ev AreaEvent) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if !run.IsActive {
		return nil // silently drop events for finished runs
	}
	return s.repo.RecordEvent(ctx, runID, EventAreaEntered, map[string]string{
		"area": ev.AreaName,
	})
}

// FinishRun marks a run as completed.
func (s *Service) FinishRun(ctx context.Context, runID int) error {
	return s.repo.FinishRun(ctx, runID)
}

// ListRuns returns all runs for a given guide, most recent first.
func (s *Service) ListRuns(ctx context.Context, guideID int) ([]RunSession, error) {
	return s.repo.ListRuns(ctx, guideID)
}

// AbandonRun marks a run as abandoned.
func (s *Service) AbandonRun(ctx context.Context, runID int) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if !run.IsActive {
		return fmt.Errorf("run: run %d is not active", runID)
	}
	return s.repo.AbandonRun(ctx, runID)
}

// SkipStep records a step as skipped in the current run.
func (s *Service) SkipStep(ctx context.Context, runID, stepID int) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if !run.IsActive {
		return fmt.Errorf("run: run %d is not active", runID)
	}
	step, err := s.guideRepo.GetStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("run: step %d: %w", stepID, err)
	}
	if step.GuideID != run.GuideID {
		return fmt.Errorf("run: step %d does not belong to guide %d", stepID, run.GuideID)
	}
	return s.repo.SkipStep(ctx, runID, stepID)
}

// UndoStep removes a step confirmation from the current run.
func (s *Service) UndoStep(ctx context.Context, runID, stepID int) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if !run.IsActive {
		return fmt.Errorf("run: run %d is not active", runID)
	}
	return s.repo.UndoStep(ctx, runID, stepID)
}

// UpsertCharacter inserts or updates the character record for a run.
func (s *Service) UpsertCharacter(ctx context.Context, c *Character) error {
	if _, err := s.repo.GetRun(ctx, c.RunID); err != nil {
		return err
	}
	return s.repo.UpsertCharacter(ctx, c)
}

// GetCharacter returns the character record for a run.
func (s *Service) GetCharacter(ctx context.Context, runID int) (*Character, error) {
	return s.repo.GetCharacter(ctx, runID)
}

// ListSnapshots returns all character snapshots for a run.
func (s *Service) ListSnapshots(ctx context.Context, runID int) ([]CharacterSnapshot, error) {
	if _, err := s.repo.GetRun(ctx, runID); err != nil {
		return nil, err
	}
	return s.repo.ListSnapshots(ctx, runID)
}

// CreateSnapshot adds a new manual character snapshot for a run.
func (s *Service) CreateSnapshot(ctx context.Context, snap *CharacterSnapshot) error {
	if _, err := s.repo.GetRun(ctx, snap.RunID); err != nil {
		return err
	}
	snap.Source = SnapshotManual
	return s.repo.CreateSnapshot(ctx, snap)
}

// ListEvents returns recent events for a run (up to limit entries).
func (s *Service) ListEvents(ctx context.Context, runID, limit int) ([]Event, error) {
	if _, err := s.repo.GetRun(ctx, runID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListEvents(ctx, runID, limit)
}

// RecordAreaEvent processes an area event from the logtail watcher.
// It does not automatically confirm steps — it only records the event.
func (s *Service) RecordAreaEvent(ctx context.Context, runID int, ev AreaEvent) error {
	return s.HandleAreaEvent(ctx, runID, ev)
}

// ListSplits returns timing splits for a run.
func (s *Service) ListSplits(ctx context.Context, runID int) ([]Split, error) {
	if _, err := s.repo.GetRun(ctx, runID); err != nil {
		return nil, err
	}
	return s.repo.ListSplits(ctx, runID)
}

// RecordSplit records a timing split for a step in a run.
func (s *Service) RecordSplit(ctx context.Context, runID, stepID int, splitMs int64) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if !run.IsActive {
		return fmt.Errorf("run: run %d is not active", runID)
	}
	return s.repo.RecordSplit(ctx, runID, stepID, splitMs)
}

// ListPendingChecks returns all unanswered manual checks for a run.
func (s *Service) ListPendingChecks(ctx context.Context, runID int) ([]ManualCheck, error) {
	if _, err := s.repo.GetRun(ctx, runID); err != nil {
		return nil, err
	}
	return s.repo.ListPendingChecks(ctx, runID)
}

// AnswerCheck confirms a manual check with the given response.
func (s *Service) AnswerCheck(ctx context.Context, checkID int, responseValue string) (*ManualCheck, error) {
	return s.repo.AnswerCheck(ctx, checkID, responseValue)
}
