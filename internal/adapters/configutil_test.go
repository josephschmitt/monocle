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
