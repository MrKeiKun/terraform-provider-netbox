package netboxapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Get reads one object. path is the collection path, for example
// "extras/tags/".
func Get[T any](ctx context.Context, c *Client, path string, id int64) (*T, error) {
	var out T
	if err := c.Do(ctx, http.MethodGet, c.endpoint(objectPath(path, id), nil), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create sends a POST and returns the created object.
func Create[T any](ctx context.Context, c *Client, path string, in any) (*T, error) {
	var out T
	if err := c.Do(ctx, http.MethodPost, c.endpoint(path, nil), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update sends a PUT and returns the updated object.
func Update[T any](ctx context.Context, c *Client, path string, id int64, in any) (*T, error) {
	var out T
	if err := c.Do(ctx, http.MethodPut, c.endpoint(objectPath(path, id), nil), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Patch sends a PATCH and returns the updated object.
func Patch[T any](ctx context.Context, c *Client, path string, id int64, in any) (*T, error) {
	var out T
	if err := c.Do(ctx, http.MethodPatch, c.endpoint(objectPath(path, id), nil), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Request sends a request to an API path and decodes the response. It is for
// endpoints that do not follow the usual object layout, for example actions on
// an object. path is relative to the API root, for example
// "ipam/prefixes/5/available-prefixes/".
func Request[T any](ctx context.Context, c *Client, method, path string, query url.Values, body any) (*T, error) {
	var out T
	if err := c.Do(ctx, method, c.endpoint(path, query), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes an object.
func Delete(ctx context.Context, c *Client, path string, id int64) error {
	return c.Do(ctx, http.MethodDelete, c.endpoint(objectPath(path, id), nil), nil, nil)
}

// Page is one page of a list response.
type Page[T any] struct {
	Count   int64   `json:"count"`
	Next    *string `json:"next"`
	Results []T     `json:"results"`
}

// ListPage fetches a single page. Use limit and offset in query to choose it.
func ListPage[T any](ctx context.Context, c *Client, path string, query url.Values) (*Page[T], error) {
	var p Page[T]
	if err := c.Do(ctx, http.MethodGet, c.endpoint(path, query), nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// List returns every object matching query, following pagination.
func List[T any](ctx context.Context, c *Client, path string, query url.Values) ([]T, error) {
	var all []T
	next := c.endpoint(path, query)
	for next != "" {
		var p Page[T]
		if err := c.Do(ctx, http.MethodGet, next, nil, &p); err != nil {
			return nil, err
		}
		all = append(all, p.Results...)
		next = ""
		if p.Next != nil {
			next = *p.Next
		}
	}
	return all, nil
}

func objectPath(path string, id int64) string {
	return fmt.Sprintf("%s/%d/", strings.TrimRight(path, "/"), id)
}
