package fwresource

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ScopeOption is one of the attributes that can point at the object of a Scope.
type ScopeOption struct {
	// Attr is the attribute that holds the ID of the object.
	Attr string
	// Type is the NetBox object type the attribute stands for, for example dcim.site.
	Type string
	// TypeAttr names the attribute that holds the object type, for an Attr that can point at
	// objects of several types. Type is not used then.
	TypeAttr string
}

// Scope maps a group of attributes to the pair of fields NetBox uses for an object of any type: one
// with the type and one with the ID. Only one of the attributes is set.
type Scope struct {
	TypeKey string
	IDKey   string
	Options []ScopeOption
	// ByPrior reads only the attributes that are set in the configuration, and none on import,
	// since the object type alone does not tell which attribute was used.
	ByPrior bool
	// GenericFallback sets the attribute for any type when none of the attributes was set, as on import.
	GenericFallback bool
	// BlankType sends an empty string instead of null as the type when none of the attributes is set.
	BlankType bool
}

// virtualAttrs returns the attributes that are handled by scopes, and not like other attributes.
func (s *Spec) virtualAttrs() map[string]bool {
	out := map[string]bool{}
	for _, sc := range s.Scopes {
		for _, o := range sc.Options {
			out[o.Attr] = true
			if o.TypeAttr != "" {
				out[o.TypeAttr] = true
			}
		}
	}
	return out
}

func (s *Spec) attr(name string) (Attr, bool) {
	for _, a := range s.Attrs {
		if a.Name == name {
			return a, true
		}
	}
	return Attr{}, false
}

func intOf(v attr.Value) (int64, bool) {
	if n, ok := v.(types.Int64); ok && !n.IsNull() && !n.IsUnknown() && n.ValueInt64() != 0 {
		return n.ValueInt64(), true
	}
	return 0, false
}

func stringOf(v attr.Value) string {
	if x, ok := v.(types.String); ok && !x.IsNull() && !x.IsUnknown() {
		return x.ValueString()
	}
	return ""
}

// scopeBody adds the type and ID fields of all scopes to the body.
func (s *Spec) scopeBody(plan Values, body map[string]any) {
	for _, sc := range s.Scopes {
		var typ any
		var id any
		if sc.BlankType {
			typ = ""
		}
		for _, o := range sc.Options {
			n, ok := intOf(plan[o.Attr])
			if !ok {
				continue
			}
			id = n
			if o.TypeAttr != "" {
				typ = stringOf(plan[o.TypeAttr])
			} else {
				typ = o.Type
			}
			break
		}
		body[sc.TypeKey] = typ
		body[sc.IDKey] = id
	}
}

// scopeState sets the attributes of all scopes from the object.
func (s *Spec) scopeState(obj map[string]any, prior, out Values) {
	for _, sc := range s.Scopes {
		for _, o := range sc.Options {
			if a, ok := s.attr(o.Attr); ok {
				out[o.Attr] = types.Int64Value(0)
				_ = a
			}
			if o.TypeAttr != "" {
				out[o.TypeAttr] = types.StringValue("")
			}
		}
		actualType, _ := obj[sc.TypeKey].(string)
		id, hasID := intFromJSON(obj[sc.IDKey])
		if !hasID || actualType == "" {
			continue
		}

		priorHas := false
		for _, o := range sc.Options {
			if _, set := intOf(prior[o.Attr]); set {
				priorHas = true
			}
		}
		for _, o := range sc.Options {
			_, set := intOf(prior[o.Attr])
			switch {
			case o.TypeAttr != "":
				if set || (sc.GenericFallback && !priorHas) {
					out[o.Attr] = types.Int64Value(id)
					out[o.TypeAttr] = types.StringValue(actualType)
				}
			case o.Type == actualType && (!sc.ByPrior || set):
				out[o.Attr] = types.Int64Value(id)
			}
		}
	}
}
