package fwresource

import (
	"testing"
)

func TestOmitQuery(t *testing.T) {
	if q := omitQuery("dcim/sites/", nil); q != nil {
		t.Fatalf("sites: got %v, want no parameters", q)
	}
	q := omitQuery("dcim/devices/", nil)
	if q.Get("omit") != "config_context" || q.Get("exclude") != "config_context" {
		t.Fatalf("devices: got %v", q)
	}
	if q := omitQuery("dcim/devices/", func(string) bool { return true }); q != nil {
		t.Fatalf("a field that is read must stay: got %v", q)
	}
}

func TestReadQueryKeepsFieldsOfTheSpec(t *testing.T) {
	plain := &Spec{Path: "virtualization/virtual-machines/", Attrs: []Attr{{Name: "name"}}}
	if q := plain.readQuery(); q.Get("omit") != "config_context" {
		t.Fatalf("got %v", q)
	}
	for _, a := range []Attr{
		{Name: "config_context", Kind: JSON, Computed: true},
		{Name: "context", Key: "config_context"},
		{Name: "context_value", Path: []string{"config_context", "value"}},
	} {
		s := &Spec{Path: "virtualization/virtual-machines/", Attrs: []Attr{{Name: "name"}, a}}
		if q := s.readQuery(); q != nil {
			t.Fatalf("attribute %s reads the config context: got %v", a.Name, q)
		}
	}
}
