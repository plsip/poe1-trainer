package recommendation

import "github.com/poe1-trainer/internal/guide"

// Priority of a recommendation.
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// Recommendation is a suggestion shown to the player with an explanation.
type Recommendation struct {
	ID       string   `json:"id"`
	Text     string   `json:"text"`
	Reason   string   `json:"reason"`
	Priority Priority `json:"priority"`
	// StepID is the guide step this recommendation refers to (0 = general).
	StepID int `json:"step_id,omitempty"`
}

// Input collects everything the engine needs to produce recommendations.
type Input struct {
	CurrentStep     *guide.Step
	NextStep        *guide.Step
	ConfirmedStepIDs map[int]bool
	ActualAct       int
	ElapsedMs       int64
}
