package providerstate

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
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
	// BatchReads reads objects that are asked for at the same time with one list request.
	BatchReads bool
	// UserAgent is sent with every request. Empty leaves the default of the HTTP library.
	UserAgent string
}

// customHeaderTransport is a transport that adds the specified headers on
// every request.
type customHeaderTransport struct {
	original http.RoundTripper
	headers  map[string]interface{}
}

// hasScheme matches an address that starts with a scheme, as in "https://netbox". The port in
// "localhost:8080" is not one.
var hasScheme = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:/`)

// parseServerURL checks the address of the NetBox server. An address without a scheme, for example
// "localhost:8080", is taken as http.
func parseServerURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if !hasScheme.MatchString(raw) {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q, use http or https", u.Scheme)
	}
	if u.Hostname() == "" {
		return nil, errors.New("the URL has no host")
	}
	if u.User != nil {
		return nil, errors.New("the URL must not hold a user name or password, NetBox takes the API token")
	}
	return u, nil
}

// Client creates the client that talks to the Netbox REST API.
func (cfg *Config) Client() (*netboxapi.Client, error) {
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("missing netbox API key")
	}

	parsedURL, err := parseServerURL(cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("error while trying to parse URL: %w", err)
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
	client.SetUserAgent(cfg.UserAgent)
	client.SetBatchReads(cfg.BatchReads)
	return client, nil
}

// maxIdleConnsPerHost is how many open connections to NetBox are kept for the next request. The
// default of the HTTP library is 2, while Terraform works on 10 resources at a time by default, so
// most requests would open a new connection, with a new TLS handshake.
const maxIdleConnsPerHost = 32

// httpClient builds the HTTP client: TLS options, proxy, custom headers and timeout.
func (cfg *Config) httpClient() (*http.Client, error) {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
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
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSClientConfig:       tlsConfig,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}

	var roundTripper http.RoundTripper = transport
	if len(cfg.Headers) > 0 {
		roundTripper = customHeaderTransport{
			original: transport,
			headers:  cfg.Headers,
		}
	}

	return &http.Client{
		Transport:     roundTripper,
		Timeout:       time.Second * time.Duration(cfg.RequestTimeout),
		CheckRedirect: checkRedirect,
	}, nil
}

// checkRedirect follows a redirect only when it stays on the NetBox host and does not go from https
// to http. The API token and the custom headers go out with every request, so following a redirect
// to another host, or to an unencrypted address, would hand them to whoever answers there.
func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	first := via[0].URL
	if !strings.EqualFold(req.URL.Hostname(), first.Hostname()) {
		return fmt.Errorf("NetBox redirected the request to another host (%s). The provider does not follow it, since that would send the API token there. Set server_url to the address NetBox is served from", req.URL.Host)
	}
	if first.Scheme == "https" && req.URL.Scheme != "https" {
		return errors.New("NetBox redirected the request from https to http. The provider does not follow it, since that would send the API token without encryption")
	}
	return nil
}

// RoundTrip adds the headers specified in the transport on every request.
func (t customHeaderTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// A round tripper must not change the request it is given.
	r = r.Clone(r.Context())
	for key, value := range t.headers {
		r.Header.Set(key, fmt.Sprintf("%v", value))
	}

	return t.original.RoundTrip(r)
}
