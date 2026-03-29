package adapters

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed all:skills
var skillsFS embed.FS

// SkillNames lists the skill directories embedded in the binary.
var SkillNames = []string{"get-feedback", "review-plan", "review-plan-wait"}

// InstallSkills writes the embedded skill directories to the given parent directory.
// Each skill is written as skillsDir/<name>/SKILL.md.
func InstallSkills(skillsDir string) error {
	for _, name := range SkillNames {
		content, err := skillsFS.ReadFile(filepath.Join("skills", name, "SKILL.md"))
		if err != nil {
			return fmt.Errorf("read embedded skill %s: %w", name, err)
		}
		dest := filepath.Join(skillsDir, name, "SKILL.md")
		if err := WriteFileAtomic(dest, content); err != nil {
			return fmt.Errorf("write skill %s: %w", name, err)
		}
	}
	return nil
}

// RemoveSkills removes installed skill directories from the given parent directory.
func RemoveSkills(skillsDir string) {
	for _, name := range SkillNames {
		dir := filepath.Join(skillsDir, name)
		_ = RemoveFileIfExists(filepath.Join(dir, "SKILL.md"))
		_ = os.Remove(dir) // remove dir if empty, ignore errors
	}
}

// SkillPaths returns the paths of installed skill files relative to skillsDir.
func SkillPaths(skillsDir string) []string {
	var paths []string
	for _, name := range SkillNames {
		paths = append(paths, filepath.Join(skillsDir, name, "SKILL.md"))
	}
	return paths
}
