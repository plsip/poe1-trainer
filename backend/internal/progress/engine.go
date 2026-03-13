package progress

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/poe1-trainer/internal/guide"
	runpkg "github.com/poe1-trainer/internal/run"
)

// Engine is the progress evaluation core.
// It is stateless — all run state is passed in and returned as proposed updates.
// Caller (run.Service) is responsible for persisting the updates.
type Engine struct {
	matchers []StepMatcher
}

// NewEngine creates an Engine loaded with the default set of matchers.
// Matcher precedence within a single step follows the order of this slice:
// ManualMatcher > AreaMatcher > LevelMatcher > QuestMatcher.
// The first matcher that returns ok=true wins for a given (event, step) pair.
func NewEngine() *Engine {
	return &Engine{
		matchers: []StepMatcher{
			ManualMatcher{},
			AreaMatcher{},
			LevelMatcher{},
			QuestMatcher{},
		},
	}
}

// ProcessEvent evaluates a domain event against all steps that are not yet
// completed or skipped, and returns proposed StepUpdates.
//
// Steps are evaluated in ascending SortOrder. A step may only receive one
// update per event (first matching rule wins).
//
// progress maps step_id -> current StepProgressStatus.
// Completed and skipped steps are never re-evaluated.
func (e *Engine) ProcessEvent(
	event DomainEvent,
	steps []guide.Step,
	progress map[int]runpkg.StepProgressStatus,
) []StepUpdate {
	sorted := stepsSorted(steps)
	var updates []StepUpdate
	for _, step := range sorted {
		status := progress[step.ID]
		if status == runpkg.StepCompleted || status == runpkg.StepSkipped {
			continue
		}
		for _, matcher := range e.matchers {
			if upd, ok := matcher.Match(event, step); ok {
				updates = append(updates, upd)
				break
			}
		}
	}
	return updates
}

// NextStep returns the first non-completed, non-skipped step in SortOrder.
// Returns nil when all steps are completed or skipped (run is finished).
func (e *Engine) NextStep(
	steps []guide.Step,
	progress map[int]runpkg.StepProgressStatus,
) *guide.Step {
	sorted := stepsSorted(steps)
	for i := range sorted {
		s := progress[sorted[i].ID]
		if s != runpkg.StepCompleted && s != runpkg.StepSkipped {
			return &sorted[i]
		}
	}
	return nil
}

// GenerateAlerts produces actionable hints for the active step.
// snapshot may be nil when no GGG character data is available.
func (e *Engine) GenerateAlerts(step *guide.Step, snapshot *SnapshotPayload) []Alert {
	if step == nil {
		return nil
	}
	var alerts []Alert
	alerts = append(alerts, e.gemAlerts(step)...)
	alerts = append(alerts, e.gearCheckAlerts(step)...)
	if snapshot != nil {
		alerts = append(alerts, e.resistAlerts(step, snapshot)...)
	}
	return alerts
}

// ─── private helpers ─────────────────────────────────────────────────────────

// gemAlerts emits buy_gem or receive_gem alerts for every gem requirement on the step.
func (e *Engine) gemAlerts(step *guide.Step) []Alert {
	if len(step.GemRequirements) == 0 {
		return nil
	}
	var alerts []Alert
	for _, gem := range step.GemRequirements {
		kind := AlertBuyGem
		verb := "Kup gem"
		if step.StepType == guide.StepTypeQuestReward {
			kind = AlertReceiveGem
			verb = "Odbierz gem z nagrody questa"
		}
		reason := gem.Note
		if reason == "" {
			reason = "Wymagany na tym etapie buildu."
		}
		alerts = append(alerts, Alert{
			Kind:        kind,
			StepID:      step.ID,
			GemName:     gem.GemName,
			Description: fmt.Sprintf("%s: %s (%s)", verb, gem.GemName, gemColorLabel(gem.Color)),
			Reason:      reason,
			Priority:    "high",
		})
	}
	return alerts
}

// gearCheckAlerts emits a check_gear alert for steps of type gear_check.
func (e *Engine) gearCheckAlerts(step *guide.Step) []Alert {
	if step.StepType != guide.StepTypeGearCheck {
		return nil
	}
	desc := step.Description
	if desc == "" {
		desc = "Sprawdź i ulepsz ekwipunek."
	}
	return []Alert{{
		Kind:        AlertCheckGear,
		StepID:      step.ID,
		Description: "Sprawdź ekwipunek: " + step.Title,
		Reason:      desc,
		Priority:    "medium",
	}}
}

// resistAlerts warns when resistances from the latest snapshot fall below 30%.
// Low resistances are dangerous in Acts 2+ and indicate the player should trade.
func (e *Engine) resistAlerts(step *guide.Step, snap *SnapshotPayload) []Alert {
	var alerts []Alert
	type res struct {
		name string
		val  int
	}
	checks := []res{
		{"ogień", snap.FireRes},
		{"zimno", snap.ColdRes},
		{"błyskawice", snap.LightRes},
	}
	for _, r := range checks {
		if r.val < 30 {
			alerts = append(alerts, Alert{
				Kind:        AlertCheckRes,
				StepID:      step.ID,
				Description: "Niska odporność — " + r.name + ": " + strconv.Itoa(r.val) + "%",
				Reason:      "Doświadczeni gracze utrzymują odporności powyżej 30% już w Akcie 2. Sprawdź przedmioty lub kup flakę odporności.",
				Priority:    "medium",
			})
		}
	}
	return alerts
}

func gemColorLabel(color string) string {
	switch color {
	case "red":
		return "czerwony (STR)"
	case "green":
		return "zielony (DEX)"
	case "blue":
		return "niebieski (INT)"
	default:
		return color
	}
}

// stepsSorted returns a copy of steps sorted ascending by SortOrder.
func stepsSorted(steps []guide.Step) []guide.Step {
	sorted := make([]guide.Step, len(steps))
	copy(sorted, steps)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SortOrder < sorted[j].SortOrder
	})
	return sorted
}
