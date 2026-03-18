package adapters

import (
	"fmt"
	"os/exec"
)

// DetectAgents returns all agents that appear to be installed on the system.
func DetectAgents() []AgentInstaller {
	var found []AgentInstaller
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
func InstallAgents(agents []string, global bool, scope string) []InstallResult {
	installers := resolveInstallers(agents, global)
	hookBin := resolveHookBinary()

	var results []InstallResult
	for _, inst := range installers {
		result := InstallResult{Agent: inst.Name()}

		configPath, supported := inst.ConfigPath(global)
		if !supported {
			result.Err = fmt.Errorf("%s does not support %s config", inst.Name(), scopeLabel(global))
			results = append(results, result)
			continue
		}
		result.ConfigPath = configPath

		installed, err := inst.IsInstalled(configPath)
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

		if err := inst.Install(configPath, InstallOptions{HookBinaryPath: hookBin, Scope: scope}); err != nil {
			result.Err = fmt.Errorf("install %s: %w", inst.Name(), err)
		} else {
			result.Installed = true
		}
		results = append(results, result)
	}
	return results
}

// UninstallAgents removes hooks for specified agents.
// If agents is empty or contains "auto", it operates on all agents that have hooks installed.
func UninstallAgents(agents []string, global bool) []InstallResult {
	installers := resolveInstallers(agents, global)

	var results []InstallResult
	for _, inst := range installers {
		result := InstallResult{Agent: inst.Name()}

		configPath, supported := inst.ConfigPath(global)
		if !supported {
			result.Err = fmt.Errorf("%s does not support %s config", inst.Name(), scopeLabel(global))
			results = append(results, result)
			continue
		}
		result.ConfigPath = configPath

		installed, err := inst.IsInstalled(configPath)
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

		if err := inst.Uninstall(configPath); err != nil {
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
func resolveInstallers(agents []string, global bool) []AgentInstaller {
	if isAuto(agents) {
		detected := DetectAgents()
		if global {
			return detected
		}
		// For project-level, filter to those supporting project config
		var supported []AgentInstaller
		for _, inst := range detected {
			if _, ok := inst.ConfigPath(false); ok {
				supported = append(supported, inst)
			}
		}
		return supported
	}

	var installers []AgentInstaller
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

// resolveHookBinary finds the monocle binary path.
func resolveHookBinary() string {
	if path, err := exec.LookPath("monocle"); err == nil {
		return path
	}
	return "monocle"
}
