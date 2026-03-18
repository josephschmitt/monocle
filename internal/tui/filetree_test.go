package tui

import (
	"testing"

	"github.com/anthropics/monocle/internal/types"
)

func makeFiles(paths ...string) []types.ChangedFile {
	files := make([]types.ChangedFile, len(paths))
	for i, p := range paths {
		files[i] = types.ChangedFile{Path: p, Status: types.FileModified}
	}
	return files
}

func TestBuildFileTree_Empty(t *testing.T) {
	roots := buildFileTree(nil)
	if len(roots) != 0 {
		t.Fatalf("expected 0 roots, got %d", len(roots))
	}
}

func TestBuildFileTree_SingleRootFile(t *testing.T) {
	roots := buildFileTree(makeFiles("main.go"))
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].Name != "main.go" {
		t.Errorf("expected name main.go, got %s", roots[0].Name)
	}
	if roots[0].File == nil {
		t.Error("expected file node, got directory")
	}
}

func TestBuildFileTree_SingleNestedFile(t *testing.T) {
	roots := buildFileTree(makeFiles("a/b/c.go"))
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	// Compressed: a/b should be one dir node
	r := roots[0]
	if r.Name != "a/b" {
		t.Errorf("expected compressed name a/b, got %s", r.Name)
	}
	if r.File != nil {
		t.Error("expected directory node")
	}
	if len(r.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(r.Children))
	}
	if r.Children[0].Name != "c.go" {
		t.Errorf("expected child name c.go, got %s", r.Children[0].Name)
	}
}

func TestBuildFileTree_MultipleFilesInSameDir(t *testing.T) {
	roots := buildFileTree(makeFiles("src/a.go", "src/b.go"))
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].Name != "src" {
		t.Errorf("expected name src, got %s", roots[0].Name)
	}
	if len(roots[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(roots[0].Children))
	}
}

func TestBuildFileTree_DeepSingleChildChain(t *testing.T) {
	roots := buildFileTree(makeFiles("a/b/c/d/e.go"))
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	r := roots[0]
	if r.Name != "a/b/c/d" {
		t.Errorf("expected compressed name a/b/c/d, got %s", r.Name)
	}
	if len(r.Children) != 1 || r.Children[0].Name != "e.go" {
		t.Error("expected single child e.go")
	}
}

func TestBuildFileTree_MixedDepths(t *testing.T) {
	roots := buildFileTree(makeFiles("a.go", "pkg/b.go", "pkg/sub/c.go"))
	// Should have 2 roots: pkg/ dir and a.go file
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	// Dirs first: pkg should come before a.go
	if roots[0].Name != "pkg" {
		t.Errorf("expected first root pkg (dir first), got %s", roots[0].Name)
	}
	if roots[1].Name != "a.go" {
		t.Errorf("expected second root a.go, got %s", roots[1].Name)
	}
	// pkg should have 2 children: sub/ dir and b.go file
	pkg := roots[0]
	if len(pkg.Children) != 2 {
		t.Fatalf("expected 2 children in pkg, got %d", len(pkg.Children))
	}
	// Dir first: sub before b.go
	if pkg.Children[0].Name != "sub" {
		t.Errorf("expected first child sub, got %s", pkg.Children[0].Name)
	}
	if pkg.Children[1].Name != "b.go" {
		t.Errorf("expected second child b.go, got %s", pkg.Children[1].Name)
	}
}

func TestBuildFileTree_CompressionStopsAtBranch(t *testing.T) {
	roots := buildFileTree(makeFiles("a/b/x.go", "a/b/y.go"))
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	// a/b compressed since a has only child b
	if roots[0].Name != "a/b" {
		t.Errorf("expected compressed a/b, got %s", roots[0].Name)
	}
	if len(roots[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(roots[0].Children))
	}
}

func TestBuildFileTree_NoCompressionMultipleChildren(t *testing.T) {
	roots := buildFileTree(makeFiles("a/x.go", "b/y.go"))
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	if roots[0].Name != "a" {
		t.Errorf("expected first root a, got %s", roots[0].Name)
	}
	if roots[1].Name != "b" {
		t.Errorf("expected second root b, got %s", roots[1].Name)
	}
}

func TestBuildFileTree_DirsSortedBeforeFiles(t *testing.T) {
	roots := buildFileTree(makeFiles("zebra.go", "alpha/file.go"))
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	if roots[0].File != nil {
		t.Error("expected first root to be a directory")
	}
	if roots[1].File == nil {
		t.Error("expected second root to be a file")
	}
}

func TestFlattenTree_AllExpanded(t *testing.T) {
	roots := buildFileTree(makeFiles("src/a.go", "src/b.go", "main.go"))
	items := flattenTree(roots, nil)
	// Expected: src/ (dir), a.go, b.go, main.go = 4 items
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}
	if !items[0].isDir {
		t.Error("expected first item to be dir")
	}
	if items[0].depth != 0 {
		t.Errorf("expected depth 0, got %d", items[0].depth)
	}
	if items[1].depth != 1 {
		t.Errorf("expected depth 1, got %d", items[1].depth)
	}
}

func TestFlattenTree_CollapsedDir(t *testing.T) {
	roots := buildFileTree(makeFiles("src/a.go", "src/b.go", "main.go"))
	collapsed := map[string]bool{"src": true}
	items := flattenTree(roots, collapsed)
	// Expected: src/ (collapsed), main.go = 2 items
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if !items[0].isDir || items[0].node.Name != "src" {
		t.Error("expected collapsed src dir as first item")
	}
	if items[1].node.Name != "main.go" {
		t.Errorf("expected main.go as second item, got %s", items[1].node.Name)
	}
}

func TestFlattenTree_NestedCollapse(t *testing.T) {
	roots := buildFileTree(makeFiles("a/b/c.go", "a/d.go", "z.go"))
	// a has children: b/ (dir) and d.go. b is not compressed since a has 2 children.
	collapsed := map[string]bool{"a": true}
	items := flattenTree(roots, collapsed)
	// Expected: a/ (collapsed), z.go = 2 items
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestFlattenTree_EmptyCollapsedMap(t *testing.T) {
	roots := buildFileTree(makeFiles("a.go", "b.go"))
	items := flattenTree(roots, map[string]bool{})
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestCompressTree_AlreadyLeaf(t *testing.T) {
	node := &fileTreeNode{
		Name: "main.go",
		Path: "main.go",
		File: &types.ChangedFile{Path: "main.go"},
	}
	compressTree(node)
	if node.Name != "main.go" {
		t.Errorf("expected name unchanged, got %s", node.Name)
	}
}

func TestSortChildren_DirsFirst(t *testing.T) {
	nodes := []*fileTreeNode{
		{Name: "z.go", File: &types.ChangedFile{}},
		{Name: "a", Children: []*fileTreeNode{}},
		{Name: "m.go", File: &types.ChangedFile{}},
		{Name: "b", Children: []*fileTreeNode{}},
	}
	sortChildren(nodes)
	expected := []string{"a", "b", "m.go", "z.go"}
	for i, n := range nodes {
		if n.Name != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], n.Name)
		}
	}
}

func TestBuildFileTree_PreservesFileReference(t *testing.T) {
	files := makeFiles("src/main.go")
	files[0].Status = types.FileAdded
	roots := buildFileTree(files)
	// Navigate to the file node
	if len(roots) != 1 || len(roots[0].Children) != 1 {
		t.Fatal("unexpected tree structure")
	}
	leaf := roots[0].Children[0]
	if leaf.File == nil {
		t.Fatal("expected file reference")
	}
	if leaf.File.Status != types.FileAdded {
		t.Errorf("expected FileAdded status, got %s", leaf.File.Status)
	}
}
