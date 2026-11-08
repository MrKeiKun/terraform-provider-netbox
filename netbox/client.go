package netbox

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/goware/urlx"
	log "github.com/sirupsen/logrus"
)

// Config struct for the netbox provider
type Config struct {
	APIToken                    string
	ServerURL                   string
	AllowInsecureHTTPS          bool
	Headers                     map[string]interface{}
	RequestTimeout              int
	MaxRetries                  int
	StripTrailingSlashesFromURL bool
	CACertFile                  string
}

// customHeaderTransport is a transport that adds the specified headers on
// every request.
type customHeaderTransport struct {
	original http.RoundTripper
	headers  map[string]interface{}
}

// Client creates the client that talks to the Netbox REST API.
func (cfg *Config) Client() (*netboxapi.Client, error) {
	log.WithFields(log.Fields{
		"server_url": cfg.ServerURL,
	}).Debug("Initializing Netbox client")

	if cfg.APIToken == "" {
		return nil, fmt.Errorf("missing netbox API key")
	}

	// urlx accepts URLs without a scheme, for example "localhost:8080"
	parsedURL, err := urlx.Parse(cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("error while trying to parse URL: %s", err)
	}

	httpClient, err := cfg.httpClient()
	if err != nil {
		return nil, err
	}

	client, err := netboxapi.New(parsedURL.String(), cfg.APIToken, httpClient)
	if err != nil {
		return nil, err
	}
	client.SetMaxRetries(cfg.MaxRetries)
	return client, nil
}

// httpClient builds the HTTP client: TLS options, proxy, custom headers and timeout.
func (cfg *Config) httpClient() (*http.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.AllowInsecureHTTPS, //nolint:gosec // the user asked for it with allow_insecure_https
	}

	if cfg.CACertFile != "" {
		pem, err := os.ReadFile(cfg.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("could not read the CA certificate file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("no certificate found in the CA certificate file %q", cfg.CACertFile)
		}
		tlsConfig.RootCAs = pool
	}

	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}

	var roundTripper http.RoundTripper = transport
	if len(cfg.Headers) > 0 {
		log.WithFields(log.Fields{
			"custom_headers": cfg.Headers,
		}).Debug("Setting custom headers on every request to Netbox")

		roundTripper = customHeaderTransport{
			original: transport,
			headers:  cfg.Headers,
		}
	}

	return &http.Client{
		Transport: roundTripper,
		Timeout:   time.Second * time.Duration(cfg.RequestTimeout),
	}, nil
}

// RoundTrip adds the headers specified in the transport on every request.
func (t customHeaderTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	for key, value := range t.headers {
		r.Header.Add(key, fmt.Sprintf("%v", value))
	}

	resp, err := t.original.RoundTrip(r)
	return resp, err
}
