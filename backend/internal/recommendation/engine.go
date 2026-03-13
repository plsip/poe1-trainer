package recommendation

import (
	"fmt"
	"strings"

	"github.com/poe1-trainer/internal/guide"
	runpkg "github.com/poe1-trainer/internal/run"
)

// Engine produces recommendations from a run's current state and its guide.
// It contains only deterministic rules — no LLM or external calls.
type Engine struct{}

// NewEngine creates a new Engine.
func NewEngine() *Engine { return &Engine{} }

// Produce returns recommendations for the current run state.
func (e *Engine) Produce(g *guide.Guide, state *runpkg.CurrentState) []Recommendation {
	if g == nil || state == nil {
		return nil
	}

	confirmedSet := make(map[int]bool, len(state.ConfirmedStepIDs))
	for _, id := range state.ConfirmedStepIDs {
		confirmedSet[id] = true
	}

	// Find current and next steps.
	var currentStep, nextStep *guide.Step
	for i := range g.Steps {
		step := &g.Steps[i]
		if step.ID == state.CurrentStepID {
			currentStep = step
			if i+1 < len(g.Steps) {
				nextStep = &g.Steps[i+1]
			}
			break
		}
	}

	input := Input{
		CurrentStep:      currentStep,
		NextStep:         nextStep,
		ConfirmedStepIDs: confirmedSet,
		ElapsedMs:        state.ElapsedMs,
	}

	var recs []Recommendation
	recs = append(recs, e.currentStepRec(input)...)
	recs = append(recs, e.gemRecs(input)...)
	recs = append(recs, e.checkpointRec(input)...)
	recs = append(recs, e.paceRec(input)...)
	return recs
}

// currentStepRec produces the primary "what to do now" recommendation.
func (e *Engine) currentStepRec(in Input) []Recommendation {
	if in.CurrentStep == nil {
		return []Recommendation{{
			ID:       "all_done",
			Text:     "Wszystkie kroki zostały potwierdzone.",
			Reason:   "Guide ukończony.",
			Priority: PriorityLow,
		}}
	}
	return []Recommendation{{
		ID:       fmt.Sprintf("step_%d", in.CurrentStep.ID),
		Text:     in.CurrentStep.Title,
		Reason:   fmt.Sprintf("Akt %d — krok %d.", in.CurrentStep.Act, in.CurrentStep.StepNumber),
		Priority: PriorityHigh,
		StepID:   in.CurrentStep.ID,
	}}
}

// gemRecs suggests gems that should be obtained at the current step.
func (e *Engine) gemRecs(in Input) []Recommendation {
	if in.CurrentStep == nil || len(in.CurrentStep.GemRequirements) == 0 {
		return nil
	}
	var recs []Recommendation
	for _, gem := range in.CurrentStep.GemRequirements {
		colorLabel := gemColorLabel(gem.Color)
		recs = append(recs, Recommendation{
			ID:       fmt.Sprintf("gem_%d_%s", in.CurrentStep.ID, sanitize(gem.GemName)),
			Text:     fmt.Sprintf("Zdobądź gem: %s (%s)", gem.GemName, colorLabel),
			Reason:   fmt.Sprintf("Wymagany na tym etapie buildu. %s", gem.Note),
			Priority: PriorityHigh,
			StepID:   in.CurrentStep.ID,
		})
	}
	return recs
}

// checkpointRec adds a high-priority reminder for manual confirmation steps.
func (e *Engine) checkpointRec(in Input) []Recommendation {
	if in.CurrentStep == nil || !in.CurrentStep.IsCheckpoint {
		return nil
	}
	return []Recommendation{{
		ID:       fmt.Sprintf("confirm_%d", in.CurrentStep.ID),
		Text:     "Potwierdź manualnie ukończenie tego kamienia milowego.",
		Reason:   "Ten krok nie może być wykryty automatycznie — wymaga twojego potwierdzenia.",
		Priority: PriorityHigh,
		StepID:   in.CurrentStep.ID,
	}}
}

// paceRec provides a general pace hint based on elapsed time.
func (e *Engine) paceRec(in Input) []Recommendation {
	if in.ElapsedMs == 0 || in.CurrentStep == nil {
		return nil
	}
	// Rough heuristic: if it's been more than 30 minutes and we're still in Act 1.
	if in.CurrentStep.Act == 1 && in.ElapsedMs > 30*60*1000 {
		return []Recommendation{{
			ID:       "pace_act1",
			Text:     "Jesteś już ponad 30 minut w Akcie 1. Rozważ szybsze tempo.",
			Reason:   "Doświadczeni gracze kończą Akt 1 w ~15-20 minut. Unikaj farmienia bez potrzeby.",
			Priority: PriorityLow,
		}}
	}
	return nil
}

func gemColorLabel(color string) string {
	switch strings.ToLower(color) {
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

func sanitize(s string) string {
	return strings.NewReplacer(" ", "_", "'", "", "/", "_").Replace(strings.ToLower(s))
}
