package netbox

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const customFieldsKey = "custom_fields"

var customFieldsSchema = &schema.Schema{
	Type:        schema.TypeMap,
	Description: "Custom field values of the object, as a map of field name to value.",
	Optional:    true,
	Default:     nil,
	Elem: &schema.Schema{
		Type:    schema.TypeString,
		Default: nil,
	},
	DiffSuppressFunc: suppressCustomFieldsDiff,
}

// customFieldValueToString renders a custom field value as returned by the
// API into the string form used in the Terraform configuration:
//   - null becomes ""
//   - select values ({"value","label"}) become the value
//   - object references become the id
//   - lists of choices or references become a JSON list of values or ids
//   - numbers and booleans become their plain text form
//   - anything else (for example a JSON field) becomes compact JSON
func customFieldValueToString(value interface{}) string {
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
	case map[string]interface{}:
		if s, ok := customFieldReference(v); ok {
			return s
		}
	case []interface{}:
		reduced := make([]interface{}, len(v))
		changed := false
		for i, e := range v {
			if m, ok := e.(map[string]interface{}); ok {
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

// customFieldReference recognises the objects NetBox returns for select
// choices ({"value","label"}) and object references ({"id","url","display"}).
// JSON custom fields that merely have similar keys are left alone.
func customFieldReference(m map[string]interface{}) (string, bool) {
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

// getCustomFields prepares the custom fields of an API response for the
// Terraform state: every value is a string and fields that are not set are
// left out.
func getCustomFields(cf interface{}) map[string]interface{} {
	cfm, ok := cf.(map[string]interface{})
	if !ok || len(cfm) == 0 {
		return nil
	}

	result := make(map[string]interface{})
	for key, value := range cfm {
		if s := customFieldValueToString(value); s != "" {
			result[key] = s
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// managedCustomFields returns the custom fields to send to NetBox: only the
// ones set in the configuration. Fields that are only in the state (they come
// from NetBox, for example defaults) are not sent again, since the field may
// not exist any more.
func managedCustomFields(d *schema.ResourceData) (map[string]interface{}, bool) {
	all, ok := d.Get(customFieldsKey).(map[string]interface{})
	if !ok || len(all) == 0 {
		return nil, false
	}
	cfg := d.GetRawConfig()
	if cfg.IsNull() || !cfg.IsKnown() {
		return all, true
	}
	raw := cfg.GetAttr(customFieldsKey)
	if raw.IsNull() || !raw.IsKnown() {
		return nil, false
	}
	managed := raw.AsValueMap()
	result := make(map[string]interface{}, len(managed))
	for name, v := range all {
		if _, ok := managed[name]; ok {
			result[name] = v
		}
	}
	if len(result) == 0 {
		return nil, false
	}
	return result, true
}

// customFieldValuesEqual compares two custom field values from state and
// configuration. Empty means unset, and JSON values are compared by content,
// so formatting differences are not drift.
func customFieldValuesEqual(a, b string) bool {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if a == b {
		return true
	}
	var ja, jb interface{}
	if json.Unmarshal([]byte(a), &ja) == nil && json.Unmarshal([]byte(b), &jb) == nil {
		return reflect.DeepEqual(normalizeCustomFieldJSON(ja), normalizeCustomFieldJSON(jb))
	}
	return false
}

// normalizeCustomFieldJSON makes lists of scalars comparable: the order does
// not matter and 1 equals "1" (ids written as strings in the configuration
// come back from NetBox as numbers).
func normalizeCustomFieldJSON(v interface{}) interface{} {
	list, ok := v.([]interface{})
	if !ok {
		return v
	}
	out := make([]string, 0, len(list))
	for _, e := range list {
		switch e.(type) {
		case map[string]interface{}, []interface{}:
			return v
		}
		out = append(out, customFieldValueToString(e))
	}
	sort.Strings(out)
	return out
}

// suppressCustomFieldsDiff hides differences that are not real:
//   - custom fields that are not in the configuration are not managed, NetBox
//     returns every defined field and some have defaults
//   - an empty string and an unset field are the same
//   - JSON values that only differ in formatting
func suppressCustomFieldsDiff(k, oldValue, newValue string, d *schema.ResourceData) bool {
	cfg := d.GetRawConfig()
	if cfg.IsNull() || !cfg.IsKnown() {
		return false
	}
	raw := cfg.GetAttr(customFieldsKey)
	if raw.IsNull() {
		// Not configured at all: nothing to manage.
		return true
	}
	if !raw.IsWhollyKnown() {
		return false
	}
	configured := map[string]string{}
	for name, v := range raw.AsValueMap() {
		if v.IsNull() {
			continue
		}
		configured[name] = v.AsString()
	}

	if k == customFieldsKey+".%" {
		o, _ := d.GetChange(customFieldsKey)
		state, _ := o.(map[string]interface{})
		for name, want := range configured {
			have, _ := state[name].(string)
			if !customFieldValuesEqual(have, want) {
				return false
			}
		}
		return true
	}

	name := strings.TrimPrefix(k, customFieldsKey+".")
	want, managed := configured[name]
	if !managed {
		return true
	}
	return customFieldValuesEqual(oldValue, want)
}

// flattenCustomFields converts custom fields to a map where all values are strings.
// Complex nested objects (like IP address references) are converted to JSON strings.
func flattenCustomFields(cf interface{}) map[string]interface{} {
	cfm, ok := cf.(map[string]interface{})
	if !ok || len(cfm) == 0 {
		return nil
	}

	result := make(map[string]interface{})
	for key, value := range cfm {
		if value == nil {
			result[key] = ""
			continue
		}

		switch v := value.(type) {
		case string, bool, float64, int, int64:
			result[key] = customFieldValueToString(v)
		default:
			// For complex types (maps, arrays, objects), convert to JSON string
			if jsonBytes, err := json.Marshal(value); err == nil {
				result[key] = string(jsonBytes)
			} else {
				// Fallback to string representation
				result[key] = fmt.Sprintf("%v", value)
			}
		}
	}

	return result
}

// customFieldsRequest returns the custom fields for a request body, or nil when
// none are set in the configuration, so that the field is left out.
func customFieldsRequest(d *schema.ResourceData) interface{} {
	if ct, ok := managedCustomFields(d); ok {
		return ct
	}
	return nil
}
