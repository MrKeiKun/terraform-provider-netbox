package fwresource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// customFieldValueToString renders a custom field value as returned by the API into the string
// form used in the Terraform configuration:
//   - null becomes ""
//   - select values (value and label) become the value
//   - object references become the id
//   - lists of choices or references become a JSON list of values or ids
//   - numbers and booleans become their plain text form
//   - anything else (for example a JSON field) becomes compact JSON
func customFieldValueToString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		return v.String()
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case map[string]any:
		if s, ok := customFieldReference(v); ok {
			return s
		}
	case []any:
		reduced := make([]any, len(v))
		changed := false
		for i, e := range v {
			if m, ok := e.(map[string]any); ok {
				if s, ok := customFieldReference(m); ok {
					if _, isChoice := m["label"]; isChoice {
						reduced[i] = s
					} else if id, err := strconv.ParseInt(s, 10, 64); err == nil {
						reduced[i] = id
					} else {
						reduced[i] = s
					}
					changed = true
					continue
				}
			}
			reduced[i] = e
		}
		if changed {
			v = reduced
		}
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v)
	}
	if b, err := json.Marshal(value); err == nil {
		return string(b)
	}
	return fmt.Sprintf("%v", value)
}

// customFieldReference recognises the objects NetBox returns for select choices (value and label)
// and object references (id, url, display). JSON custom fields that merely have similar keys are
// left alone.
func customFieldReference(m map[string]any) (string, bool) {
	if val, ok := m["value"].(string); ok && len(m) == 2 {
		if _, hasLabel := m["label"]; hasLabel {
			return val, true
		}
	}
	if id, ok := m["id"].(float64); ok {
		_, hasURL := m["url"]
		_, hasDisplay := m["display"]
		if hasURL || hasDisplay {
			return strconv.FormatInt(int64(id), 10), true
		}
	}
	return "", false
}

// customFieldsFromObject prepares the custom fields of an API response for the state. Every value
// is a string and fields that are not set are left out. With only, just those fields are kept.
func customFieldsFromObject(cf any, only map[string]string, adoptAll, keepMissing bool) map[string]string {
	cfm, _ := cf.(map[string]any)
	result := map[string]string{}
	for key, value := range cfm {
		s := customFieldValueToString(value)
		if prior, managed := only[key]; managed {
			// keep the formatting of the configuration when the content is the same
			if customFieldValuesEqual(prior, s) {
				s = prior
			}
			result[key] = s
			continue
		}
		if adoptAll && s != "" {
			result[key] = s
		}
	}
	if keepMissing {
		// fields NetBox no longer reports (for example, the field was deleted in the same run) keep their value
		for k, v := range only {
			if _, ok := result[k]; !ok {
				result[k] = v
			}
		}
	}
	return result
}

// customFieldValuesEqual compares two custom field values from state and configuration. Empty means
// unset, and JSON values are compared by content, so formatting differences are not drift.
func customFieldValuesEqual(a, b string) bool {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if a == b {
		return true
	}
	var ja, jb any
	if json.Unmarshal([]byte(a), &ja) == nil && json.Unmarshal([]byte(b), &jb) == nil {
		return reflect.DeepEqual(normalizeCustomFieldJSON(ja), normalizeCustomFieldJSON(jb))
	}
	return false
}

// normalizeCustomFieldJSON makes lists of scalars comparable: the order does not matter and 1
// equals "1" (ids written as strings in the configuration come back from NetBox as numbers).
func normalizeCustomFieldJSON(v any) any {
	list, ok := v.([]any)
	if !ok {
		return v
	}
	out := make([]string, 0, len(list))
	for _, e := range list {
		switch e.(type) {
		case map[string]any, []any:
			return v
		}
		out = append(out, customFieldValueToString(e))
	}
	sort.Strings(out)
	return out
}
