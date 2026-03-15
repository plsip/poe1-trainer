package run

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/poe1-trainer/internal/game"
	"github.com/poe1-trainer/internal/guide"
)

// Service orchestrates run lifecycle and step confirmation.
type Service struct {
	repo      *Repository
	guideRepo *guide.Repository
	gameRepo  game.AreaRepository
	mu        sync.Mutex
	areaHints map[int]areaHint
}

type areaHint struct {
	areaCode   string
	areaLevel  int
	occurredAt time.Time
}

// NewService creates a new Service.
func NewService(repo *Repository, guideRepo *guide.Repository, gameRepo game.AreaRepository) *Service {
	return &Service{repo: repo, guideRepo: guideRepo, gameRepo: gameRepo, areaHints: map[int]areaHint{}}
}

// CreateRun starts a new run for the given guide.
// When autoStart=true the run timer does not begin until the first logtail area event.
func (s *Service) CreateRun(ctx context.Context, guideID int, characterName, league string, autoStart bool) (*RunSession, error) {
	// Verify the guide exists.
	if _, err := s.guideRepo.GetByID(ctx, guideID); err != nil {
		return nil, fmt.Errorf("run: guide %d not found: %w", guideID, err)
	}
	run, err := s.repo.CreateRun(ctx, guideID, characterName, league, autoStart)
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

// ConfirmAct confirms all unconfirmed steps for a specific act.
func (s *Service) ConfirmAct(ctx context.Context, runID, act int, by ConfirmedBy) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if !run.IsActive {
		return fmt.Errorf("run: run %d is not active", runID)
	}

	guide, err := s.guideRepo.GetByID(ctx, run.GuideID)
	if err != nil {
		return err
	}

	checkpoints, err := s.repo.ListCheckpoints(ctx, runID)
	if err != nil {
		return err
	}
	confirmed := make(map[int]bool)
	for _, cp := range checkpoints {
		confirmed[cp.StepID] = true
	}

	for _, step := range guide.Steps {
		if step.Act == act && !confirmed[step.ID] {
			if _, err := s.repo.ConfirmStep(ctx, runID, step.ID, by); err != nil {
				return err
			}
			_ = s.repo.RecordEvent(ctx, runID, EventStepConfirmed, map[string]string{
				"step_id": fmt.Sprint(step.ID),
				"by":      string(by),
				"act":     fmt.Sprint(act),
			})
		}
	}
	return nil
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
	g, err := s.loadGuideForRun(ctx, run)
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

	stepTimings, err := s.buildStepTimings(ctx, run, g, confirmedSet, currentStepID)
	if err != nil {
		return nil, err
	}

	var elapsedMs int64
	timerStart := run.StartedAt
	if run.TimerStartedAt != nil {
		timerStart = *run.TimerStartedAt
	}
	if run.IsActive && run.TimerStartedAt != nil {
		raw := time.Since(timerStart).Milliseconds() - run.TotalPausedMs
		if run.PausedAt != nil {
			// don't advance elapsed while paused
			raw = timerStart.Sub(timerStart).Milliseconds()
			// elapsed up to pause moment
			raw = run.PausedAt.Sub(timerStart).Milliseconds() - run.TotalPausedMs
		}
		if raw < 0 {
			raw = 0
		}
		elapsedMs = raw
	} else if !run.IsActive && run.FinishedAt != nil && run.TimerStartedAt != nil {
		elapsedMs = run.FinishedAt.Sub(timerStart).Milliseconds() - run.TotalPausedMs
		if elapsedMs < 0 {
			elapsedMs = 0
		}
	}

	return &CurrentState{
		Run:              *run,
		CurrentStepID:    currentStepID,
		ConfirmedStepIDs: confirmedIDs,
		ElapsedMs:        elapsedMs,
		StepTimings:      stepTimings,
	}, nil
}

// HandleAreaEvent processes an event emitted by the logtail watcher.
// It does not automatically confirm steps — it only records the event.
// The run service treats logtail signals as informational, not authoritative.
// Side effect: if timer_started_at is NULL (auto-start mode), it is set to now.
func (s *Service) HandleAreaEvent(ctx context.Context, runID int, ev AreaEvent) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if !run.IsActive {
		return nil // silently drop events for finished runs
	}
	processedAt := ev.OccurredAt
	if processedAt.IsZero() {
		processedAt = time.Now()
	}
	if ev.AreaCode == "" && ev.AreaName != "" {
		if hint, ok := s.takeAreaHint(runID, processedAt); ok {
			ev.AreaCode = hint.areaCode
			ev.AreaLevel = hint.areaLevel
		} else {
			if areas, err := s.gameRepo.GetByName(ctx, ev.AreaName); err == nil && len(areas) == 1 {
				ev.AreaCode = areas[0].AreaCode
			}
		}
	} else if ev.AreaCode != "" && ev.AreaName == "" {
		if area, err := s.gameRepo.GetByCode(ctx, ev.AreaCode); err == nil && area != nil {
			ev.AreaName = area.Name
		}
	}
	// Auto-start: first area event starts the timer.
	if run.TimerStartedAt == nil {
		if err := s.repo.SetTimerStartedAt(ctx, runID, processedAt); err != nil {
			return err
		}
		run.TimerStartedAt = &processedAt
	}
	if err := s.repo.RecordEvent(ctx, runID, EventAreaEntered, map[string]string{
		"area":      ev.AreaName,
		"area_code": ev.AreaCode,
	}); err != nil {
		return err
	}

	g, err := s.loadGuideForRun(ctx, run)
	if err != nil {
		return err
	}
	splits, err := s.repo.ListSplits(ctx, runID)
	if err != nil {
		return err
	}
	recorded := make(map[int]bool, len(splits))
	for _, split := range splits {
		recorded[split.StepID] = true
	}
	checkpoints, err := s.repo.ListCheckpoints(ctx, runID)
	if err != nil {
		return err
	}
	confirmedSet := make(map[int]bool, len(checkpoints))
	for _, cp := range checkpoints {
		confirmedSet[cp.StepID] = true
	}
	act := 0
	if area, err := s.gameRepo.GetByCode(ctx, ev.AreaCode); err == nil && area != nil && area.Act != nil {
		act = *area.Act
	} else {
		act = actFromAreaCode(ev.AreaCode)
	}
	// Use the confirmed-checkpoint watermark so that a prematurely recorded
	// split for a "future" step (e.g. step 53 when the user first enters
	// Forest Encampment in act 2) does not raise minSortOrd and block correct
	// matching for earlier logtail steps (e.g. steps 47-49 in act 2).
	minSortOrd := lastRecordedSortOrder(g.Steps, confirmedSet)
	step := firstAreaStepForArea(g.Steps, ev.AreaName, act, recorded, minSortOrd)
	slog.Info("logtail: area split probe",
		"run_id", runID,
		"area", ev.AreaName,
		"area_code", ev.AreaCode,
		"act", act,
		"min_sort_order", minSortOrd,
		"recorded_count", len(recorded),
		"step_found", step != nil,
	)
	if step == nil {
		return nil
	}

	splitMs := elapsedAt(run, processedAt)
	slog.Info("logtail: recording split",
		"run_id", runID,
		"step_id", step.ID,
		"sort_order", step.SortOrder,
		"split_ms", splitMs,
	)
	if err := s.repo.RecordSplit(ctx, runID, step.ID, splitMs); err != nil {
		return err
	}

	for _, splitID := range staleAreaSplitIDs(g.Steps, splits, step, ev.AreaName) {
		if err := s.repo.DeleteSplit(ctx, runID, splitID); err != nil {
			return err
		}
	}

	return nil
}

// HandleAreaGenerated stores the latest generated area context for the run.
func (s *Service) HandleAreaGenerated(ctx context.Context, runID int, areaCode string, areaLevel int, occurredAt time.Time) error {
	payload := map[string]string{
		"area_code":  areaCode,
		"area_level": fmt.Sprint(areaLevel),
	}
	if area, err := s.gameRepo.GetByCode(ctx, areaCode); err == nil && area != nil {
		payload["area"] = area.Name
	}
	if err := s.RecordLogEvent(ctx, runID, EventAreaGenerated, payload); err != nil {
		return err
	}
	s.mu.Lock()
	s.areaHints[runID] = areaHint{areaCode: areaCode, areaLevel: areaLevel, occurredAt: occurredAt}
	s.mu.Unlock()
	return nil
}

// RecordLogEvent stores a non-authoritative logtail event for later analysis.
// Unlike step confirmations, these events never mutate progress on their own.
func (s *Service) RecordLogEvent(ctx context.Context, runID int, eventType EventType, payload map[string]string) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if !run.IsActive {
		return nil
	}
	if payload == nil {
		payload = map[string]string{}
	}
	return s.repo.RecordEvent(ctx, runID, eventType, payload)
}

// FinishRun marks a run as completed and recomputes the local ranking.
func (s *Service) FinishRun(ctx context.Context, runID int) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	// Resume first if paused so the final elapsed time is accurate.
	if run.PausedAt != nil {
		_ = s.repo.ResumeRun(ctx, runID)
	}
	if err := s.repo.FinishRun(ctx, runID); err != nil {
		return err
	}
	// Compute ranking — best-effort, failure is non-fatal.
	if run.TimerStartedAt != nil {
		_ = s.repo.ComputeAndUpsertRanking(ctx, runID, run.GuideID)
	}
	return nil
}

// PauseRun pauses the run timer. Safe to call on already-paused runs (no-op).
func (s *Service) PauseRun(ctx context.Context, runID int) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if !run.IsActive {
		return fmt.Errorf("run: %d is not active", runID)
	}
	return s.repo.PauseRun(ctx, runID)
}

// ResumeRun resumes a paused run timer. Safe to call on running runs (no-op).
func (s *Service) ResumeRun(ctx context.Context, runID int) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if !run.IsActive {
		return fmt.Errorf("run: %d is not active", runID)
	}
	return s.repo.ResumeRun(ctx, runID)
}

// GetRunDeltas returns per-split timing deltas versus the PB run and the most
// recently finished previous run for the same guide.
func (s *Service) GetRunDeltas(ctx context.Context, runID int) (*RunDeltasResponse, error) {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}

	currentSplits, err := s.repo.ListSplits(ctx, runID)
	if err != nil {
		return nil, err
	}

	pbRunID, pbSplits, _ := s.repo.GetPBSplitsForGuide(ctx, run.GuideID)
	prevRunID, prevSplits, _ := s.repo.GetPrevRunSplits(ctx, run.GuideID, runID)

	pbMap := make(map[int]int64, len(pbSplits))
	for _, sp := range pbSplits {
		pbMap[sp.StepID] = sp.SplitMs
	}
	prevMap := make(map[int]int64, len(prevSplits))
	for _, sp := range prevSplits {
		prevMap[sp.StepID] = sp.SplitMs
	}

	deltas := make([]SplitDelta, 0, len(currentSplits))
	for _, sp := range currentSplits {
		d := SplitDelta{
			StepID:  sp.StepID,
			SplitMs: sp.SplitMs,
		}
		if pbMs, ok := pbMap[sp.StepID]; ok && pbRunID != runID {
			delta := sp.SplitMs - pbMs
			d.DeltaPBMs = &delta
		}
		if prevMs, ok := prevMap[sp.StepID]; ok {
			delta := sp.SplitMs - prevMs
			d.DeltaPrevMs = &delta
		}
		deltas = append(deltas, d)
	}

	resp := &RunDeltasResponse{
		RunID:  runID,
		Splits: deltas,
	}
	if pbRunID != 0 {
		resp.PBRunID = &pbRunID
	}
	if prevRunID != 0 {
		resp.PrevRunID = &prevRunID
	}
	return resp, nil
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

// CreateGGGSnapshot stores a snapshot sourced from the GGG API.
// Persists JSONB cache fields (equipped_items, skills, raw_response) in addition to scalars.
func (s *Service) CreateGGGSnapshot(ctx context.Context, snap *CharacterSnapshot) error {
	if _, err := s.repo.GetRun(ctx, snap.RunID); err != nil {
		return err
	}
	snap.Source = SnapshotGGG
	return s.repo.CreateGGGSnapshot(ctx, snap)
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

func (s *Service) buildStepTimings(ctx context.Context, run *RunSession, g *guide.Guide, confirmedSet map[int]bool, currentStepID int) ([]StepTiming, error) {
	splits, err := s.repo.ListSplits(ctx, run.ID)
	if err != nil {
		return nil, err
	}
	if len(splits) == 0 {
		return nil, nil
	}

	pbRunID, pbSplits, err := s.repo.GetPBSplitsForGuide(ctx, run.GuideID)
	if err != nil {
		return nil, err
	}

	splitMap := make(map[int]int64, len(splits))
	for _, split := range splits {
		splitMap[split.StepID] = split.SplitMs
	}
	pbMap := make(map[int]int64, len(pbSplits))
	for _, split := range pbSplits {
		pbMap[split.StepID] = split.SplitMs
	}

	stepTimings := make([]StepTiming, 0, len(splits))
	for _, step := range g.Steps {
		// Only show timings for confirmed steps or the current step;
		// auto-recorded area splits for future steps must not be surfaced.
		if !confirmedSet[step.ID] && step.ID != currentStepID {
			continue
		}
		splitMs, ok := splitMap[step.ID]
		if !ok {
			continue
		}
		timing := StepTiming{
			StepID:  step.ID,
			SplitMs: splitMs,
		}
		if pbMs, ok := pbMap[step.ID]; ok {
			delta := splitMs - pbMs
			if pbRunID != 0 {
				timing.DeltaPBMs = &delta
			}
		}
		stepTimings = append(stepTimings, timing)
	}

	return stepTimings, nil
}

func (s *Service) loadGuideForRun(ctx context.Context, run *RunSession) (*guide.Guide, error) {
	if run.GuideRevision > 0 {
		return s.guideRepo.GetByIDRevision(ctx, run.GuideID, run.GuideRevision)
	}
	return s.guideRepo.GetByID(ctx, run.GuideID)
}

func firstAreaStepForArea(steps []guide.Step, areaName string, act int, recorded map[int]bool, minSortOrder int) *guide.Step {
	entered := normalizeAreaName(areaName)
	for i := range steps {
		step := &steps[i]
		if recorded[step.ID] {
			continue
		}
		// Allow +/- 1 act difference for boundary zones (e.g. City of Sarn = A3 in game, A2 in guide)
		if act > 0 && step.Act > 0 && (step.Act < act-1 || step.Act > act+1) {
			continue
		}
		if step.SortOrder < minSortOrder {
			continue
		}
		if stepMatchesArea(step, entered) {
			return step
		}
	}
	return nil
}

func (s *Service) takeAreaHint(runID int, occurredAt time.Time) (areaHint, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hint, ok := s.areaHints[runID]
	if !ok {
		return areaHint{}, false
	}
	if !occurredAt.IsZero() {
		delta := occurredAt.Sub(hint.occurredAt)
		if delta < 0 || delta > 10*time.Second {
			return areaHint{}, false
		}
	}
	delete(s.areaHints, runID)
	return hint, true
}

func actFromAreaCode(areaCode string) int {
	parts := strings.Split(strings.TrimSpace(areaCode), "_")
	if len(parts) < 2 {
		return 0
	}
	act := 0
	_, _ = fmt.Sscanf(parts[1], "%d", &act)
	if act < 1 || act > 10 {
		return 0
	}
	return act
}

func staleAreaSplitIDs(steps []guide.Step, splits []Split, selected *guide.Step, areaName string) []int {
	if selected == nil {
		return nil
	}

	entered := normalizeAreaName(areaName)
	stepByID := make(map[int]guide.Step, len(steps))
	for _, step := range steps {
		stepByID[step.ID] = step
	}

	staleIDs := make([]int, 0)
	for _, split := range splits {
		step, ok := stepByID[split.StepID]
		if !ok {
			continue
		}
		if step.ID == selected.ID {
			continue
		}
		if step.Act != selected.Act {
			continue
		}
		if step.SortOrder >= selected.SortOrder {
			continue
		}
		if !stepMatchesArea(&step, entered) {
			continue
		}
		staleIDs = append(staleIDs, split.ID)
	}

	return staleIDs
}

func lastRecordedSortOrder(steps []guide.Step, recorded map[int]bool) int {
	lastSortOrder := 0
	for _, step := range steps {
		if !recorded[step.ID] {
			continue
		}
		if step.SortOrder > lastSortOrder {
			lastSortOrder = step.SortOrder
		}
	}
	return lastSortOrder
}

func stepMatchesArea(step *guide.Step, areaName string) bool {
	if step.CompletionMode != guide.CompletionLogtail && step.CompletionMode != guide.CompletionLogtailAsk {
		return false
	}
	hasAreaCondition := false
	for _, cond := range step.Conditions {
		if cond.ConditionType != guide.ConditionLogtailArea {
			continue
		}
		hasAreaCondition = true
		if normalizeAreaName(cond.Payload["area"]) == areaName {
			return true
		}
	}
	if hasAreaCondition {
		return false
	}
	return normalizeAreaName(step.Area) == areaName
}

func normalizeAreaName(areaName string) string {
	normalized := strings.ToLower(strings.TrimSpace(areaName))
	normalized = strings.TrimPrefix(normalized, "the ")
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

func elapsedAt(run *RunSession, at time.Time) int64 {
	if run.TimerStartedAt == nil {
		return 0
	}
	effectiveAt := at
	if run.PausedAt != nil && at.After(*run.PausedAt) {
		effectiveAt = *run.PausedAt
	}
	elapsedMs := effectiveAt.Sub(*run.TimerStartedAt).Milliseconds() - run.TotalPausedMs
	if elapsedMs < 0 {
		return 0
	}
	return elapsedMs
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
