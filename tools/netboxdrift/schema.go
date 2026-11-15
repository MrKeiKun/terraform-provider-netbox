package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// endpoint is one API collection, for example dcim/sites/.
type endpoint struct {
	Path string // without the /api/ prefix, with a trailing slash
	// Fields are the writable fields of the POST body.
	Fields map[string]bool
	// Enums are the choices of the writable fields that only take a fixed set of strings.
	Enums map[string][]string
	// Query are the query parameters of the list operation.
	Query map[string]bool
	// CanCreate tells whether the collection accepts POST.
	CanCreate bool
	// HasDetail tells whether /{id}/ exists, which is what makes a collection an object type.
	HasDetail bool
}

type openAPI struct {
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Schemas map[string]*schemaObj `json:"schemas"`
	} `json:"components"`
}

type schemaObj struct {
	Ref        string                `json:"$ref"`
	Type       string                `json:"type"`
	Properties map[string]*schemaObj `json:"properties"`
	Items      *schemaObj            `json:"items"`
	AllOf      []*schemaObj          `json:"allOf"`
	OneOf      []*schemaObj          `json:"oneOf"`
	AnyOf      []*schemaObj          `json:"anyOf"`
	ReadOnly   bool                  `json:"readOnly"`
	Enum       []any                 `json:"enum"`
}

type operation struct {
	Parameters  []parameter `json:"parameters"`
	RequestBody struct {
		Content map[string]struct {
			Schema *schemaObj `json:"schema"`
		} `json:"content"`
	} `json:"requestBody"`
}

type parameter struct {
	Name string `json:"name"`
	In   string `json:"in"`
}

// loadSchema reads the OpenAPI document from a file, or from the URL of a NetBox server.
func loadSchema(file, serverURL, token string) ([]byte, error) {
	if file != "" {
		return os.ReadFile(file)
	}
	if serverURL == "" {
		return nil, fmt.Errorf("give -schema or -url")
	}
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(serverURL, "/")+"/api/schema/?format=json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if token != "" {
		scheme := "Token"
		if strings.HasPrefix(token, "nbt_") {
			scheme = "Bearer"
		}
		req.Header.Set("Authorization", scheme+" "+token)
	}
	resp, err := (&http.Client{Timeout: 120 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", req.URL, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// parseEndpoints turns the OpenAPI document into the list of API collections.
func parseEndpoints(data []byte) ([]*endpoint, error) {
	var doc openAPI
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}
	var out []*endpoint
	for p, methods := range doc.Paths {
		trimmed := strings.TrimPrefix(p, "/api/")
		if trimmed == p || strings.Contains(trimmed, "{") || !strings.HasSuffix(trimmed, "/") {
			continue
		}
		ep := &endpoint{Path: trimmed, Fields: map[string]bool{}, Enums: map[string][]string{}, Query: map[string]bool{}}
		_, ep.HasDetail = doc.Paths[p+"{id}/"]
		if raw, ok := methods["get"]; ok {
			var op operation
			if err := json.Unmarshal(raw, &op); err != nil {
				return nil, fmt.Errorf("parse GET %s: %w", p, err)
			}
			for _, par := range op.Parameters {
				if par.In == "query" {
					ep.Query[par.Name] = true
				}
			}
		} else {
			continue
		}
		if raw, ok := methods["post"]; ok {
			var op operation
			if err := json.Unmarshal(raw, &op); err != nil {
				return nil, fmt.Errorf("parse POST %s: %w", p, err)
			}
			ep.CanCreate = true
			for _, c := range op.RequestBody.Content {
				collectFields(&doc, c.Schema, ep.Fields, ep.Enums, 0)
			}
		}
		out = append(out, ep)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func resolve(doc *openAPI, s *schemaObj) *schemaObj {
	const prefix = "#/components/schemas/"
	for s != nil && strings.HasPrefix(s.Ref, prefix) {
		s = doc.Components.Schemas[strings.TrimPrefix(s.Ref, prefix)]
	}
	return s
}

// collectFields adds the writable properties of a request body schema. POST bodies can be a single
// object or a list of them, for bulk creation.
func collectFields(doc *openAPI, s *schemaObj, out map[string]bool, enums map[string][]string, depth int) {
	s = resolve(doc, s)
	if s == nil || depth > 6 {
		return
	}
	for name, prop := range s.Properties {
		if r := resolve(doc, prop); (prop != nil && prop.ReadOnly) || (r != nil && r.ReadOnly) {
			continue
		}
		out[name] = true
		if choices := enumChoices(doc, prop, 0); len(choices) > 0 {
			enums[name] = choices
		}
	}
	collectFields(doc, s.Items, out, enums, depth+1)
	for _, list := range [][]*schemaObj{s.AllOf, s.OneOf, s.AnyOf} {
		for _, sub := range list {
			collectFields(doc, sub, out, enums, depth+1)
		}
	}
}

// enumChoices returns the strings that a property accepts, when it is limited to a fixed set. The
// empty string and null are left out, they only say that the field may be blank.
func enumChoices(doc *openAPI, s *schemaObj, depth int) []string {
	s = resolve(doc, s)
	if s == nil || depth > 4 {
		return nil
	}
	var out []string
	for _, v := range s.Enum {
		if str, ok := v.(string); ok && str != "" {
			out = append(out, str)
		}
	}
	for _, list := range [][]*schemaObj{s.AllOf, s.OneOf, s.AnyOf} {
		for _, sub := range list {
			out = append(out, enumChoices(doc, sub, depth+1)...)
		}
	}
	sort.Strings(out)
	return out
}
