package fwresource

import (
	"strings"
	"testing"
)

func TestFilterParam(t *testing.T) {
	p := &Plural{
		Filters: map[string]string{"name": "name", "tags": "tag[]", "status": "status", "weird": "weird"},
		Lookups: map[string][]string{"name": {"ic", "isw", "n"}, "tag": {"n"}, "status": {"n"}},
	}
	for name, want := range map[string]string{
		"name":      "name",
		"tags":      "tag[]",
		"name__ic":  "name__ic",
		"name__isw": "name__isw",
		// the lookup goes onto the query parameter, not onto the name of the filter
		"tags__n":   "tag__n[]",
		"status__n": "status__n",
	} {
		got, err := p.filterParam(name)
		if err != nil || got != want {
			t.Errorf("%s: got %q, %v, want %q", name, got, err, want)
		}
	}
	// NetBox ignores a parameter it does not know, so these must never be sent.
	for name, want := range map[string]string{
		"bogus":       "not a supported filter",
		"bogus__ic":   "not a supported filter",
		"name__bogus": "has the lookup expressions ic, isw, n",
		"status__ic":  "has the lookup expressions n",
		"weird__n":    "has no lookup expressions",
		"name__":      "has the lookup expressions",
	} {
		if _, err := p.filterParam(name); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v, want an error with %q", name, err, want)
		}
	}
}

func TestFilterSuggestion(t *testing.T) {
	p := &Plural{Filters: map[string]string{"site_id": "site_id", "tags": "tag[]", "name": "name"}}
	for name, want := range map[string]string{
		"site": "Did you mean 'site_id'?",
		"tag":  "Did you mean 'tags'?",
	} {
		if _, err := p.filterParam(name); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v, want %q", name, err, want)
		}
	}
	if _, err := p.filterParam("bogus"); err == nil || strings.Contains(err.Error(), "Did you mean") {
		t.Errorf("a name without a similar filter got a suggestion: %v", err)
	}
}
