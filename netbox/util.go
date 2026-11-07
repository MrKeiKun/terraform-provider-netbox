package netbox

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func strToPtr(str string) *string {
	return &str
}

func int64ToPtr(i int64) *int64 {
	return &i
}

func float64ToPtr(i float64) *float64 {
	return &i
}

func toStringList(a interface{}) []string {
	strList := []string{}
	for _, str := range a.(*schema.Set).List() {
		strList = append(strList, str.(string))
	}
	return strList
}

func toInt64List(a interface{}) []int64 {
	intList := []int64{}
	for _, number := range a.(*schema.Set).List() {
		if n, ok := number.(int); ok {
			intList = append(intList, int64(n))
		} else if n, ok := number.(int64); ok {
			intList = append(intList, n)
		}
	}
	return intList
}

func toInt64PtrList(a interface{}) []*int64 {
	intList := []*int64{}
	if set, ok := a.(*schema.Set); ok {
		for _, number := range set.List() {
			if n, ok := number.(int); ok {
				intList = append(intList, int64ToPtr(int64(n)))
			} else if n, ok := number.(int64); ok {
				intList = append(intList, int64ToPtr(n))
			}
		}
	}
	return intList
}

func joinStringWithFinalConjunction(elems []string, sep, con string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}

	var b strings.Builder
	b.WriteString(strings.Join(elems[0:len(elems)-1], sep))
	fmt.Fprintf(&b, " %s %s", con, elems[len(elems)-1])
	return b.String()
}

func buildValidValueDescription(options []string) string {
	var quoted []string
	for _, option := range options {
		quoted = append(quoted, fmt.Sprintf("`%s`", option))
	}
	return "Valid values are " + joinStringWithFinalConjunction(quoted, ", ", "and")
}

func getOptionalStr(d *schema.ResourceData, key string, useSpace bool) string {
	strVal := ""
	// check if key is set
	strValInterface, ok := d.GetOk(key)
	if ok || d.HasChange(key) {
		if !ok && useSpace {
			// Setting an space string deletes the value
			// Ideally we would have the ability to determine if a struct value was set to the zero-value or not
			// The API often supports setting null to clear a value, but this is a Go/Swagger limitation
			strVal = " "
		} else if ok {
			strVal = strValInterface.(string)
		}
	}
	return strVal
}

func getOptionalVal[SchemaT int | float64, ApiT int64 | float64](d *schema.ResourceData, key string) *ApiT {
	var apiPtr *ApiT
	schemaValIf, ok := d.GetOk(key)
	if ok {
		schemaVal, _ := schemaValIf.(SchemaT)
		apiVal := ApiT(schemaVal)
		apiPtr = &apiVal
	}
	return apiPtr
}

func getOptionalInt(d *schema.ResourceData, key string) *int64 {
	return getOptionalVal[int, int64](d, key)
}

func getOptionalFloat(d *schema.ResourceData, key string) *float64 {
	return getOptionalVal[float64, float64](d, key)
}

// jsonSemanticCompare returns true when 2 json strings encode the same
// structure, regardless of whitespace differences. This can be used in
// DiffSuppressFunc implementations to prevent terraform showing whitespace
// changes as differences on refresh.
func jsonSemanticCompare(a, b string) (equal bool, err error) {
	var aDecoded, bDecoded any

	err = json.Unmarshal([]byte(a), &aDecoded)
	if err != nil {
		return false, fmt.Errorf("could not decode a: %w", err)
	}

	err = json.Unmarshal([]byte(b), &bDecoded)
	if err != nil {
		return false, fmt.Errorf("could not decode b: %w", err)
	}

	return reflect.DeepEqual(aDecoded, bDecoded), nil
}

// extractNetboxVersionFromString extracts the first semantic versioning string
// from a given string. The string must be preceded with "v".
// This is needed since netbox-docker recently started displaying its version string
// as, for example, v4.2.8-Docker-3.2.1 and we want to ignore the Docker part
// ref: https://github.com/MrKeiKun/terraform-provider-netbox/issues/729
func extractSemanticVersionFromString(s string) (version string, err error) {
	re := regexp.MustCompile(`^v?(\d+\.\d+\.\d+)`)

	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return "", fmt.Errorf("no semantic version found in version string")
	}
	return matches[1], nil
}

// getOptionalStrPtr returns the string of an optional attribute, or nil when it
// is not set. The nil is sent as null, which clears the value in NetBox.
func getOptionalStrPtr(d *schema.ResourceData, key string) *string {
	if v, ok := d.GetOk(key); ok {
		s := v.(string)
		return &s
	}
	return nil
}

// getIDSet returns the IDs in a set of numbers. The result is never nil, so an
// empty set is sent as an empty list and clears the relation.
func getIDSet(d *schema.ResourceData, key string) []int64 {
	ids := []int64{}
	if set, ok := d.Get(key).(*schema.Set); ok {
		for _, v := range set.List() {
			ids = append(ids, int64(v.(int)))
		}
	}
	return ids
}

// getStringSet returns the strings in a set of strings. The result is never
// nil, so an empty set is sent as an empty list.
func getStringSet(d *schema.ResourceData, key string) []string {
	out := []string{}
	if set, ok := d.Get(key).(*schema.Set); ok {
		for _, v := range set.List() {
			out = append(out, v.(string))
		}
	}
	return out
}

// suppressEquivalentJSON hides differences between JSON strings that only
// differ in formatting.
func suppressEquivalentJSON(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	equal, _ := jsonSemanticCompare(oldValue, newValue)
	return equal
}

// jsonRawFromResourceData returns the JSON in a string attribute. An empty
// attribute gives nil, which is sent as null and clears the value.
func jsonRawFromResourceData(d *schema.ResourceData, key string) (*json.RawMessage, error) {
	v, ok := d.GetOk(key)
	if !ok || v.(string) == "" {
		return nil, nil
	}
	raw := json.RawMessage(v.(string))
	return &raw, nil
}

// jsonString turns a JSON value of a response into the string that is stored in
// the state. A missing or null value gives an empty string.
func jsonString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	return string(raw)
}

// getOptionalStrClear returns the string of an optional attribute. It returns
// nil when the attribute is not set and did not change, so that the field is
// left out of the request, and a pointer to an empty string when the attribute
// was removed, which clears the value.
func getOptionalStrClear(d *schema.ResourceData, key string) *string {
	if v, ok := d.GetOk(key); ok {
		s := v.(string)
		return &s
	}
	if d.HasChange(key) {
		s := ""
		return &s
	}
	return nil
}

// setBodyStr adds a string attribute to a request body that is a map. The key
// is left out when the attribute is not set and did not change; when the
// attribute was removed it is sent as an empty string to clear the value.
func setBodyStr(body map[string]interface{}, d *schema.ResourceData, key, apiKey string) {
	if v := getOptionalStrClear(d, key); v != nil {
		body[apiKey] = *v
	}
}

// getOptionalBool returns the value of an optional boolean attribute, or nil
// when it is not set, so that NetBox applies its own default.
func getOptionalBool(d *schema.ResourceData, key string) *bool {
	if v, ok := d.GetOkExists(key); ok { //nolint:staticcheck // no other way to tell false from unset
		b := v.(bool)
		return &b
	}
	return nil
}
