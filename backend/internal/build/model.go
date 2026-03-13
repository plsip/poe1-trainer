package build

import "time"

// Build is the top-level archetype of a character build (e.g., "Storm Burst Totems").
type Build struct {
	ID          int       `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Class       string    `json:"class"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Version is a specific revision of a Build, tied to a game patch or meta shift.
// One Build can have many Versions; one Version has exactly one Guide.
type Version struct {
	ID         int       `json:"id"`
	BuildID    int       `json:"build_id"`
	Version    string    `json:"version"`
	PatchTag   string    `json:"patch_tag"`
	Notes      string    `json:"notes"`
	IsCurrent  bool      `json:"is_current"`
	ReleasedAt *time.Time `json:"released_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
