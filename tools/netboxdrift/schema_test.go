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

const enumSample = `{
 "paths": {
  "/api/dcim/sites/": {
   "get": {},
   "post": {"requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/SiteRequest"}}}}}
  },
  "/api/dcim/sites/{id}/": {"get": {}}
 },
 "components": {"schemas": {
  "SiteRequest": {"type": "object", "properties": {
   "status": {"$ref": "#/components/schemas/StatusEnum"},
   "face": {"oneOf": [{"$ref": "#/components/schemas/FaceEnum"}, {"$ref": "#/components/schemas/BlankEnum"}]},
   "name": {"type": "string"}}},
  "StatusEnum": {"type": "string", "enum": ["planned", "active"]},
  "FaceEnum": {"type": "string", "enum": ["front", "rear"]},
  "BlankEnum": {"enum": [""]}
 }}
}`

func TestParseEnums(t *testing.T) {
	eps, err := parseEndpoints([]byte(enumSample))
	if err != nil {
		t.Fatal(err)
	}
	var site *endpoint
	for _, e := range eps {
		if e.Path == "dcim/sites/" {
			site = e
		}
	}
	if site == nil {
		t.Fatal("sites endpoint not parsed")
	}
	if got := site.Enums["status"]; len(got) != 2 || got[0] != "active" || got[1] != "planned" {
		t.Errorf("status choices = %v", got)
	}
	if got := site.Enums["face"]; len(got) != 2 || got[0] != "front" {
		t.Errorf("face choices = %v, the blank choice has to be left out", got)
	}
	if _, ok := site.Enums["name"]; ok {
		t.Error("a plain string has no choices")
	}
}
