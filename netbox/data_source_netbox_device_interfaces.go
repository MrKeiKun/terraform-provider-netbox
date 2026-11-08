package netbox

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxDeviceInterfaces() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxDeviceInterfaceRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves all device interfaces from NetBox. Narrow the result with filter blocks and cap it with limit.`,
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
				Description:  "The name regex of the device interfaces.",
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
			"interfaces": {
				Type:        schema.TypeList,
				Description: "The interfaces of the device interfaces.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the device interfaces.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the device interfaces.",
							Computed:    true,
						},
						"mgmt_only": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this interface is used only for out-of-band management.",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Whether the device interfaces is enabled.",
							Computed:    true,
						},
						"mac_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The MAC address as string from the first MAC address assigned to this interface, if any.",
						},
						"mac_addresses": {
							Type:        schema.TypeSet,
							Description: "The mac addresses of the device interfaces.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Description: "The numeric ID of the device interfaces.",
										Computed:    true,
									},
									"mac_address": {
										Type:        schema.TypeString,
										Description: "The MAC address of the device interfaces.",
										Computed:    true,
									},
									"description": {
										Type:        schema.TypeString,
										Description: "A description of the device interfaces.",
										Computed:    true,
									},
								},
							},
						},
						"mode": {
							Type:        schema.TypeMap,
							Description: "The mode of the device interfaces.",
							Computed:    true,
							Elem: &schema.Schema{
								Type:     schema.TypeString,
								Computed: true,
							},
						},
						"mtu": {
							Type:        schema.TypeInt,
							Description: "The MTU of the device interfaces.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the device interfaces.",
							Computed:    true,
						},
						"tag_ids": {
							Type:        schema.TypeList,
							Description: "The IDs of the tag associated with the device interfaces.",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"tagged_vlans": {
							Type:        schema.TypeList,
							Description: "The tagged vlans of the device interfaces.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Description: "The numeric ID of the device interfaces.",
										Computed:    true,
									},
									"vid": {
										Type:        schema.TypeInt,
										Description: "The VLAN ID (1-4094).",
										Computed:    true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "The name of the device interfaces.",
										Computed:    true,
									},
								},
							},
						},
						// Do as a TypeList due to limitation of TypeMap
						"untagged_vlan": {
							Type:        schema.TypeList,
							Description: "The untagged vlan of the device interfaces.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Description: "The numeric ID of the device interfaces.",
										Computed:    true,
									},
									"vid": {
										Type:        schema.TypeInt,
										Description: "The VLAN ID (1-4094).",
										Computed:    true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "The name of the device interfaces.",
										Computed:    true,
									},
								},
							},
						},
						"device_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the device associated with the device interfaces.",
							Computed:    true,
						},
						"lag_device_interface_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the LAG interface this interface is a member of, if any.",
						},
						"type": {
							Type:        schema.TypeString,
							Description: "The type of the device interfaces.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxDeviceInterfaceRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"mac_address": "mac_address",
		"name":        "name",
		"tag":         "tag",
		"device_id":   "device_id",
		"lag_id":      "lag_id",
	}

	// Fetch everything and apply the regular expression before the limit
	var keep func(map[string]interface{}) bool
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		keep = func(o map[string]interface{}) bool { return r.MatchString(jStr(o, "name")) }
	}

	results, err := listFilteredWith[map[string]interface{}](d, m, "dcim/interfaces/", "interfaces", allowed, nil, keep)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["id"] = jInt(v, "id")
		mapping["description"] = jStr(v, "description")
		mapping["enabled"] = jBool(v, "enabled")
		mapping["mgmt_only"] = jBool(v, "mgmt_only")
		mapping["mac_address"] = jFirst(v, "mac_addresses", "mac_address")
		mapping["mac_addresses"] = jObjectsList(v, "mac_addresses", "id", "description", "mac_address")
		mapping["mode"] = jChoiceMap(v, "mode")
		mapping["mtu"] = jIntP(v, "mtu")
		mapping["name"] = jStr(v, "name")
		mapping["tagged_vlans"] = jObjectsList(v, "tagged_vlans", "id", "vid", "name")
		mapping["tag_ids"] = jTagIDs(v)
		mapping["untagged_vlan"] = jObjectList(v, "untagged_vlan", "id", "vid", "name")
		mapping["type"] = jChoice(v, "type")
		mapping["device_id"] = jID(v, "device")
		mapping["lag_device_interface_id"] = jID(v, "lag")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("interfaces", s)
}
