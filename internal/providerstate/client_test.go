package providerstate

import (
	"context"
	"encoding/pem"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/stretchr/testify/assert"
)

func TestValidClientWithAllData(t *testing.T) {
	config := Config{
		APIToken:  "07b12b765127747e4afd56cb531b7bf9c61f3c30",
		ServerURL: "https://localhost:8080",
	}

	client, err := config.Client()
	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestURLMissingSchemaShouldWork(t *testing.T) {
	config := Config{
		APIToken:  "07b12b765127747e4afd56cb531b7bf9c61f3c30",
		ServerURL: "localhost:8080",
	}

	client, err := config.Client()
	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestURLMaleformedUrlShouldFail(t *testing.T) {
	config := Config{
		APIToken:  "07b12b765127747e4afd56cb531b7bf9c61f3c30",
		ServerURL: "xyz:/localhost:8080",
	}

	_, err := config.Client()
	assert.Error(t, err)
}

func TestURLMissingPortShouldWork(t *testing.T) {
	config := Config{
		APIToken:  "07b12b765127747e4afd56cb531b7bf9c61f3c30",
		ServerURL: "http://localhost",
	}

	client, err := config.Client()
	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestURLMissingAccessKey(t *testing.T) {
	config := Config{
		APIToken:  "",
		ServerURL: "http://localhost",
	}

	_, err := config.Client()
	assert.Error(t, err)
}

func TestAdditionalHeadersSet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vals, ok := r.Header["Hello"]

		assert.True(t, ok)
		assert.Len(t, vals, 1)
		assert.Equal(t, vals[0], "World!")
	}))
	defer ts.Close()

	config := Config{
		APIToken:  "07b12b765127747e4afd56cb531b7bf9c61f3c30",
		ServerURL: ts.URL,
		Headers: map[string]interface{}{
			"Hello": "World!",
		},
	}

	client, err := config.Client()
	assert.NoError(t, err)

	_, _ = netboxapi.Request[map[string]interface{}](context.Background(), client, http.MethodGet, "status/", nil, nil)
}

func TestV1TokenUsesTokenScheme(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Token 07b12b765127747e4afd56cb531b7bf9c61f3c30", auth)
	}))
	defer ts.Close()

	config := Config{
		APIToken:  "07b12b765127747e4afd56cb531b7bf9c61f3c30",
		ServerURL: ts.URL,
	}

	client, err := config.Client()
	assert.NoError(t, err)

	_, _ = netboxapi.Request[map[string]interface{}](context.Background(), client, http.MethodGet, "status/", nil, nil)
}

func TestV2TokenUsesBearerScheme(t *testing.T) {
	v2Token := "nbt_abc1234567890abcdef.checksum1234"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+v2Token, auth)
	}))
	defer ts.Close()

	config := Config{
		APIToken:  v2Token,
		ServerURL: ts.URL,
	}

	client, err := config.Client()
	assert.NoError(t, err)

	_, _ = netboxapi.Request[map[string]interface{}](context.Background(), client, http.MethodGet, "status/", nil, nil)
}

func TestRequestTimeoutIsUsed(t *testing.T) {
	config := Config{
		APIToken:       "07b12b765127747e4afd56cb531b7bf9c61f3c30",
		ServerURL:      "http://localhost",
		RequestTimeout: 90,
	}

	httpClient, err := config.httpClient()
	assert.NoError(t, err)
	assert.Equal(t, 90*time.Second, httpClient.Timeout)
}

// newTLSTestServer starts a server with a self-signed certificate that answers
// every request with an empty JSON object.
func newTLSTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}))
	t.Cleanup(ts.Close)

	return ts
}

func requestStatus(t *testing.T, config Config) error {
	t.Helper()

	client, err := config.Client()
	assert.NoError(t, err)

	_, err = netboxapi.Request[map[string]interface{}](context.Background(), client, http.MethodGet, "status/", nil, nil)
	return err
}

func TestInvalidHttpsCertificate(t *testing.T) {
	ts := newTLSTestServer(t)

	err := requestStatus(t, Config{
		APIToken:  "07b12b765127747e4afd56cb531b7bf9c61f3c30",
		ServerURL: ts.URL,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "certificate")
}

func TestAllowInsecureHTTPSAcceptsInvalidCertificate(t *testing.T) {
	ts := newTLSTestServer(t)

	err := requestStatus(t, Config{
		APIToken:           "07b12b765127747e4afd56cb531b7bf9c61f3c30",
		ServerURL:          ts.URL,
		AllowInsecureHTTPS: true,
	})

	assert.NoError(t, err)
}

func TestCACertFileTrustsServerCertificate(t *testing.T) {
	ts := newTLSTestServer(t)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ts.Certificate().Raw})
	caFile := filepath.Join(t.TempDir(), "ca.pem")
	assert.NoError(t, os.WriteFile(caFile, certPEM, 0o600))

	err := requestStatus(t, Config{
		APIToken:   "07b12b765127747e4afd56cb531b7bf9c61f3c30",
		ServerURL:  ts.URL,
		CACertFile: caFile,
	})

	assert.NoError(t, err)
}

// The transport must not change the request it gets, and a configured header replaces the old one.
func TestCustomHeaderTransportClonesRequest(t *testing.T) {
	var seen http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = r.Header.Clone() }))
	defer ts.Close()

	rt := customHeaderTransport{original: http.DefaultTransport, headers: map[string]interface{}{"X-Proxy": "new"}}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL, nil)
	req.Header.Set("X-Proxy", "old")
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if got := seen.Values("X-Proxy"); len(got) != 1 || got[0] != "new" {
		t.Errorf("X-Proxy = %v", got)
	}
	if req.Header.Get("X-Proxy") != "old" {
		t.Error("the caller's request was changed")
	}
}

func TestParseServerURL(t *testing.T) {
	for raw, want := range map[string]string{
		"localhost:8080":             "http://localhost:8080",
		" https://netbox.example ":   "https://netbox.example",
		"netbox.example/netbox":      "http://netbox.example/netbox",
		"http://[::1]:8000":          "http://[::1]:8000",
		"ftp://netbox.example":       "",
		"xyz:/localhost:8080":        "",
		"https://":                   "",
		"https://user:pw@netbox.com": "",
	} {
		u, err := parseServerURL(raw)
		if want == "" {
			assert.Error(t, err, raw)
			continue
		}
		if assert.NoError(t, err, raw) {
			assert.Equal(t, want, u.String(), raw)
		}
	}
}

// The API token must not follow a redirect to another host or from https to http.
func TestRedirectsThatLeakTheTokenAreRefused(t *testing.T) {
	var leaked atomic.Bool
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" || r.Header.Get("X-Secret") != "" {
			leaked.Store(true)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer other.Close()
	// 127.0.0.1 and localhost are the same server but not the same host name.
	elsewhere := strings.Replace(other.URL, "127.0.0.1", "localhost", 1)

	var target string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/moved/" {
			_, _ = w.Write([]byte(`{"ok": true}`))
			return
		}
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
	}))
	defer srv.Close()

	config := Config{APIToken: "secret", ServerURL: srv.URL, RequestTimeout: 5, Headers: map[string]interface{}{"X-Secret": "s"}}
	client, err := config.Client()
	assert.NoError(t, err)

	target = elsewhere + "/api/status/"
	_, err = netboxapi.Request[map[string]any](context.Background(), client, http.MethodGet, "status/", nil, nil)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "another host")
	}
	assert.False(t, leaked.Load(), "the token or a custom header was sent to another host")

	// A redirect that stays on the host is followed.
	target = srv.URL + "/api/moved/"
	out, err := netboxapi.Request[map[string]any](context.Background(), client, http.MethodGet, "status/", nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, true, (*out)["ok"])
	}
}

func TestCheckRedirectRefusesDowngrade(t *testing.T) {
	req := func(raw string) *http.Request {
		r, err := http.NewRequestWithContext(context.Background(), http.MethodGet, raw, nil)
		assert.NoError(t, err)
		return r
	}
	first := []*http.Request{req("https://netbox.example/api/status/")}
	assert.NoError(t, checkRedirect(req("https://NETBOX.example:8443/api/status/"), first))
	assert.ErrorContains(t, checkRedirect(req("http://netbox.example/api/status/"), first), "https to http")
	assert.NoError(t, checkRedirect(req("https://netbox.example/api/status/"), []*http.Request{req("http://netbox.example/api/status/")}))
	assert.ErrorContains(t, checkRedirect(req("https://netbox.example/"), make([]*http.Request, 10)), "10 redirects")
}

// Terraform works on many resources at once, their requests must share connections.
func TestConnectionsAreReused(t *testing.T) {
	var conns atomic.Int32
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			conns.Add(1)
		}
	}
	srv.Start()
	defer srv.Close()

	config := Config{APIToken: "t", ServerURL: srv.URL, RequestTimeout: 5}
	client, err := config.Client()
	assert.NoError(t, err)

	const parallel = 10
	for round := 0; round < 5; round++ {
		var wg sync.WaitGroup
		for i := 0; i < parallel; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := netboxapi.Request[map[string]any](context.Background(), client, http.MethodGet, "status/", nil, nil)
				assert.NoError(t, err)
			}()
		}
		wg.Wait()
	}
	assert.LessOrEqual(t, conns.Load(), int32(parallel), "requests opened new connections instead of reusing idle ones")
}

func TestIsLoopback(t *testing.T) {
	for host, want := range map[string]bool{"localhost": true, "127.0.0.1": true, "::1": true, "netbox": false, "10.0.0.1": false} {
		assert.Equal(t, want, isLoopback(host), host)
	}
}

// A web server that sends http to https answers a POST with a 301, which turns it into a GET. The
// provider must not follow that: nothing would be created.
func TestRedirectThatTurnsAWriteIntoAReadIsRefused(t *testing.T) {
	var posts, gets atomic.Int32
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posts.Add(1)
		} else {
			gets.Add(1)
		}
		_, _ = w.Write([]byte(`{"count": 0, "results": []}`))
	}))
	defer final.Close()
	var hits atomic.Int32
	front := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Redirect(w, r, final.URL+r.URL.RequestURI(), http.StatusMovedPermanently)
	}))
	defer front.Close()

	config := Config{APIToken: "t", ServerURL: front.URL, RequestTimeout: 5, MaxRetries: 3}
	client, err := config.Client()
	assert.NoError(t, err)

	_, err = netboxapi.Create[map[string]any](context.Background(), client, "tenancy/tenants/", map[string]any{"name": "a"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Set server_url to the address NetBox redirects to, probably "+final.URL)
		assert.False(t, netboxapi.MaybeProcessed(err), "nothing was written, the error must not say the object may exist")
	}
	assert.Equal(t, int32(1), hits.Load(), "a refused redirect was tried again")
	assert.Equal(t, int32(0), posts.Load()+gets.Load(), "the redirect was followed")

	// A read is still followed, it stays a read.
	_, err = netboxapi.ListPage[map[string]any](context.Background(), client, "tenancy/tenants/", nil)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), gets.Load())
}

// 307 and 308 keep the method and the body, so a write can follow them.
func TestRedirectThatKeepsTheMethodIsFollowed(t *testing.T) {
	var body string
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = r.Method + " " + string(b)
		_, _ = w.Write([]byte(`{"id": 1}`))
	}))
	defer final.Close()
	front := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL+r.URL.RequestURI(), http.StatusPermanentRedirect)
	}))
	defer front.Close()

	config := Config{APIToken: "t", ServerURL: front.URL, RequestTimeout: 5}
	client, err := config.Client()
	assert.NoError(t, err)
	_, err = netboxapi.Create[map[string]any](context.Background(), client, "tenancy/tenants/", map[string]any{"name": "a"})
	assert.NoError(t, err)
	assert.Equal(t, `POST {"name":"a"}`, body)
}
