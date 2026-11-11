// Package netboxapi is a small REST client for the NetBox API.
//
// It does not try to model the whole API. Resources define the fields they
// need and use the generic helpers in this package to talk to NetBox.
package netboxapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client sends requests to the NetBox REST API.
type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client

	maxRetries int

	cfMu    sync.Mutex
	cfTypes map[string]string
}

// SetMaxRetries sets how often a failed idempotent request is sent again.
// Zero turns retries off.
func (c *Client) SetMaxRetries(n int) {
	if n < 0 {
		n = 0
	}
	c.maxRetries = n
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
	return &Client{baseURL: u, token: token, httpClient: httpClient, maxRetries: DefaultMaxRetries}, nil
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

// DefaultMaxRetries is how often an idempotent request is sent again when the
// server or the connection fails temporarily.
const DefaultMaxRetries = 3

// retryDelay is the base wait between attempts. It grows with every attempt.
var retryDelay = time.Second

// idempotent reports whether a request can safely be sent again after a
// temporary failure. POST and PATCH are left out, since a lost response could
// create a duplicate. A database deadlock is the exception, see isDeadlock.
func idempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete:
		return true
	}
	return false
}

// isDeadlock reports whether NetBox answered with a database deadlock. Postgres
// rolls the transaction back in that case, so the request changed nothing and
// can be sent again. NetBox reports it as a 500 error.
func isDeadlock(status int, body []byte) bool {
	return status == http.StatusInternalServerError &&
		bytes.Contains(bytes.ToLower(body), []byte("deadlock detected"))
}

// isDeadlockError reports whether err is the deadlock answer from NetBox.
func isDeadlockError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && isDeadlock(apiErr.StatusCode, []byte(apiErr.Body))
}

// Do sends a request to rawURL. body is JSON encoded when not nil. When out
// is not nil the JSON response is decoded into it. Idempotent requests are
// sent again (see SetMaxRetries) when the connection fails or NetBox (or a proxy
// in front of it) answers 502, 503 or 504. Any request is sent again when the
// NetBox database reports a deadlock.
func (c *Client) Do(ctx context.Context, method, rawURL string, body, out any) error {
	var buf []byte
	if body != nil {
		var err error
		buf, err = c.encodeBody(ctx, method, rawURL, body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
	}

	attempts := 1 + c.maxRetries
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		var retry bool
		retry, err = c.doOnce(ctx, method, rawURL, buf, out)
		if err == nil || !retry || attempt == attempts {
			return err
		}
		if !idempotent(method) && !isDeadlockError(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return err
		case <-time.After(time.Duration(attempt) * retryDelay):
		}
	}
	return err
}

// doOnce sends the request a single time. The result tells whether a failure
// is temporary.
func (c *Client) doOnce(ctx context.Context, method, rawURL string, buf []byte, out any) (retry bool, err error) {
	var reader io.Reader
	if buf != nil {
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, reader)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Accept", "application/json")
	if buf != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ctx.Err() == nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx.Err() == nil, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		temporary := resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout ||
			isDeadlock(resp.StatusCode, data)
		return temporary, &APIError{Method: method, URL: rawURL, StatusCode: resp.StatusCode, Body: string(data)}
	}
	if out == nil || len(data) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return false, fmt.Errorf("decoding response from %s: %w", rawURL, err)
	}
	return false, nil
}
