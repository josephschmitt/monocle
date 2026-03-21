package core

import (
	"encoding/json"
	"fmt"
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
		IgnorePatterns: []string{},
		DiffStyle:      "unified",
		SidebarStyle:   "flat",
		Theme:          "default",
		ReviewFormat: types.ReviewFormatConfig{
			IncludeSnippets: true,
			MaxSnippetLines: 10,
			IncludeSummary:  true,
		},
	}
}

// SaveConfig writes the configuration to the global config path.
func SaveConfig(cfg *types.Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func configPath() string {
	cfgDir := os.Getenv("XDG_CONFIG_HOME")
	if cfgDir == "" {
		home, _ := os.UserHomeDir()
		cfgDir = filepath.Join(home, ".config")
	}
	return filepath.Join(cfgDir, "monocle", "config.json")
}
