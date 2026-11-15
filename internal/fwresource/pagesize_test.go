package fwresource

import (
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
)

func TestPageSize(t *testing.T) {
	d := &DataSource{state: &providerstate.State{PageSize: 500}}
	for name, tt := range map[string]struct {
		limit  int64
		client bool
		want   int
	}{
		"no limit":                       {0, false, 500},
		"a limit above the page size":    {900, false, 500},
		"a small limit":                  {5, false, 5},
		"a small limit and client sides": {5, true, 500},
	} {
		if got := d.pageSize(tt.limit, tt.client); got != tt.want {
			t.Errorf("%s: page size = %d, want %d", name, got, tt.want)
		}
	}

	// without a setting the default applies
	d = &DataSource{state: &providerstate.State{}}
	if got := d.pageSize(0, false); got != providerstate.DefaultPageSize {
		t.Errorf("default page size = %d, want %d", got, providerstate.DefaultPageSize)
	}
	d = &DataSource{}
	if got := d.pageSize(0, false); got != providerstate.DefaultPageSize {
		t.Errorf("page size without state = %d, want %d", got, providerstate.DefaultPageSize)
	}
}
