package fwprovider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// manualResources are the resources that do not fit the usual pattern and are described by hand.
var manualResources = []*fwresource.Spec{
	{
		TypeName:     "netbox_vlan_group",
		Path:         "ipam/vlan-groups/",
		Description:  ":meta:subcategory:IP Address Management (IPAM):\n> A VLAN Group represents a collection of VLANs. Generally, these are limited by one of a number of scopes such as \"Site\" or \"Virtualization Cluster\".",
		Tags:         true,
		CustomFields: true,
		Put:          true,
		// NetBox 4.4 reports the end of every VLAN ID range one too low in the response to a write.
		Reload:       true,
		RequiredWith: [][]string{{"scope_id", "scope_type"}},
		Attrs: []fwresource.Attr{
			{Name: "comments", Kind: fwresource.String, Optional: true, Description: "Free-form comments about the vlan group."},
			{Name: "description", Kind: fwresource.String, Optional: true, Description: "A description of the vlan group. Defaults to `\"\"`."},
			{Name: "name", Kind: fwresource.String, Required: true, Description: "The name of the vlan group."},
			{Name: "owner_id", Kind: fwresource.Int, Key: "owner", Ref: true, Nullable: true, Optional: true, Description: "The ID of the owner associated with the vlan group."},
			{Name: "scope_id", Kind: fwresource.Int, Nullable: true, Optional: true, Description: "The ID of the scope associated with the vlan group. Required when `scope_type` is set."},
			{Name: "scope_type", Kind: fwresource.String, Nullable: true, Optional: true, Description: "Valid values are `dcim.location`, `dcim.site`, `dcim.sitegroup`, `dcim.region`, `dcim.rack`, `virtualization.cluster` and `virtualization.clustergroup`.", OneOf: []string{"dcim.location", "dcim.site", "dcim.sitegroup", "dcim.region", "dcim.rack", "virtualization.cluster", "virtualization.clustergroup"}},
			{Name: "slug", Kind: fwresource.String, Required: true, Description: "The URL-friendly slug of the vlan group.", MinLen: 1, MaxLen: 100},
			{Name: "tenant_id", Kind: fwresource.Int, Key: "tenant", Ref: true, Nullable: true, Optional: true, Description: "The ID of the tenant associated with the vlan group."},
			{Name: "vid_ranges", Kind: fwresource.IntLists, Required: true, Description: "The vid ranges of the vlan group."},
		},
	},
	{
		TypeName:     "netbox_custom_field_choice_set",
		Path:         "extras/custom-field-choice-sets/",
		Description:  ":meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/extras/customfieldchoiceset/):\nSingle- and multi-selection custom fields must define a set of valid choices from which the user may choose when defining the field value. These choices are defined as sets that may be reused among multiple custom fields.\nA choice set must define a base choice set and/or a set of arbitrary extra choices.",
		CustomFields: true,
		AtLeastOneOf: [][]string{{"base_choices", "extra_choices"}},
		// NetBox sorts the choices by their value when order_alphabetically is on
		ValidateConfig: choiceSetOrder,
		ModifyBody: func(_ context.Context, plan, _ fwresource.Values, _ map[string]any) error {
			if l, ok := plan["extra_choices"].(types.List); ok && !l.IsNull() && !l.IsUnknown() {
				for _, e := range l.Elements() {
					if inner, ok := e.(types.List); ok && len(inner.Elements()) != 2 {
						return fmt.Errorf("length of inner lists must be exactly two for custom field choice sets")
					}
				}
			}
			return nil
		},
		Attrs: []fwresource.Attr{
			{Name: "base_choices", Kind: fwresource.String, Nullable: true, Optional: true, Description: "Valid values are `IATA`, `ISO_3166` and `UN_LOCODE`.", OneOf: []string{"IATA", "ISO_3166", "UN_LOCODE"}},
			{Name: "choice_colors", Kind: fwresource.StringMap, Optional: true, Description: "The color of each choice, as a map of the choice to a color name such as `blue`, `red` or `green`."},
			{Name: "description", Kind: fwresource.String, Optional: true, Description: "A description of the custom field choice set."},
			{Name: "extra_choices", Kind: fwresource.StringLists, Optional: true, Description: "This length of the inner lists must be exactly two, where the first value is the value of a choice and the second value is the label of the choice. With `order_alphabetically` NetBox sorts the choices, so write them in that order."},
			{Name: "name", Kind: fwresource.String, Required: true, Description: "The name of the custom field choice set."},
			{Name: "order_alphabetically", Kind: fwresource.Bool, Optional: true, Default: false, Description: "NetBox keeps `extra_choices` sorted by the value of each choice. Write the choices in that order, or the plan stops with an error."},
			{Name: "owner_id", Kind: fwresource.Int, Key: "owner", Ref: true, Nullable: true, Optional: true, Description: "The ID of the owner associated with the custom field choice set."},
		},
	},
}

// choiceSetOrder checks that extra_choices is written in the order that NetBox keeps it in when
// order_alphabetically is on: sorted by the value of each choice. NetBox sorts the list itself, and
// a list in another order would come back changed, which Terraform reports as an inconsistent result.
func choiceSetOrder(_ context.Context, cfg fwresource.Values) diag.Diagnostics {
	on, ok := cfg["order_alphabetically"].(types.Bool)
	if !ok || on.IsNull() || on.IsUnknown() || !on.ValueBool() {
		return nil
	}
	list, ok := cfg["extra_choices"].(types.List)
	if !ok || list.IsNull() || list.IsUnknown() {
		return nil
	}
	var values []string
	for _, e := range list.Elements() {
		inner, ok := e.(types.List)
		if !ok || inner.IsNull() || inner.IsUnknown() || len(inner.Elements()) == 0 {
			return nil
		}
		first, ok := inner.Elements()[0].(types.String)
		if !ok || first.IsNull() || first.IsUnknown() {
			return nil
		}
		values = append(values, first.ValueString())
	}
	for i := 1; i < len(values); i++ {
		if values[i] >= values[i-1] {
			continue
		}
		sorted := append([]string(nil), values...)
		sort.Strings(sorted)
		var d diag.Diagnostics
		d.AddAttributeError(path.Root("extra_choices"), "extra_choices is not in the order NetBox keeps",
			fmt.Sprintf("order_alphabetically is true, so NetBox keeps extra_choices sorted by the value of each choice, and the choice %q would move before %q. "+
				"Terraform would then report a difference after the apply.\n\n"+
				"Write the choices in this order of their values: %s.\n\n"+
				"Or set order_alphabetically to false to keep your own order.",
				values[i], values[i-1], strings.Join(sorted, ", ")))
		return d
	}
	return nil
}
