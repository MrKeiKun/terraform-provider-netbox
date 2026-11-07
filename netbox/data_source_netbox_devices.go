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
		Read:        dataSourceNetboxDevicesRead,
		Description: ":meta:subcategory:Data Center Inventory Management (DCIM):",
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"limit": {
				Type:             schema.TypeInt,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
			},
			"devices": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"asset_tag": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"comments": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"config_context": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"local_context_data": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"custom_fields": {
							Type:     schema.TypeMap,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"device_type_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"location_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"manufacturer_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"model": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"platform_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"site_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"tenant_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"role_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"serial": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"rack_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"rack_face": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"rack_position": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"primary_ipv4": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"primary_ipv6": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"oob_ip": {
							Type:     schema.TypeString,
							Computed: true,
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
