package providerstate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
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

// A wrong token is reported at once, not after several retries.
func TestConnectDoesNotRetryAWrongToken(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	start := time.Now()
	state, diags := Connect(context.Background(), Settings{ServerURL: ts.URL, APIToken: "0123456789abcdef0123456789abcdef01234567", RequestTimeout: 10, MaxRetries: 3})
	if state != nil || len(diags) == 0 {
		t.Fatal("a wrong token was accepted")
	}
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Errorf("the status endpoint was called %d times, want 1", n)
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("took %s", time.Since(start))
	}
}

func statusServer(t *testing.T, version string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/status/") {
			_, _ = w.Write([]byte(`{"netbox-version":"` + version + `"}`))
			return
		}
		// two tags with the same name, to make the name ambiguous
		_, _ = w.Write([]byte(`{"count":2,"next":null,"results":[{"name":"dup","slug":"dup"},{"name":"dup","slug":"dup2"}]}`))
	}))
	t.Cleanup(ts.Close)
	return ts
}

func connect(t *testing.T, s Settings) (*State, []Diag) {
	t.Helper()
	s.APIToken = "0123456789abcdef0123456789abcdef01234567"
	s.RequestTimeout = 10
	return Connect(context.Background(), s)
}

func hasWarning(diags []Diag, summary string) bool {
	for _, d := range diags {
		if !d.Error && strings.Contains(d.Summary, summary) {
			return true
		}
	}
	return false
}

func TestConnectStripsTrailingSlashes(t *testing.T) {
	ts := statusServer(t, "4.4.0")
	state, diags := connect(t, Settings{ServerURL: ts.URL + "//", StripTrailingSlashesFromURL: true})
	assert.NotNil(t, state)
	assert.True(t, hasWarning(diags, "Stripped trailing slashes"), "%v", diags)
}

func TestConnectVersionProblems(t *testing.T) {
	ts := statusServer(t, "dev")
	state, diags := connect(t, Settings{ServerURL: ts.URL})
	assert.Nil(t, state)
	assert.NotEmpty(t, diags)

	ts = statusServer(t, "9.9.9")
	state, diags = connect(t, Settings{ServerURL: ts.URL})
	assert.NotNil(t, state, "an unsupported version is only a warning")
	assert.True(t, hasWarning(diags, "Possibly unsupported"), "%v", diags)
}

func TestConnectRejectsBadSettings(t *testing.T) {
	ts := statusServer(t, "4.4.0")
	state, diags := connect(t, Settings{ServerURL: ts.URL, CACertFile: filepath.Join(t.TempDir(), "missing.pem")})
	assert.Nil(t, state)
	assert.NotEmpty(t, diags)

	garbage := filepath.Join(t.TempDir(), "garbage.pem")
	assert.NoError(t, os.WriteFile(garbage, []byte("not a certificate"), 0o600))
	state, diags = connect(t, Settings{ServerURL: ts.URL, CACertFile: garbage})
	assert.Nil(t, state)
	assert.NotEmpty(t, diags)

	state, diags = connect(t, Settings{ServerURL: "http://[::1"})
	assert.Nil(t, state)
	assert.NotEmpty(t, diags)
}

func TestTagLookupErrors(t *testing.T) {
	ts := statusServer(t, "4.4.0")
	state, _ := connect(t, Settings{ServerURL: ts.URL})
	if state == nil {
		t.Fatal("no state")
	}
	_, err := state.TagRef(context.Background(), "dup")
	assert.Error(t, err, "an ambiguous tag name was accepted")

	gone := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	gone.Close()
	state.Rest, _ = netboxapi.New(gone.URL, "token", nil)
	state.Rest.SetMaxRetries(0)
	_, err = state.TagRef(context.Background(), "x")
	assert.Error(t, err, "an unreachable NetBox was accepted")
}
