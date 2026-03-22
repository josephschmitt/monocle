package core

import "github.com/anthropics/monocle/internal/types"

// gitStub implements GitAPI for unit tests without requiring a real git repo.
type gitStub struct {
	repoRoot   string
	currentRef string
	files      []types.ChangedFile
	diffResult *types.DiffResult
	commits    []LogEntry
}

func (g *gitStub) RepoRoot() string { return g.repoRoot }

func (g *gitStub) CurrentRef() (string, error) {
	return g.currentRef, nil
}

func (g *gitStub) Diff(_ string) ([]types.ChangedFile, error) {
	return g.files, nil
}

func (g *gitStub) FileDiff(_, _ string, _ int) (*types.DiffResult, error) {
	if g.diffResult != nil {
		return g.diffResult, nil
	}
	return &types.DiffResult{}, nil
}

func (g *gitStub) FileContent(_, _ string) (string, error) {
	return "", nil
}

func (g *gitStub) RecentCommits(_ int) ([]LogEntry, error) {
	return g.commits, nil
}

func (g *gitStub) ResolveRef(ref string) (string, error) {
	return g.currentRef, nil
}
