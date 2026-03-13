package progress

import "time"

// ─── Domain Events ───────────────────────────────────────────────────────────

// DomainEventKind identifies the source and nature of a domain event.
type DomainEventKind string

const (
	// KindAreaEntered jest emitowany przez logtail gdy gracz wchodzi do nowej strefy.
	KindAreaEntered DomainEventKind = "area_entered"
	// KindLevelUp jest emitowany gdy postać awansuje na wyższy poziom.
	KindLevelUp DomainEventKind = "level_up"
	// KindQuestCompleted jest emitowany gdy quest zmienia status na Finished.
	KindQuestCompleted DomainEventKind = "quest_completed"
	// KindCharSnapshot jest emitowany po pobraniu migawki postaci z GGG API.
	KindCharSnapshot DomainEventKind = "char_snapshot"
	// KindManualConfirm jest emitowany gdy gracz ręcznie potwierdza krok w UI.
	KindManualConfirm DomainEventKind = "manual_confirm"
)

// DomainEvent is a normalized event produced from a raw input source.
// Exactly one payload field is non-nil per event instance.
type DomainEvent struct {
	Kind       DomainEventKind
	OccurredAt time.Time
	RunID      int

	// Payloads — dokładnie jedno non-nil pole na zdarzenie.
	Area     *AreaPayload
	Level    *LevelPayload
	Quest    *QuestPayload
	Snapshot *SnapshotPayload
	Manual   *ManualPayload
}

// AreaPayload carries data for KindAreaEntered events.
type AreaPayload struct {
	AreaName string
}

// LevelPayload carries data for KindLevelUp events.
type LevelPayload struct {
	Level int
}

// QuestPayload carries data for KindQuestCompleted events.
type QuestPayload struct {
	QuestID string
	State   string // "Finished" | "Failed" | etc.
}

// SnapshotPayload carries character state from KindCharSnapshot events (GGG API).
// Źródło: GGG character API. Używany do generowania alertów, nie do matchowania kroków.
type SnapshotPayload struct {
	Level             int
	LifeMax           int
	FireRes           int
	ColdRes           int
	LightRes          int
	ChaosRes          int
	CompletedQuestIDs map[string]string // quest_id -> state
}

// ManualPayload carries data for KindManualConfirm events.
type ManualPayload struct {
	StepID int
}

// ─── Constructors ────────────────────────────────────────────────────────────

// NewAreaEnteredEvent tworzy zdarzenie wejścia do strefy z logtail.
func NewAreaEnteredEvent(runID int, areaName string, at time.Time) DomainEvent {
	return DomainEvent{
		Kind:       KindAreaEntered,
		RunID:      runID,
		OccurredAt: at,
		Area:       &AreaPayload{AreaName: areaName},
	}
}

// NewLevelUpEvent tworzy zdarzenie awansu postaci.
func NewLevelUpEvent(runID, level int, at time.Time) DomainEvent {
	return DomainEvent{
		Kind:       KindLevelUp,
		RunID:      runID,
		OccurredAt: at,
		Level:      &LevelPayload{Level: level},
	}
}

// NewQuestCompletedEvent tworzy zdarzenie ukończenia questa.
func NewQuestCompletedEvent(runID int, questID, state string, at time.Time) DomainEvent {
	return DomainEvent{
		Kind:       KindQuestCompleted,
		RunID:      runID,
		OccurredAt: at,
		Quest:      &QuestPayload{QuestID: questID, State: state},
	}
}

// NewCharSnapshotEvent tworzy zdarzenie migawki postaci z GGG API.
func NewCharSnapshotEvent(runID int, snap SnapshotPayload, at time.Time) DomainEvent {
	return DomainEvent{
		Kind:       KindCharSnapshot,
		RunID:      runID,
		OccurredAt: at,
		Snapshot:   &snap,
	}
}

// NewManualConfirmEvent tworzy zdarzenie ręcznego potwierdzenia kroku przez gracza.
func NewManualConfirmEvent(runID, stepID int, at time.Time) DomainEvent {
	return DomainEvent{
		Kind:       KindManualConfirm,
		RunID:      runID,
		OccurredAt: at,
		Manual:     &ManualPayload{StepID: stepID},
	}
}

// ─── Evidence model ──────────────────────────────────────────────────────────

// Confidence is a normalized score [0..1] expressing certainty of step completion.
type Confidence float64

const (
	// ConfidenceHigh — pewność wynikająca z jawnego potwierdzenia lub jednoznacznego sygnału API.
	ConfidenceHigh Confidence = 1.0
	// ConfidenceMedium — pewność wynikająca z inferencji logtail (wejście do strefy).
	ConfidenceMedium Confidence = 0.7
	// ConfidenceLow — słaby sygnał; wymaga dodatkowej weryfikacji.
	ConfidenceLow Confidence = 0.4
)

// Evidence records the audit trail for why a step's status was changed.
// Persisted as JSONB in run_step_progress.evidence.
type Evidence struct {
	EventKind   DomainEventKind `json:"event_kind"`
	Confidence  Confidence      `json:"confidence"`
	Description string          `json:"description"`
	OccurredAt  time.Time       `json:"occurred_at"`
}

// ─── Alert model ─────────────────────────────────────────────────────────────

// AlertKind identifies the type of actionable alert shown to the player.
type AlertKind string

const (
	// AlertBuyGem — gracz powinien kupić gem u handlarza.
	AlertBuyGem AlertKind = "buy_gem"
	// AlertReceiveGem — gracz powinien odebrać gem jako nagrodę questa.
	AlertReceiveGem AlertKind = "receive_gem"
	// AlertLevelGem — gracz powinien już mieć gem na określonym poziomie.
	AlertLevelGem AlertKind = "level_gem"
	// AlertCheckGear — gracz powinien sprawdzić i zaktualizować ekwipunek.
	AlertCheckGear AlertKind = "check_gear"
	// AlertCheckFlask — gracz powinien sprawdzić flakony.
	AlertCheckFlask AlertKind = "check_flask"
	// AlertCheckRes — gracz powinien sprawdzić odporności.
	AlertCheckRes AlertKind = "check_resist"
)

// Alert is an actionable hint displayed to the player alongside the current step.
type Alert struct {
	Kind        AlertKind `json:"kind"`
	StepID      int       `json:"step_id,omitempty"`
	GemName     string    `json:"gem_name,omitempty"`
	Description string    `json:"description"`
	Reason      string    `json:"reason"`
	Priority    string    `json:"priority"` // high | medium | low
}
