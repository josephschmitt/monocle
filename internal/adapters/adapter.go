package adapters

// SkillInstaller manages skill file installation for a specific agent.
type SkillInstaller interface {
	Name() string
	Detect() bool
	SkillPath(global bool) (string, bool)
	IsInstalled(skillPath string) (bool, error)
	Install(skillPath string) error
	Uninstall(skillPath string) error
}

// InstallResult reports what happened for a single agent.
type InstallResult struct {
	Agent       string
	SkillPath   string
	Installed   bool
	AlreadyDone bool
	Err         error
	Details     []string // additional installation info lines
}

// AllInstallers returns installers for all supported agents.
func AllInstallers() []SkillInstaller {
	return []SkillInstaller{
		&ClaudeAdapter{},
		&GeminiAdapter{},
		&CodexAdapter{},
		&OpenCodeAdapter{},
	}
}

// GetInstaller returns the installer for the given agent name, or nil.
func GetInstaller(name string) SkillInstaller {
	for _, i := range AllInstallers() {
		if i.Name() == name {
			return i
		}
	}
	return nil
}
