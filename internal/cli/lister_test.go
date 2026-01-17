package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/wetwire-neo4j-go/internal/discover"
)

func TestNewLister(t *testing.T) {
	l := NewLister()
	if l == nil {
		t.Fatal("NewLister returned nil")
	}
	if l.scanner == nil {
		t.Error("scanner is nil")
	}
}

func TestLister_List_EmptyDir(t *testing.T) {
	l := NewLister()

	// Create temp directory with no Go files
	tmpDir, err := os.MkdirTemp("", "lister-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err = l.List(tmpDir, "table")
	if err != nil {
		t.Errorf("List on empty dir failed: %v", err)
	}
}

func TestLister_List_TableFormat(t *testing.T) {
	l := NewLister()

	// Create temp directory with a Go file
	tmpDir, err := os.MkdirTemp("", "lister-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	goFile := filepath.Join(tmpDir, "schema.go")
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = schema.NodeType{
	Label: "Person",
}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err = l.List(tmpDir, "table")
	if err != nil {
		t.Errorf("List with table format failed: %v", err)
	}
}

func TestLister_List_JSONFormat(t *testing.T) {
	l := NewLister()

	// Create temp directory with a Go file
	tmpDir, err := os.MkdirTemp("", "lister-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	goFile := filepath.Join(tmpDir, "schema.go")
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = schema.NodeType{
	Label: "Person",
}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err = l.List(tmpDir, "json")
	if err != nil {
		t.Errorf("List with json format failed: %v", err)
	}
}

func TestLister_List_InvalidFormat(t *testing.T) {
	l := NewLister()

	tmpDir, err := os.MkdirTemp("", "lister-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	goFile := filepath.Join(tmpDir, "schema.go")
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = schema.NodeType{
	Label: "Person",
}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err = l.List(tmpDir, "invalid")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestLister_ListByKind_NodeType(t *testing.T) {
	l := NewLister()

	tmpDir, err := os.MkdirTemp("", "lister-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	goFile := filepath.Join(tmpDir, "schema.go")
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = schema.NodeType{
	Label: "Person",
}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err = l.ListByKind(tmpDir, discover.KindNodeType, "table")
	if err != nil {
		t.Errorf("ListByKind failed: %v", err)
	}
}

func TestLister_ListByKind_EmptyResult(t *testing.T) {
	l := NewLister()

	tmpDir, err := os.MkdirTemp("", "lister-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create file with NodeType but search for Algorithm
	goFile := filepath.Join(tmpDir, "schema.go")
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = schema.NodeType{
	Label: "Person",
}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err = l.ListByKind(tmpDir, discover.KindAlgorithm, "table")
	if err != nil {
		t.Errorf("ListByKind failed: %v", err)
	}
}

func TestLister_ListDependencies(t *testing.T) {
	l := NewLister()

	tmpDir, err := os.MkdirTemp("", "lister-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	goFile := filepath.Join(tmpDir, "schema.go")
	content := `package schema

import "github.com/lex00/wetwire-neo4j-go/pkg/neo4j/schema"

var Person = schema.NodeType{
	Label: "Person",
}

var Employee = schema.NodeType{
	Label: "Employee",
}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err = l.ListDependencies(tmpDir)
	if err != nil {
		t.Errorf("ListDependencies failed: %v", err)
	}
}

func TestLister_ListDependencies_EmptyDir(t *testing.T) {
	l := NewLister()

	tmpDir, err := os.MkdirTemp("", "lister-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err = l.ListDependencies(tmpDir)
	if err != nil {
		t.Errorf("ListDependencies on empty dir failed: %v", err)
	}
}

func TestShortenPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "relative to cwd",
			input: filepath.Join(cwd, "subdir", "file.go"),
			want:  filepath.Join("subdir", "file.go"),
		},
		{
			name:  "absolute path outside cwd",
			input: "/some/other/path/file.go",
			want:  "/some/other/path/file.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortenPath(tt.input)
			if got != tt.want {
				t.Errorf("shortenPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
