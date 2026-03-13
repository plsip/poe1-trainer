package gem

// Color of a socket / gem.
type Color string

const (
	ColorRed   Color = "red"
	ColorGreen Color = "green"
	ColorBlue  Color = "blue"
	ColorWhite Color = "white"
)

// Gem is a dictionary entry for a Path of Exile active or support gem.
// Names are the canonical English in-game names (source of truth).
type Gem struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Color     Color  `json:"color"`
	IsSkill   bool   `json:"is_skill"`
	IsSupport bool   `json:"is_support"`
	WikiURL   string `json:"wiki_url,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

// AcquisitionType describes how a gem can be obtained.
type AcquisitionType string

const (
	AcquisitionVendor AcquisitionType = "vendor"
	AcquisitionDrop   AcquisitionType = "drop"
	AcquisitionQuest  AcquisitionType = "quest"
)

// AvailabilityRule records the earliest point in a guide where a gem can be obtained.
type AvailabilityRule struct {
	ID                 int             `json:"id"`
	GuideID            int             `json:"guide_id"`
	GemID              int             `json:"gem_id"`
	ActFirstAvailable  int             `json:"act_first_available"`
	VendorName         string          `json:"vendor_name,omitempty"`
	VendorAct          *int            `json:"vendor_act,omitempty"`
	AcquisitionType    AcquisitionType `json:"acquisition_type"`
	QuestName          string          `json:"quest_name,omitempty"`
	Notes              string          `json:"notes,omitempty"`
}

// ActionType describes what should be done with a gem at a guide step.
type ActionType string

const (
	ActionSocket  ActionType = "socket"
	ActionLink    ActionType = "link"
	ActionQuality ActionType = "quality"
	ActionVaal    ActionType = "vaal"
	ActionSwap    ActionType = "swap"
	ActionBuy     ActionType = "buy"

	// Setup transition actions (Prompt 07).
	ActionStartLevel  ActionType = "start_level"   // Begin leveling a gem (typically on weapon swap).
	ActionWeaponSwap  ActionType = "weapon_swap"   // Place gem on weapon swap for passive leveling.
	ActionFullSwitch  ActionType = "full_switch"   // Perform a full setup switch to this gem/skill.
	ActionQuestReward ActionType = "quest_reward"  // Receive gem as a quest reward.
	ActionVendorFallback ActionType = "vendor_fallback" // Acquire from a fallback vendor in a later act.
)

// UpgradeRule defines a gem-related action required at a specific guide step.
type UpgradeRule struct {
	ID          int        `json:"id"`
	GuideID     int        `json:"guide_id"`
	StepID      int        `json:"step_id"`
	GemID       int        `json:"gem_id"`
	ActionType  ActionType `json:"action_type"`
	TargetValue string     `json:"target_value,omitempty"` // "4L", "B-B-R-R", "20/20"
	Priority    int        `json:"priority"`
	Notes       string     `json:"notes,omitempty"`
}
