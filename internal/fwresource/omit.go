package fwresource

import (
	"net/url"
	"strings"
)

// heavyFields are fields that cost NetBox extra work for every object it returns, by API path. The
// config context of a device or virtual machine is put together from all config contexts that apply
// to it, and can be much larger than the rest of the object.
var heavyFields = map[string][]string{
	"dcim/devices/":                    {"config_context"},
	"virtualization/virtual-machines/": {"config_context"},
}

// omitQuery returns the query parameters that ask NetBox to leave out the heavy fields of the
// objects at path that the caller does not read. used tells whether a field is read. The result is
// nil when there is nothing to leave out.
//
// NetBox 4.5 and later take the fields in omit, older versions take them in exclude. A version
// ignores the parameter it does not know, so both are sent.
func omitQuery(path string, used func(field string) bool) url.Values {
	var omit []string
	for _, f := range heavyFields[path] {
		if used == nil || !used(f) {
			omit = append(omit, f)
		}
	}
	if len(omit) == 0 {
		return nil
	}
	fields := strings.Join(omit, ",")
	return url.Values{"omit": {fields}, "exclude": {fields}}
}

// usesField reports whether one of the attributes reads the API field.
func usesField(attrs []Attr, field string) bool {
	for _, a := range attrs {
		if a.key() == field || (len(a.Path) > 0 && a.Path[0] == field) {
			return true
		}
	}
	return false
}

// readQuery returns the query parameters the resource sends with every request for its object.
func (s *Spec) readQuery() url.Values {
	return omitQuery(s.Path, func(f string) bool { return usesField(s.Attrs, f) })
}
