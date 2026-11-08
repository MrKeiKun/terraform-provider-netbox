package netboxapi

import "testing"

func TestNewRejectsURLWithoutScheme(t *testing.T) {
	for _, u := range []string{"netbox.example.com", "localhost:8001", "", "/api"} {
		if _, err := New(u, "token", nil); err == nil {
			t.Errorf("New(%q) accepted a URL without scheme and host", u)
		}
	}
	for _, u := range []string{"http://localhost:8001", "https://netbox.example.com/"} {
		if _, err := New(u, "token", nil); err != nil {
			t.Errorf("New(%q) failed: %v", u, err)
		}
	}
}
