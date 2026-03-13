package run

import "time"

// Status of a single run.
type Status string

const (
	StatusActive   Status = "active"
	StatusFinished Status = "finished"
)

// RunSession represents one character's run through the campaign.
type RunSession struct {
	ID            int        `json:"id"`
	GuideID       int        `json:"guide_id"`
	CharacterName string     `json:"character_name"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	IsActive      bool       `json:"is_active"`
}

// ConfirmedBy describes how a checkpoint was confirmed.
type ConfirmedBy string

const (
	ConfirmedByManual  ConfirmedBy = "manual"
	ConfirmedByLogtail ConfirmedBy = "logtail"
	ConfirmedByGGG     ConfirmedBy = "ggg"
)

// Checkpoint records that a guide step was completed in a particular run.
type Checkpoint struct {
	ID          int         `json:"id"`
	RunID       int         `json:"run_id"`
	StepID      int         `json:"step_id"`
	ConfirmedAt time.Time   `json:"confirmed_at"`
	ConfirmedBy ConfirmedBy `json:"confirmed_by"`
}

// EventType identifies the kind of RunEvent.
type EventType string

const (
	EventAreaEntered    EventType = "area_entered"
	EventStepConfirmed  EventType = "step_confirmed"
	EventExternalHint   EventType = "hint"
)

// Event is an immutable record of something that happened during a run.
type Event struct {
	ID         int64             `json:"id"`
	RunID      int               `json:"run_id"`
	EventType  EventType         `json:"event_type"`
	Payload    map[string]string `json:"payload"`
	OccurredAt time.Time         `json:"occurred_at"`
}

// AreaEvent is emitted by the logtail watcher when the player enters an area.
type AreaEvent struct {
	AreaName string
}

// CurrentState is the aggregated state of an active run returned by the API.
type CurrentState struct {
	Run             RunSession   `json:"run"`
	CurrentStepID   int          `json:"current_step_id"`
	ConfirmedStepIDs []int       `json:"confirmed_step_ids"`
	ElapsedMs       int64        `json:"elapsed_ms"`
}
