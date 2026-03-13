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
func (s *Service) CreateRun(ctx context.Context, guideID int, characterName string) (*RunSession, error) {
	// Verify the guide exists.
	if _, err := s.guideRepo.GetByID(ctx, guideID); err != nil {
		return nil, fmt.Errorf("run: guide %d not found: %w", guideID, err)
	}
	run, err := s.repo.CreateRun(ctx, guideID, characterName)
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
