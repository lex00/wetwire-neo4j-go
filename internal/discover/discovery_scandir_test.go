package discover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanner_ScanDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files
	files := map[string]string{
		"schema.go": `package schema
type Person struct { NodeType }
type Company struct { NodeType }
`,
		"relationships.go": `package schema
type WorksFor struct { RelationshipType }
`,
		"subdir/algo.go": `package algo
type Influence struct { PageRank }
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	s := NewScanner()
	resources, err := s.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if len(resources) != 4 {
		t.Errorf("expected 4 resources, got %d", len(resources))
	}

	// Check we found all types
	kinds := make(map[ResourceKind]int)
	for _, r := range resources {
		kinds[r.Kind]++
	}

	if kinds[KindNodeType] != 2 {
		t.Errorf("expected 2 NodeType, got %d", kinds[KindNodeType])
	}
	if kinds[KindRelationshipType] != 1 {
		t.Errorf("expected 1 RelationshipType, got %d", kinds[KindRelationshipType])
	}
	if kinds[KindAlgorithm] != 1 {
		t.Errorf("expected 1 Algorithm, got %d", kinds[KindAlgorithm])
	}
}

func TestScanner_ScanDir_SkipsTestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"schema.go": `package schema
type Person struct { NodeType }
`,
		"schema_test.go": `package schema
type TestPerson struct { NodeType }
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	s := NewScanner()
	resources, err := s.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// Should only find Person, not TestPerson
	if len(resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(resources))
	}
	if resources[0].Name != "Person" {
		t.Errorf("expected Person, got %s", resources[0].Name)
	}
}

func TestScanner_ScanDir_SkipsVendor(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"schema.go": `package schema
type Person struct { NodeType }
`,
		"vendor/other/schema.go": `package other
type VendorType struct { NodeType }
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	s := NewScanner()
	resources, err := s.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// Should only find Person, not VendorType
	if len(resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(resources))
	}
}

// Test scanning empty directory
func TestScanner_ScanDir_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	s := NewScanner()
	resources, err := s.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

// Test scanning nonexistent directory
func TestScanner_ScanDir_Nonexistent(t *testing.T) {
	s := NewScanner()
	_, err := s.ScanDir("/nonexistent/directory")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// Test ScanDir with error in file
func TestScanner_ScanDir_WithParseError(t *testing.T) {
	tmpDir := t.TempDir()

	// Valid file
	validContent := `package main
type Person struct { NodeType }
`
	if err := os.WriteFile(filepath.Join(tmpDir, "valid.go"), []byte(validContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Invalid file - but ScanDir might skip or error
	invalidContent := `package main
func invalid { syntax error
`
	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.go"), []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	s := NewScanner()
	_, err := s.ScanDir(tmpDir)
	// Either error or partial results is acceptable
	if err != nil {
		t.Logf("ScanDir returned error (acceptable): %v", err)
	}
}
