// Copyright (c) 2022 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package netbox

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxDevices() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxDevicesRead),
		Description: ":meta:subcategory:Data Center Inventory Management (DCIM):Retrieves all devices from NetBox. Narrow the result with filter blocks and cap it with limit.",
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeSet,
				Description: "Filters that select which items are returned. Each filter is a name/value pair.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the field to filter by.",
							Required:    true,
						},
						"value": {
							Type:        schema.TypeString,
							Description: "The value of the field to filter by.",
							Required:    true,
						},
					},
				},
			},
			"name_regex": {
				Type:         schema.TypeString,
				Description:  "The name regex of the devices.",
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"limit": {
				Type:             schema.TypeInt,
				Description:      "The maximum number of items to return. Defaults to 0, which returns all items.",
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
			},
			"devices": {
				Type:        schema.TypeList,
				Description: "The devices of the devices.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"asset_tag": {
							Type:        schema.TypeString,
							Description: "The asset tag of the devices.",
							Computed:    true,
						},
						"cluster_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the cluster associated with the devices.",
							Computed:    true,
						},
						"comments": {
							Type:        schema.TypeString,
							Description: "Free-form comments about the devices.",
							Computed:    true,
						},
						"config_context": {
							Type:        schema.TypeString,
							Description: "The config context of the devices.",
							Computed:    true,
						},
						"local_context_data": {
							Type:        schema.TypeString,
							Description: "The local context data of the devices.",
							Computed:    true,
						},
						"custom_fields": {
							Type:        schema.TypeMap,
							Description: "Custom field values of the devices, as a map of field name to value.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the devices.",
							Computed:    true,
						},
						"device_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the device associated with the devices.",
							Computed:    true,
						},
						"device_type_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the device type associated with the devices.",
							Computed:    true,
						},
						"location_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the location associated with the devices.",
							Computed:    true,
						},
						"manufacturer_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the manufacturer associated with the devices.",
							Computed:    true,
						},
						"model": {
							Type:        schema.TypeString,
							Description: "The model of the devices.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the devices.",
							Computed:    true,
						},
						"platform_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the platform associated with the devices.",
							Computed:    true,
						},
						"site_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the site associated with the devices.",
							Computed:    true,
						},
						"tenant_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the tenant associated with the devices.",
							Computed:    true,
						},
						"role_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the role associated with the devices.",
							Computed:    true,
						},
						"serial": {
							Type:        schema.TypeString,
							Description: "The serial number of the devices.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "The status of the devices.",
							Computed:    true,
						},
						"rack_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the rack associated with the devices.",
							Computed:    true,
						},
						"rack_face": {
							Type:        schema.TypeString,
							Description: "The rack face of the devices.",
							Computed:    true,
						},
						"rack_position": {
							Type:        schema.TypeFloat,
							Description: "The rack position of the devices.",
							Computed:    true,
						},
						"primary_ipv4": {
							Type:        schema.TypeString,
							Description: "The primary ipv4 of the devices.",
							Computed:    true,
						},
						"primary_ipv6": {
							Type:        schema.TypeString,
							Description: "The primary ipv6 of the devices.",
							Computed:    true,
						},
						"oob_ip": {
							Type:        schema.TypeString,
							Description: "The oob ip of the devices.",
							Computed:    true,
						},
						"tags": tagsSchemaRead,
					},
				},
			},
		},
	}
}

func dataSourceNetboxDevicesRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"asset_tag":      "asset_tag",
		"cluster_id":     "cluster_id",
		"device_type_id": "device_type_id",
		"name":           "name",
		"region":         "region",
		"role_id":        "role_id",
		"site_id":        "site_id",
		"location_id":    "location_id",
		"rack_id":        "rack_id",
		"tenant_id":      "tenant_id",
		"tags":           "tag[]",
		"status":         "status",
	}

	// Fetch everything and apply the regular expression before the limit
	var keep func(map[string]interface{}) bool
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		keep = func(o map[string]interface{}) bool { return r.MatchString(jStr(o, "name")) }
	}

	results, err := listFilteredWith[map[string]interface{}](d, m, "dcim/devices/", "devices", allowed, nil, keep)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["asset_tag"] = jStr(v, "asset_tag")
		mapping["cluster_id"] = jID(v, "cluster")
		mapping["comments"] = jStr(v, "comments")
		mapping["description"] = jStr(v, "description")
		mapping["config_context"] = jJSON(v, "config_context")
		mapping["local_context_data"] = jJSON(v, "local_context_data")
		mapping["device_id"] = jInt(v, "id")
		mapping["device_type_id"] = jID(v, "device_type")
		mapping["manufacturer_id"] = jDeepID(v, "device_type", "manufacturer")
		mapping["model"] = jNested(v, "device_type", "model")
		mapping["name"] = jStr(v, "name")
		mapping["location_id"] = jID(v, "location")
		mapping["platform_id"] = jID(v, "platform")
		mapping["site_id"] = jID(v, "site")
		mapping["tenant_id"] = jID(v, "tenant")
		mapping["role_id"] = jID(v, "role")
		mapping["serial"] = jStr(v, "serial")
		mapping["status"] = jChoice(v, "status")
		mapping["custom_fields"] = flattenCustomFields(v["custom_fields"])
		mapping["rack_id"] = jID(v, "rack")
		mapping["rack_position"] = jFloatP(v, "position")
		mapping["rack_face"] = jChoice(v, "face")
		mapping["tags"] = jTags(v)
		mapping["primary_ipv4"] = jIPAddress(v, "primary_ip4")
		mapping["primary_ipv6"] = jIPAddress(v, "primary_ip6")
		mapping["oob_ip"] = jIPAddress(v, "oob_ip")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("devices", s)
}
