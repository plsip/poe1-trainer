package api

// ─── Builds ──────────────────────────────────────────────────────────────────

// CreateBuildRequest is the body for POST /builds.
type CreateBuildRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Description string `json:"description"`
}

// CreateVersionRequest is the body for POST /builds/{id}/versions.
type CreateVersionRequest struct {
	Version    string  `json:"version"`
	PatchTag   string  `json:"patch_tag"`
	Notes      string  `json:"notes"`
	IsCurrent  bool    `json:"is_current"`
	ReleasedAt *string `json:"released_at,omitempty"` // ISO 8601 date
}

// ─── Guide import ─────────────────────────────────────────────────────────────

// ImportGuideRequest is the body for POST /guides/import.
type ImportGuideRequest struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	BuildName string `json:"build_name"`
	Version   string `json:"version"`
	Content   string `json:"content"` // raw Markdown text of the guide
}

// ─── Runs ─────────────────────────────────────────────────────────────────────

// CreateRunRequest is the body for POST /runs.
type CreateRunRequest struct {
	GuideID       int    `json:"guide_id"`
	CharacterName string `json:"character_name"`
	League        string `json:"league"`
}

// ─── Characters & snapshots ───────────────────────────────────────────────────

// UpsertCharacterRequest is the body for PUT /runs/{id}/character.
type UpsertCharacterRequest struct {
	CharacterName  string `json:"character_name"`
	CharacterClass string `json:"character_class"`
	League         string `json:"league"`
	LevelAtStart   int    `json:"level_at_start"`
}

// CreateSnapshotRequest is the body for POST /runs/{id}/snapshots.
// All stat fields are optional; omit what is unknown at the time of the snapshot.
type CreateSnapshotRequest struct {
	Level        int  `json:"level"`
	LifeMax      *int `json:"life_max,omitempty"`
	ManaMax      *int `json:"mana_max,omitempty"`
	ResFire      *int `json:"res_fire,omitempty"`
	ResCold      *int `json:"res_cold,omitempty"`
	ResLightning *int `json:"res_lightning,omitempty"`
	ResChaos     *int `json:"res_chaos,omitempty"`
}

// ─── Events ───────────────────────────────────────────────────────────────────

// RecordEventRequest is the body for POST /runs/{id}/events.
type RecordEventRequest struct {
	EventType string            `json:"event_type"`
	Payload   map[string]string `json:"payload"`
}

// ─── Splits ───────────────────────────────────────────────────────────────────

// RecordSplitRequest is the body for POST /runs/{id}/steps/{step_id}/split.
type RecordSplitRequest struct {
	SplitMs int64 `json:"split_ms"`
}

// ─── Manual checks ────────────────────────────────────────────────────────────

// AnswerCheckRequest is the body for POST /runs/{id}/checks/{check_id}/answer.
type AnswerCheckRequest struct {
	ResponseValue string `json:"response_value"`
}

// ─── Alerts ───────────────────────────────────────────────────────────────────

// Alert aggregates a single gem or gear suggestion for the current step.
type Alert struct {
	Kind        string `json:"kind"`              // "gem" | "gear"
	Priority    string `json:"priority"`          // "high" | "medium" | "low"
	Slot        string `json:"slot,omitempty"`    // gear slot (gear alerts only)
	Description string `json:"description"`
	StepID      int    `json:"step_id,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// AlertsResponse is the response body for GET /runs/{id}/alerts.
type AlertsResponse struct {
	StepID int     `json:"step_id"`
	Alerts []Alert `json:"alerts"`
}
