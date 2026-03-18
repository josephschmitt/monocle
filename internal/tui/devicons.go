package tui

import (
	"path/filepath"
	"strings"
)

// fileIcon returns a Nerd Font devicon for the given file path.
func fileIcon(path string) string {
	base := filepath.Base(path)

	// Exact filename matches
	if icon, ok := nameIcons[strings.ToLower(base)]; ok {
		return icon
	}

	// Extension matches
	ext := strings.ToLower(filepath.Ext(path))
	if icon, ok := extIcons[ext]; ok {
		return icon
	}

	// Default
	return "\uf15b" //
}

var nameIcons = map[string]string{
	"makefile":         "\ue779", //
	"dockerfile":       "\ue7b0", //
	"docker-compose.yml": "\ue7b0",
	".gitignore":       "\ue702", //
	".gitconfig":       "\ue702",
	".gitmodules":      "\ue702",
	"go.mod":           "\ue626", //
	"go.sum":           "\ue626",
	"package.json":     "\ue71e", //
	"tsconfig.json":    "\ue628", //
	"license":          "\uf718", //
	"readme.md":        "\uf48a", //
	"changelog.md":     "\uf48a",
	".env":             "\uf462", //
	".env.local":       "\uf462",
	"devbox.json":      "\uf489", //
	"devbox.lock":      "\uf489",
}

var extIcons = map[string]string{
	// Go
	".go": "\ue626", //

	// Web
	".js":   "\ue74e", //
	".mjs":  "\ue74e",
	".jsx":  "\ue7ba", //
	".ts":   "\ue628", //
	".tsx":  "\ue7ba",
	".html": "\ue736", //
	".css":  "\ue749", //
	".scss": "\ue749",
	".vue":  "\ue6a0", //

	// Data
	".json": "\ue60b", //
	".yaml": "\ue60b",
	".yml":  "\ue60b",
	".toml": "\ue60b",
	".xml":  "\ue619", //
	".csv":  "\uf1c3", //

	// Scripting
	".py":   "\ue73c", //
	".rb":   "\ue791", //
	".sh":   "\ue795", //
	".bash": "\ue795",
	".zsh":  "\ue795",
	".fish": "\ue795",
	".lua":  "\ue620", //

	// Systems
	".rs":   "\ue7a8", //
	".c":    "\ue61e", //
	".h":    "\ue61e",
	".cpp":  "\ue61d", //
	".hpp":  "\ue61d",
	".java": "\ue738", //
	".kt":   "\ue634", //
	".swift": "\ue755", //

	// Config
	".sql":  "\ue706", //
	".db":   "\ue706",
	".graphql": "\ue662",

	// Docs
	".md":   "\ue73e", //
	".txt":  "\uf15c", //
	".pdf":  "\uf1c1", //

	// Images
	".png":  "\uf1c5", //
	".jpg":  "\uf1c5",
	".jpeg": "\uf1c5",
	".gif":  "\uf1c5",
	".svg":  "\uf1c5",
	".ico":  "\uf1c5",

	// Archives
	".zip":  "\uf1c6", //
	".tar":  "\uf1c6",
	".gz":   "\uf1c6",

	// Lock files
	".lock": "\uf023", //
}
