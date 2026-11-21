package fwresource

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// groupKind is the rule of a group of attributes.
type groupKind int

const (
	groupConflicts groupKind = iota
	groupExactlyOne
	groupAtLeastOne
)

// groupValidator checks which attributes of a group are set. An optional attribute that has the
// zero value (an empty string, 0 or false) counts as not set, because that is what the state and
// the configuration generated from it hold for an attribute that is not set.
type groupValidator struct {
	kind  groupKind
	names []string
	zero  map[string]bool
}

func (v groupValidator) Description(context.Context) string {
	switch v.kind {
	case groupConflicts:
		return fmt.Sprintf("at most one of %s can be set", strings.Join(v.names, ", "))
	case groupExactlyOne:
		return fmt.Sprintf("exactly one of %s must be set", strings.Join(v.names, ", "))
	}
	return fmt.Sprintf("at least one of %s must be set", strings.Join(v.names, ", "))
}

func (v groupValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v groupValidator) ValidateResource(_ context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var obj map[string]tftypes.Value
	if err := req.Config.Raw.As(&obj); err != nil {
		return
	}
	var set []string
	for _, n := range v.names {
		val, ok := obj[n]
		if !ok {
			continue
		}
		if !val.IsKnown() {
			// the value is not known yet, so the group cannot be judged
			return
		}
		if isSetValue(val, v.zero[n]) {
			set = append(set, n)
		}
	}
	list := strings.Join(v.names, ",")
	switch v.kind {
	case groupConflicts:
		if len(set) > 1 {
			resp.Diagnostics.AddAttributeError(path.Root(set[0]), "Invalid Attribute Combination",
				fmt.Sprintf("These attributes cannot be configured together: [%s]", list))
		}
	case groupExactlyOne:
		if len(set) != 1 {
			resp.Diagnostics.AddAttributeError(path.Root(v.names[0]), "Invalid Attribute Combination",
				fmt.Sprintf("Exactly one of these attributes must be configured: [%s]", list))
		}
	case groupAtLeastOne:
		if len(set) == 0 {
			resp.Diagnostics.AddAttributeError(path.Root(v.names[0]), "Missing Attribute Configuration",
				fmt.Sprintf("At least one of these attributes must be configured: [%s]", list))
		}
	}
}

// isSetValue tells whether a configuration value counts as set. With zero, the zero value of the
// type counts as not set.
func isSetValue(v tftypes.Value, zero bool) bool {
	if v.IsNull() {
		return false
	}
	if !zero {
		return true
	}
	switch {
	case v.Type().Is(tftypes.String):
		var s string
		return v.As(&s) != nil || s != ""
	case v.Type().Is(tftypes.Number):
		var n big.Float
		return v.As(&n) != nil || n.Sign() != 0
	case v.Type().Is(tftypes.Bool):
		var b bool
		return v.As(&b) != nil || b
	case v.Type().Is(tftypes.Set{}), v.Type().Is(tftypes.List{}), v.Type().Is(tftypes.Map{}):
		var elems []tftypes.Value
		if v.Type().Is(tftypes.Map{}) {
			var m map[string]tftypes.Value
			return v.As(&m) != nil || len(m) > 0
		}
		return v.As(&elems) != nil || len(elems) > 0
	}
	return true
}
