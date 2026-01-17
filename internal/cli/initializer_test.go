package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/wetwire-neo4j-go/internal/discover"
)

func TestInitializer_Interface(t *testing.T) {
}

func TestInitializer_Init(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "my-neo4j-project"
	projectPath := filepath.Join(tmpDir, projectName)

	init := NewInitializer()
	err := init.Init(context.Background(), projectPath, InitOptions{})
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

	// Verify go.mod exists
	goModPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Error("expected go.mod to exist")
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
	err := init.Init(context.Background(), projectPath, InitOptions{Force: false})
	if err == nil {
		t.Error("expected error when directory exists without force")
	}

	// With force, should succeed
	err = init.Init(context.Background(), projectPath, InitOptions{Force: true})
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
	err := init.Init(context.Background(), projectPath, InitOptions{Template: "gds"})
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
	err := init.Init(context.Background(), projectPath, InitOptions{Template: "graphrag"})
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
	err := init.Init(context.Background(), "", InitOptions{})
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestInitializer_E2E_InitThenList(t *testing.T) {
	// End-to-end test: init project → list finds resources
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "e2e-test")

	init := NewInitializer()
	err := init.Init(context.Background(), projectPath, InitOptions{})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify go.mod exists (required for proper Go project)
	goModPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Fatal("go.mod not created - list will fail")
	}

	// Now list should find resources
	scanner := discover.NewScanner()
	resources, err := scanner.ScanDir(projectPath)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// Should find at least Person, Company, WorksFor from default template
	if len(resources) < 3 {
		t.Errorf("expected at least 3 resources, got %d", len(resources))
	}

	// Verify we found Person
	foundPerson := false
	for _, r := range resources {
		if r.Name == "Person" {
			foundPerson = true
			break
		}
	}
	if !foundPerson {
		t.Error("list did not find Person node type after init")
	}
}

func TestE2E_InitImportList(t *testing.T) {
	// Full e2e test: init → import → list
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "full-e2e")

	// Step 1: Init project
	init := NewInitializer()
	if err := init.Init(context.Background(), projectPath, InitOptions{}); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify go.mod exists
	if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); os.IsNotExist(err) {
		t.Fatal("go.mod not created")
	}

	// Step 2: Create cypher file and import additional schema
	cypherContent := `CREATE CONSTRAINT order_id FOR (o:Order) REQUIRE o.id IS UNIQUE;
CREATE CONSTRAINT product_sku FOR (p:Product) REQUIRE p.sku IS UNIQUE;`
	cypherFile := filepath.Join(projectPath, "additional.cypher")
	if err := os.WriteFile(cypherFile, []byte(cypherContent), 0644); err != nil {
		t.Fatalf("failed to write cypher file: %v", err)
	}

	// Import to a new schema file
	importer := NewImporterCLI()
	outputFile := filepath.Join(projectPath, "schema", "orders.go")
	if err := importer.ImportToFile(cypherFile, "", "", "", "", "schema", outputFile); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify the imported file exists and has .go extension
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("imported file not created at %s", outputFile)
	}

	// Step 3: List should find resources from both init and import
	scanner := discover.NewScanner()
	resources, err := scanner.ScanDir(projectPath)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// Should find: Person, Company, WorksFor (from init) + order, product (from import)
	if len(resources) < 5 {
		t.Errorf("expected at least 5 resources, got %d", len(resources))
	}

	// Verify we found resources from init
	foundPerson := false
	foundCompany := false
	// Verify we found resources from import
	foundOrder := false
	foundProduct := false

	for _, r := range resources {
		switch r.Name {
		case "Person":
			foundPerson = true
		case "Company":
			foundCompany = true
		case "Order":
			foundOrder = true
		case "Product":
			foundProduct = true
		}
	}

	if !foundPerson {
		t.Error("list did not find Person (from init)")
	}
	if !foundCompany {
		t.Error("list did not find Company (from init)")
	}
	if !foundOrder {
		t.Error("list did not find Order (from import)")
	}
	if !foundProduct {
		t.Error("list did not find Product (from import)")
	}
}
