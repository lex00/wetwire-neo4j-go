package discover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDependencyGraph_TopologicalSort(t *testing.T) {
	resources := []DiscoveredResource{
		{Name: "C", Kind: KindNodeType, Dependencies: []string{"A", "B"}},
		{Name: "A", Kind: KindNodeType, Dependencies: []string{}},
		{Name: "B", Kind: KindNodeType, Dependencies: []string{"A"}},
	}

	g := NewDependencyGraph(resources)
	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(sorted))
	}

	// A should come before B, B should come before C
	positions := make(map[string]int)
	for i, r := range sorted {
		positions[r.Name] = i
	}

	if positions["A"] > positions["B"] {
		t.Error("A should come before B")
	}
	if positions["B"] > positions["C"] {
		t.Error("B should come before C")
	}
	if positions["A"] > positions["C"] {
		t.Error("A should come before C")
	}
}

func TestDependencyGraph_TopologicalSort_Cycle(t *testing.T) {
	resources := []DiscoveredResource{
		{Name: "A", Kind: KindNodeType, Dependencies: []string{"B"}},
		{Name: "B", Kind: KindNodeType, Dependencies: []string{"C"}},
		{Name: "C", Kind: KindNodeType, Dependencies: []string{"A"}},
	}

	g := NewDependencyGraph(resources)
	_, err := g.TopologicalSort()
	if err == nil {
		t.Error("expected error for circular dependency")
	}
}

func TestDependencyGraph_HasCycle(t *testing.T) {
	t.Run("no cycle", func(t *testing.T) {
		resources := []DiscoveredResource{
			{Name: "A", Kind: KindNodeType, Dependencies: []string{}},
			{Name: "B", Kind: KindNodeType, Dependencies: []string{"A"}},
		}
		g := NewDependencyGraph(resources)
		if g.HasCycle() {
			t.Error("should not have cycle")
		}
	})

	t.Run("has cycle", func(t *testing.T) {
		resources := []DiscoveredResource{
			{Name: "A", Kind: KindNodeType, Dependencies: []string{"B"}},
			{Name: "B", Kind: KindNodeType, Dependencies: []string{"A"}},
		}
		g := NewDependencyGraph(resources)
		if !g.HasCycle() {
			t.Error("should have cycle")
		}
	})
}

func TestDependencyGraph_GetDependencies(t *testing.T) {
	resources := []DiscoveredResource{
		{Name: "A", Kind: KindNodeType, Dependencies: []string{}},
		{Name: "B", Kind: KindNodeType, Dependencies: []string{"A"}},
		{Name: "C", Kind: KindNodeType, Dependencies: []string{"B"}},
	}

	g := NewDependencyGraph(resources)

	t.Run("leaf node", func(t *testing.T) {
		deps := g.GetDependencies("A")
		if len(deps) != 0 {
			t.Errorf("expected 0 deps, got %d", len(deps))
		}
	})

	t.Run("one level", func(t *testing.T) {
		deps := g.GetDependencies("B")
		if len(deps) != 1 || deps[0] != "A" {
			t.Errorf("expected [A], got %v", deps)
		}
	})

	t.Run("recursive", func(t *testing.T) {
		deps := g.GetDependencies("C")
		if len(deps) != 2 {
			t.Errorf("expected 2 deps, got %d", len(deps))
		}
	})
}

// Test GetDependencies with unknown resource
func TestDependencyGraph_GetDependencies_Unknown(t *testing.T) {
	resources := []DiscoveredResource{
		{Name: "A", Kind: KindNodeType, Dependencies: []string{}},
	}

	g := NewDependencyGraph(resources)
	deps := g.GetDependencies("Unknown")

	if len(deps) != 0 {
		t.Errorf("expected 0 deps for unknown resource, got %d", len(deps))
	}
}

// Test dependency extraction with various field types
func TestScanner_ExtractDependencies(t *testing.T) {
	content := `package main

type Location struct {
	NodeType
}

type Person struct {
	NodeType
	WorksAt *Location
	Name string
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "deps.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	// Find Person and check its dependencies
	var person *DiscoveredResource
	for i := range resources {
		if resources[i].Name == "Person" {
			person = &resources[i]
			break
		}
	}

	if person == nil {
		t.Fatal("Person resource not found")
	}

	// Should have Location as dependency
	found := false
	for _, dep := range person.Dependencies {
		if dep == "Location" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected Location dependency, got %v", person.Dependencies)
	}
}

// Test complex struct with various field references to exercise walkExprForDeps
func TestScanner_ScanFile_ComplexDependencies(t *testing.T) {
	content := `package main

type Address struct {
	NodeType
}

type Company struct {
	NodeType
	HeadOffice Address
}

type Person struct {
	NodeType
	Employer *Company
	Addresses []Address
	Properties map[string]Address
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "complex_deps.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if len(resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(resources))
	}

	// Find Person and verify dependencies
	var person *DiscoveredResource
	for i := range resources {
		if resources[i].Name == "Person" {
			person = &resources[i]
			break
		}
	}

	if person == nil {
		t.Fatal("Person not found")
	}

	// Person should depend on Company and Address
	hasDeps := len(person.Dependencies) >= 1
	if !hasDeps {
		t.Errorf("expected Person to have dependencies, got %v", person.Dependencies)
	}
}

// Test extractCompositeLitDependencies through variable declarations
func TestScanner_ScanFile_CompositeLitDependencies(t *testing.T) {
	content := `package main

var person = &NodeType{
	Label: "Person",
}

var worksFor = &RelationshipType{
	Label: "WORKS_FOR",
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "lit_deps.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	s := NewScanner()
	resources, err := s.ScanFile(tmpFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
}
