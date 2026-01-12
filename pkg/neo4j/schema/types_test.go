package schema

import (
	"testing"
	"time"
)

func TestPropertyTypes(t *testing.T) {
	tests := []struct {
		name     string
		propType PropertyType
		want     string
	}{
		{"STRING type", STRING, "STRING"},
		{"INTEGER type", INTEGER, "INTEGER"},
		{"FLOAT type", FLOAT, "FLOAT"},
		{"BOOLEAN type", BOOLEAN, "BOOLEAN"},
		{"DATE type", DATE, "DATE"},
		{"DATETIME type", DATETIME, "DATETIME"},
		{"POINT type", POINT, "POINT"},
		{"LIST_STRING type", LIST_STRING, "LIST_STRING"},
		{"LIST_INTEGER type", LIST_INTEGER, "LIST_INTEGER"},
		{"LIST_FLOAT type", LIST_FLOAT, "LIST_FLOAT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.propType); got != tt.want {
				t.Errorf("PropertyType = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCardinality(t *testing.T) {
	tests := []struct {
		name        string
		cardinality Cardinality
		want        string
	}{
		{"ONE_TO_ONE", ONE_TO_ONE, "ONE_TO_ONE"},
		{"ONE_TO_MANY", ONE_TO_MANY, "ONE_TO_MANY"},
		{"MANY_TO_ONE", MANY_TO_ONE, "MANY_TO_ONE"},
		{"MANY_TO_MANY", MANY_TO_MANY, "MANY_TO_MANY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.cardinality); got != tt.want {
				t.Errorf("Cardinality = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraintType(t *testing.T) {
	tests := []struct {
		name           string
		constraintType ConstraintType
		want           string
	}{
		{"UNIQUE constraint", UNIQUE, "UNIQUE"},
		{"EXISTS constraint", EXISTS, "EXISTS"},
		{"NODE_KEY constraint", NODE_KEY, "NODE_KEY"},
		{"REL_KEY constraint", REL_KEY, "REL_KEY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.constraintType); got != tt.want {
				t.Errorf("ConstraintType = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndexType(t *testing.T) {
	tests := []struct {
		name      string
		indexType IndexType
		want      string
	}{
		{"BTREE index", BTREE, "BTREE"},
		{"TEXT index", TEXT, "TEXT"},
		{"FULLTEXT index", FULLTEXT, "FULLTEXT"},
		{"POINT index", POINT_INDEX, "POINT"},
		{"VECTOR index", VECTOR, "VECTOR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.indexType); got != tt.want {
				t.Errorf("IndexType = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeType(t *testing.T) {
	node := &NodeType{
		Label: "Person",
		Properties: []Property{
			{Name: "name", Type: STRING, Required: true, Description: "The person's full name"},
			{Name: "age", Type: INTEGER, Required: false},
			{Name: "email", Type: STRING, Unique: true},
		},
		Constraints: []Constraint{
			{Name: "person_email_unique", Type: UNIQUE, Properties: []string{"email"}},
		},
		Indexes: []Index{
			{Name: "person_name_idx", Type: BTREE, Properties: []string{"name"}},
		},
		Description: "Represents a person in the system",
	}

	t.Run("ResourceType", func(t *testing.T) {
		if got := node.ResourceType(); got != "NodeType" {
			t.Errorf("NodeType.ResourceType() = %v, want NodeType", got)
		}
	})

	t.Run("ResourceName", func(t *testing.T) {
		if got := node.ResourceName(); got != "Person" {
			t.Errorf("NodeType.ResourceName() = %v, want Person", got)
		}
	})

	t.Run("Properties count", func(t *testing.T) {
		if len(node.Properties) != 3 {
			t.Errorf("Properties count = %v, want 3", len(node.Properties))
		}
	})

	t.Run("Required property", func(t *testing.T) {
		if !node.Properties[0].Required {
			t.Error("name property should be required")
		}
	})

	t.Run("Unique property", func(t *testing.T) {
		if !node.Properties[2].Unique {
			t.Error("email property should be unique")
		}
	})
}

func TestRelationshipType(t *testing.T) {
	rel := &RelationshipType{
		Label:       "WORKS_FOR",
		Source:      "Person",
		Target:      "Company",
		Cardinality: MANY_TO_ONE,
		Properties: []Property{
			{Name: "since", Type: DATE, Required: true},
			{Name: "role", Type: STRING},
		},
		Description: "Employment relationship",
	}

	t.Run("ResourceType", func(t *testing.T) {
		if got := rel.ResourceType(); got != "RelationshipType" {
			t.Errorf("RelationshipType.ResourceType() = %v, want RelationshipType", got)
		}
	})

	t.Run("ResourceName", func(t *testing.T) {
		if got := rel.ResourceName(); got != "WORKS_FOR" {
			t.Errorf("RelationshipType.ResourceName() = %v, want WORKS_FOR", got)
		}
	})

	t.Run("Source and Target", func(t *testing.T) {
		if rel.Source != "Person" {
			t.Errorf("Source = %v, want Person", rel.Source)
		}
		if rel.Target != "Company" {
			t.Errorf("Target = %v, want Company", rel.Target)
		}
	})

	t.Run("Cardinality", func(t *testing.T) {
		if rel.Cardinality != MANY_TO_ONE {
			t.Errorf("Cardinality = %v, want MANY_TO_ONE", rel.Cardinality)
		}
	})
}

func TestProperty(t *testing.T) {
	prop := Property{
		Name:         "score",
		Type:         FLOAT,
		Required:     true,
		Unique:       false,
		Description:  "User score",
		DefaultValue: 0.0,
	}

	if prop.Name != "score" {
		t.Errorf("Name = %v, want score", prop.Name)
	}
	if prop.Type != FLOAT {
		t.Errorf("Type = %v, want FLOAT", prop.Type)
	}
	if !prop.Required {
		t.Error("Required should be true")
	}
	if prop.Unique {
		t.Error("Unique should be false")
	}
}

func TestConstraint(t *testing.T) {
	constraint := Constraint{
		Name:       "person_key",
		Type:       NODE_KEY,
		Properties: []string{"name", "email"},
	}

	if constraint.Name != "person_key" {
		t.Errorf("Name = %v, want person_key", constraint.Name)
	}
	if constraint.Type != NODE_KEY {
		t.Errorf("Type = %v, want NODE_KEY", constraint.Type)
	}
	if len(constraint.Properties) != 2 {
		t.Errorf("Properties count = %v, want 2", len(constraint.Properties))
	}
}

func TestIndex(t *testing.T) {
	idx := Index{
		Name:       "embedding_idx",
		Type:       VECTOR,
		Properties: []string{"embedding"},
		Options: map[string]any{
			"dimensions":          384,
			"similarity_function": "cosine",
		},
	}

	if idx.Name != "embedding_idx" {
		t.Errorf("Name = %v, want embedding_idx", idx.Name)
	}
	if idx.Type != VECTOR {
		t.Errorf("Type = %v, want VECTOR", idx.Type)
	}
	if idx.Options["dimensions"] != 384 {
		t.Errorf("dimensions = %v, want 384", idx.Options["dimensions"])
	}
}

func TestPoint(t *testing.T) {
	z := 100.0
	point := Point{
		SRID: 4326,
		X:    -122.4194,
		Y:    37.7749,
		Z:    &z,
	}

	if point.SRID != 4326 {
		t.Errorf("SRID = %v, want 4326", point.SRID)
	}
	if point.X != -122.4194 {
		t.Errorf("X = %v, want -122.4194", point.X)
	}
	if point.Y != 37.7749 {
		t.Errorf("Y = %v, want 37.7749", point.Y)
	}
	if *point.Z != 100.0 {
		t.Errorf("Z = %v, want 100.0", *point.Z)
	}
}

func TestPropertyValueConstructors(t *testing.T) {
	t.Run("NewStringValue", func(t *testing.T) {
		pv := NewStringValue("hello")
		if pv.String == nil || *pv.String != "hello" {
			t.Error("String value not set correctly")
		}
	})

	t.Run("NewIntegerValue", func(t *testing.T) {
		pv := NewIntegerValue(42)
		if pv.Integer == nil || *pv.Integer != 42 {
			t.Error("Integer value not set correctly")
		}
	})

	t.Run("NewFloatValue", func(t *testing.T) {
		pv := NewFloatValue(3.14)
		if pv.Float == nil || *pv.Float != 3.14 {
			t.Error("Float value not set correctly")
		}
	})

	t.Run("NewBooleanValue", func(t *testing.T) {
		pv := NewBooleanValue(true)
		if pv.Boolean == nil || *pv.Boolean != true {
			t.Error("Boolean value not set correctly")
		}
	})

	t.Run("NewDateValue", func(t *testing.T) {
		now := time.Now()
		pv := NewDateValue(now)
		if pv.Date == nil || !pv.Date.Equal(now) {
			t.Error("Date value not set correctly")
		}
	})

	t.Run("NewDateTimeValue", func(t *testing.T) {
		now := time.Now()
		pv := NewDateTimeValue(now)
		if pv.DateTime == nil || !pv.DateTime.Equal(now) {
			t.Error("DateTime value not set correctly")
		}
	})

	t.Run("NewPointValue", func(t *testing.T) {
		p := Point{SRID: 4326, X: 1.0, Y: 2.0}
		pv := NewPointValue(p)
		if pv.Point == nil || pv.Point.X != 1.0 {
			t.Error("Point value not set correctly")
		}
	})

	t.Run("NewListStringValue", func(t *testing.T) {
		pv := NewListStringValue([]string{"a", "b"})
		if len(pv.ListString) != 2 || pv.ListString[0] != "a" {
			t.Error("ListString value not set correctly")
		}
	})

	t.Run("NewListIntegerValue", func(t *testing.T) {
		pv := NewListIntegerValue([]int64{1, 2, 3})
		if len(pv.ListInteger) != 3 || pv.ListInteger[0] != 1 {
			t.Error("ListInteger value not set correctly")
		}
	})

	t.Run("NewListFloatValue", func(t *testing.T) {
		pv := NewListFloatValue([]float64{1.1, 2.2})
		if len(pv.ListFloat) != 2 || pv.ListFloat[0] != 1.1 {
			t.Error("ListFloat value not set correctly")
		}
	})
}

func TestResourceInterface(t *testing.T) {
	// Verify that NodeType and RelationshipType implement Resource interface
	var _ Resource = &NodeType{}
	var _ Resource = &RelationshipType{}
}
