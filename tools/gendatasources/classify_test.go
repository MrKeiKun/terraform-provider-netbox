package main

import "testing"

func TestClassifyStringThatMayBeEmpty(t *testing.T) {
	doc := obj{}
	// NetBox describes a color, an email address or a link as a string or an empty string
	union := obj{"oneOf": []any{
		obj{"type": "string", "pattern": "^[0-9a-f]{6}$", "maxLength": float64(6)},
		obj{"type": "string", "maxLength": float64(0)},
	}}
	a, ok := classify(doc, "color", union, "circuit type")
	if !ok || a.kind != "String" {
		t.Errorf("a string or empty string became %q, want String", a.kind)
	}

	// a value of any type stays JSON
	a, ok = classify(doc, "default", obj{"nullable": true}, "custom field")
	if !ok || a.kind != "JSON" {
		t.Errorf("a value of any type became %q, want JSON", a.kind)
	}

	// a union with a member that is not a string stays JSON
	mixed := obj{"oneOf": []any{obj{"type": "string"}, obj{"type": "integer"}}}
	if a, ok := classify(doc, "mixed", mixed, "thing"); !ok || a.kind != "JSON" {
		t.Errorf("a union of a string and a number became %q, want JSON", a.kind)
	}
}
