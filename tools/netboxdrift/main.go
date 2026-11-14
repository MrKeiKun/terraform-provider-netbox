// Command netboxdrift compares the NetBox OpenAPI schema with the resources and data sources of the
// provider and lists what the provider does not cover yet: collections without a resource or data
// source, writable fields without an attribute, and list filters without a data source filter.
//
//	go run ./tools/netboxdrift -url http://localhost:8001 -token $NETBOX_API_TOKEN
//	go run ./tools/netboxdrift -schema schema.json -baseline tools/netboxdrift/baseline.txt
//
// With -baseline the command only fails for drift that is not in the baseline, so known gaps do not
// break CI but new ones do. -write-baseline writes the current drift to that file.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwprovider"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
)

// ignoredFields are writable API fields that no attribute is expected for.
var ignoredFields = map[string]bool{
	"id": true, "display": true, "url": true, "created": true, "last_updated": true,
	"custom_field_data": true, "owner": false,
}

// ignoredQuery are list parameters every collection has, or that only make sense for the API.
var ignoredQuery = map[string]bool{
	"limit": true, "offset": true, "ordering": true, "brief": true, "fields": true, "omit": true,
	"id": true, "created": true, "last_updated": true, "created_by_request": true,
	"updated_by_request": true, "modified_by_request": true, "q": true, "config_template_id": false,
}

// notManaged are writable API fields that are on purpose not attributes, with the reason.
var notManaged = map[string]string{
	"netbox_device_type front_image": "file upload, the API only takes multipart form data",
	"netbox_device_type rear_image":  "file upload, the API only takes multipart form data",
	"netbox_token pepper_id":         "internal, NetBox picks the pepper",
	"netbox_token version":           "internal, NetBox sets the token version",
	"netbox_user date_joined":        "set by NetBox when the account is created",
	"netbox_user last_login":         "set by NetBox at login",
}

// notProvided are API collections that have no resource or data source on purpose, with the reason.
var notProvided = map[string]string{
	"resource extras/image-attachments/":                "file upload, the API only takes multipart form data; netbox_image_attachment data sources exist",
	"datasource virtualization/virtual-disks/ singular": "netbox_virtual_disk has always been the list data source, its name stays for compatibility",
}

// skippedCollections are API collections that are not objects a user manages.
var skippedPrefixes = []string{"core/", "extras/scripts/", "extras/script-modules/", "plugins/"}

func main() {
	var (
		schemaFile = flag.String("schema", "", "OpenAPI document to read")
		serverURL  = flag.String("url", "", "NetBox server to fetch /api/schema/ from")
		token      = flag.String("token", os.Getenv("NETBOX_API_TOKEN"), "API token for -url")
		baseline   = flag.String("baseline", "", "file with known drift; only other drift fails the run")
		writeBase  = flag.Bool("write-baseline", false, "write the current drift to -baseline")
		versions   = flag.String("versions", "", "with -write-versions: version=schema file pairs, for example 4.4.0=a.json,4.7.0=b.json")
		writeVer   = flag.String("write-versions", "", "write the table of the first NetBox version of each resource and attribute to this file")
		section    = flag.String("section", "", "only print this kind: resource, datasource, field, filter")
	)
	flag.Parse()

	if *writeVer != "" {
		files, err := parseVersions(*versions)
		if err != nil {
			fatal(err)
		}
		if err := writeVersions(files, *writeVer); err != nil {
			fatal(err)
		}
		return
	}

	data, err := loadSchema(*schemaFile, *serverURL, *token)
	if err != nil {
		fatal(err)
	}
	endpoints, err := parseEndpoints(data)
	if err != nil {
		fatal(err)
	}

	items := compare(endpoints)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Kind != items[j].Kind {
			return items[i].Kind < items[j].Kind
		}
		return items[i].Key() < items[j].Key()
	})

	if *writeBase {
		if *baseline == "" {
			fatal(fmt.Errorf("-write-baseline needs -baseline"))
		}
		var b strings.Builder
		for _, it := range items {
			b.WriteString(it.Key() + "\n")
		}
		if err := os.WriteFile(*baseline, []byte(b.String()), 0o644); err != nil {
			fatal(err)
		}
		fmt.Printf("wrote %d entries to %s\n", len(items), *baseline)
		return
	}

	known := map[string]bool{}
	if *baseline != "" {
		raw, err := os.ReadFile(*baseline)
		if err != nil {
			fatal(err)
		}
		for _, line := range strings.Split(string(raw), "\n") {
			if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
				known[line] = true
			}
		}
	}

	counts := map[string]int{}
	var fresh []item
	for _, it := range items {
		counts[it.Kind]++
		if *section != "" && it.Kind != *section {
			continue
		}
		mark := " "
		if *baseline != "" && !known[it.Key()] {
			mark = "+"
			fresh = append(fresh, it)
		}
		fmt.Printf("%s %s\n", mark, it.Key())
	}
	fmt.Printf("\nmissing resources: %d, data sources: %d, fields: %d, filters: %d\n",
		counts["resource"], counts["datasource"], counts["field"], counts["filter"])
	if len(fresh) > 0 {
		fmt.Fprintf(os.Stderr, "%d new drift entries that are not in %s\n", len(fresh), *baseline)
		os.Exit(1)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "netboxdrift:", err)
	os.Exit(2)
}

// item is one gap between NetBox and the provider.
type item struct {
	Kind string // resource, datasource, field, filter
	Path string // API collection
	Name string // field, filter, or data source flavour (singular, plural)
	Spec string // type name of the resource or data source that lacks it
}

func (i item) Key() string {
	switch i.Kind {
	case "resource":
		return fmt.Sprintf("resource %s", i.Path)
	case "datasource":
		return fmt.Sprintf("datasource %s %s", i.Path, i.Name)
	default:
		return fmt.Sprintf("%s %s %s %s", i.Kind, i.Spec, i.Path, i.Name)
	}
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

func compare(endpoints []*endpoint) []item {
	resources, dataSources := fwprovider.Inventory()

	resByPath := map[string][]*fwresource.Spec{}
	for _, r := range resources {
		resByPath[r.Path] = append(resByPath[r.Path], r)
	}
	dsByPath := map[string][]*fwresource.DataSourceSpec{}
	for _, d := range dataSources {
		dsByPath[d.Path] = append(dsByPath[d.Path], d)
	}
	byPath := map[string]*endpoint{}
	for _, ep := range endpoints {
		byPath[ep.Path] = ep
	}

	var out []item
	for _, ep := range endpoints {
		if !ep.HasDetail || skipped(ep.Path) {
			continue
		}
		if ep.CanCreate && len(resByPath[ep.Path]) == 0 && notProvided["resource "+ep.Path] == "" {
			out = append(out, item{Kind: "resource", Path: ep.Path})
		}
		var singular, plural bool
		for _, d := range dsByPath[ep.Path] {
			if d.Plural != nil {
				plural = true
			} else {
				singular = true
			}
		}
		if !singular && notProvided["datasource "+ep.Path+" singular"] == "" {
			out = append(out, item{Kind: "datasource", Path: ep.Path, Name: "singular"})
		}
		if !plural {
			out = append(out, item{Kind: "datasource", Path: ep.Path, Name: "plural"})
		}
	}

	for path, specs := range resByPath {
		ep := byPath[path]
		if ep == nil {
			continue
		}
		for _, s := range specs {
			covered := resourceKeys(s)
			for f := range ep.Fields {
				if !covered[f] && !ignoredFields[f] && notManaged[s.TypeName+" "+f] == "" {
					out = append(out, item{Kind: "field", Path: path, Name: f, Spec: s.TypeName})
				}
			}
		}
	}

	for path, specs := range dsByPath {
		ep := byPath[path]
		if ep == nil {
			continue
		}
		for _, d := range specs {
			if d.Plural == nil {
				continue // a single object lookup only needs the filters that find one object
			}
			covered := filterKeys(d)
			for q := range ep.Query {
				base := q
				if i := strings.Index(q, "__"); i >= 0 {
					continue // lookup modifiers such as __n or __ic are covered by their base filter
				}
				if ignoredQuery[base] || covered[base] || covered[strings.TrimSuffix(base, "_id")] {
					continue
				}
				// NetBox names a foreign key filter foo or foo_id; either one covers the other.
				if covered[base+"_id"] {
					continue
				}
				out = append(out, item{Kind: "filter", Path: path, Name: q, Spec: d.TypeName})
			}
		}
	}
	return out
}

// resourceKeys returns the API fields a resource has an attribute for.
func resourceKeys(s *fwresource.Spec) map[string]bool {
	out := map[string]bool{}
	addAttrs(out, s.Attrs)
	if s.Tags {
		out["tags"] = true
	}
	if s.CustomFields {
		out["custom_fields"] = true
	}
	for _, sc := range s.Scopes {
		out[sc.TypeKey] = true
		out[sc.IDKey] = true
	}
	return out
}

func addAttrs(out map[string]bool, attrs []fwresource.Attr) {
	for _, a := range attrs {
		if a.Local {
			continue
		}
		out[a.Name] = true
		if a.Key != "" {
			out[a.Key] = true
		}
		if len(a.Path) > 0 {
			out[a.Path[0]] = true
		}
	}
}

// filterKeys returns the query parameters a data source can filter on.
func filterKeys(d *fwresource.DataSourceSpec) map[string]bool {
	out := map[string]bool{}
	for _, list := range [][]fwresource.Attr{d.Attrs, d.ExtraFilters, d.FilterBlock} {
		for _, a := range list {
			if a.Computed && !a.Optional {
				continue
			}
			out[a.Name] = true
			if a.Key != "" {
				out[a.Key] = true
			}
			if a.Query != "" {
				out[strings.TrimSuffix(a.Query, "[]")] = true
			}
		}
	}
	if d.Plural != nil {
		for name, q := range d.Plural.Filters {
			out[name] = true
			out[strings.TrimSuffix(q, "[]")] = true
		}
	}
	if d.CustomFields {
		out["cf"] = true
	}
	return out
}
