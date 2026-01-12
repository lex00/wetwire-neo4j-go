package validator

import (
	"testing"

	"github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"
)

func TestConfig_Defaults(t *testing.T) {
	config := Config{
		URI:      "bolt://localhost:7687",
		Username: "neo4j",
		Password: "password",
	}

	// Database should default to "neo4j" when creating validator
	if config.Database != "" {
		t.Errorf("expected empty database, got %s", config.Database)
	}
}

func TestNew_MissingURI(t *testing.T) {
	config := Config{
		Username: "neo4j",
		Password: "password",
	}

	_, err := New(config)
	if err == nil {
		t.Error("expected error for missing URI")
	}
}

func TestValidationResult_Structure(t *testing.T) {
	r := ValidationResult{
		Type:    "schema",
		Target:  "NodeType(Person)",
		Valid:   true,
		Message: "label 'Person' exists in database",
		Details: map[string]any{"exists": true},
	}

	if r.Type != "schema" {
		t.Errorf("Type = %s, want schema", r.Type)
	}
	if !r.Valid {
		t.Error("expected Valid to be true")
	}
}

func TestHasErrors_NoErrors(t *testing.T) {
	results := []ValidationResult{
		{Valid: true},
		{Valid: true},
	}

	if HasErrors(results) {
		t.Error("HasErrors should return false for all valid results")
	}
}

func TestHasErrors_WithErrors(t *testing.T) {
	results := []ValidationResult{
		{Valid: true},
		{Valid: false},
	}

	if !HasErrors(results) {
		t.Error("HasErrors should return true when error present")
	}
}

func TestFilterInvalid(t *testing.T) {
	results := []ValidationResult{
		{Valid: true, Target: "valid1"},
		{Valid: false, Target: "invalid1"},
		{Valid: true, Target: "valid2"},
		{Valid: false, Target: "invalid2"},
	}

	invalid := FilterInvalid(results)
	if len(invalid) != 2 {
		t.Errorf("expected 2 invalid results, got %d", len(invalid))
	}

	for _, r := range invalid {
		if r.Valid {
			t.Error("FilterInvalid returned a valid result")
		}
	}
}

func TestFormatResults_Empty(t *testing.T) {
	output := FormatResults(nil)
	if output != "No validations performed" {
		t.Errorf("unexpected output for empty results: %s", output)
	}
}

func TestFormatResults_WithResults(t *testing.T) {
	results := []ValidationResult{
		{Type: "schema", Target: "Person", Valid: true, Message: "exists"},
		{Type: "schema", Target: "Unknown", Valid: false, Message: "not found"},
	}

	output := FormatResults(results)
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestGDSInfo_Structure(t *testing.T) {
	info := GDSInfo{
		Installed: true,
		Version:   "2.5.0",
		Edition:   "community",
	}

	if !info.Installed {
		t.Error("expected Installed to be true")
	}
	if info.Version != "2.5.0" {
		t.Errorf("Version = %s, want 2.5.0", info.Version)
	}
}

func TestDatabaseInfo_Structure(t *testing.T) {
	info := DatabaseInfo{
		Version:           "5.15.0",
		Edition:           "enterprise",
		NodeLabels:        []string{"Person", "Company"},
		RelationshipTypes: []string{"WORKS_FOR", "KNOWS"},
	}

	if len(info.NodeLabels) != 2 {
		t.Errorf("expected 2 node labels, got %d", len(info.NodeLabels))
	}
	if len(info.RelationshipTypes) != 2 {
		t.Errorf("expected 2 relationship types, got %d", len(info.RelationshipTypes))
	}
}

func TestExistsMsg(t *testing.T) {
	tests := []struct {
		exists bool
		want   string
	}{
		{true, "exists"},
		{false, "does not exist"},
	}

	for _, tt := range tests {
		got := existsMsg(tt.exists)
		if got != tt.want {
			t.Errorf("existsMsg(%v) = %s, want %s", tt.exists, got, tt.want)
		}
	}
}

// MockValidator simulates validation without a Neo4j connection
type MockValidator struct {
	dbInfo  *DatabaseInfo
	gdsInfo *GDSInfo
}

func NewMockValidator(dbInfo *DatabaseInfo, gdsInfo *GDSInfo) *MockValidator {
	return &MockValidator{
		dbInfo:  dbInfo,
		gdsInfo: gdsInfo,
	}
}

func (m *MockValidator) labelExists(label string) bool {
	for _, l := range m.dbInfo.NodeLabels {
		if l == label {
			return true
		}
	}
	return false
}

func (m *MockValidator) relationshipTypeExists(relType string) bool {
	for _, rt := range m.dbInfo.RelationshipTypes {
		if rt == relType {
			return true
		}
	}
	return false
}

func (m *MockValidator) ValidateNodeType(node *schema.NodeType) []ValidationResult {
	var results []ValidationResult

	labelExists := m.labelExists(node.Label)
	results = append(results, ValidationResult{
		Type:    "schema",
		Target:  "NodeType(" + node.Label + ")",
		Valid:   true,
		Message: "label '" + node.Label + "' " + existsMsg(labelExists) + " in database",
		Details: map[string]any{"exists": labelExists},
	})

	return results
}

func TestMockValidator_ValidateNodeType_Exists(t *testing.T) {
	dbInfo := &DatabaseInfo{
		NodeLabels: []string{"Person", "Company"},
	}
	gdsInfo := &GDSInfo{Installed: true}

	m := NewMockValidator(dbInfo, gdsInfo)

	node := &schema.NodeType{Label: "Person"}
	results := m.ValidateNodeType(node)

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	if !results[0].Valid {
		t.Error("expected validation to pass")
	}

	exists, ok := results[0].Details["exists"].(bool)
	if !ok || !exists {
		t.Error("expected exists to be true")
	}
}

func TestMockValidator_ValidateNodeType_NotExists(t *testing.T) {
	dbInfo := &DatabaseInfo{
		NodeLabels: []string{"Company"},
	}
	gdsInfo := &GDSInfo{Installed: true}

	m := NewMockValidator(dbInfo, gdsInfo)

	node := &schema.NodeType{Label: "Person"}
	results := m.ValidateNodeType(node)

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	exists, ok := results[0].Details["exists"].(bool)
	if !ok || exists {
		t.Error("expected exists to be false")
	}
}

func TestMockValidator_RelationshipTypeExists(t *testing.T) {
	dbInfo := &DatabaseInfo{
		RelationshipTypes: []string{"KNOWS", "WORKS_FOR"},
	}
	gdsInfo := &GDSInfo{Installed: true}

	m := NewMockValidator(dbInfo, gdsInfo)

	tests := []struct {
		relType string
		want    bool
	}{
		{"KNOWS", true},
		{"WORKS_FOR", true},
		{"UNKNOWN", false},
	}

	for _, tt := range tests {
		got := m.relationshipTypeExists(tt.relType)
		if got != tt.want {
			t.Errorf("relationshipTypeExists(%s) = %v, want %v", tt.relType, got, tt.want)
		}
	}
}

func TestFormatResults_Mixed(t *testing.T) {
	results := []ValidationResult{
		{Type: "schema", Target: "NodeType(Person)", Valid: true, Message: "label exists"},
		{Type: "algorithm", Target: "PageRank", Valid: false, Message: "GDS not installed"},
		{Type: "projection", Target: "my-graph", Valid: true, Message: "projection valid"},
	}

	output := FormatResults(results)

	// Check all results are formatted
	if output == "" {
		t.Error("expected non-empty output")
	}

	// Check status indicators
	if !contains(output, "✓") {
		t.Error("expected checkmark for valid results")
	}
	if !contains(output, "✗") {
		t.Error("expected X for invalid results")
	}

	// Check targets are included
	if !contains(output, "NodeType(Person)") {
		t.Error("expected NodeType(Person) in output")
	}
	if !contains(output, "PageRank") {
		t.Error("expected PageRank in output")
	}
}

func TestFilterInvalid_AllValid(t *testing.T) {
	results := []ValidationResult{
		{Valid: true, Target: "a"},
		{Valid: true, Target: "b"},
		{Valid: true, Target: "c"},
	}

	invalid := FilterInvalid(results)
	if len(invalid) != 0 {
		t.Errorf("expected 0 invalid results, got %d", len(invalid))
	}
}

func TestFilterInvalid_AllInvalid(t *testing.T) {
	results := []ValidationResult{
		{Valid: false, Target: "a"},
		{Valid: false, Target: "b"},
	}

	invalid := FilterInvalid(results)
	if len(invalid) != 2 {
		t.Errorf("expected 2 invalid results, got %d", len(invalid))
	}
}

func TestFilterInvalid_Empty(t *testing.T) {
	var results []ValidationResult

	invalid := FilterInvalid(results)
	if len(invalid) != 0 {
		t.Errorf("expected 0 results, got %d", len(invalid))
	}
}

func TestHasErrors_Empty(t *testing.T) {
	var results []ValidationResult

	if HasErrors(results) {
		t.Error("HasErrors should return false for empty results")
	}
}

func TestValidationResult_Details(t *testing.T) {
	r := ValidationResult{
		Type:    "schema",
		Target:  "Test",
		Valid:   true,
		Message: "test message",
		Details: map[string]any{
			"exists":      true,
			"count":       42,
			"name":        "test",
			"nested":      map[string]any{"key": "value"},
		},
	}

	// Test details access
	if _, ok := r.Details["exists"]; !ok {
		t.Error("expected exists key in details")
	}
	if count, ok := r.Details["count"].(int); !ok || count != 42 {
		t.Error("expected count to be 42")
	}
	if name, ok := r.Details["name"].(string); !ok || name != "test" {
		t.Error("expected name to be 'test'")
	}
}

func TestConfig_AllFields(t *testing.T) {
	config := Config{
		URI:      "neo4j+s://example.com:7687",
		Username: "admin",
		Password: "secret123",
		Database: "mydb",
	}

	if config.URI != "neo4j+s://example.com:7687" {
		t.Errorf("URI = %s, want neo4j+s://example.com:7687", config.URI)
	}
	if config.Username != "admin" {
		t.Errorf("Username = %s, want admin", config.Username)
	}
	if config.Password != "secret123" {
		t.Errorf("Password = %s, want secret123", config.Password)
	}
	if config.Database != "mydb" {
		t.Errorf("Database = %s, want mydb", config.Database)
	}
}

func TestGDSInfo_NotInstalled(t *testing.T) {
	info := GDSInfo{
		Installed: false,
	}

	if info.Installed {
		t.Error("expected Installed to be false")
	}
	if info.Version != "" {
		t.Errorf("Version should be empty, got %s", info.Version)
	}
	if info.Edition != "" {
		t.Errorf("Edition should be empty, got %s", info.Edition)
	}
}

func TestDatabaseInfo_Empty(t *testing.T) {
	info := DatabaseInfo{}

	if len(info.NodeLabels) != 0 {
		t.Errorf("expected 0 node labels, got %d", len(info.NodeLabels))
	}
	if len(info.RelationshipTypes) != 0 {
		t.Errorf("expected 0 relationship types, got %d", len(info.RelationshipTypes))
	}
}

func TestMockValidator_LabelExists(t *testing.T) {
	dbInfo := &DatabaseInfo{
		NodeLabels: []string{"Person", "Company", "Location"},
	}
	gdsInfo := &GDSInfo{Installed: true}

	m := NewMockValidator(dbInfo, gdsInfo)

	tests := []struct {
		label string
		want  bool
	}{
		{"Person", true},
		{"Company", true},
		{"Location", true},
		{"Unknown", false},
		{"", false},
		{"person", false}, // case sensitive
	}

	for _, tt := range tests {
		got := m.labelExists(tt.label)
		if got != tt.want {
			t.Errorf("labelExists(%s) = %v, want %v", tt.label, got, tt.want)
		}
	}
}

func TestValidationResult_TypeValues(t *testing.T) {
	types := []string{"schema", "algorithm", "projection"}

	for _, typ := range types {
		r := ValidationResult{Type: typ}
		if r.Type != typ {
			t.Errorf("Type = %s, want %s", r.Type, typ)
		}
	}
}

func TestFormatResults_SingleValid(t *testing.T) {
	results := []ValidationResult{
		{Type: "schema", Target: "Test", Valid: true, Message: "passed"},
	}

	output := FormatResults(results)
	if !contains(output, "✓") {
		t.Error("expected checkmark in output")
	}
	if contains(output, "✗") {
		t.Error("should not contain X for valid-only results")
	}
}

func TestFormatResults_SingleInvalid(t *testing.T) {
	results := []ValidationResult{
		{Type: "schema", Target: "Test", Valid: false, Message: "failed"},
	}

	output := FormatResults(results)
	if contains(output, "✓") {
		t.Error("should not contain checkmark for invalid-only results")
	}
	if !contains(output, "✗") {
		t.Error("expected X in output")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
