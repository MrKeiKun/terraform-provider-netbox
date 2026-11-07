package netbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// addStrQuery adds a string attribute to the query when it is set.
func addStrQuery(q url.Values, d *schema.ResourceData, key, param string) {
	if v, ok := d.Get(key).(string); ok && v != "" {
		q.Set(param, v)
	}
}

// addIntQuery adds a number attribute to the query when it is not zero.
func addIntQuery(q url.Values, d *schema.ResourceData, key, param string) {
	if v, ok := d.Get(key).(int); ok && v != 0 {
		q.Set(param, strconv.Itoa(v))
	}
}

// findOne looks up the single object that matches query. It fails when there is
// no match or more than one. noun names the object in the error messages.
func findOne[T any](m interface{}, path string, query url.Values, noun string) (*T, error) {
	return findOneWith[T](m, path, query,
		fmt.Sprintf("more than one %s returned, specify a more narrow filter", noun),
		fmt.Sprintf("no %s found matching filter", noun))
}

// findOneWith is findOne with explicit error messages for more than one and for
// no match.
func findOneWith[T any](m interface{}, path string, query url.Values, moreThanOne, none string) (*T, error) {
	api := m.(*providerState)

	// Two results are enough to know that the filter is not narrow enough.
	query.Set("limit", "2")

	page, err := netboxapi.ListPage[T](context.Background(), api.rest, path, query)
	if err != nil {
		return nil, err
	}

	if page.Count > 1 {
		return nil, errors.New(moreThanOne)
	}
	if page.Count == 0 || len(page.Results) == 0 {
		return nil, errors.New(none)
	}
	return &page.Results[0], nil
}

// listFiltered fetches every object for a plural data source. The "filter"
// blocks and the "limit" attribute of the data source are translated into the
// query. allowed maps each accepted filter name to the query parameter it is
// sent as; a parameter ending in [] takes a comma separated list. No other
// filter names are accepted, so that a typo in a filter cannot silently return
// everything. extra holds fixed query parameters.
func listFiltered[T any](d *schema.ResourceData, m interface{}, path, what string, allowed map[string]string, extra url.Values) ([]T, error) {
	return listFilteredWith[T](d, m, path, what, allowed, extra, nil)
}

// listFilteredWith is listFiltered with an additional filter that is applied
// on the client side, for example for a name_regex attribute. When keep is set,
// all pages are fetched before it is applied and the limit comes after it.
func listFilteredWith[T any](d *schema.ResourceData, m interface{}, path, what string, allowed map[string]string, extra url.Values, keep func(T) bool) ([]T, error) {
	api := m.(*providerState)

	query := url.Values{}
	for k, vs := range extra {
		for _, v := range vs {
			query.Add(k, v)
		}
	}

	if filter, ok := d.GetOk("filter"); ok {
		for _, f := range filter.(*schema.Set).List() {
			name := f.(map[string]interface{})["name"].(string)
			value := f.(map[string]interface{})["value"].(string)
			param, ok := allowed[name]
			if !ok {
				return nil, fmt.Errorf("'%s' is not a supported filter parameter", name)
			}
			if value == "" {
				// an empty value is not sent, same as before
				continue
			}
			if list, isList := strings.CutSuffix(param, "[]"); isList {
				// the value is a comma separated list, sent as repeated parameters
				for _, part := range strings.Split(value, ",") {
					query.Add(list, part)
				}
				continue
			}
			query.Add(param, value)
		}
	}

	var userLimit int64
	if limit, ok := d.GetOk("limit"); ok {
		userLimit = int64(limit.(int))
	}

	var all []T
	for {
		query.Set("limit", strconv.Itoa(DefaultPageSize))
		query.Set("offset", strconv.Itoa(len(all)))

		page, err := netboxapi.ListPage[T](context.Background(), api.rest, path, query)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s at offset %d: %w", what, len(all), err)
		}
		all = append(all, page.Results...)

		if len(page.Results) == 0 || page.Next == nil || *page.Next == "" {
			break
		}
		if keep == nil && userLimit > 0 && int64(len(all)) >= userLimit {
			break
		}
	}

	if keep != nil {
		var kept []T
		for _, o := range all {
			if keep(o) {
				kept = append(kept, o)
			}
		}
		all = kept
	}

	if userLimit > 0 && int64(len(all)) > userLimit {
		all = all[:userLimit]
	}
	return all, nil
}

// addCustomFieldQuery adds a cf_<name> filter for every string custom field value.
func addCustomFieldQuery(q url.Values, cfm map[string]interface{}) {
	for k, v := range cfm {
		if vs, ok := v.(string); ok {
			q.Set("cf_"+k, vs)
		}
	}
}

// The plural data sources read the objects as plain JSON and copy the fields
// they expose into the schema. These helpers read single fields from such an
// object; a missing or null field gives the zero value.

func jStr(o map[string]interface{}, k string) string {
	s, _ := o[k].(string)
	return s
}

func jInt(o map[string]interface{}, k string) int64 {
	f, _ := o[k].(float64)
	return int64(f)
}

func jBool(o map[string]interface{}, k string) bool {
	b, _ := o[k].(bool)
	return b
}

// jID returns the ID of a nested object, or nil when there is none.
func jID(o map[string]interface{}, k string) interface{} {
	nested, ok := o[k].(map[string]interface{})
	if !ok {
		return nil
	}
	if f, ok := nested["id"].(float64); ok {
		return int64(f)
	}
	return nil
}

// jChoice returns the value of a nested choice object, or an empty string.
func jChoice(o map[string]interface{}, k string) string {
	nested, ok := o[k].(map[string]interface{})
	if !ok {
		return ""
	}
	s, _ := nested["value"].(string)
	return s
}

// jTags returns the names of the tags of an object.
func jTags(o map[string]interface{}) []string {
	names := []string{}
	list, _ := o["tags"].([]interface{})
	for _, item := range list {
		if nested, ok := item.(map[string]interface{}); ok {
			if name, ok := nested["name"].(string); ok {
				names = append(names, name)
			}
		}
	}
	return names
}

// jChoiceInt returns the value of a nested choice object that holds a number,
// or nil when there is none.
func jChoiceInt(o map[string]interface{}, k string) interface{} {
	nested, ok := o[k].(map[string]interface{})
	if !ok {
		return nil
	}
	if f, ok := nested["value"].(float64); ok {
		return int64(f)
	}
	return nil
}

// jTime returns a timestamp in the format the provider has always used.
func jTime(o map[string]interface{}, k string) string {
	s, _ := o[k].(string)
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

// jNested returns a field of a nested object, numbers as int64. It returns nil
// when the nested object or the field is missing.
func jNested(o map[string]interface{}, k, field string) interface{} {
	nested, ok := o[k].(map[string]interface{})
	if !ok {
		return nil
	}
	v := nested[field]
	if f, ok := v.(float64); ok {
		return int64(f)
	}
	return v
}

// jTagIDs returns the IDs of the tags of an object.
func jTagIDs(o map[string]interface{}) []int64 {
	var ids []int64
	list, _ := o["tags"].([]interface{})
	for _, item := range list {
		if nested, ok := item.(map[string]interface{}); ok {
			if f, ok := nested["id"].(float64); ok {
				ids = append(ids, int64(f))
			}
		}
	}
	return ids
}

func jNumbers(v interface{}) interface{} {
	if f, ok := v.(float64); ok {
		return int64(f)
	}
	return v
}

// jObjectList returns a nested object as a list with one map that holds the
// named fields, or nil when the object is missing. Numbers become int64.
func jObjectList(o map[string]interface{}, k string, fields ...string) []map[string]interface{} {
	nested, ok := o[k].(map[string]interface{})
	if !ok {
		return nil
	}
	return []map[string]interface{}{pickFields(nested, fields)}
}

// jObjectsList returns a list of nested objects as a list of maps that hold the
// named fields.
func jObjectsList(o map[string]interface{}, k string, fields ...string) []map[string]interface{} {
	var out []map[string]interface{}
	list, _ := o[k].([]interface{})
	for _, item := range list {
		if nested, ok := item.(map[string]interface{}); ok {
			out = append(out, pickFields(nested, fields))
		}
	}
	return out
}

func pickFields(nested map[string]interface{}, fields []string) map[string]interface{} {
	m := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		m[f] = jNumbers(nested[f])
	}
	return m
}

// jJSON returns a field as a JSON string, or nil when the field is missing or null.
func jJSON(o map[string]interface{}, k string) interface{} {
	if o[k] == nil {
		return nil
	}
	b, err := json.Marshal(o[k])
	if err != nil {
		return nil
	}
	return string(b)
}

// jIntP, jFloatP and jBoolP are like jInt, jFloat and jBool, but a missing or
// null field stays unset instead of becoming zero.
func jIntP(o map[string]interface{}, k string) interface{} {
	if f, ok := o[k].(float64); ok {
		return int64(f)
	}
	return nil
}

func jFloatP(o map[string]interface{}, k string) interface{} {
	if f, ok := o[k].(float64); ok {
		return f
	}
	return nil
}

// jDeepID returns the ID of an object nested inside a nested object.
func jDeepID(o map[string]interface{}, k1, k2 string) interface{} {
	nested, ok := o[k1].(map[string]interface{})
	if !ok {
		return nil
	}
	return jID(nested, k2)
}

// jIPAddress returns the address, without the prefix length, of a nested IP
// address object, or nil when there is none or it cannot be parsed.
func jIPAddress(o map[string]interface{}, k string) interface{} {
	nested, ok := o[k].(map[string]interface{})
	if !ok {
		return nil
	}
	address, _ := nested["address"].(string)
	ip, _, err := net.ParseCIDR(address)
	if err != nil {
		return nil
	}
	return ip.String()
}

// jChoiceMap returns a nested choice object as a map with label and value, or
// nil when there is none.
func jChoiceMap(o map[string]interface{}, k string) interface{} {
	nested, ok := o[k].(map[string]interface{})
	if !ok {
		return nil
	}
	label, _ := nested["label"].(string)
	value, _ := nested["value"].(string)
	return map[string]string{"label": label, "value": value}
}

// jFirst returns a field of the first object in a list of nested objects.
func jFirst(o map[string]interface{}, k, field string) interface{} {
	list, _ := o[k].([]interface{})
	if len(list) == 0 {
		return nil
	}
	first, ok := list[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return first[field]
}

// jRanges turns a list of [start, end] pairs into a list of maps.
func jRanges(o map[string]interface{}, k string) []map[string]int64 {
	ranges := []map[string]int64{}
	list, _ := o[k].([]interface{})
	for _, item := range list {
		pair, ok := item.([]interface{})
		if !ok || len(pair) < 2 {
			continue
		}
		start, _ := pair[0].(float64)
		end, _ := pair[1].(float64)
		ranges = append(ranges, map[string]int64{"start": int64(start), "end": int64(end)})
	}
	return ranges
}

// jIDs returns the IDs of a list of nested objects.
func jIDs(o map[string]interface{}, k string) []int64 {
	ids := []int64{}
	list, _ := o[k].([]interface{})
	for _, item := range list {
		if nested, ok := item.(map[string]interface{}); ok {
			if f, ok := nested["id"].(float64); ok {
				ids = append(ids, int64(f))
			}
		}
	}
	return ids
}

// jInts returns the numbers in a list field.
func jInts(o map[string]interface{}, k string) []int64 {
	out := []int64{}
	list, _ := o[k].([]interface{})
	for _, item := range list {
		if f, ok := item.(float64); ok {
			out = append(out, int64(f))
		}
	}
	return out
}

// jStrs returns the strings in a list field.
func jStrs(o map[string]interface{}, k string) []string {
	out := []string{}
	list, _ := o[k].([]interface{})
	for _, item := range list {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
