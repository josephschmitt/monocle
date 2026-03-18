package core

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/anthropics/monocle/internal/types"
)

// LoadConfig loads configuration from XDG-compliant paths.
// It checks ~/.config/monocle/config.json first, then .monocle/config.json in cwd.
func LoadConfig() (*types.Config, error) {
	cfg := DefaultConfig()

	// Global config
	globalPath := configPath()
	if data, err := os.ReadFile(globalPath); err == nil {
		json.Unmarshal(data, cfg) //nolint:errcheck
	}

	// Project-level config
	if data, err := os.ReadFile(".monocle/config.json"); err == nil {
		json.Unmarshal(data, cfg) //nolint:errcheck
	}

	return cfg, nil
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *types.Config {
	return &types.Config{
		DefaultAgent:   "claude",
		IgnorePatterns: []string{},
		DiffStyle:      "unified",
		SidebarStyle:   "flat",
		Theme:          "default",
		ReviewFormat: types.ReviewFormatConfig{
			IncludeSnippets: true,
			MaxSnippetLines: 10,
			IncludeSummary:  true,
		},
		Hooks: types.HooksConfig{
			TimeoutMs: 30000,
		},
	}
}

func configPath() string {
	cfgDir := os.Getenv("XDG_CONFIG_HOME")
	if cfgDir == "" {
		home, _ := os.UserHomeDir()
		cfgDir = filepath.Join(home, ".config")
	}
	return filepath.Join(cfgDir, "monocle", "config.json")
}

