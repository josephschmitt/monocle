package types

import (
	"encoding/json"
	"time"
)

type Config struct {
	IgnorePatterns []string          `json:"ignore_patterns"`
	Keybindings    map[string]string `json:"keybindings"`
	DiffStyle      string            `json:"diff_style"`
	SidebarStyle   string            `json:"sidebar_style"`
	Theme          string            `json:"theme"`
	Layout         string            `json:"layout"`
	Wrap           bool              `json:"wrap"`
	TabSize        int               `json:"tab_size"`
	ContextLines   int               `json:"context_lines"`
	ReviewFormat      ReviewFormatConfig `json:"review_format"`
	AutoFocusMode     bool               `json:"auto_focus_mode"`
	Mouse              *bool `json:"mouse"`
	MinDiffWidth       int   `json:"min_diff_width"`
	CommentExpand         *bool  `json:"comment_expand"`
	CommentExpandDelay    int    `json:"comment_expand_delay"`
	MarkReviewedOnSubmit  string `json:"mark_reviewed_on_submit"` // "all" (default), "commented", "manual"
	ReviewTracking        bool   `json:"review_tracking"`         // enable reviewed state, snapshots, change detection (default: true)
	// IdleTimeout controls how long `monocle serve` stays running after the
	// last client disconnects (minus a 60s grace window). Serialised as a
	// Go duration string (e.g. "30m", "1h"); an empty/zero value uses the
	// default 30 min, and a negative value disables idle shutdown.
	IdleTimeout Duration `json:"idle_timeout,omitempty"`
}

// Duration is a time.Duration alias with JSON support so config files can
// hold human-readable values like "30m" or "1h" instead of nanosecond ints.
type Duration time.Duration

// UnmarshalJSON accepts either a quoted duration string ("30m") or a raw
// integer-nanosecond value so hand-written JSON and marshalled output both
// round-trip.
func (d *Duration) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		if s == "" {
			*d = 0
			return nil
		}
		dur, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*d = Duration(dur)
		return nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*d = Duration(n)
	return nil
}

// MarshalJSON writes the duration as a human-readable string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

type ReviewFormatConfig struct {
	IncludeSnippets bool `json:"include_snippets"`
	MaxSnippetLines int  `json:"max_snippet_lines"`
	IncludeSummary  bool `json:"include_summary"`
}
