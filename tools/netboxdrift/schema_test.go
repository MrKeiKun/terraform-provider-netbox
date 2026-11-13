package main

import "testing"

const sample = `{
 "paths": {
  "/api/dcim/sites/": {
   "get": {"parameters": [{"name": "name", "in": "query"}, {"name": "limit", "in": "query"}]},
   "post": {"requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/SiteRequest"}}}}}
  },
  "/api/dcim/sites/{id}/": {"get": {}},
  "/api/status/": {"get": {}}
 },
 "components": {"schemas": {"SiteRequest": {"type": "object", "properties": {
   "name": {"type": "string"}, "id": {"type": "integer", "readOnly": true}}}}}
}`

func TestParseEndpoints(t *testing.T) {
	eps, err := parseEndpoints([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	var site *endpoint
	for _, e := range eps {
		if e.Path == "dcim/sites/" {
			site = e
		}
	}
	if site == nil || !site.HasDetail || !site.CanCreate {
		t.Fatalf("sites endpoint not parsed: %+v", site)
	}
	if !site.Fields["name"] || site.Fields["id"] {
		t.Errorf("fields = %v, want name only", site.Fields)
	}
	if !site.Query["name"] {
		t.Errorf("query = %v", site.Query)
	}
}
