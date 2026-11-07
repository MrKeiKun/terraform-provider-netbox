package netboxapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// customFieldDef is the part of a NetBox custom field definition needed to
// send values of the right JSON type.
type customFieldDef struct {
	Name string `json:"name"`
	Type struct {
		Value string `json:"value"`
	} `json:"type"`
}

// customFieldTypes returns the type of each custom field by name. The result
// is cached; it is refreshed when a name that is not cached is asked for, so
// fields created earlier in the same run are found.
func (c *Client) customFieldTypes(ctx context.Context, names []string) (map[string]string, error) {
	c.cfMu.Lock()
	defer c.cfMu.Unlock()

	missing := c.cfTypes == nil
	for _, n := range names {
		if _, ok := c.cfTypes[n]; !ok {
			missing = true
		}
	}
	if missing {
		defs, err := List[customFieldDef](ctx, c, "extras/custom-fields/", url.Values{"limit": {"1000"}})
		if err != nil {
			return nil, err
		}
		c.cfTypes = make(map[string]string, len(defs))
		for _, d := range defs {
			c.cfTypes[d.Name] = d.Type.Value
		}
	}
	return c.cfTypes, nil
}

// encodeBody marshals a request body. When the body has custom_fields whose
// values are strings (the provider keeps them as map(string)), they are
// converted to the JSON type NetBox expects for each field.
func (c *Client) encodeBody(ctx context.Context, method, rawURL string, body any) ([]byte, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
		return buf, nil
	}
	if strings.Contains(rawURL, "/extras/custom-fields/") {
		return buf, nil
	}

	var obj map[string]json.RawMessage
	if json.Unmarshal(buf, &obj) != nil {
		return buf, nil
	}
	rawCF, ok := obj["custom_fields"]
	if !ok {
		return buf, nil
	}
	var cf map[string]any
	if json.Unmarshal(rawCF, &cf) != nil || len(cf) == 0 {
		return buf, nil
	}
	names := make([]string, 0, len(cf))
	needTypes := false
	for k, v := range cf {
		names = append(names, k)
		if _, isString := v.(string); isString {
			needTypes = true
		}
	}
	if !needTypes {
		return buf, nil
	}
	types, err := c.customFieldTypes(ctx, names)
	if err != nil {
		return nil, err
	}
	for k, v := range cf {
		if s, isString := v.(string); isString {
			cf[k] = TypedCustomFieldValue(types[k], s)
		}
	}
	newCF, err := json.Marshal(cf)
	if err != nil {
		return nil, err
	}
	obj["custom_fields"] = newCF
	return json.Marshal(obj)
}

// TypedCustomFieldValue converts the string form of a custom field value to
// the JSON value NetBox expects for the given custom field type. An empty
// string clears the field. A value that cannot be converted is returned as
// is, so NetBox reports the problem.
func TypedCustomFieldValue(fieldType, s string) any {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	switch fieldType {
	case "integer":
		if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return i
		}
	case "decimal":
		if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return json.Number(trimmed)
		}
	case "boolean":
		if b, err := strconv.ParseBool(trimmed); err == nil {
			return b
		}
	case "json":
		if json.Valid([]byte(trimmed)) {
			return json.RawMessage(trimmed)
		}
	case "select":
		var choice struct {
			Value *string `json:"value"`
			Label *string `json:"label"`
		}
		if strings.HasPrefix(trimmed, "{") && json.Unmarshal([]byte(trimmed), &choice) == nil && choice.Value != nil {
			return *choice.Value
		}
	case "multiselect":
		if l, ok := listValue(trimmed, func(m map[string]any) (any, bool) {
			v, ok := m["value"].(string)
			return v, ok
		}); ok {
			return l
		}
	case "object":
		if id, ok := objectID(trimmed); ok {
			return id
		}
	case "multiobject":
		if l, ok := listValue(trimmed, func(m map[string]any) (any, bool) {
			f, ok := m["id"].(float64)
			return int64(f), ok
		}); ok {
			ids := make([]any, len(l))
			for i, v := range l {
				if id, ok := toID(v); ok {
					ids[i] = id
				} else {
					return s
				}
			}
			return ids
		}
	}
	return s
}

// listValue parses a JSON list or a comma separated list. Elements that are
// objects are reduced with pick.
func listValue(s string, pick func(map[string]any) (any, bool)) ([]any, bool) {
	var raw []any
	if strings.HasPrefix(s, "[") {
		if json.Unmarshal([]byte(s), &raw) != nil {
			return nil, false
		}
	} else {
		for _, p := range strings.Split(s, ",") {
			raw = append(raw, strings.TrimSpace(p))
		}
	}
	out := make([]any, 0, len(raw))
	for _, e := range raw {
		if m, ok := e.(map[string]any); ok {
			v, ok := pick(m)
			if !ok {
				return nil, false
			}
			out = append(out, v)
			continue
		}
		out = append(out, e)
	}
	return out, true
}

func toID(v any) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case float64:
		return int64(t), true
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		return i, err == nil
	}
	return 0, false
}

// objectID reads the id of an object reference: either a plain id or the
// nested object NetBox returns.
func objectID(s string) (int64, bool) {
	if strings.HasPrefix(s, "{") {
		var m map[string]any
		if json.Unmarshal([]byte(s), &m) != nil {
			return 0, false
		}
		return toID(m["id"])
	}
	return toID(s)
}
