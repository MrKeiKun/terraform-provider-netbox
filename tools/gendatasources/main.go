// Command gendatasources writes the data sources that the provider does not have yet, from the NetBox
// OpenAPI schema, to internal/fwprovider/datasources_generated.go. A collection that already has a
// hand written data source of one kind (single or plural) only gets the other kind.
//
//	go run ./tools/gendatasources -url http://localhost:8001
//	go run ./tools/gendatasources -schema schema.json
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwprovider"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
)

var (
	schemaFile   = flag.String("schema", "", "OpenAPI document to read")
	serverURL    = flag.String("url", "", "NetBox server to fetch /api/schema/ from")
	outFile      = flag.String("out", "internal/fwprovider/datasources_generated.go", "file to write")
	extrasOut    = flag.String("extras-out", "internal/fwprovider/datasources_fields_generated.go", "file to write the extra attributes of hand written data sources to")
	resExtrasOut = flag.String("resource-extras-out", "internal/fwprovider/resources_fields_generated.go", "file to write the read-only attributes of the resources to")
	filtersOut   = flag.String("filters-out", "internal/fwprovider/datasources_filters_generated.go", "file to write the extra filters of hand written plural data sources to")
	examplesDir  = flag.String("examples", "examples/data-sources", "directory to write data source examples to, empty for none")
	examplesList = flag.String("examples-list", "tools/gendatasources/generated-examples.txt", "list of the examples this command writes")
)

type obj = map[string]any

func main() {
	flag.Parse()
	data, err := load()
	if err != nil {
		fatal(err)
	}
	var doc obj
	if err := json.Unmarshal(data, &doc); err != nil {
		fatal(err)
	}
	src, filters, extras, resExtras, err := generate(doc)
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*outFile, src, 0o644); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*filtersOut, filters, 0o644); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*extrasOut, extras, 0o644); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*resExtrasOut, resExtras, 0o644); err != nil {
		fatal(err)
	}
	if *examplesDir != "" {
		if err := writeExamples(*examplesDir, *examplesList); err != nil {
			fatal(err)
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "gendatasources:", err)
	os.Exit(1)
}

func load() ([]byte, error) {
	if *schemaFile != "" {
		return os.ReadFile(*schemaFile)
	}
	if *serverURL == "" {
		return nil, fmt.Errorf("give -schema or -url")
	}
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(*serverURL, "/")+"/api/schema/?format=json", nil)
	if err != nil {
		return nil, err
	}
	if token := os.Getenv("NETBOX_API_TOKEN"); token != "" {
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

// skippedPrefixes are collections that are not objects a user manages.
var skippedPrefixes = []string{"core/", "extras/scripts/", "extras/script-modules/", "plugins/"}

// names that differ from the name made from the path. They are the names of the objects in the NetBox UI.
var nameOverrides = map[string][2]string{
	"dcim/console-ports/":            {"device_console_port", "device_console_ports"},
	"dcim/console-server-ports/":     {"device_console_server_port", "device_console_server_ports"},
	"dcim/front-ports/":              {"device_front_port", "device_front_ports"},
	"dcim/rear-ports/":               {"device_rear_port", "device_rear_ports"},
	"dcim/power-ports/":              {"device_power_port", "device_power_ports"},
	"dcim/power-outlets/":            {"device_power_outlet", "device_power_outlets"},
	"dcim/interfaces/":               {"device_interface", "device_interfaces"},
	"dcim/module-bays/":              {"device_module_bay", "device_module_bays"},
	"dcim/virtual-chassis/":          {"virtual_chassis", "virtual_chassis_list"},
	"virtualization/interfaces/":     {"virtual_machine_interface", "virtual_machine_interfaces"},
	"circuits/providers/":            {"circuit_provider", "circuit_providers"},
	"circuits/provider-accounts/":    {"circuit_provider_account", "circuit_provider_accounts"},
	"ipam/roles/":                    {"ipam_role", "ipam_roles"},
	"vpn/tunnel-groups/":             {"vpn_tunnel_group", "vpn_tunnel_groups"},
	"vpn/tunnel-terminations/":       {"vpn_tunnel_termination", "vpn_tunnel_terminations"},
	"extras/config-context-profiles": {"config_context_profile", "config_context_profiles"},
}

// scopeFilters are the foreign keys that narrow down a single object lookup, such as the device of an interface.
var scopeFilters = map[string]bool{
	"device": true, "virtual_machine": true, "module": true, "circuit": true, "provider": true,
	"device_type": true, "module_type": true, "rack": true, "site": true, "cluster": true,
	"policy": true, "group": true, "l2vpn": true, "cable": true, "tunnel": true,
}

// notLookups are properties that are never used to find a single object.
var notLookups = map[string]bool{
	"id": true, "url": true, "display": true, "display_url": true, "created": true, "last_updated": true,
	"description": true, "comments": true, "tags": true, "custom_fields": true,
}

// ignoredQuery are list parameters that are not offered as filters.
var ignoredQuery = map[string]bool{
	"limit": true, "offset": true, "ordering": true, "brief": true, "fields": true, "omit": true, "expand": true,
}

// secretNames are properties that can hold a secret. Their data source attributes are sensitive.
var secretNames = map[string]bool{"key": true, "token": true, "secret": true, "password": true}

var validName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type collection struct {
	path     string
	listOp   obj
	item     obj             // read schema of one object
	params   map[string]bool // query parameters, true when the parameter is a list
	singular string
	plural   string
	subcat   string
}

func generate(doc obj) ([]byte, []byte, []byte, []byte, error) {
	paths, _ := doc["paths"].(obj)
	hand := fwprovider.HandwrittenDataSources()
	takenNames := map[string]bool{}
	extraFilters := map[string]map[string]string{}
	byPath := map[string][]bool{}
	handPlural := map[string][]*fwresource.DataSourceSpec{}
	handByPath := map[string][]*fwresource.DataSourceSpec{}
	extras := map[string][]attr{}
	resExtras := map[string][]attr{}
	resByPath := map[string][]*fwresource.Spec{}
	resources, _ := fwprovider.Inventory()
	for _, r := range resources {
		resByPath[r.Path] = append(resByPath[r.Path], r)
	}
	subcats := map[string]string{}
	for _, d := range hand {
		takenNames[d.TypeName] = true
		handByPath[d.Path] = append(handByPath[d.Path], d)
		if d.Plural != nil {
			handPlural[d.Path] = append(handPlural[d.Path], d)
		}
		isPlural := d.Plural != nil
		have := byPath[d.Path]
		if have == nil {
			have = []bool{false, false}
		}
		if isPlural {
			have[1] = true
		} else {
			have[0] = true
		}
		byPath[d.Path] = have
		app := strings.SplitN(d.Path, "/", 2)[0]
		if _, ok := subcats[app]; !ok {
			if rest, found := strings.CutPrefix(d.Description, ":meta:subcategory:"); found {
				if i := strings.Index(rest, ":"); i > 0 {
					subcats[app] = rest[:i]
				}
			}
		}
	}

	var keys []string
	for p := range paths {
		keys = append(keys, p)
	}
	sort.Strings(keys)

	var out bytes.Buffer
	out.WriteString("// Code generated by tools/gendatasources from the NetBox API schema. DO NOT EDIT.\n\n")
	out.WriteString("package fwprovider\n\nimport \"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource\"\n\n")
	out.WriteString("// generatedDataSources are the data sources that are made from the NetBox API schema.\n")
	out.WriteString("var generatedDataSources = []*fwresource.DataSourceSpec{\n")

	for _, p := range keys {
		trimmed := strings.TrimPrefix(p, "/api/")
		if trimmed == p || strings.Contains(trimmed, "{") || !strings.HasSuffix(trimmed, "/") {
			continue
		}
		if _, ok := paths[p+"{id}/"]; !ok || skipped(trimmed) {
			continue
		}
		methods, _ := paths[p].(obj)
		listOp, _ := methods["get"].(obj)
		if listOp == nil {
			continue
		}
		c := &collection{path: trimmed, listOp: listOp, params: map[string]bool{}}
		c.item = resolve(doc, listItem(doc, listOp))
		if c.item == nil {
			continue
		}
		for _, raw := range asList(listOp["parameters"]) {
			par, _ := raw.(obj)
			if par["in"] != "query" {
				continue
			}
			name, _ := par["name"].(string)
			sch, _ := par["schema"].(obj)
			c.params[name] = sch["type"] == "array"
		}
		for _, d := range handPlural[trimmed] {
			extraFilters[d.TypeName] = missingFilters(d, c.params)
		}
		c.names()
		for _, d := range handByPath[trimmed] {
			extras[d.TypeName] = c.extraAttrs(doc, d, addedNames(d.TypeName))
		}
		for _, r := range resByPath[trimmed] {
			resExtras[r.TypeName] = c.resourceExtraAttrs(doc, r, addedResourceNames(r.TypeName))
		}
		c.subcat = subcats[strings.SplitN(trimmed, "/", 2)[0]]
		if c.subcat == "" {
			c.subcat = "Extras"
		}
		have := byPath[trimmed]
		if have == nil {
			have = []bool{false, false}
		}
		if !have[0] && !takenNames["netbox_"+c.singular] {
			takenNames["netbox_"+c.singular] = true
			c.writeSingular(&out, doc)
		}
		if !have[1] && !takenNames["netbox_"+c.plural] {
			takenNames["netbox_"+c.plural] = true
			c.writePlural(&out, doc)
		}
	}
	out.WriteString("}\n")
	src, err := format.Source(out.Bytes())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	filters, err := format.Source(filtersSource(extraFilters))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	extrasSrc, err := extrasSource(extras)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	resSrc, err := resourceExtrasSource(resExtras)
	return src, filters, extrasSrc, resSrc, err
}

func skipped(path string) bool {
	if path == "core/data-sources/" {
		return false
	}
	for _, p := range skippedPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func (c *collection) names() {
	if n, ok := nameOverrides[c.path]; ok {
		c.singular, c.plural = n[0], n[1]
		return
	}
	segs := strings.Split(strings.Trim(c.path, "/"), "/")
	plural := strings.ReplaceAll(segs[len(segs)-1], "-", "_")
	c.plural = plural
	switch {
	case strings.HasSuffix(plural, "ies"):
		c.singular = strings.TrimSuffix(plural, "ies") + "y"
	case strings.HasSuffix(plural, "sses"):
		c.singular = strings.TrimSuffix(plural, "es")
	case strings.HasSuffix(plural, "ss"):
		c.singular = plural
	default:
		c.singular = strings.TrimSuffix(plural, "s")
	}
}

func human(s string) string { return strings.ReplaceAll(s, "_", " ") }

// attr is one attribute in the generated code.
type attr struct {
	name, kind, key, desc, elemKey string
	query                          string
	computed, optional, required   bool
	sensitive                      bool
	zero                           bool
}

func (a attr) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "{Name: %q, Kind: fwresource.%s", a.name, a.kind)
	if a.required {
		b.WriteString(", Required: true")
	}
	if a.optional {
		b.WriteString(", Optional: true")
	}
	if a.computed {
		b.WriteString(", Computed: true")
	}
	if a.zero {
		b.WriteString(", Zero: true")
	}
	if a.sensitive {
		b.WriteString(", Sensitive: true")
	}
	if a.key != "" && a.key != a.name {
		fmt.Fprintf(&b, ", Key: %q", a.key)
	}
	if a.elemKey != "" {
		fmt.Fprintf(&b, ", ElemKey: %q", a.elemKey)
	}
	if a.query != "" {
		fmt.Fprintf(&b, ", Query: %q", a.query)
	}
	fmt.Fprintf(&b, ", Description: %q}", a.desc)
	return b.String()
}

// attrs turns the properties of the read schema into attributes, in name order.
func (c *collection) attrs(doc obj, noun string, plural bool) []attr {
	props, _ := c.item["properties"].(obj)
	var names []string
	for n := range props {
		names = append(names, n)
	}
	sort.Strings(names)
	seen := map[string]bool{}
	var out []attr
	for _, n := range names {
		if n == "url" || n == "display" || n == "display_url" || !validName.MatchString(n) || (n == "id" && !plural) {
			continue
		}
		a, ok := classify(doc, n, resolve(doc, props[n]), noun)
		if !ok || seen[a.name] {
			continue
		}
		seen[a.name] = true
		a.computed = !plural
		a.sensitive = secretNames[a.name]
		a.zero = a.zero && true
		out = append(out, a)
	}
	return out
}

// classify maps one property of the API to an attribute.
func classify(doc obj, name string, p obj, noun string) (attr, bool) {
	if p == nil {
		return attr{}, false
	}
	desc := fmt.Sprintf("The %s of the %s.", human(name), noun)
	typ, _ := p["type"].(string)
	props, _ := p["properties"].(obj)
	switch {
	case name == "created":
		return attr{name: name, kind: "String", zero: true, desc: "When the " + noun + " was created."}, true
	case name == "last_updated":
		return attr{name: name, kind: "String", zero: true, desc: "When the " + noun + " was last changed."}, true
	case strings.HasSuffix(name, "_count") && typ == "integer":
		return attr{name: name, kind: "Int", zero: true, desc: "The number of " + human(strings.TrimSuffix(name, "_count")) + " objects that refer to the " + noun + "."}, true
	}
	switch name {
	case "tags":
		return attr{name: "tags", kind: "StringSet", elemKey: "name", desc: "Tags assigned to the object, by name."}, true
	case "custom_fields":
		return attr{name: name, kind: "StringMap", desc: "Custom field values of the " + noun + ", as a map of field name to value."}, true
	case "comments":
		return attr{name: name, kind: "String", zero: true, desc: "Free-form comments about the " + noun + "."}, true
	case "description":
		return attr{name: name, kind: "String", zero: true, desc: "A description of the " + noun + "."}, true
	}
	switch typ {
	case "string":
		return attr{name: name, kind: "String", zero: true, desc: desc}, true
	case "integer":
		return attr{name: name, kind: "Int", zero: true, desc: desc}, true
	case "number":
		return attr{name: name, kind: "Float", zero: true, desc: desc}, true
	case "boolean":
		return attr{name: name, kind: "Bool", zero: true, desc: desc}, true
	case "array":
		item := resolve(doc, p["items"])
		itemProps, _ := item["properties"].(obj)
		switch {
		case item["type"] == "string":
			return attr{name: name, kind: "StringList", desc: desc}, true
		case item["type"] == "integer":
			return attr{name: name, kind: "IntList", desc: desc}, true
		case itemProps["id"] != nil:
			return attr{name: name, kind: "IntList", desc: "The IDs of the " + human(name) + " of the " + noun + "."}, true
		}
		return attr{}, false
	case "object", "":
		if props["value"] != nil && props["label"] != nil {
			return attr{name: name, kind: "String", zero: true, desc: desc}, true
		}
		if props["id"] != nil {
			return attr{name: name + "_id", kind: "Int", key: name, zero: true, desc: fmt.Sprintf("The ID of the %s associated with the %s.", human(name), noun)}, true
		}
		if len(props) == 0 {
			return attr{name: name, kind: "JSON", desc: desc}, true
		}
	}
	return attr{}, false
}

func (c *collection) hasProp(name string) bool {
	props, _ := c.item["properties"].(obj)
	return props[name] != nil
}

func (c *collection) writeSingular(out *bytes.Buffer, doc obj) {
	noun := human(c.singular)
	attrs := c.attrs(doc, noun, false)
	byName := map[string]int{}
	for i, a := range attrs {
		byName[a.name] = i
	}
	var lookup []string
	idLookup := false
	atLeast := ""
	setLookup := func(name string, optional, required bool) {
		i, ok := byName[name]
		if !ok {
			return
		}
		attrs[i].computed = !required
		attrs[i].optional = optional
		attrs[i].required = required
		attrs[i].zero = false
		lookup = append(lookup, name)
	}
	switch {
	case c.hasProp("name") && c.hasProp("slug"):
		setLookup("name", true, false)
		setLookup("slug", true, false)
		atLeast = `{"id", "name", "slug"}`
		idLookup = true
	case c.hasProp("name"):
		setLookup("name", false, true)
	default:
		// no name: any plain value of the object that NetBox can filter on finds it
		var names []string
		for _, a := range attrs {
			base := strings.TrimSuffix(a.name, "_id")
			switch {
			case notLookups[a.name], a.kind != "String" && a.kind != "Int" && a.kind != "Bool":
			case a.key != "" && c.params[a.key+"_id"]:
				names = append(names, a.name)
			case a.key == "" && c.params[base]:
				names = append(names, a.name)
			}
		}
		for _, n := range names {
			i := byName[n]
			attrs[i].query = ""
			if attrs[i].key != "" {
				attrs[i].query = attrs[i].key + "_id"
			}
			setLookup(n, true, false)
		}
		idLookup = true
		if len(lookup) > 0 {
			parts := make([]string, 0, len(lookup)+1)
			parts = append(parts, `"id"`)
			for _, n := range lookup {
				parts = append(parts, fmt.Sprintf("%q", n))
			}
			atLeast = "{" + strings.Join(parts, ", ") + "}"
		} else {
			atLeast = `{"id"}`
		}
	}
	if idLookup {
		// every object can be found by its number, which is the only way for one that has no name
		id := attr{name: "id", kind: "Int", optional: true, computed: true, desc: fmt.Sprintf("The numeric ID of the %s.", noun)}
		attrs = append([]attr{id}, attrs...)
		for i := range attrs {
			byName[attrs[i].name] = i
		}
	}
	// foreign keys that narrow the search
	for i := range attrs {
		if attrs[i].key != "" && scopeFilters[attrs[i].key] && c.params[attrs[i].key+"_id"] && !contains(lookup, attrs[i].name) {
			attrs[i].optional, attrs[i].computed, attrs[i].zero = true, true, false
			attrs[i].query = attrs[i].key + "_id"
			lookup = append(lookup, attrs[i].name)
		}
	}
	for i := range attrs {
		if attrs[i].query == attrs[i].name {
			attrs[i].query = ""
		}
	}

	fmt.Fprintf(out, "\t{\n\t\tTypeName: %q,\n\t\tPath: %q,\n\t\tNoun: %q,\n", "netbox_"+c.singular, c.path, noun)
	fmt.Fprintf(out, "\t\tDescription: %q,\n", fmt.Sprintf(":meta:subcategory:%s:Retrieves a single %s from NetBox. The lookup fails if no match or more than one match is found.", c.subcat, noun))
	if atLeast != "" {
		fmt.Fprintf(out, "\t\tAtLeastOneOf: [][]string{%s},\n", atLeast)
	}
	out.WriteString("\t\tAttrs: []fwresource.Attr{\n")
	for _, a := range attrs {
		out.WriteString("\t\t\t" + a.String() + ",\n")
	}
	out.WriteString("\t\t},\n\t},\n")
}

func (c *collection) writePlural(out *bytes.Buffer, doc obj) {
	noun := human(c.plural)
	attrs := c.attrs(doc, noun, true)
	var names []string
	for q := range c.params {
		if ignoredQuery[q] || strings.Contains(q, "__") || strings.HasPrefix(q, "cf_") || !validName.MatchString(q) {
			continue
		}
		names = append(names, q)
	}
	sort.Strings(names)

	fmt.Fprintf(out, "\t{\n\t\tTypeName: %q,\n\t\tPath: %q,\n", "netbox_"+c.plural, c.path)
	fmt.Fprintf(out, "\t\tDescription: %q,\n", fmt.Sprintf(":meta:subcategory:%s:Retrieves all %s from NetBox. Narrow the result with filter blocks and cap it with limit.", c.subcat, noun))
	fmt.Fprintf(out, "\t\tPlural: &fwresource.Plural{Result: %q, What: %q, Filters: map[string]string{", c.plural, noun)
	for i, q := range names {
		if i > 0 {
			out.WriteString(", ")
		}
		param := q
		if c.params[q] {
			param += "[]"
		}
		fmt.Fprintf(out, "%q: %q", q, param)
	}
	out.WriteString("}},\n\t\tAttrs: []fwresource.Attr{\n")
	for _, a := range attrs {
		out.WriteString("\t\t\t" + a.String() + ",\n")
	}
	out.WriteString("\t\t},\n\t},\n")
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func asList(v any) []any {
	l, _ := v.([]any)
	return l
}

// resolve follows a $ref and merges allOf.
func resolve(doc obj, v any) obj {
	m, _ := v.(obj)
	for m != nil {
		if ref, ok := m["$ref"].(string); ok {
			name := strings.TrimPrefix(ref, "#/components/schemas/")
			comps, _ := doc["components"].(obj)
			schemas, _ := comps["schemas"].(obj)
			m, _ = schemas[name].(obj)
			continue
		}
		if all := asList(m["allOf"]); len(all) > 0 {
			merged := obj{}
			for k, val := range m {
				if k != "allOf" {
					merged[k] = val
				}
			}
			props := obj{}
			for _, part := range all {
				p := resolve(doc, part)
				if pp, ok := p["properties"].(obj); ok {
					for k, val := range pp {
						props[k] = val
					}
				}
				if t, ok := p["type"]; ok && merged["type"] == nil {
					merged["type"] = t
				}
			}
			if own, ok := m["properties"].(obj); ok {
				for k, val := range own {
					props[k] = val
				}
			}
			merged["properties"] = props
			return merged
		}
		break
	}
	return m
}

// listItem finds the schema of one object in the response of a list operation.
func listItem(doc obj, op obj) obj {
	responses, _ := op["responses"].(obj)
	ok, _ := responses["200"].(obj)
	content, _ := ok["content"].(obj)
	js, _ := content["application/json"].(obj)
	page := resolve(doc, js["schema"])
	props, _ := page["properties"].(obj)
	results, _ := props["results"].(obj)
	item, _ := results["items"].(obj)
	return item
}
