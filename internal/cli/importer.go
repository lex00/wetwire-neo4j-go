package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/lex00/wetwire-neo4j-go/internal/importer"
)

// ImporterCLI provides CLI functionality for importing Neo4j schemas.
type ImporterCLI struct{}

// NewImporterCLI creates a new ImporterCLI.
func NewImporterCLI() *ImporterCLI {
	return &ImporterCLI{}
}

// DefaultPackageName returns the default Go package name for generated code.
func (i *ImporterCLI) DefaultPackageName() string {
	return "schema"
}

// ImportFromCypher imports schema definitions from a Cypher file.
func (i *ImporterCLI) ImportFromCypher(filePath, packageName string, w io.Writer) error {
	if packageName == "" {
		packageName = i.DefaultPackageName()
	}

	imp := importer.NewCypherImporter(filePath)
	result, err := imp.Import(context.Background())
	if err != nil {
		return fmt.Errorf("failed to import from Cypher file: %w", err)
	}

	return i.generateOutput(result, packageName, w)
}

// ImportFromNeo4j imports schema definitions from a live Neo4j instance.
func (i *ImporterCLI) ImportFromNeo4j(uri, username, password, database, packageName string, w io.Writer) error {
	if uri == "" {
		return fmt.Errorf("URI is required for Neo4j import")
	}

	if packageName == "" {
		packageName = i.DefaultPackageName()
	}

	ctx := context.Background()
	imp, err := importer.NewNeo4jImporter(ctx, importer.Neo4jConfig{
		URI:      uri,
		Username: username,
		Password: password,
		Database: database,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Neo4j: %w", err)
	}
	defer func() { _ = imp.Close(ctx) }()

	result, err := imp.Import(ctx)
	if err != nil {
		return fmt.Errorf("failed to import from Neo4j: %w", err)
	}

	return i.generateOutput(result, packageName, w)
}

// ImportToFile imports schema definitions and writes to a file.
func (i *ImporterCLI) ImportToFile(cypherFile, uri, username, password, database, packageName, outputFile string) error {
	if packageName == "" {
		packageName = i.DefaultPackageName()
	}

	var result *importer.ImportResult
	var err error
	ctx := context.Background()

	if cypherFile != "" {
		imp := importer.NewCypherImporter(cypherFile)
		result, err = imp.Import(ctx)
		if err != nil {
			return fmt.Errorf("failed to import from Cypher file: %w", err)
		}
	} else if uri != "" {
		imp, connErr := importer.NewNeo4jImporter(ctx, importer.Neo4jConfig{
			URI:      uri,
			Username: username,
			Password: password,
			Database: database,
		})
		if connErr != nil {
			return fmt.Errorf("failed to connect to Neo4j: %w", connErr)
		}
		defer func() { _ = imp.Close(ctx) }()

		result, err = imp.Import(ctx)
		if err != nil {
			return fmt.Errorf("failed to import from Neo4j: %w", err)
		}
	} else {
		return fmt.Errorf("either --file or --uri must be provided")
	}

	// Generate code
	gen := importer.NewGenerator(packageName)
	code, err := gen.Generate(result)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func (i *ImporterCLI) generateOutput(result *importer.ImportResult, packageName string, w io.Writer) error {
	gen := importer.NewGenerator(packageName)
	code, err := gen.Generate(result)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	_, err = w.Write([]byte(code))
	return err
}
