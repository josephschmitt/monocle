package adapters

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONFile_Missing(t *testing.T) {
	m, err := ReadJSONFile(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %v", m)
	}
}

func TestReadJSONFile_Empty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.json")
	os.WriteFile(path, []byte(""), 0644)

	m, err := ReadJSONFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %v", m)
	}
}

func TestReadJSONFile_Valid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.json")
	os.WriteFile(path, []byte(`{"key": "value", "nested": {"a": 1}}`), 0644)

	m, err := ReadJSONFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["key"] != "value" {
		t.Fatalf("expected key=value, got %v", m["key"])
	}
}

func TestReadJSONFile_Invalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(path, []byte(`not json`), 0644)

	_, err := ReadJSONFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWriteJSONFile_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "test.json")

	data := map[string]any{"hello": "world"}
	if err := WriteJSONFile(path, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, err := ReadJSONFile(path)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if m["hello"] != "world" {
		t.Fatalf("expected hello=world, got %v", m["hello"])
	}
}

func TestWriteJSONFile_Roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "roundtrip.json")

	data := map[string]any{
		"string": "value",
		"number": float64(42),
		"nested": map[string]any{
			"inner": true,
		},
	}

	if err := WriteJSONFile(path, data); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	m, err := ReadJSONFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if m["string"] != "value" {
		t.Fatalf("string mismatch: %v", m["string"])
	}
	nested := m["nested"].(map[string]any)
	if nested["inner"] != true {
		t.Fatalf("nested.inner mismatch: %v", nested["inner"])
	}
}

func TestGetNestedKey(t *testing.T) {
	m := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "deep",
			},
		},
		"top": "level",
	}

	tests := []struct {
		key    string
		want   any
		wantOK bool
	}{
		{"top", "level", true},
		{"a.b.c", "deep", true},
		{"a.b", map[string]any{"c": "deep"}, true},
		{"missing", nil, false},
		{"a.missing", nil, false},
		{"a.b.c.d", nil, false},
	}

	for _, tt := range tests {
		got, ok := GetNestedKey(m, tt.key)
		if ok != tt.wantOK {
			t.Errorf("GetNestedKey(%q): ok=%v, want %v", tt.key, ok, tt.wantOK)
		}
		if tt.wantOK && ok {
			// Compare string values directly
			if s, isStr := tt.want.(string); isStr {
				if got != s {
					t.Errorf("GetNestedKey(%q) = %v, want %v", tt.key, got, tt.want)
				}
			}
		}
	}
}

func TestSetNestedKey(t *testing.T) {
	m := map[string]any{}

	SetNestedKey(m, "a.b.c", "value")

	got, ok := GetNestedKey(m, "a.b.c")
	if !ok || got != "value" {
		t.Fatalf("expected a.b.c=value, got %v (ok=%v)", got, ok)
	}

	// Setting again should overwrite
	SetNestedKey(m, "a.b.c", "updated")
	got, _ = GetNestedKey(m, "a.b.c")
	if got != "updated" {
		t.Fatalf("expected a.b.c=updated, got %v", got)
	}

	// Should preserve sibling keys
	SetNestedKey(m, "a.b.d", "sibling")
	got, _ = GetNestedKey(m, "a.b.c")
	if got != "updated" {
		t.Fatalf("sibling set clobbered a.b.c: %v", got)
	}
	got, _ = GetNestedKey(m, "a.b.d")
	if got != "sibling" {
		t.Fatalf("expected a.b.d=sibling, got %v", got)
	}
}

func TestDeleteNestedKey(t *testing.T) {
	m := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "value",
				"d": "other",
			},
		},
	}

	if !DeleteNestedKey(m, "a.b.c") {
		t.Fatal("expected delete to return true")
	}
	if _, ok := GetNestedKey(m, "a.b.c"); ok {
		t.Fatal("a.b.c should be deleted")
	}
	// Sibling should remain
	if got, ok := GetNestedKey(m, "a.b.d"); !ok || got != "other" {
		t.Fatal("a.b.d should still exist")
	}

	// Deleting non-existent key returns false
	if DeleteNestedKey(m, "x.y.z") {
		t.Fatal("expected false for non-existent key")
	}
}
