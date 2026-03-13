package run

import "time"

// Status of a single run.
type Status string

const (
	StatusActive    Status = "active"
	StatusFinished  Status = "finished"
	StatusAbandoned Status = "abandoned"
)

// RunSession represents one character's run through the campaign.
type RunSession struct {
	ID             int        `json:"id"`
	GuideID        int        `json:"guide_id"`
	CharacterName  string     `json:"character_name"`
	League         string     `json:"league"`
	Status         Status     `json:"status"`
	Notes          string     `json:"notes,omitempty"`
	StartedAt      time.Time  `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	IsActive       bool       `json:"is_active"`
	// Timer fields (migration 005)
	TimerStartedAt *time.Time `json:"timer_started_at,omitempty"`
	PausedAt       *time.Time `json:"paused_at,omitempty"`
	TotalPausedMs  int64      `json:"total_paused_ms"`
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
	Run              RunSession `json:"run"`
	CurrentStepID    int        `json:"current_step_id"`
	ConfirmedStepIDs []int      `json:"confirmed_step_ids"`
	ElapsedMs        int64      `json:"elapsed_ms"`
}

// ─── run_characters ──────────────────────────────────────────────────────────

// Character holds per-character details for a run.
// level_current is a cached value updated from the latest CharacterSnapshot.
type Character struct {
	ID             int        `json:"id"`
	RunID          int        `json:"run_id"`
	CharacterName  string     `json:"character_name"`
	CharacterClass string     `json:"character_class"`
	League         string     `json:"league"`
	LevelAtStart   int        `json:"level_at_start"`
	LevelCurrent   int        `json:"level_current"` // cached — derived from latest snapshot
	LastSeenAt     *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ─── run_step_progress ───────────────────────────────────────────────────────

// StepProgressStatus is the lifecycle state of a single step within a run.
type StepProgressStatus string

const (
	StepPending             StepProgressStatus = "pending"
	StepInProgress          StepProgressStatus = "in_progress"
	StepNeedsConfirmation   StepProgressStatus = "needs_confirmation"
	StepCompleted           StepProgressStatus = "completed"
	StepSkipped             StepProgressStatus = "skipped"
)

// StepProgress is the authoritative record of a step's state in a run.
// status = "completed" can be re-derived from completed_at IS NOT NULL.
// Source of truth: completed_at + confirmed_by + confirmed_at.
type StepProgress struct {
	ID          int                `json:"id"`
	RunID       int                `json:"run_id"`
	StepID      int                `json:"step_id"`
	Status      StepProgressStatus `json:"status"`
	CompletedAt *time.Time         `json:"completed_at,omitempty"`
	ConfirmedBy ConfirmedBy        `json:"confirmed_by"`
	ConfirmedAt *time.Time         `json:"confirmed_at,omitempty"`
	Notes       string             `json:"notes,omitempty"`
}

// ─── character_snapshots ─────────────────────────────────────────────────────

// SnapshotSource describes the origin of a CharacterSnapshot.
type SnapshotSource string

const (
	SnapshotManual SnapshotSource = "manual"
	SnapshotGGG    SnapshotSource = "ggg"
)

// CharacterSnapshot captures the state of a character at a point in time.
//
// Scalar fields (Level, LifeMax, Res*) are the source of truth used by alert rules.
// EquippedItems, Skills, and RawResponse are cached from the GGG API (derived).
// Snapshots are immutable — a new state always creates a new row.
type CharacterSnapshot struct {
	ID            int64          `json:"id"`
	RunID         int            `json:"run_id"`
	CapturedAt    time.Time      `json:"captured_at"`
	Source        SnapshotSource `json:"source"`
	Level         int            `json:"level"`
	LifeMax       *int           `json:"life_max,omitempty"`
	ManaMax       *int           `json:"mana_max,omitempty"`
	ResFire       *int           `json:"res_fire,omitempty"`
	ResCold       *int           `json:"res_cold,omitempty"`
	ResLightning  *int           `json:"res_lightning,omitempty"`
	ResChaos      *int           `json:"res_chaos,omitempty"`
	EquippedItems map[string]any `json:"equipped_items"` // cached from GGG API
	Skills        map[string]any `json:"skills"`         // cached socketed gems
	RawResponse   map[string]any `json:"raw_response"`   // full GGG API response cache
}

// ─── manual_checks ───────────────────────────────────────────────────────────

// CheckType categorises a manual confirmation question.
type CheckType string

const (
	CheckGear     CheckType = "gear"
	CheckGem      CheckType = "gem"
	CheckLevel    CheckType = "level"
	CheckResist   CheckType = "resist"
	CheckFlask    CheckType = "flask"
	CheckQuest    CheckType = "quest"
	CheckFreeForm CheckType = "free_form"
)

// ManualCheck is a question posed to the player that requires an explicit answer.
// Generated by the recommendation engine or the guide importer.
// StepID == nil means the check is not step-specific.
type ManualCheck struct {
	ID            int        `json:"id"`
	RunID         int        `json:"run_id"`
	StepID        *int       `json:"step_id,omitempty"`
	CheckType     CheckType  `json:"check_type"`
	Prompt        string     `json:"prompt"`
	IsConfirmed   bool       `json:"is_confirmed"`
	ResponseValue string     `json:"response_value,omitempty"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ─── run_splits ──────────────────────────────────────────────────────────────

// Split records the elapsed time (in ms from run start) when a step was completed.
type Split struct {
	ID         int       `json:"id"`
	RunID      int       `json:"run_id"`
	StepID     int       `json:"step_id"`
	SplitMs    int64     `json:"split_ms"`
	RecordedAt time.Time `json:"recorded_at"`
}

// ─── local_rankings ──────────────────────────────────────────────────────────

// LocalRanking is a pre-computed summary of a finished run's performance.
//
// All fields are derived:
//   - TotalMs   from runs.finished_at − runs.timer_started_at
//   - ActSplits from run_splits aggregated per act
//   - Rank      from ordering all finished runs of the same guide by TotalMs
//
// Recomputed after every run status change to "finished".
type LocalRanking struct {
	ID         int            `json:"id"`
	GuideID    int            `json:"guide_id"`
	RunID      int            `json:"run_id"`
	TotalMs    int64          `json:"total_ms"`
	ActSplits  map[string]int `json:"act_splits"` // {"1": 120000, "2": 345000, …}
	Rank       int            `json:"rank"`
	ComputedAt time.Time      `json:"computed_at"`
}

// DetailedRankingEntry is a full ranking row returned by GET /guides/{slug}/ranking.
// Combines local_rankings with run metadata.
type DetailedRankingEntry struct {
	Rank          int            `json:"rank"`
	RunID         int            `json:"run_id"`
	CharacterName string         `json:"character_name"`
	StartedAt     time.Time      `json:"started_at"`
	TotalMs       int64          `json:"total_ms"`
	ActSplits     map[string]int `json:"act_splits"` // {"1": 120000, …}
	IsPB          bool           `json:"is_pb"`
}

// RankingStats aggregates key timing statistics for a guide's runs.
type RankingStats struct {
	Count     int    `json:"count"`
	PBMs      *int64 `json:"pb_ms,omitempty"`
	MedianMs  *int64 `json:"median_ms,omitempty"`
	LastRunMs *int64 `json:"last_run_ms,omitempty"`
	LastRunID *int   `json:"last_run_id,omitempty"`
}

// SplitDelta describes a split in the current run relative to the PB run and
// the most recent previous run.
type SplitDelta struct {
	StepID      int    `json:"step_id"`
	SplitMs     int64  `json:"split_ms"`
	DeltaPBMs   *int64 `json:"delta_pb_ms,omitempty"`   // positive = slower than PB
	DeltaPrevMs *int64 `json:"delta_prev_ms,omitempty"` // positive = slower than prev run
}

// RunDeltasResponse is the body of GET /runs/{id}/split-deltas.
type RunDeltasResponse struct {
	RunID     int          `json:"run_id"`
	PBRunID   *int         `json:"pb_run_id,omitempty"`
	PrevRunID *int         `json:"prev_run_id,omitempty"`
	Splits    []SplitDelta `json:"splits"`
}
