package progress

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/poe1-trainer/internal/guide"
	runpkg "github.com/poe1-trainer/internal/run"
)

// StepMatcher evaluates a DomainEvent against a guide.Step.
// If the event satisfies a condition on the step, it returns a proposed
// StepUpdate and true. Otherwise it returns the zero value and false.
//
// Matchers are stateless and safe for concurrent use.
type StepMatcher interface {
	Match(event DomainEvent, step guide.Step) (StepUpdate, bool)
}

// StepUpdate is a proposed status change for a guide step produced by a matcher.
type StepUpdate struct {
	StepID      int
	NewStatus   runpkg.StepProgressStatus
	ConfirmedBy runpkg.ConfirmedBy
	Evidence    Evidence
}

// ─── AreaMatcher ─────────────────────────────────────────────────────────────

// AreaMatcher matches KindAreaEntered events against ConditionLogtailArea conditions.
// Comparison is case-insensitive and trims surrounding whitespace.
//
// When the step's CompletionMode is logtail_ask, the proposed status is
// needs_confirmation instead of completed — the player must still verify.
type AreaMatcher struct{}

func (m AreaMatcher) Match(event DomainEvent, step guide.Step) (StepUpdate, bool) {
	if event.Kind != KindAreaEntered || event.Area == nil {
		return StepUpdate{}, false
	}
	entered := strings.TrimSpace(event.Area.AreaName)
	for _, cond := range step.Conditions {
		if cond.ConditionType != guide.ConditionLogtailArea {
			continue
		}
		expected, ok := cond.Payload["area"]
		if !ok {
			continue
		}
		if !strings.EqualFold(entered, strings.TrimSpace(expected)) {
			continue
		}
		newStatus := runpkg.StepCompleted
		if step.CompletionMode == guide.CompletionLogtailAsk {
			newStatus = runpkg.StepNeedsConfirmation
		}
		return StepUpdate{
			StepID:      step.ID,
			NewStatus:   newStatus,
			ConfirmedBy: runpkg.ConfirmedByLogtail,
			Evidence: Evidence{
				EventKind:   KindAreaEntered,
				Confidence:  ConfidenceMedium,
				Description: fmt.Sprintf("Gracz wszedł do strefy: %q.", entered),
				OccurredAt:  event.OccurredAt,
			},
		}, true
	}
	return StepUpdate{}, false
}

// ─── LevelMatcher ────────────────────────────────────────────────────────────

// LevelMatcher matches KindLevelUp events against ConditionGGGLevel conditions.
// A condition is satisfied when the event's level is >= the min_level in the payload.
type LevelMatcher struct{}

func (m LevelMatcher) Match(event DomainEvent, step guide.Step) (StepUpdate, bool) {
	if event.Kind != KindLevelUp || event.Level == nil {
		return StepUpdate{}, false
	}
	for _, cond := range step.Conditions {
		if cond.ConditionType != guide.ConditionGGGLevel {
			continue
		}
		minStr, ok := cond.Payload["min_level"]
		if !ok {
			continue
		}
		minLevel, err := strconv.Atoi(minStr)
		if err != nil || event.Level.Level < minLevel {
			continue
		}
		newStatus := runpkg.StepCompleted
		if step.CompletionMode == guide.CompletionGGGAPIAsk {
			newStatus = runpkg.StepNeedsConfirmation
		}
		return StepUpdate{
			StepID:      step.ID,
			NewStatus:   newStatus,
			ConfirmedBy: runpkg.ConfirmedByGGG,
			Evidence: Evidence{
				EventKind:   KindLevelUp,
				Confidence:  ConfidenceHigh,
				Description: fmt.Sprintf("Postać osiągnęła poziom %d (wymagane min. %d).", event.Level.Level, minLevel),
				OccurredAt:  event.OccurredAt,
			},
		}, true
	}
	return StepUpdate{}, false
}

// ─── QuestMatcher ────────────────────────────────────────────────────────────

// QuestMatcher matches KindQuestCompleted events against ConditionGGGQuest conditions.
// If the payload specifies a "state", the match is case-insensitive. If "state"
// is absent it matches any non-empty state.
type QuestMatcher struct{}

func (m QuestMatcher) Match(event DomainEvent, step guide.Step) (StepUpdate, bool) {
	if event.Kind != KindQuestCompleted || event.Quest == nil {
		return StepUpdate{}, false
	}
	for _, cond := range step.Conditions {
		if cond.ConditionType != guide.ConditionGGGQuest {
			continue
		}
		wantID, ok := cond.Payload["quest_id"]
		if !ok || wantID != event.Quest.QuestID {
			continue
		}
		wantState := cond.Payload["state"]
		if wantState != "" && !strings.EqualFold(wantState, event.Quest.State) {
			continue
		}
		newStatus := runpkg.StepCompleted
		if step.CompletionMode == guide.CompletionGGGAPIAsk {
			newStatus = runpkg.StepNeedsConfirmation
		}
		return StepUpdate{
			StepID:      step.ID,
			NewStatus:   newStatus,
			ConfirmedBy: runpkg.ConfirmedByGGG,
			Evidence: Evidence{
				EventKind:   KindQuestCompleted,
				Confidence:  ConfidenceHigh,
				Description: fmt.Sprintf("Quest %q ukończony ze statusem %q.", event.Quest.QuestID, event.Quest.State),
				OccurredAt:  event.OccurredAt,
			},
		}, true
	}
	return StepUpdate{}, false
}

// ─── ManualMatcher ───────────────────────────────────────────────────────────

// ManualMatcher matches KindManualConfirm events against the target step.
// The event's Manual.StepID must equal the step's ID — this prevents one
// confirm event from accidentally completing an unrelated step.
type ManualMatcher struct{}

func (m ManualMatcher) Match(event DomainEvent, step guide.Step) (StepUpdate, bool) {
	if event.Kind != KindManualConfirm || event.Manual == nil {
		return StepUpdate{}, false
	}
	if event.Manual.StepID != step.ID {
		return StepUpdate{}, false
	}
	return StepUpdate{
		StepID:      step.ID,
		NewStatus:   runpkg.StepCompleted,
		ConfirmedBy: runpkg.ConfirmedByManual,
		Evidence: Evidence{
			EventKind:   KindManualConfirm,
			Confidence:  ConfidenceHigh,
			Description: "Gracz ręcznie potwierdził ukończenie kroku.",
			OccurredAt:  event.OccurredAt,
		},
	}, true
}
