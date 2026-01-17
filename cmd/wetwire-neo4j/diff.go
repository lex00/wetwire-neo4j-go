// Command diff compares two Neo4j configurations and shows differences.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/lex00/wetwire-neo4j-go/internal/discover"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <path1> <path2>",
	Short: "Compare two Neo4j configurations",
	Long: `Semantically compare two Neo4j configurations and show differences.

Shows added resources, removed resources, modified resources, and dependency changes.

Supported output formats:
  - text (default): Human-readable output
  - json: Machine-readable JSON format
  - markdown: Markdown formatted diff report

Examples:
  # Compare two Go packages
  wetwire-neo4j diff ./old-schema ./new-schema

  # JSON output for automation
  wetwire-neo4j diff ./v1 ./v2 --format json

  # Markdown report
  wetwire-neo4j diff ./v1 ./v2 --format markdown`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

func init() {
	diffCmd.Flags().String("format", "text", "Output format: text, json, markdown")
}

// resourceDiff represents changes to a single resource.
type resourceDiff struct {
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
	Changes []string `json:"changes,omitempty"`
}

// diffResult holds the comparison results.
type diffResult struct {
	Success           bool           `json:"success"`
	Message           string         `json:"message,omitempty"`
	AddedResources    []resourceDiff `json:"added_resources,omitempty"`
	RemovedResources  []resourceDiff `json:"removed_resources,omitempty"`
	ModifiedResources []resourceDiff `json:"modified_resources,omitempty"`
}

func newDiffCmd() *cobra.Command {
	return diffCmd
}

func runDiff(cmd *cobra.Command, args []string) error {
	outputFormat, _ := cmd.Flags().GetString("format")

	path1 := args[0]
	path2 := args[1]

	// Discover resources from both paths
	scanner := discover.NewScanner()

	resources1, err := scanner.ScanDir(path1)
	if err != nil {
		return outputDiffError(outputFormat, fmt.Errorf("scan %s: %w", path1, err))
	}

	resources2, err := scanner.ScanDir(path2)
	if err != nil {
		return outputDiffError(outputFormat, fmt.Errorf("scan %s: %w", path2, err))
	}

	// Compare the configurations
	result := compareResources(resources1, resources2)
	result.Success = true

	// Output the result
	switch outputFormat {
	case "json":
		return outputDiffJSON(result)
	case "markdown":
		outputDiffMarkdown(result)
	default:
		outputDiffText(result)
	}

	return nil
}

func compareResources(old, new []discover.DiscoveredResource) diffResult {
	result := diffResult{}

	// Create maps for easier lookup
	oldMap := make(map[string]discover.DiscoveredResource)
	newMap := make(map[string]discover.DiscoveredResource)

	for _, r := range old {
		oldMap[r.Name] = r
	}
	for _, r := range new {
		newMap[r.Name] = r
	}

	// Find added resources
	for _, newRes := range new {
		if _, exists := oldMap[newRes.Name]; !exists {
			result.AddedResources = append(result.AddedResources, resourceDiff{
				Name: newRes.Name,
				Kind: string(newRes.Kind),
			})
		}
	}

	// Find removed resources
	for _, oldRes := range old {
		if _, exists := newMap[oldRes.Name]; !exists {
			result.RemovedResources = append(result.RemovedResources, resourceDiff{
				Name: oldRes.Name,
				Kind: string(oldRes.Kind),
			})
		}
	}

	// Find modified resources
	for _, newRes := range new {
		oldRes, exists := oldMap[newRes.Name]
		if !exists {
			continue // Resource was added, not modified
		}

		var changes []string

		// Check kind changes
		if oldRes.Kind != newRes.Kind {
			changes = append(changes, fmt.Sprintf("kind: %s -> %s", oldRes.Kind, newRes.Kind))
		}

		// Check source/target changes (for relationships)
		if oldRes.Source != newRes.Source {
			changes = append(changes, fmt.Sprintf("source: %s -> %s", oldRes.Source, newRes.Source))
		}
		if oldRes.Target != newRes.Target {
			changes = append(changes, fmt.Sprintf("target: %s -> %s", oldRes.Target, newRes.Target))
		}

		// Check property count changes
		if len(oldRes.Properties) != len(newRes.Properties) {
			changes = append(changes, fmt.Sprintf("properties: %d -> %d", len(oldRes.Properties), len(newRes.Properties)))
		}

		// Check constraint count changes
		if len(oldRes.Constraints) != len(newRes.Constraints) {
			changes = append(changes, fmt.Sprintf("constraints: %d -> %d", len(oldRes.Constraints), len(newRes.Constraints)))
		}

		// Check index count changes
		if len(oldRes.Indexes) != len(newRes.Indexes) {
			changes = append(changes, fmt.Sprintf("indexes: %d -> %d", len(oldRes.Indexes), len(newRes.Indexes)))
		}

		// Check dependency changes
		oldDeps := make(map[string]bool)
		newDeps := make(map[string]bool)
		for _, dep := range oldRes.Dependencies {
			oldDeps[dep] = true
		}
		for _, dep := range newRes.Dependencies {
			newDeps[dep] = true
		}

		var addedDeps, removedDeps []string
		for dep := range newDeps {
			if !oldDeps[dep] {
				addedDeps = append(addedDeps, dep)
			}
		}
		for dep := range oldDeps {
			if !newDeps[dep] {
				removedDeps = append(removedDeps, dep)
			}
		}

		if len(addedDeps) > 0 {
			sort.Strings(addedDeps)
			changes = append(changes, fmt.Sprintf("added dependencies: %v", addedDeps))
		}
		if len(removedDeps) > 0 {
			sort.Strings(removedDeps)
			changes = append(changes, fmt.Sprintf("removed dependencies: %v", removedDeps))
		}

		// If there are any changes, add to modified resources
		if len(changes) > 0 {
			result.ModifiedResources = append(result.ModifiedResources, resourceDiff{
				Name:    newRes.Name,
				Kind:    string(newRes.Kind),
				Changes: changes,
			})
		}
	}

	// Sort results for consistent output
	sort.Slice(result.AddedResources, func(i, j int) bool {
		return result.AddedResources[i].Name < result.AddedResources[j].Name
	})
	sort.Slice(result.RemovedResources, func(i, j int) bool {
		return result.RemovedResources[i].Name < result.RemovedResources[j].Name
	})
	sort.Slice(result.ModifiedResources, func(i, j int) bool {
		return result.ModifiedResources[i].Name < result.ModifiedResources[j].Name
	})

	return result
}

func outputDiffError(format string, err error) error {
	result := diffResult{
		Success: false,
		Message: err.Error(),
	}

	if format == "json" {
		return outputDiffJSON(result)
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	return nil
}

func outputDiffText(result diffResult) {
	hasChanges := len(result.AddedResources) > 0 ||
		len(result.RemovedResources) > 0 ||
		len(result.ModifiedResources) > 0

	if !hasChanges {
		fmt.Println("No differences found.")
		return
	}

	fmt.Println("Neo4j Configuration Diff")
	fmt.Println("========================")
	fmt.Println()

	if len(result.AddedResources) > 0 {
		fmt.Println("Added Resources:")
		for _, r := range result.AddedResources {
			fmt.Printf("  + %s (%s)\n", r.Name, r.Kind)
		}
		fmt.Println()
	}

	if len(result.RemovedResources) > 0 {
		fmt.Println("Removed Resources:")
		for _, r := range result.RemovedResources {
			fmt.Printf("  - %s (%s)\n", r.Name, r.Kind)
		}
		fmt.Println()
	}

	if len(result.ModifiedResources) > 0 {
		fmt.Println("Modified Resources:")
		for _, r := range result.ModifiedResources {
			fmt.Printf("  ~ %s (%s)\n", r.Name, r.Kind)
			for _, change := range r.Changes {
				fmt.Printf("    - %s\n", change)
			}
		}
		fmt.Println()
	}
}

func outputDiffJSON(result diffResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return nil
	}
	fmt.Println(string(data))
	return nil
}

func outputDiffMarkdown(result diffResult) {
	hasChanges := len(result.AddedResources) > 0 ||
		len(result.RemovedResources) > 0 ||
		len(result.ModifiedResources) > 0

	fmt.Println("## Neo4j Configuration Diff")
	fmt.Println()

	if !hasChanges {
		fmt.Println("No differences found.")
		return
	}

	if len(result.AddedResources) > 0 {
		fmt.Println("### Added Resources")
		fmt.Println()
		for _, r := range result.AddedResources {
			fmt.Printf("- `%s` (%s)\n", r.Name, r.Kind)
		}
		fmt.Println()
	}

	if len(result.RemovedResources) > 0 {
		fmt.Println("### Removed Resources")
		fmt.Println()
		for _, r := range result.RemovedResources {
			fmt.Printf("- `%s` (%s)\n", r.Name, r.Kind)
		}
		fmt.Println()
	}

	if len(result.ModifiedResources) > 0 {
		fmt.Println("### Modified Resources")
		fmt.Println()
		for _, r := range result.ModifiedResources {
			fmt.Printf("#### `%s` (%s)\n\n", r.Name, r.Kind)
			for _, change := range r.Changes {
				fmt.Printf("- %s\n", change)
			}
			fmt.Println()
		}
	}
}
