package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/lex00/wetwire-neo4j-go/internal/discovery"
)

// Lister lists discovered Neo4j definitions.
type Lister struct {
	scanner *discovery.Scanner
}

// NewLister creates a new Lister.
func NewLister() *Lister {
	return &Lister{
		scanner: discovery.NewScanner(),
	}
}

// List discovers and displays all Neo4j definitions in the specified path.
func (l *Lister) List(path string, format string) error {
	resources, err := l.scanner.ScanDir(path)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(resources) == 0 {
		fmt.Println("No definitions found")
		return nil
	}

	// Sort resources by kind then name
	sort.Slice(resources, func(i, j int) bool {
		if resources[i].Kind != resources[j].Kind {
			return resources[i].Kind < resources[j].Kind
		}
		return resources[i].Name < resources[j].Name
	})

	switch format {
	case "json":
		return l.listJSON(resources)
	case "table":
		return l.listTable(resources)
	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json)", format)
	}
}

// listTable outputs resources in a table format.
func (l *Lister) listTable(resources []discovery.DiscoveredResource) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	_, _ = fmt.Fprintln(w, "TYPE\tNAME\tFILE\tLINE")
	_, _ = fmt.Fprintln(w, "----\t----\t----\t----")

	for _, r := range resources {
		// Shorten file path for display
		shortFile := shortenPath(r.File)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", r.Kind, r.Name, shortFile, r.Line)
	}

	_ = w.Flush()

	// Summary
	fmt.Println()
	l.printSummary(resources)

	return nil
}

// listJSON outputs resources in JSON format.
func (l *Lister) listJSON(resources []discovery.DiscoveredResource) error {
	output := make(map[string][]map[string]any)

	for _, r := range resources {
		entry := map[string]any{
			"name":    r.Name,
			"file":    r.File,
			"line":    r.Line,
			"package": r.Package,
		}
		if len(r.Dependencies) > 0 {
			entry["dependencies"] = r.Dependencies
		}

		key := string(r.Kind)
		output[key] = append(output[key], entry)
	}

	// Add summary
	summary := make(map[string]int)
	for kind, items := range output {
		summary[kind] = len(items)
	}
	output["_summary"] = []map[string]any{{"counts": summary, "total": len(resources)}}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// printSummary prints a summary of discovered resources.
func (l *Lister) printSummary(resources []discovery.DiscoveredResource) {
	counts := make(map[discovery.ResourceKind]int)
	for _, r := range resources {
		counts[r.Kind]++
	}

	fmt.Printf("Total: %d definitions\n", len(resources))

	// Sort kinds for consistent output
	var kinds []discovery.ResourceKind
	for kind := range counts {
		kinds = append(kinds, kind)
	}
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i] < kinds[j]
	})

	for _, kind := range kinds {
		fmt.Printf("  %s: %d\n", kind, counts[kind])
	}
}

// shortenPath shortens a file path for display.
func shortenPath(path string) string {
	// Try to make path relative to current directory
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}

	if strings.HasPrefix(path, cwd) {
		rel := strings.TrimPrefix(path, cwd)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			rel = "."
		}
		return rel
	}

	return path
}

// ListByKind lists resources filtered by kind.
func (l *Lister) ListByKind(path string, kind discovery.ResourceKind, format string) error {
	resources, err := l.scanner.ScanDir(path)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Filter by kind
	var filtered []discovery.DiscoveredResource
	for _, r := range resources {
		if r.Kind == kind {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) == 0 {
		fmt.Printf("No %s definitions found\n", kind)
		return nil
	}

	switch format {
	case "json":
		return l.listJSON(filtered)
	case "table":
		return l.listTable(filtered)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// ListDependencies shows the dependency graph for resources.
func (l *Lister) ListDependencies(path string) error {
	resources, err := l.scanner.ScanDir(path)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(resources) == 0 {
		fmt.Println("No definitions found")
		return nil
	}

	graph := discovery.NewDependencyGraph(resources)

	fmt.Println("Dependency Graph:")
	fmt.Println("-----------------")

	for _, r := range resources {
		deps := graph.GetDependencies(r.Name)
		if len(deps) > 0 {
			fmt.Printf("%s -> %s\n", r.Name, strings.Join(deps, ", "))
		}
	}

	// Check for cycles
	if graph.HasCycle() {
		fmt.Println("\nWarning: Circular dependencies detected!")
	} else {
		sorted, _ := graph.TopologicalSort()
		fmt.Println("\nBuild order:")
		for i, r := range sorted {
			fmt.Printf("  %d. %s (%s)\n", i+1, r.Name, r.Kind)
		}
	}

	return nil
}
