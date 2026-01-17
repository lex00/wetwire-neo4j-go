package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lex00/wetwire-neo4j-go/internal/discover"
)

// GraphCLI provides CLI functionality for visualizing resource dependencies.
type GraphCLI struct{}

// NewGraphCLI creates a new GraphCLI.
func NewGraphCLI() *GraphCLI {
	return &GraphCLI{}
}

// Generate generates a dependency graph visualization from discovered resources.
func (g *GraphCLI) Generate(path, format string, w io.Writer) error {
	// Discover resources
	scanner := discover.NewScanner()
	resources, err := scanner.ScanDir(path)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Build dependency graph
	graph := discover.NewDependencyGraph(resources)

	switch strings.ToLower(format) {
	case "dot", "graphviz":
		return g.generateDOT(resources, graph, w)
	case "mermaid":
		return g.generateMermaid(resources, graph, w)
	default:
		return fmt.Errorf("unsupported format: %s (use 'dot' or 'mermaid')", format)
	}
}

// generateDOT outputs the graph in DOT (Graphviz) format.
func (g *GraphCLI) generateDOT(resources []discover.DiscoveredResource, graph *discover.DependencyGraph, w io.Writer) error {
	_, _ = fmt.Fprintln(w, "digraph dependencies {")
	_, _ = fmt.Fprintln(w, "  rankdir=TB;")
	_, _ = fmt.Fprintln(w, "  node [shape=box];")
	_, _ = fmt.Fprintln(w)

	// Define node styles by kind
	kindColors := map[discover.ResourceKind]string{
		discover.KindNodeType:         "lightblue",
		discover.KindRelationshipType: "lightgreen",
		discover.KindAlgorithm:        "lightyellow",
		discover.KindPipeline:         "lightpink",
		discover.KindRetriever:        "lavender",
	}

	// Sort resources for deterministic output
	sortedResources := make([]discover.DiscoveredResource, len(resources))
	copy(sortedResources, resources)
	sort.Slice(sortedResources, func(i, j int) bool {
		return sortedResources[i].Name < sortedResources[j].Name
	})

	// Define nodes
	for _, r := range sortedResources {
		color := kindColors[r.Kind]
		if color == "" {
			color = "white"
		}
		label := fmt.Sprintf("%s\\n[%s]", r.Name, r.Kind)
		_, _ = fmt.Fprintf(w, "  %q [label=%q, style=filled, fillcolor=%s];\n",
			r.Name, label, color)
	}

	_, _ = fmt.Fprintln(w)

	// Define edges (dependencies)
	for _, r := range sortedResources {
		deps := graph.GetDependencies(r.Name)
		sort.Strings(deps)
		for _, dep := range deps {
			_, _ = fmt.Fprintf(w, "  %q -> %q;\n", r.Name, dep)
		}
	}

	_, _ = fmt.Fprintln(w, "}")
	return nil
}

// generateMermaid outputs the graph in Mermaid format.
func (g *GraphCLI) generateMermaid(resources []discover.DiscoveredResource, graph *discover.DependencyGraph, w io.Writer) error {
	_, _ = fmt.Fprintln(w, "graph TD")

	// Sort resources for deterministic output
	sortedResources := make([]discover.DiscoveredResource, len(resources))
	copy(sortedResources, resources)
	sort.Slice(sortedResources, func(i, j int) bool {
		return sortedResources[i].Name < sortedResources[j].Name
	})

	// Define nodes with kind as suffix
	for _, r := range sortedResources {
		nodeID := sanitizeMermaidID(r.Name)
		label := fmt.Sprintf("%s [%s]", r.Name, r.Kind)
		_, _ = fmt.Fprintf(w, "  %s[%q]\n", nodeID, label)
	}

	_, _ = fmt.Fprintln(w)

	// Define edges (dependencies)
	for _, r := range sortedResources {
		deps := graph.GetDependencies(r.Name)
		sort.Strings(deps)
		for _, dep := range deps {
			fromID := sanitizeMermaidID(r.Name)
			toID := sanitizeMermaidID(dep)
			_, _ = fmt.Fprintf(w, "  %s --> %s\n", fromID, toID)
		}
	}

	return nil
}

// sanitizeMermaidID converts a name to a valid Mermaid node ID.
func sanitizeMermaidID(name string) string {
	// Replace any characters that might cause issues
	result := strings.ReplaceAll(name, "-", "_")
	result = strings.ReplaceAll(result, " ", "_")
	return result
}
