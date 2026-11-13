package fwprovider

import (
	"context"
	"fmt"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
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
			{Name: "extra_choices", Kind: fwresource.StringLists, Optional: true, Description: "This length of the inner lists must be exactly two, where the first value is the value of a choice and the second value is the label of the choice."},
			{Name: "name", Kind: fwresource.String, Required: true, Description: "The name of the custom field choice set."},
			{Name: "order_alphabetically", Kind: fwresource.Bool, Optional: true, Default: false, Description: "experimental"},
			{Name: "owner_id", Kind: fwresource.Int, Key: "owner", Ref: true, Nullable: true, Optional: true, Description: "The ID of the owner associated with the custom field choice set."},
		},
	},
}
