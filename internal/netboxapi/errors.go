package netboxapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"
	"unicode/utf8"
)

// APIError is returned when NetBox answers with a non-2xx status.
type APIError struct {
	Method     string
	URL        string
	StatusCode int
	Body       string
	// RetryAfter is the wait the server asked for with a Retry-After header, zero when it did not.
	RetryAfter time.Duration
}

// maxErrorBody is how much of a response goes into an error message. A proxy in front of NetBox can
// answer with a whole HTML page, which would bury the rest of the Terraform output.
const maxErrorBody = 4096

func (e *APIError) Error() string {
	body := e.Body
	if len(body) > maxErrorBody {
		cut := maxErrorBody
		for cut > 0 && !utf8.RuneStart(body[cut]) {
			cut--
		}
		body = fmt.Sprintf("%s... (%d more bytes)", body[:cut], len(body)-cut)
	}
	return fmt.Sprintf("netbox: %s %s returned %d: %s", e.Method, e.URL, e.StatusCode, body)
}

// IsNotFound reports whether err is a 404 from NetBox.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}
