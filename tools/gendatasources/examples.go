package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwprovider"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
)

const exampleFile = "data-source.tf"

// writeExamples writes an example for every data source that has none. Examples that an earlier run
// wrote are written again, since the data source can have changed. Examples that someone wrote are
// never touched. The names of the ones that are made here are in the list file.
func writeExamples(dir, listFile string) error {
	generated := map[string]bool{}
	if f, err := os.Open(listFile); err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if line := strings.TrimSpace(sc.Text()); line != "" {
				generated[line] = true
			}
		}
		f.Close()
	}

	_, dataSources := fwprovider.Inventory()
	next := map[string]bool{}
	for _, d := range dataSources {
		path := filepath.Join(dir, d.TypeName, exampleFile)
		_, err := os.Stat(path)
		exists := err == nil
		if exists && !generated[d.TypeName] {
			continue // written by hand
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(exampleFor(d)), 0o644); err != nil {
			return err
		}
		next[d.TypeName] = true
	}

	var names []string
	for n := range next {
		names = append(names, n)
	}
	sort.Strings(names)
	var b bytes.Buffer
	b.WriteString("# The data sources whose examples tools/gendatasources writes. Examples that are not listed here are\n")
	b.WriteString("# written by hand and are left as they are.\n")
	for _, n := range names {
		b.WriteString(n + "\n")
	}
	return os.WriteFile(listFile, b.Bytes(), 0o644)
}

func exampleValue(a fwresource.Attr) (string, bool) {
	switch a.Kind {
	case fwresource.String:
		if len(a.OneOf) > 0 {
			return fmt.Sprintf("%q", a.OneOf[0]), true
		}
		return `"example"`, true
	case fwresource.Int:
		return "1", true
	case fwresource.Float:
		return "1", true
	case fwresource.Bool:
		return "true", true
	}
	return "", false
}

// exampleFor makes the example of one data source.
func exampleFor(d *fwresource.DataSourceSpec) string {
	var b strings.Builder
	switch {
	case d.Plural != nil:
		writePluralExample(&b, d)
	case len(d.FilterBlock) > 0:
		fmt.Fprintf(&b, "data %q \"example\" {\n  filter {\n", d.TypeName)
		for _, a := range d.FilterBlock {
			if v, ok := exampleValue(a); ok {
				fmt.Fprintf(&b, "    %s = %s\n", a.Name, v)
				break
			}
		}
		b.WriteString("  }\n}\n")
	default:
		writeSingularExample(&b, d)
	}
	return b.String()
}

func writePluralExample(b *strings.Builder, d *fwresource.DataSourceSpec) {
	p := d.Plural
	fmt.Fprintf(b, "data %q \"example\" {\n", d.TypeName)
	if !p.NoFilter && len(p.Filters) > 0 {
		var names []string
		for n := range p.Filters {
			names = append(names, n)
		}
		sort.Strings(names)
		name := names[0]
		for _, preferred := range []string{"name", "slug", "status"} {
			if _, ok := p.Filters[preferred]; ok {
				name = preferred
				break
			}
		}
		fmt.Fprintf(b, "  filter {\n    name  = %q\n    value = \"example\"\n  }\n", name)
	}
	if !p.NoFilter {
		b.WriteString("  limit = 10\n")
	}
	b.WriteString("}\n")
	for _, a := range d.Attrs {
		if a.Name == "id" {
			fmt.Fprintf(b, "\noutput \"%s_ids\" {\n  value = data.%s.example.%s[*].id\n}\n", strings.TrimPrefix(d.TypeName, "netbox_"), d.TypeName, p.Result)
			break
		}
	}
}

func writeSingularExample(b *strings.Builder, d *fwresource.DataSourceSpec) {
	var lines []string
	chosen := map[string]bool{}
	for _, a := range d.Attrs {
		if a.Required {
			if v, ok := exampleValue(a); ok {
				lines = append(lines, fmt.Sprintf("%s = %s", a.Name, v))
				chosen[a.Name] = true
			}
		}
	}
	if len(lines) == 0 {
		// one of the values that find the object
		var want []string
		if len(d.AtLeastOneOf) > 0 {
			want = d.AtLeastOneOf[0]
		} else {
			for _, a := range d.Attrs {
				if a.Optional {
					want = append(want, a.Name)
				}
			}
		}
		for _, name := range want {
			for _, a := range d.Attrs {
				if a.Name != name {
					continue
				}
				if v, ok := exampleValue(a); ok {
					lines = append(lines, fmt.Sprintf("%s = %s", a.Name, v))
					chosen[a.Name] = true
				}
			}
			if len(lines) > 0 {
				break
			}
		}
	}
	// the parent that narrows the search, such as the device of an interface
	for _, parent := range []string{"device", "virtual_machine", "module", "circuit", "provider", "rack", "site", "cluster", "policy", "group", "l2vpn"} {
		added := false
		for _, a := range d.Attrs {
			if a.Optional && !chosen[a.Name] && a.Kind == fwresource.Int && a.Key == parent && len(lines) > 0 {
				lines = append(lines, fmt.Sprintf("%s = 1", a.Name))
				added = true
				break
			}
		}
		if added {
			break
		}
	}
	// the equals signs line up
	width := 0
	for _, l := range lines {
		if i := strings.Index(l, " = "); i > width {
			width = i
		}
	}
	for i, l := range lines {
		k := strings.Index(l, " = ")
		lines[i] = l[:k] + strings.Repeat(" ", width-k) + l[k:]
	}
	fmt.Fprintf(b, "data %q \"example\" {", d.TypeName)
	if len(lines) == 0 {
		b.WriteString("}\n")
	} else {
		b.WriteString("\n")
		for _, l := range lines {
			b.WriteString("  " + l + "\n")
		}
		b.WriteString("}\n")
	}
	fmt.Fprintf(b, "\noutput \"%s_id\" {\n  value = data.%s.example.id\n}\n", strings.TrimPrefix(d.TypeName, "netbox_"), d.TypeName)
}
