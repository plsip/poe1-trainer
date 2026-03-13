package guide

import "time"

// Guide represents a parsed campaign guide for a specific build.
type Guide struct {
	ID             int       `json:"id"`
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	BuildName      string    `json:"build_name"`
	Version        string    `json:"version"`
	BuildVersionID *int      `json:"build_version_id,omitempty"`
	Steps          []Step    `json:"steps"`
	CreatedAt      time.Time `json:"created_at"`
}

// CompletionMode controls how a step's completion is detected.
type CompletionMode string

const (
	CompletionManual      CompletionMode = "manual"       // player clicks Confirm
	CompletionLogtail     CompletionMode = "logtail"      // inferred from Client.txt, no confirmation
	CompletionLogtailAsk  CompletionMode = "logtail_ask"  // inferred from Client.txt + requires confirmation
	CompletionGGGAPI      CompletionMode = "ggg_api"      // inferred from GGG API, no confirmation
	CompletionGGGAPIAsk   CompletionMode = "ggg_api_ask"  // inferred from GGG API + requires confirmation
)

// StepType classifies the primary action of a guide step.
// Used by the recommendation engine to fire the right rule and by the
// parser heuristics to assign completion mode and inference conditions.
type StepType string

const (
	StepTypeGeneral      StepType = "general"       // default fallback
	StepTypeBossKill     StepType = "boss_kill"      // killing a named boss
	StepTypeLabyrinth    StepType = "labyrinth"      // Labyrinth Trial or completion
	StepTypeQuestReward  StepType = "quest_reward"   // collecting a quest reward
	StepTypeGemAcquire   StepType = "gem_acquire"    // buying or receiving a gem
	StepTypeNavigation   StepType = "navigation"     // traveling to an area
	StepTypeVendorRecipe StepType = "vendor_recipe"  // crafting at vendor
	StepTypeCraft        StepType = "craft"          // crafting bench operation
	StepTypePortal       StepType = "portal"         // using portal or logout shortcut
	StepTypeGearCheck    StepType = "gear_check"     // evaluating or upgrading gear/flasks
)

// Step is a single actionable item in the guide.
type Step struct {
	ID              int              `json:"id"`
	GuideID         int              `json:"guide_id"`
	StepNumber      int              `json:"step_number"`
	Act             int              `json:"act"`
	Section         string           `json:"section,omitempty"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Area            string           `json:"area"`
	QuestName       string           `json:"quest_name,omitempty"`
	StepType        StepType         `json:"step_type"`
	IsCheckpoint    bool             `json:"is_checkpoint"`
	RequiresManual  bool             `json:"requires_manual"`
	CompletionMode  CompletionMode   `json:"completion_mode"`
	SortOrder       int              `json:"sort_order"`
	GemRequirements []GemRequirement `json:"gem_requirements,omitempty"`
	Conditions      []StepCondition  `json:"conditions,omitempty"`
}

// GemRequirement describes a gem that should be obtained or equipped at a step.
type GemRequirement struct {
	ID      int    `json:"id"`
	StepID  int    `json:"step_id"`
	GemName string `json:"gem_name"`
	Color   string `json:"color"` // red | green | blue
	Note    string `json:"note"`
}

// ConditionType identifies the inference mechanism for a step condition.
type ConditionType string

const (
	ConditionLogtailArea  ConditionType = "logtail_area"   // payload: {"area": "..."}
	ConditionLogtailEvent ConditionType = "logtail_event"  // payload: {"pattern": "..."}
	ConditionGGGLevel     ConditionType = "ggg_level"      // payload: {"min_level": N}
	ConditionGGGQuest     ConditionType = "ggg_quest"      // payload: {"quest_id": "...", "state": "Finished"}
	ConditionManualConfirm ConditionType = "manual_confirm" // no payload required
)

// StepCondition encodes one inference rule attached to a Step.
// Multiple conditions on the same step are evaluated in ascending Priority order;
// the first satisfied condition triggers the status change.
type StepCondition struct {
	ID            int               `json:"id"`
	StepID        int               `json:"step_id"`
	ConditionType ConditionType     `json:"condition_type"`
	Payload       map[string]string `json:"payload"`
	Priority      int               `json:"priority"`
	Notes         string            `json:"notes,omitempty"`
}

// GearHintPriority ranks the importance of a gear hint.
type GearHintPriority string

const (
	GearHintHigh   GearHintPriority = "high"
	GearHintMedium GearHintPriority = "medium"
	GearHintLow    GearHintPriority = "low"
)

// GearHintRule is a piece of gear advice tied to a guide step (or the whole guide).
// StepID == 0 means the hint is global (not step-specific).
type GearHintRule struct {
	ID          int              `json:"id"`
	GuideID     int              `json:"guide_id"`
	StepID      *int             `json:"step_id,omitempty"`
	Slot        string           `json:"slot"`   // helmet | chest | weapon | ring | …
	Description string           `json:"description"`
	MinLife     *int             `json:"min_life,omitempty"`
	MinRes      *int             `json:"min_res,omitempty"`
	Priority    GearHintPriority `json:"priority"`
	Notes       string           `json:"notes,omitempty"`
}
