package fwprovider

import "github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"

// customDataSources are data sources that look up their objects with code of their own.
var customDataSources = []*fwresource.DataSourceSpec{
	{
		TypeName:    "netbox_available_prefix",
		Description: ":meta:subcategory:IP Address Management (IPAM):Retrieves a single available prefix from NetBox. The lookup fails if no match or more than one match is found.",
		ExtraFilters: []fwresource.Attr{
			{Name: "prefix_id", Kind: fwresource.Int, Required: true, Query: "-", Description: "The ID of the prefix to list the available prefixes of."},
		},
		Plural: &fwresource.Plural{Result: "prefixes_available", What: "prefixes_available", Filters: map[string]string{}},
		Attrs: []fwresource.Attr{
			{Name: "family", Kind: fwresource.Int, Zero: true, Description: "The IP family, 4 or 6."},
			{Name: "prefix", Kind: fwresource.String, Zero: true, Description: "The prefix in CIDR notation."},
			{Name: "vrf_id", Kind: fwresource.Int, Zero: true, Description: "The ID of the vrf associated with the available prefix."},
		},
	},
	{
		TypeName:    "netbox_device_render_config",
		Description: ":meta:subcategory:Data Center Inventory Management (DCIM):Render the configuration template assigned to a device using the device's config context.",
		Noun:        "device_render_config",
		Attrs: []fwresource.Attr{
			{Name: "config_template_id", Kind: fwresource.Int, Computed: true, Description: "The ID of the config template that was used for rendering."},
			{Name: "config_template_name", Kind: fwresource.String, Computed: true, Description: "The name of the config template that was used for rendering."},
			{Name: "content", Kind: fwresource.String, Computed: true, Description: "The rendered configuration content."},
			{Name: "device_id", Kind: fwresource.Int, Required: true, Description: "The ID of the device to render configuration for."},
		},
	},
	{
		TypeName:    "netbox_ip_address",
		Description: ":meta:subcategory:IP Address Management (IPAM):Retrieves a single ip address from NetBox. The lookup fails if no match or more than one match is found.",
		Noun:        "ip_address",
		Attrs: []fwresource.Attr{
			{Name: "address_family", Kind: fwresource.String, Computed: true, Description: "The address family of the ip address."},
			{Name: "created", Kind: fwresource.String, Computed: true, Description: "The created of the ip address."},
			{Name: "custom_fields", Kind: fwresource.StringMap, Computed: true, Description: "Custom field values of the ip address, as a map of field name to value."},
			{Name: "description", Kind: fwresource.String, Computed: true, Description: "A description of the ip address."},
			{Name: "dns_name", Kind: fwresource.String, Computed: true, Description: "The DNS name of the ip address."},
			{Name: "id", Kind: fwresource.Int, Required: true, Description: "The numeric ID of the ip address."},
			{Name: "ip_address", Kind: fwresource.String, Computed: true, Description: "The IP address of the ip address, in CIDR notation."},
			{Name: "last_updated", Kind: fwresource.String, Computed: true, Description: "The last updated of the ip address."},
			{Name: "role", Kind: fwresource.String, Computed: true, Description: "The role of the ip address."},
			{Name: "status", Kind: fwresource.String, Computed: true, Description: "The status of the ip address."},
			{Name: "tags", Kind: fwresource.NestedList, Computed: true, Sub: []fwresource.Attr{
				{Name: "display", Kind: fwresource.String, Zero: true, Description: "The display of the ip address."},
				{Name: "id", Kind: fwresource.Int, Zero: true, Description: "The numeric ID of the ip address."},
				{Name: "name", Kind: fwresource.String, Zero: true, Description: "The name of the ip address."},
				{Name: "slug", Kind: fwresource.String, Zero: true, Description: "The URL-friendly slug of the ip address."},
			}, Description: "Tags to assign to the ip address."},
			{Name: "tenant", Kind: fwresource.NestedList, Computed: true, Sub: []fwresource.Attr{
				{Name: "id", Kind: fwresource.Int, Zero: true, Description: "The numeric ID of the ip address."},
				{Name: "name", Kind: fwresource.String, Zero: true, Description: "The name of the ip address."},
				{Name: "slug", Kind: fwresource.String, Zero: true, Description: "The URL-friendly slug of the ip address."},
			}, Description: "The tenant of the ip address."},
		},
	},
	{
		TypeName:    "netbox_notification_group",
		Description: ":meta:subcategory:Extras:Retrieves a single notification group from NetBox. The lookup fails if no match or more than one match is found.",
		Noun:        "notification_group",
		Attrs: []fwresource.Attr{
			{Name: "description", Kind: fwresource.String, Computed: true, Description: "A description of the notification group."},
			{Name: "group_ids", Kind: fwresource.IntSet, Computed: true, Description: "The IDs of the group associated with the notification group."},
			{Name: "name", Kind: fwresource.String, Optional: true, Query: "-", Computed: true, Description: "The name of the notification group. At least one of `name` must be given."},
			{Name: "user_ids", Kind: fwresource.IntSet, Computed: true, Description: "The IDs of the user associated with the notification group."},
		},
	},
	{
		TypeName:    "netbox_rack_type",
		Description: ":meta:subcategory:Data Center Inventory Management (DCIM):Look up a rack type by `model` or `slug`.",
		Noun:        "rack_type",
		Attrs: []fwresource.Attr{
			{Name: "comments", Kind: fwresource.String, Computed: true, Description: "Free-form comments about the rack type."},
			{Name: "description", Kind: fwresource.String, Computed: true, Description: "A description of the rack type."},
			{Name: "form_factor", Kind: fwresource.String, Computed: true, Description: "The form factor of the rack type."},
			{Name: "manufacturer_id", Kind: fwresource.Int, Computed: true, Description: "The ID of the manufacturer associated with the rack type."},
			{Name: "max_weight", Kind: fwresource.Int, Computed: true, Description: "The max weight of the rack type."},
			{Name: "model", Kind: fwresource.String, Optional: true, Query: "-", Computed: true, Description: "The model of the rack type. At least one of `model` or `slug` must be given."},
			{Name: "mounting_depth_mm", Kind: fwresource.Int, Computed: true, Description: "The mounting depth mm of the rack type."},
			{Name: "outer_depth", Kind: fwresource.Int, Computed: true, Description: "The outer depth of the rack type."},
			{Name: "outer_unit", Kind: fwresource.String, Computed: true, Description: "The outer unit of the rack type."},
			{Name: "outer_width", Kind: fwresource.Int, Computed: true, Description: "The outer width of the rack type."},
			{Name: "slug", Kind: fwresource.String, Optional: true, Query: "-", Computed: true, Description: "The URL-friendly slug of the rack type. At least one of `model` or `slug` must be given."},
			{Name: "starting_unit", Kind: fwresource.Int, Computed: true, Description: "The starting unit of the rack type."},
			{Name: "u_height", Kind: fwresource.Int, Computed: true, Description: "The height of the rack type in rack units."},
			{Name: "weight", Kind: fwresource.Float, Computed: true, Description: "The weight of the rack type."},
			{Name: "weight_unit", Kind: fwresource.String, Computed: true, Description: "The unit of the weight."},
			{Name: "width", Kind: fwresource.Int, Computed: true, Description: "The width of the rack type."},
		},
	},
	{
		TypeName:    "netbox_virtual_disk",
		Description: ":meta:subcategory:Virtualization:Retrieves all virtual disk from NetBox. Narrow the result with filter blocks and cap it with limit.",
		ExtraFilters: []fwresource.Attr{
			{Name: "name_regex", Kind: fwresource.String, Optional: true, Query: "-", Description: "The name regex of the virtual disk."},
			{Name: "ordering", Kind: fwresource.StringList, Optional: true, Query: "-", Description: "One or more fields to use for ordering the results. Prefix a field name with `-` to reverse the order. E.g. `id`, `-name`."},
		},
		Plural: &fwresource.Plural{Result: "virtual_disks", What: "virtual_disks", Filters: map[string]string{}},
		Attrs: []fwresource.Attr{
			{Name: "custom_fields", Kind: fwresource.StringMap},
			{Name: "description", Kind: fwresource.String, Zero: true, Description: "A description of the virtual disk."},
			{Name: "id", Kind: fwresource.Int, Zero: true, Description: "The numeric ID of the virtual disk."},
			{Name: "name", Kind: fwresource.String, Zero: true, Description: "The name of the virtual disk."},
			{Name: "size_mb", Kind: fwresource.Int, Zero: true, Description: "The size mb of the virtual disk."},
			{Name: "tags", Kind: fwresource.StringList},
			{Name: "virtual_machine_id", Kind: fwresource.Int, Zero: true, Description: "The ID of the virtual machine associated with the virtual disk."},
		},
	},
}
