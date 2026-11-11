package providerstate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConnect_nonJSONStatusResponse verifies that when the Netbox version check (GET /api/status/)
// gets a non-JSON response - for example from a misconfigured server_url or a proxy/ingress
// returning an HTML error page - Connect returns a clear, actionable diagnostic instead of the raw
// JSON decoding error. See
// https://github.com/MrKeiKun/terraform-provider-netbox/issues/263,
// https://github.com/MrKeiKun/terraform-provider-netbox/issues/606, and
// others reporting this same unhelpful error message.
func TestConnect_nonJSONStatusResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>not json</body></html>"))
	}))
	defer ts.Close()

	settings := Settings{}
	settings.ServerURL = ts.URL
	settings.APIToken = "0123456789abcdef0123456789abcdef01234567"
	settings.RequestTimeout = 10

	state, diags := Connect(context.Background(), settings)

	assert.Nil(t, state)
	found := false
	for _, d := range diags {
		if !d.Error {
			continue
		}
		found = true
		assert.Equal(t, "Failed to determine Netbox version", d.Summary)
		assert.Contains(t, d.Detail, "skip_version_check")
		assert.Contains(t, d.Detail, ts.URL+"/api/status/")
		assert.True(t, strings.Contains(d.Detail, "Original error:"),
			"expected the original error to still be included for debugging, prefixed by an explanation")
	}
	assert.True(t, found, "expected an error diagnostic")
}
