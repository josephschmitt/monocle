package tui

import (
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

type iconInfo struct {
	glyph string
	color string
}

var defaultIcon = iconInfo{"\uf15b", "7"}

// fileIcon returns a styled Nerd Font devicon for the given file path.
func fileIcon(path string) string {
	info := iconLookup(path)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(info.color)).Render(info.glyph)
}

func iconLookup(path string) iconInfo {
	base := filepath.Base(path)

	if info, ok := nameIcons[strings.ToLower(base)]; ok {
		return info
	}

	ext := strings.ToLower(filepath.Ext(path))
	if info, ok := extIcons[ext]; ok {
		return info
	}

	return defaultIcon
}

var nameIcons = map[string]iconInfo{
	"makefile":           {"\ue779", "#6d8086"},
	"dockerfile":         {"\ue7b0", "#384d54"},
	"docker-compose.yml": {"\ue7b0", "#384d54"},
	".gitignore":         {"\ue702", "#f54d27"},
	".gitconfig":         {"\ue702", "#f54d27"},
	".gitmodules":        {"\ue702", "#f54d27"},
	"go.mod":             {"\ue626", "#00acd7"},
	"go.sum":             {"\ue626", "#00acd7"},
	"package.json":       {"\ue71e", "#e8274b"},
	"tsconfig.json":      {"\ue628", "#519aba"},
	"license":            {"\uf718", "#d0bf41"},
	"readme.md":          {"\uf48a", "#519aba"},
	"changelog.md":       {"\uf48a", "#519aba"},
	".env":               {"\uf462", "#faf743"},
	".env.local":         {"\uf462", "#faf743"},
	"devbox.json":        {"\uf489", "#a074c4"},
	"devbox.lock":        {"\uf489", "#a074c4"},
}

var extIcons = map[string]iconInfo{
	// Go
	".go": {"\ue626", "#00acd7"},

	// Web
	".js":   {"\ue74e", "#cbcb41"},
	".mjs":  {"\ue74e", "#cbcb41"},
	".jsx":  {"\ue7ba", "#20c2e3"},
	".ts":   {"\ue628", "#519aba"},
	".tsx":  {"\ue7ba", "#519aba"},
	".html": {"\ue736", "#e44d26"},
	".css":  {"\ue749", "#42a5f5"},
	".scss": {"\ue749", "#f55385"},
	".vue":  {"\ue6a0", "#8dc149"},

	// Data
	".json": {"\ue60b", "#cbcb41"},
	".yaml": {"\ue60b", "#6d8086"},
	".yml":  {"\ue60b", "#6d8086"},
	".toml": {"\ue60b", "#6d8086"},
	".xml":  {"\ue619", "#e44d26"},
	".csv":  {"\uf1c3", "#89e051"},

	// Scripting
	".py":   {"\ue73c", "#ffbc03"},
	".rb":   {"\ue791", "#e52002"},
	".sh":   {"\ue795", "#4eaa25"},
	".bash": {"\ue795", "#4eaa25"},
	".zsh":  {"\ue795", "#4eaa25"},
	".fish": {"\ue795", "#4eaa25"},
	".lua":  {"\ue620", "#51a0cf"},

	// Systems
	".rs":    {"\ue7a8", "#dea584"},
	".c":     {"\ue61e", "#599eff"},
	".h":     {"\ue61e", "#a074c4"},
	".cpp":   {"\ue61d", "#f34b7d"},
	".hpp":   {"\ue61d", "#a074c4"},
	".java":  {"\ue738", "#cc3e44"},
	".kt":    {"\ue634", "#7f52ff"},
	".swift": {"\ue755", "#e37933"},

	// Config
	".sql":     {"\ue706", "#dad8d8"},
	".db":      {"\ue706", "#dad8d8"},
	".graphql": {"\ue662", "#e535ab"},

	// Docs
	".md":  {"\ue73e", "#519aba"},
	".txt": {"\uf15c", "#89e051"},
	".pdf": {"\uf1c1", "#b30b00"},

	// Images
	".png":  {"\uf1c5", "#a074c4"},
	".jpg":  {"\uf1c5", "#a074c4"},
	".jpeg": {"\uf1c5", "#a074c4"},
	".gif":  {"\uf1c5", "#a074c4"},
	".svg":  {"\uf1c5", "#ffb13b"},
	".ico":  {"\uf1c5", "#cbcb41"},

	// Archives
	".zip": {"\uf1c6", "#eca517"},
	".tar": {"\uf1c6", "#eca517"},
	".gz":  {"\uf1c6", "#eca517"},

	// Lock files
	".lock": {"\uf023", "#6d8086"},
}
