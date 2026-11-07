// Package netboxapi is a small REST client for the NetBox API.
//
// It does not try to model the whole API. Resources define the fields they
// need and use the generic helpers in this package to talk to NetBox.
package netboxapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// Client sends requests to the NetBox REST API.
type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client

	cfMu    sync.Mutex
	cfTypes map[string]string
}

// New creates a client. baseURL is the NetBox root URL (for example
// https://netbox.example.com), without the /api suffix.
func New(baseURL, token string, httpClient *http.Client) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("missing netbox API key")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error while trying to parse URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid netbox URL %q: scheme and host are required", baseURL)
	}
	u.Path = strings.TrimRight(u.Path, "/")
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: u, token: token, httpClient: httpClient}, nil
}

// authHeader returns the Authorization header value. v2 tokens (nbt_ prefix)
// use the Bearer scheme, legacy tokens use Token.
func (c *Client) authHeader() string {
	if strings.HasPrefix(c.token, "nbt_") {
		return "Bearer " + c.token
	}
	return "Token " + c.token
}

// endpoint builds the full URL for an API path such as "extras/tags/".
func (c *Client) endpoint(path string, query url.Values) string {
	u := *c.baseURL
	u.Path = u.Path + "/api/" + strings.TrimLeft(path, "/")
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	return u.String()
}

// Do sends a request to rawURL. body is JSON encoded when not nil. When out
// is not nil the JSON response is decoded into it.
func (c *Client) Do(ctx context.Context, method, rawURL string, body, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := c.encodeBody(ctx, method, rawURL, body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &APIError{Method: method, URL: rawURL, StatusCode: resp.StatusCode, Body: string(data)}
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decoding response from %s: %w", rawURL, err)
	}
	return nil
}
