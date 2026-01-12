package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/wetwire-core-go/cmd"
)

func TestInitializer_Interface(t *testing.T) {
	// Verify Initializer implements cmd.Initializer interface
	var _ cmd.Initializer = (*Initializer)(nil)
}

func TestInitializer_Init(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "my-neo4j-project"
	projectPath := filepath.Join(tmpDir, projectName)

	init := NewInitializer()
	err := init.Init(context.Background(), projectPath, cmd.InitOptions{})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify directory structure
	expectedDirs := []string{
		"",
		"schema",
		"algorithms",
		"pipelines",
		"retrievers",
		"kg",
	}
	for _, dir := range expectedDirs {
		path := filepath.Join(projectPath, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", path)
		}
	}

	// Verify main.go exists
	mainPath := filepath.Join(projectPath, "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		t.Error("expected main.go to exist")
	}

	// Verify schema file exists
	schemaPath := filepath.Join(projectPath, "schema", "schema.go")
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Error("expected schema/schema.go to exist")
	}
}

func TestInitializer_InitForce(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "existing-project")

	// Create existing directory with a file
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	existingFile := filepath.Join(projectPath, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	init := NewInitializer()

	// Without force, should fail
	err := init.Init(context.Background(), projectPath, cmd.InitOptions{Force: false})
	if err == nil {
		t.Error("expected error when directory exists without force")
	}

	// With force, should succeed
	err = init.Init(context.Background(), projectPath, cmd.InitOptions{Force: true})
	if err != nil {
		t.Errorf("Init with force failed: %v", err)
	}

	// Verify main.go was created
	mainPath := filepath.Join(projectPath, "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		t.Error("expected main.go to exist after force init")
	}
}

func TestInitializer_InitWithTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "template-project")

	init := NewInitializer()

	// Test with "gds" template
	err := init.Init(context.Background(), projectPath, cmd.InitOptions{Template: "gds"})
	if err != nil {
		t.Fatalf("Init with gds template failed: %v", err)
	}

	// Should have algorithms
	algoPath := filepath.Join(projectPath, "algorithms", "algorithms.go")
	if _, err := os.Stat(algoPath); os.IsNotExist(err) {
		t.Error("expected algorithms/algorithms.go to exist with gds template")
	}
}

func TestInitializer_InitGraphRAGTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "graphrag-project")

	init := NewInitializer()

	// Test with "graphrag" template
	err := init.Init(context.Background(), projectPath, cmd.InitOptions{Template: "graphrag"})
	if err != nil {
		t.Fatalf("Init with graphrag template failed: %v", err)
	}

	// Should have retrievers and kg
	retrieverPath := filepath.Join(projectPath, "retrievers", "retrievers.go")
	if _, err := os.Stat(retrieverPath); os.IsNotExist(err) {
		t.Error("expected retrievers/retrievers.go to exist with graphrag template")
	}

	kgPath := filepath.Join(projectPath, "kg", "kg.go")
	if _, err := os.Stat(kgPath); os.IsNotExist(err) {
		t.Error("expected kg/kg.go to exist with graphrag template")
	}
}

func TestInitializer_InvalidPath(t *testing.T) {
	init := NewInitializer()

	// Empty path should fail
	err := init.Init(context.Background(), "", cmd.InitOptions{})
	if err == nil {
		t.Error("expected error for empty path")
	}
}
