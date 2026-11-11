package providerstate

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
