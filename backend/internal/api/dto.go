package api

// ─── Integration status ──────────────────────────────────────────────────────

// IntegrationStatusResponse is the body for GET /integration/status.
type IntegrationStatusResponse struct {
	// LogWatcher reflects the current logtail.Status, or "disabled" when LOG_PATH is not set.
	LogWatcher string `json:"log_watcher"`
}

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
	// AutoStart: gdy true, timer nie startuje w momencie stworzenia runu,
	// lecz czeka na pierwszy zdarzenie area_entered z logtaila.
	AutoStart bool `json:"auto_start"`
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

// ─── Log replay ─────────────────────────────────────────────────────────────────

// ReplayLogRequest is the optional body for POST /runs/{id}/replay-log.
type ReplayLogRequest struct {
	// LogPath overrides the server-configured log path for this request only.
	// Leave empty to use the server default (LOG_PATH env / auto-detected path).
	LogPath string `json:"log_path,omitempty"`
	// LogTZ overrides the timezone used to parse log timestamps, e.g. "Europe/Warsaw".
	// Leave empty to use the server-configured LOG_TZ or time.Local.
	LogTZ string `json:"log_tz,omitempty"`
}

// ReplayLogResponse is the body returned by POST /runs/{id}/replay-log.
type ReplayLogResponse struct {
	LinesRead        int   `json:"lines_read"`
	EventsDispatched int   `json:"events_dispatched"`
	ParseErrors      int   `json:"parse_errors"`
	DurationMs       int64 `json:"duration_ms"`
}

// ─── Alerts ───────────────────────────────────────────────────────────────────

// Alert aggregates a single gem or gear suggestion for the current step or campaign phase.
type Alert struct {
	Kind        string `json:"kind"`              // "gem" | "gear"
	Priority    string `json:"priority"`          // "high" | "medium" | "low"
	Slot        string `json:"slot,omitempty"`    // gear slot (gear alerts only)
	Description string `json:"description"`
	StepID      int    `json:"step_id,omitempty"`
	Notes       string `json:"notes,omitempty"`
	// Extended fields (Prompt 07): fine-grained classification and explanation.
	GemName    string `json:"gem_name,omitempty"`    // gem name (gem alerts only)
	ActionType string `json:"action_type,omitempty"` // fine-grained sub-type (vendor | weapon_swap | full_switch | …)
	Reason     string `json:"reason,omitempty"`      // explanation of why this alert matters now
	Source     string `json:"source,omitempty"`      // "step" = step-specific | "rule" = campaign-phase rule
}

// AlertsResponse is the response body for GET /runs/{id}/alerts.
type AlertsResponse struct {
	StepID int     `json:"step_id"`
	Alerts []Alert `json:"alerts"`
}

// ─── GGG Integration ──────────────────────────────────────────────────────────

// GGGStatusResponse is the response body for GET /ggg/status.
type GGGStatusResponse struct {
	// Configured is true when GGG_CLIENT_ID and GGG_CLIENT_SECRET are set.
	Configured bool `json:"configured"`
	// Available is true when there is a valid OAuth token stored on disk.
	Available bool `json:"available"`
	// Username is the GGG account name from the stored token (empty when unavailable).
	Username string `json:"username,omitempty"`
}

// GGGSyncSnapshotRequest is the body for POST /runs/{id}/snapshots/ggg.
type GGGSyncSnapshotRequest struct {
	CharacterName string `json:"character_name"`
	// Realm: "pc" (default), "xbox", "sony".
	Realm string `json:"realm,omitempty"`
}

