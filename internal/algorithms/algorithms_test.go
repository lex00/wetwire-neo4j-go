package algorithms

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPageRank_Interface(t *testing.T) {
	pr := &PageRank{
		BaseAlgorithm: BaseAlgorithm{
			Name:      "influence",
			GraphName: "social",
			Mode:      Stream,
		},
		DampingFactor: 0.85,
		MaxIterations: 50,
	}

	if pr.AlgorithmName() != "influence" {
		t.Errorf("AlgorithmName() = %v, want influence", pr.AlgorithmName())
	}
	if pr.AlgorithmType() != "gds.pageRank" {
		t.Errorf("AlgorithmType() = %v, want gds.pageRank", pr.AlgorithmType())
	}
	if pr.AlgorithmCategory() != Centrality {
		t.Errorf("AlgorithmCategory() = %v, want Centrality", pr.AlgorithmCategory())
	}
	if pr.GetGraphName() != "social" {
		t.Errorf("GetGraphName() = %v, want social", pr.GetGraphName())
	}
	if pr.GetMode() != Stream {
		t.Errorf("GetMode() = %v, want stream", pr.GetMode())
	}
}

func TestBaseAlgorithm_DefaultMode(t *testing.T) {
	base := &BaseAlgorithm{Name: "test", GraphName: "graph"}
	if base.GetMode() != Stream {
		t.Errorf("GetMode() default = %v, want stream", base.GetMode())
	}
}

func TestAlgorithmCategories(t *testing.T) {
	tests := []struct {
		algo     Algorithm
		category Category
	}{
		{&PageRank{}, Centrality},
		{&ArticleRank{}, Centrality},
		{&Betweenness{}, Centrality},
		{&Degree{}, Centrality},
		{&Closeness{}, Centrality},
		{&Louvain{}, Community},
		{&Leiden{}, Community},
		{&LabelPropagation{}, Community},
		{&WCC{}, Community},
		{&TriangleCount{}, Community},
		{&KCore{}, Community},
		{&NodeSimilarity{}, Similarity},
		{&KNN{}, Similarity},
		{&FastRP{}, Embeddings},
		{&Node2Vec{}, Embeddings},
		{&GraphSAGE{}, Embeddings},
		{&HashGNN{}, Embeddings},
		{&Dijkstra{}, PathFinding},
		{&AStar{}, PathFinding},
		{&BFS{}, PathFinding},
		{&DFS{}, PathFinding},
	}

	for _, tt := range tests {
		t.Run(tt.algo.AlgorithmType(), func(t *testing.T) {
			if tt.algo.AlgorithmCategory() != tt.category {
				t.Errorf("Category = %v, want %v", tt.algo.AlgorithmCategory(), tt.category)
			}
		})
	}
}

func TestAlgorithmTypes(t *testing.T) {
	tests := []struct {
		algo Algorithm
		typ  string
	}{
		{&PageRank{}, "gds.pageRank"},
		{&ArticleRank{}, "gds.articleRank"},
		{&Betweenness{}, "gds.betweenness"},
		{&Degree{}, "gds.degree"},
		{&Closeness{}, "gds.closeness"},
		{&Louvain{}, "gds.louvain"},
		{&Leiden{}, "gds.leiden"},
		{&LabelPropagation{}, "gds.labelPropagation"},
		{&WCC{}, "gds.wcc"},
		{&TriangleCount{}, "gds.triangleCount"},
		{&KCore{}, "gds.kcore"},
		{&NodeSimilarity{}, "gds.nodeSimilarity"},
		{&KNN{}, "gds.knn"},
		{&FastRP{}, "gds.fastRP"},
		{&Node2Vec{}, "gds.node2vec"},
		{&GraphSAGE{}, "gds.beta.graphSage"},
		{&HashGNN{}, "gds.hashgnn"},
		{&Dijkstra{}, "gds.shortestPath.dijkstra"},
		{&AStar{}, "gds.shortestPath.astar"},
		{&BFS{}, "gds.bfs"},
		{&DFS{}, "gds.dfs"},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			if tt.algo.AlgorithmType() != tt.typ {
				t.Errorf("AlgorithmType() = %v, want %v", tt.algo.AlgorithmType(), tt.typ)
			}
		})
	}
}

func TestNewAlgorithmSerializer(t *testing.T) {
	s := NewAlgorithmSerializer()
	if s == nil {
		t.Fatal("NewAlgorithmSerializer returned nil")
	}
	if s.templates == nil {
		t.Error("templates is nil")
	}
}

func TestAlgorithmSerializer_ToCypher_PageRank(t *testing.T) {
	s := NewAlgorithmSerializer()
	pr := &PageRank{
		BaseAlgorithm: BaseAlgorithm{
			Name:      "influence",
			GraphName: "social_graph",
			Mode:      Stream,
		},
		DampingFactor: 0.85,
		MaxIterations: 50,
		Tolerance:     0.0000001,
	}

	result, err := s.ToCypher(pr)
	if err != nil {
		t.Fatalf("ToCypher failed: %v", err)
	}

	if !strings.Contains(result, "CALL gds.pageRank.stream") {
		t.Errorf("expected gds.pageRank.stream, got: %s", result)
	}
	if !strings.Contains(result, "'social_graph'") {
		t.Errorf("expected graph name, got: %s", result)
	}
	if !strings.Contains(result, "dampingFactor: 0.85") {
		t.Errorf("expected dampingFactor, got: %s", result)
	}
	if !strings.Contains(result, "maxIterations: 50") {
		t.Errorf("expected maxIterations, got: %s", result)
	}
	if !strings.Contains(result, "YIELD nodeId, score") {
		t.Errorf("expected YIELD clause, got: %s", result)
	}
}

func TestAlgorithmSerializer_ToCypher_Louvain(t *testing.T) {
	s := NewAlgorithmSerializer()
	louvain := &Louvain{
		BaseAlgorithm: BaseAlgorithm{
			GraphName: "my_graph",
			Mode:      Write,
		},
		MaxLevels:     10,
		MaxIterations: 10,
		WriteProperty: "communityId",
	}

	result, err := s.ToCypher(louvain)
	if err != nil {
		t.Fatalf("ToCypher failed: %v", err)
	}

	if !strings.Contains(result, "CALL gds.louvain.write") {
		t.Errorf("expected gds.louvain.write, got: %s", result)
	}
	if !strings.Contains(result, "writeProperty: 'communityId'") {
		t.Errorf("expected writeProperty, got: %s", result)
	}
	if !strings.Contains(result, "YIELD nodePropertiesWritten") {
		t.Errorf("expected write YIELD, got: %s", result)
	}
}

func TestAlgorithmSerializer_ToCypher_FastRP(t *testing.T) {
	s := NewAlgorithmSerializer()
	fastrp := &FastRP{
		BaseAlgorithm: BaseAlgorithm{
			GraphName: "my_graph",
			Mode:      Mutate,
		},
		EmbeddingDimension: 256,
		IterationWeights:   []float64{0.0, 1.0, 1.0},
		MutateProperty:     "embedding",
	}

	result, err := s.ToCypher(fastrp)
	if err != nil {
		t.Fatalf("ToCypher failed: %v", err)
	}

	if !strings.Contains(result, "CALL gds.fastRP.mutate") {
		t.Errorf("expected gds.fastRP.mutate, got: %s", result)
	}
	if !strings.Contains(result, "embeddingDimension: 256") {
		t.Errorf("expected embeddingDimension, got: %s", result)
	}
	if !strings.Contains(result, "[0, 1, 1]") {
		t.Errorf("expected iterationWeights, got: %s", result)
	}
}

func TestAlgorithmSerializer_ToCypher_NodeSimilarity(t *testing.T) {
	s := NewAlgorithmSerializer()
	ns := &NodeSimilarity{
		BaseAlgorithm: BaseAlgorithm{
			GraphName: "my_graph",
			Mode:      Stats,
		},
		TopK:             10,
		SimilarityCutoff: 0.5,
		SimilarityMetric: "JACCARD",
	}

	result, err := s.ToCypher(ns)
	if err != nil {
		t.Fatalf("ToCypher failed: %v", err)
	}

	if !strings.Contains(result, "CALL gds.nodeSimilarity.stats") {
		t.Errorf("expected gds.nodeSimilarity.stats, got: %s", result)
	}
	if !strings.Contains(result, "YIELD nodeCount, relationshipCount") {
		t.Errorf("expected stats YIELD, got: %s", result)
	}
}

func TestAlgorithmSerializer_ToJSON_PageRank(t *testing.T) {
	s := NewAlgorithmSerializer()
	pr := &PageRank{
		BaseAlgorithm: BaseAlgorithm{
			Name:        "influence",
			GraphName:   "social_graph",
			Mode:        Stream,
			Concurrency: 4,
		},
		DampingFactor: 0.85,
		MaxIterations: 50,
	}

	result, err := s.ToJSON(pr)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed["name"] != "influence" {
		t.Errorf("name = %v, want influence", parsed["name"])
	}
	if parsed["graphName"] != "social_graph" {
		t.Errorf("graphName = %v, want social_graph", parsed["graphName"])
	}
	if parsed["algorithmType"] != "gds.pageRank" {
		t.Errorf("algorithmType = %v, want gds.pageRank", parsed["algorithmType"])
	}
	if parsed["dampingFactor"] != 0.85 {
		t.Errorf("dampingFactor = %v, want 0.85", parsed["dampingFactor"])
	}
}

func TestAlgorithmSerializer_ToMap(t *testing.T) {
	s := NewAlgorithmSerializer()
	pr := &PageRank{
		BaseAlgorithm: BaseAlgorithm{
			Name:      "test",
			GraphName: "graph",
			Mode:      Write,
		},
		DampingFactor: 0.9,
		WriteProperty: "score",
	}

	result := s.ToMap(pr)

	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
	if result["writeProperty"] != "score" {
		t.Errorf("writeProperty = %v, want score", result["writeProperty"])
	}
}

func TestAlgorithmSerializer_BatchToCypher(t *testing.T) {
	s := NewAlgorithmSerializer()
	algorithms := []Algorithm{
		&PageRank{
			BaseAlgorithm: BaseAlgorithm{Name: "pr", GraphName: "g", Mode: Stream},
		},
		&Louvain{
			BaseAlgorithm: BaseAlgorithm{Name: "lv", GraphName: "g", Mode: Stream},
		},
	}

	result, err := s.BatchToCypher(algorithms)
	if err != nil {
		t.Fatalf("BatchToCypher failed: %v", err)
	}

	if !strings.Contains(result, "gds.pageRank.stream") {
		t.Errorf("expected pageRank, got: %s", result)
	}
	if !strings.Contains(result, "gds.louvain.stream") {
		t.Errorf("expected louvain, got: %s", result)
	}
	if !strings.Contains(result, "// pr") {
		t.Errorf("expected comment, got: %s", result)
	}
}

func TestAlgorithmSerializer_BatchToJSON(t *testing.T) {
	s := NewAlgorithmSerializer()
	algorithms := []Algorithm{
		&PageRank{BaseAlgorithm: BaseAlgorithm{Name: "pr", GraphName: "g"}},
		&Louvain{BaseAlgorithm: BaseAlgorithm{Name: "lv", GraphName: "g"}},
	}

	result, err := s.BatchToJSON(algorithms)
	if err != nil {
		t.Fatalf("BatchToJSON failed: %v", err)
	}

	var parsed []map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("expected 2 algorithms, got %d", len(parsed))
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"DampingFactor", "dampingFactor"},
		{"MaxIterations", "maxIterations"},
		{"Name", "name"},
		{"ID", "id"},
		{"GraphName", "graphName"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := toCamelCase(tt.input); got != tt.want {
				t.Errorf("toCamelCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"string", "hello", "'hello'"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"string slice", []string{"a", "b"}, "['a', 'b']"},
		{"float slice", []float64{1.0, 2.0}, "[1, 2]"},
		{"int slice", []int{1, 2, 3}, "[1, 2, 3]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatValue(tt.input); got != tt.want {
				t.Errorf("formatValue(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetYieldFields(t *testing.T) {
	s := NewAlgorithmSerializer()

	tests := []struct {
		algo  Algorithm
		yield string
	}{
		{&PageRank{BaseAlgorithm: BaseAlgorithm{Mode: Stream}}, "nodeId, score"},
		{&PageRank{BaseAlgorithm: BaseAlgorithm{Mode: Stats}}, "nodeCount, relationshipCount, computeMillis"},
		{&PageRank{BaseAlgorithm: BaseAlgorithm{Mode: Write}}, "nodePropertiesWritten, computeMillis"},
		{&Louvain{BaseAlgorithm: BaseAlgorithm{Mode: Stream}}, "nodeId, communityId"},
		{&NodeSimilarity{BaseAlgorithm: BaseAlgorithm{Mode: Stream}}, "node1, node2, similarity"},
		{&FastRP{BaseAlgorithm: BaseAlgorithm{Mode: Stream}}, "nodeId, embedding"},
		{&Dijkstra{BaseAlgorithm: BaseAlgorithm{Mode: Stream}}, "sourceNode, targetNode, path, totalCost"},
	}

	for _, tt := range tests {
		t.Run(tt.algo.AlgorithmType(), func(t *testing.T) {
			result := s.getYieldFields(tt.algo)
			if result != tt.yield {
				t.Errorf("getYieldFields() = %v, want %v", result, tt.yield)
			}
		})
	}
}

func TestAlgorithm_ImplementsInterface(t *testing.T) {
	// Verify all algorithms implement the Algorithm interface
	var _ Algorithm = &PageRank{}
	var _ Algorithm = &ArticleRank{}
	var _ Algorithm = &Betweenness{}
	var _ Algorithm = &Degree{}
	var _ Algorithm = &Closeness{}
	var _ Algorithm = &Louvain{}
	var _ Algorithm = &Leiden{}
	var _ Algorithm = &LabelPropagation{}
	var _ Algorithm = &WCC{}
	var _ Algorithm = &TriangleCount{}
	var _ Algorithm = &KCore{}
	var _ Algorithm = &NodeSimilarity{}
	var _ Algorithm = &KNN{}
	var _ Algorithm = &FastRP{}
	var _ Algorithm = &Node2Vec{}
	var _ Algorithm = &GraphSAGE{}
	var _ Algorithm = &HashGNN{}
	var _ Algorithm = &Dijkstra{}
	var _ Algorithm = &AStar{}
	var _ Algorithm = &BFS{}
	var _ Algorithm = &DFS{}
}
