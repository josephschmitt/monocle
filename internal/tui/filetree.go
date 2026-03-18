package tui

import (
	"sort"
	"strings"

	"github.com/anthropics/monocle/internal/types"
)

// fileTreeNode represents a node in the file tree hierarchy.
// Directory nodes have Children and nil File; file nodes have File and nil Children.
type fileTreeNode struct {
	Name     string              // Display name (may be compressed path like "internal/tui")
	Path     string              // Full path from repo root
	File     *types.ChangedFile  // Non-nil for file nodes
	Children []*fileTreeNode     // Non-nil for directory nodes
}

// visibleItem represents a single item in the flattened visible list.
type visibleItem struct {
	node  *fileTreeNode
	depth int
	isDir bool
}

// buildFileTree constructs a tree from a flat list of changed files.
// It splits paths by "/", builds a nested tree, sorts directories first
// then alphabetically, and compresses single-child directory chains.
func buildFileTree(files []types.ChangedFile) []*fileTreeNode {
	if len(files) == 0 {
		return nil
	}

	// Build an intermediate tree using a map for fast child lookup.
	type buildNode struct {
		name     string
		path     string
		file     *types.ChangedFile
		children map[string]*buildNode
		order    []string // insertion order keys
	}

	root := &buildNode{children: make(map[string]*buildNode)}

	for i := range files {
		f := &files[i]
		parts := strings.Split(f.Path, "/")
		current := root

		for j, part := range parts {
			if j == len(parts)-1 {
				// Leaf file node.
				if _, seen := current.children[part]; !seen {
					current.order = append(current.order, part)
				}
				current.children[part] = &buildNode{
					name: part,
					path: f.Path,
					file: f,
				}
			} else {
				// Directory node.
				child, exists := current.children[part]
				if !exists {
					dirPath := strings.Join(parts[:j+1], "/")
					child = &buildNode{
						name:     part,
						path:     dirPath,
						children: make(map[string]*buildNode),
					}
					current.children[part] = child
					current.order = append(current.order, part)
				}
				current = child
			}
		}
	}

	// Convert buildNode tree to fileTreeNode tree.
	var convert func(bn *buildNode) *fileTreeNode
	convert = func(bn *buildNode) *fileTreeNode {
		node := &fileTreeNode{
			Name: bn.name,
			Path: bn.path,
		}
		if bn.file != nil {
			node.File = bn.file
			return node
		}
		for _, key := range bn.order {
			child := bn.children[key]
			node.Children = append(node.Children, convert(child))
		}
		sortChildren(node.Children)
		return node
	}

	var roots []*fileTreeNode
	for _, key := range root.order {
		child := root.children[key]
		roots = append(roots, convert(child))
	}
	sortChildren(roots)

	// Compress single-child directory chains.
	for _, r := range roots {
		compressTree(r)
	}

	return roots
}

// compressTree merges single-child directory chains into one node.
// e.g., a/ -> b/ -> c/ -> file.go becomes "a/b/c" -> file.go
func compressTree(node *fileTreeNode) {
	if node.File != nil {
		return
	}

	// Recurse first so children are compressed before we check.
	for _, child := range node.Children {
		compressTree(child)
	}

	// Compress: if this dir has exactly one child and that child is also a dir,
	// merge the child into this node.
	for len(node.Children) == 1 && node.Children[0].File == nil {
		child := node.Children[0]
		node.Name = node.Name + "/" + child.Name
		node.Path = child.Path
		node.Children = child.Children
	}
}

// flattenTree produces an ordered list of visible items from the tree,
// respecting collapsed directories.
func flattenTree(roots []*fileTreeNode, collapsed map[string]bool) []visibleItem {
	var items []visibleItem
	for _, root := range roots {
		flattenNode(root, 0, collapsed, &items)
	}
	return items
}

func flattenNode(node *fileTreeNode, depth int, collapsed map[string]bool, items *[]visibleItem) {
	isDir := node.File == nil
	*items = append(*items, visibleItem{
		node:  node,
		depth: depth,
		isDir: isDir,
	})

	if isDir && !collapsed[node.Path] {
		for _, child := range node.Children {
			flattenNode(child, depth+1, collapsed, items)
		}
	}
}

// sortChildren sorts a slice of tree nodes: directories first, then files,
// alphabetical within each group.
func sortChildren(nodes []*fileTreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		iDir := nodes[i].File == nil
		jDir := nodes[j].File == nil
		if iDir != jDir {
			return iDir // dirs first
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})
}
