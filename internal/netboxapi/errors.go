package netboxapi

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError is returned when NetBox answers with a non-2xx status.
type APIError struct {
	Method     string
	URL        string
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("netbox: %s %s returned %d: %s", e.Method, e.URL, e.StatusCode, e.Body)
}

// IsNotFound reports whether err is a 404 from NetBox.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}
