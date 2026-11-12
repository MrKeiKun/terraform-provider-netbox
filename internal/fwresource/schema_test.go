package fwresource

import (
	"context"
	"testing"
)

// Every option of an attribute gives a schema that the framework accepts.
func TestSchemaOptions(t *testing.T) {
	min, max := int64(1), int64(9)
	spec := &Spec{
		TypeName: "x",
		Attrs: []Attr{
			{Name: "s1", Kind: String, Optional: true, ForceNew: true, Regex: "^a", MaxLen: 5},
			{Name: "s2", Kind: String, Optional: true, MinLen: 2},
			{Name: "i1", Kind: Int, Optional: true, ForceNew: true, Max: &max},
			{Name: "i2", Kind: Int, Optional: true, Min: &min},
			{Name: "i3", Kind: Int, Optional: true, Min: &min, Max: &max},
			{Name: "f1", Kind: Float, Optional: true, ForceNew: true, Min: &min, Max: &max},
			{Name: "f2", Kind: Float, Optional: true, Min: &min},
			{Name: "f3", Kind: Float, Optional: true, Max: &max},
			{Name: "b1", Kind: Bool, Optional: true, ForceNew: true},
			{Name: "set1", Kind: StringSet, Optional: true, ForceNew: true, ElemOneOf: []string{"a", "b"}},
			{Name: "list1", Kind: StringList, Optional: true, ForceNew: true},
			{Name: "nested", Kind: NestedList, Required: true, Sub: []Attr{{Name: "id", Kind: Int, Computed: true}}},
		},
	}
	s := spec.buildSchema()
	if diags := s.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("invalid schema: %v", diags)
	}
	if _, ok := s.Blocks["nested"]; !ok {
		t.Error("a required nested list is not a block")
	}
}

func TestDataSourceSchemaOptions(t *testing.T) {
	min, max := int64(1), int64(9)
	spec := &DataSourceSpec{
		TypeName: "x",
		Attrs: []Attr{
			{Name: "i1", Kind: Int, Optional: true, Min: &min, Max: &max},
			{Name: "i2", Kind: Int, Optional: true, Min: &min},
			{Name: "s1", Kind: String, Optional: true, OneOf: []string{"a", "b"}},
			{Name: "s2", Kind: String, Optional: true},
		},
		ExactlyOneOf: [][]string{{"i1", "s1"}},
	}
	s := spec.buildSchema()
	if diags := s.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("invalid schema: %v", diags)
	}
	if got := (&DataSource{spec: spec}).ConfigValidators(context.Background()); len(got) != 1 {
		t.Errorf("validators = %d, want 1", len(got))
	}
}

func TestScopeAttrLookup(t *testing.T) {
	s := &Spec{Attrs: []Attr{{Name: "a"}}}
	if _, ok := s.attr("a"); !ok {
		t.Error("attribute a not found")
	}
	if _, ok := s.attr("missing"); ok {
		t.Error("a missing attribute was found")
	}
}
