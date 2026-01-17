package discover

import (
	"fmt"
	"strings"
)

// FormatSchemaContext formats discovered resources as a context string for agent prompts.
// The output is a summarized view suitable for injection into system prompts.
// Returns an empty string if no resources are provided.
func FormatSchemaContext(resources []DiscoveredResource) string {
	if len(resources) == 0 {
		return ""
	}

	// Group resources by kind
	nodes := make([]DiscoveredResource, 0)
	relationships := make([]DiscoveredResource, 0)
	algorithms := make([]DiscoveredResource, 0)
	pipelines := make([]DiscoveredResource, 0)
	retrievers := make([]DiscoveredResource, 0)
	var agentContext string

	for _, r := range resources {
		switch r.Kind {
		case KindSchema:
			// Extract AgentContext from Schema
			if r.AgentContext != "" {
				agentContext = r.AgentContext
			}
		case KindNodeType:
			nodes = append(nodes, r)
		case KindRelationshipType:
			relationships = append(relationships, r)
		case KindAlgorithm:
			algorithms = append(algorithms, r)
		case KindPipeline:
			pipelines = append(pipelines, r)
		case KindRetriever:
			retrievers = append(retrievers, r)
		}
	}

	var sb strings.Builder
	sb.WriteString("## Existing Schema\n\n")

	// Write AgentContext instructions first (most important for agent behavior)
	if agentContext != "" {
		sb.WriteString("### Agent Instructions\n")
		sb.WriteString("IMPORTANT: Follow these instructions when working with this schema:\n\n")
		sb.WriteString(agentContext)
		sb.WriteString("\n\n")
	}

	sb.WriteString("The following resources already exist in this project. Use wetwire_list or read the source files for full property details.\n\n")

	// Write nodes section
	if len(nodes) > 0 {
		sb.WriteString("### Nodes\n")
		for _, n := range nodes {
			sb.WriteString(fmt.Sprintf("- %s (%s:%d)\n", n.Name, n.File, n.Line))
		}
		sb.WriteString("\n")
	}

	// Write relationships section
	if len(relationships) > 0 {
		sb.WriteString("### Relationships\n")
		for _, r := range relationships {
			sb.WriteString(fmt.Sprintf("- %s (%s:%d)\n", r.Name, r.File, r.Line))
		}
		sb.WriteString("\n")
	}

	// Write algorithms section
	if len(algorithms) > 0 {
		sb.WriteString("### Algorithms\n")
		for _, a := range algorithms {
			sb.WriteString(fmt.Sprintf("- %s (%s:%d)\n", a.Name, a.File, a.Line))
		}
		sb.WriteString("\n")
	}

	// Write pipelines section
	if len(pipelines) > 0 {
		sb.WriteString("### Pipelines\n")
		for _, p := range pipelines {
			sb.WriteString(fmt.Sprintf("- %s (%s:%d)\n", p.Name, p.File, p.Line))
		}
		sb.WriteString("\n")
	}

	// Write retrievers section
	if len(retrievers) > 0 {
		sb.WriteString("### Retrievers\n")
		for _, r := range retrievers {
			sb.WriteString(fmt.Sprintf("- %s (%s:%d)\n", r.Name, r.File, r.Line))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Extend or reference these resources. Do not recreate them.\n")

	return sb.String()
}
