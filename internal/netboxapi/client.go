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
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
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
	userAgent  string

	cfMu    sync.Mutex
	cfTypes map[string]string

	// see batch.go
	batchReads bool
	batchMu    sync.Mutex
	batches    map[string]*batchGroup
}

// SetMaxRetries sets how often a failed idempotent request is sent again.
// Zero turns retries off.
func (c *Client) SetMaxRetries(n int) {
	if n < 0 {
		n = 0
	}
	c.maxRetries = n
}

// SetUserAgent sets the User-Agent header of every request, so that the administrator of NetBox can
// tell the requests of the provider apart in the logs.
func (c *Client) SetUserAgent(ua string) {
	c.userAgent = ua
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

// retryDelay is the wait before the second attempt. It doubles with every further attempt.
var retryDelay = time.Second

// maxRetryDelay is the longest wait between two attempts, also when NetBox asks for more with a
// Retry-After header.
const maxRetryDelay = 30 * time.Second

// backoff returns how long to wait after the given failed attempt (1 is the first). The wait doubles
// with every attempt and is spread by up to a quarter in both directions, so that the requests
// Terraform runs side by side do not all come back at the same moment. A wait the server asked for
// with Retry-After is used as it is.
func backoff(attempt int, err error) time.Duration {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.RetryAfter > 0 {
		return min(apiErr.RetryAfter, maxRetryDelay)
	}
	d := retryDelay
	for i := 1; i < attempt && d < maxRetryDelay; i++ {
		d *= 2
	}
	d = min(d, maxRetryDelay)
	if quarter := int64(d / 4); quarter > 0 {
		d += time.Duration(rand.Int64N(2*quarter+1) - quarter) //nolint:gosec // spreading out retries needs no secure random numbers
	}
	return d
}

// retryAfter reads the Retry-After header, which holds a number of seconds or a date.
func retryAfter(h http.Header) time.Duration {
	v := h.Get("Retry-After")
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return max(time.Duration(secs)*time.Second, 0)
	}
	if at, err := http.ParseTime(v); err == nil {
		return max(time.Until(at), 0)
	}
	return 0
}

// idempotent reports whether a request can safely be sent again after a
// temporary failure. POST and PATCH are left out, since a lost response could
// create a duplicate. They are only sent again when NetBox did not carry out the
// request, see notProcessed.
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

// permanent reports whether err says of itself that sending the request again will not help, for
// example a redirect the HTTP client refuses to follow.
func permanent(err error) bool {
	var p interface{ Permanent() bool }
	return errors.As(err, &p) && p.Permanent()
}

// notProcessed reports whether err tells that NetBox did not carry out the request, so that it is
// safe to send again whatever its method:
//   - a database deadlock, Postgres rolled the transaction back
//   - 429 Too Many Requests, the request was turned away
//   - 503 Service Unavailable, which a proxy or the application server answers when it cannot hand
//     the request to NetBox, for example while NetBox restarts or has too much to do
//   - no connection could be made, so nothing was sent
//
// 502 and 504 are not in the list: a proxy also answers them when NetBox did the work but the
// answer got lost or came too late.
func notProcessed(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests ||
			apiErr.StatusCode == http.StatusServiceUnavailable ||
			isDeadlock(apiErr.StatusCode, []byte(apiErr.Body))
	}
	var opErr *net.OpError
	return errors.As(err, &opErr) && opErr.Op == "dial"
}

// Do sends a request to rawURL. body is JSON encoded when not nil. When out
// is not nil the JSON response is decoded into it. Idempotent requests are
// sent again (see SetMaxRetries) when the connection fails or NetBox (or a proxy
// in front of it) answers 502, 503 or 504. Any request is sent again when NetBox
// did not carry it out, see notProcessed.
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
		if !idempotent(method) && !notProcessed(err) {
			return err
		}
		timer := time.NewTimer(backoff(attempt, err))
		select {
		case <-ctx.Done():
			timer.Stop()
			return err
		case <-timer.C:
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
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if buf != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ctx.Err() == nil && !permanent(err), err
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
			resp.StatusCode == http.StatusTooManyRequests ||
			isDeadlock(resp.StatusCode, data)
		apiErr := &APIError{Method: method, URL: rawURL, StatusCode: resp.StatusCode, Body: string(data)}
		if temporary {
			apiErr.RetryAfter = retryAfter(resp.Header)
		}
		return temporary, apiErr
	}
	if out == nil || len(data) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return false, fmt.Errorf("decoding response from %s: %w", rawURL, err)
	}
	return false, nil
}
