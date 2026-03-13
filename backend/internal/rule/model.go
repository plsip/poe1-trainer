package rule

// AlertKind classifies the type of alert shown to the player.
// Subcategories allow the frontend to group and style alerts precisely.
type AlertKind string

const (
	// Gem availability — where and how to acquire a gem for the first time.
	KindGemQuestReward    AlertKind = "gem_quest_reward"    // Quest reward unlock.
	KindGemVendor         AlertKind = "gem_vendor"          // Purchase from vendor.
	KindGemVendorFallback AlertKind = "gem_vendor_fallback" // Fallback vendor in later act.

	// Setup transitions — what to do with a gem at this stage.
	KindGemStartLeveling AlertKind = "gem_start_leveling" // Start leveling a gem on swap.
	KindGemWeaponSwap    AlertKind = "gem_weapon_swap"    // Place on weapon swap for passive leveling.
	KindGemFullSwitch    AlertKind = "gem_full_switch"    // Perform full setup switch.

	// Gear alerts — equipment checks.
	KindGearMovementSpeed AlertKind = "gear_movement_speed" // Check boots for movement speed.
	KindGearResistance    AlertKind = "gear_resistance"     // Craft or buy resistances.
	KindGearFourLink      AlertKind = "gear_4link"          // Look for 4-linked item.
	KindGearWeapon        AlertKind = "gear_weapon"         // Upgrade weapon.
	KindGearCheckpoint    AlertKind = "gear_checkpoint"     // Specific checkpoint item (e.g. Sapphire Ring).
	KindGearGeneral       AlertKind = "gear_general"        // Generic gear hint.
)

// Category returns the coarse API category: "gem" or "gear".
// Used to map rule.Alert to the api.Alert kind field (backward-compatible with frontend).
func (k AlertKind) Category() string {
	switch k {
	case KindGemQuestReward, KindGemVendor, KindGemVendorFallback,
		KindGemStartLeveling, KindGemWeaponSwap, KindGemFullSwitch:
		return "gem"
	default:
		return "gear"
	}
}

// Condition constrains when a rule fires.
// All non-zero / non-empty fields are ANDed together; zero values mean no restriction.
type Condition struct {
	// MinAct / MaxAct: act range (inclusive). 0 = no restriction.
	MinAct int `json:"min_act,omitempty"`
	MaxAct int `json:"max_act,omitempty"`
	// MinLevel / MaxLevel: character level range. 0 = no restriction.
	MinLevel int `json:"min_level,omitempty"`
	MaxLevel int `json:"max_level,omitempty"`
	// StepTypes: only fire at steps whose step_type is in this list.
	// Empty slice = no restriction (fires at any step type).
	StepTypes []string `json:"step_types,omitempty"`
}

// Alert is the output of a rule evaluation — a single actionable hint.
type Alert struct {
	// Kind is the fine-grained classification; Category() gives the coarse one.
	Kind AlertKind `json:"kind"`
	// Priority ranks urgency: "high" | "medium" | "low".
	Priority string `json:"priority"`
	// Description is the short, actionable text shown to the player.
	Description string `json:"description"`
	// Reason explains why this alert matters at the current stage.
	Reason string `json:"reason,omitempty"`
	// GemName is set for gem-category alerts.
	GemName string `json:"gem_name,omitempty"`
	// ActionType is the sub-action identifier (quest_reward | vendor | weapon_swap | full_switch | …).
	// In the API DTO this is mapped to the action_type field.
	ActionType string `json:"action_type,omitempty"`
	// Slot is the gear slot for gear-category alerts (boots | chest | ring | weapon | flask | …).
	Slot string `json:"slot,omitempty"`
}

// Rule pairs a Condition with the Alert it produces when the condition is met.
type Rule struct {
	Condition Condition `json:"condition"`
	Alert     Alert     `json:"alert"`
}

// RulesFile is the top-level format of an embedded rule seed file.
// One file per guide slug; loaded at startup via go:embed.
type RulesFile struct {
	GuideSlug string `json:"guide_slug"`
	Rules     []Rule `json:"rules"`
}

// EvalContext carries the run state used to check rule conditions.
type EvalContext struct {
	Act      int    // current step's act number
	Level    int    // character level from latest snapshot (0 = unknown)
	StepType string // current step's step_type value
}
