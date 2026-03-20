package adapters

import (
	"fmt"
)

// DetectAgents returns all agents that appear to be installed on the system.
func DetectAgents() []SkillInstaller {
	var found []SkillInstaller
	for _, installer := range AllInstallers() {
		if installer.Detect() {
			found = append(found, installer)
		}
	}
	return found
}

// InstallAgents runs install for a list of agent names.
// If agents is empty or contains "auto", it auto-detects installed agents.
// Returns a result per agent.
func InstallAgents(agents []string, global bool) []InstallResult {
	installers := resolveInstallers(agents, global)

	var results []InstallResult
	for _, inst := range installers {
		result := InstallResult{Agent: inst.Name()}

		skillPath, supported := inst.SkillPath(global)
		if !supported {
			result.Err = fmt.Errorf("%s does not support %s skill installation", inst.Name(), scopeLabel(global))
			results = append(results, result)
			continue
		}
		result.SkillPath = skillPath

		installed, err := inst.IsInstalled(skillPath)
		if err != nil {
			result.Err = fmt.Errorf("check %s: %w", inst.Name(), err)
			results = append(results, result)
			continue
		}
		if installed {
			result.AlreadyDone = true
			results = append(results, result)
			continue
		}

		if err := inst.Install(skillPath); err != nil {
			result.Err = fmt.Errorf("install %s: %w", inst.Name(), err)
		} else {
			result.Installed = true
			// Collect additional details from adapters that support it
			if detailer, ok := inst.(interface{ InstallDetails() []string }); ok {
				result.Details = detailer.InstallDetails()
			}
		}
		results = append(results, result)
	}
	return results
}

// UninstallAgents removes skills for specified agents.
// If agents is empty or contains "auto", it operates on all agents that have skills installed.
func UninstallAgents(agents []string, global bool) []InstallResult {
	installers := resolveInstallers(agents, global)

	var results []InstallResult
	for _, inst := range installers {
		result := InstallResult{Agent: inst.Name()}

		skillPath, supported := inst.SkillPath(global)
		if !supported {
			result.Err = fmt.Errorf("%s does not support %s skill installation", inst.Name(), scopeLabel(global))
			results = append(results, result)
			continue
		}
		result.SkillPath = skillPath

		installed, err := inst.IsInstalled(skillPath)
		if err != nil {
			result.Err = fmt.Errorf("check %s: %w", inst.Name(), err)
			results = append(results, result)
			continue
		}
		if !installed {
			result.AlreadyDone = true
			results = append(results, result)
			continue
		}

		if err := inst.Uninstall(skillPath); err != nil {
			result.Err = fmt.Errorf("uninstall %s: %w", inst.Name(), err)
		} else {
			result.Installed = true // reusing field to indicate action taken
		}
		results = append(results, result)
	}
	return results
}

// resolveInstallers converts agent name args to installers.
// "auto" or empty list means auto-detect.
func resolveInstallers(agents []string, global bool) []SkillInstaller {
	if isAuto(agents) {
		detected := DetectAgents()
		if global {
			return detected
		}
		// For project-level, filter to those supporting project config
		var supported []SkillInstaller
		for _, inst := range detected {
			if _, ok := inst.SkillPath(false); ok {
				supported = append(supported, inst)
			}
		}
		return supported
	}

	var installers []SkillInstaller
	for _, name := range agents {
		if inst := GetInstaller(name); inst != nil {
			installers = append(installers, inst)
		}
	}
	return installers
}

func isAuto(agents []string) bool {
	if len(agents) == 0 {
		return true
	}
	return len(agents) == 1 && agents[0] == "auto"
}

func scopeLabel(global bool) string {
	if global {
		return "global"
	}
	return "project"
}
