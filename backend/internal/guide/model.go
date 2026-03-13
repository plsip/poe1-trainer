package guide

import "time"

// Guide represents a parsed campaign guide for a specific build.
type Guide struct {
	ID        int       `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	BuildName string    `json:"build_name"`
	Version   string    `json:"version"`
	Steps     []Step    `json:"steps"`
	CreatedAt time.Time `json:"created_at"`
}

// Step is a single actionable item in the guide.
type Step struct {
	ID             int              `json:"id"`
	GuideID        int              `json:"guide_id"`
	StepNumber     int              `json:"step_number"`
	Act            int              `json:"act"`
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	Area           string           `json:"area"`
	IsCheckpoint   bool             `json:"is_checkpoint"`
	RequiresManual bool             `json:"requires_manual"`
	SortOrder      int              `json:"sort_order"`
	GemRequirements []GemRequirement `json:"gem_requirements,omitempty"`
}

// GemRequirement describes a gem that should be obtained or equipped at a step.
type GemRequirement struct {
	ID      int    `json:"id"`
	StepID  int    `json:"step_id"`
	GemName string `json:"gem_name"`
	Color   string `json:"color"` // red | green | blue
	Note    string `json:"note"`
}
