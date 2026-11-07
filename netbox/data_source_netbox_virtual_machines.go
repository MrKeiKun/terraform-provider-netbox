package netbox

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxVirtualMachine() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxVirtualMachineRead,
		Description: `:meta:subcategory:Virtualization:`,
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
			"vms": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"disk_size_mb": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"local_context_data": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"memory_mb": {
							Type:     schema.TypeInt,
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
						"platform_slug": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"platform_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"primary_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"primary_ip4": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"primary_ip6": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"site_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tag_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"tenant_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vcpus": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"vm_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxVirtualMachineRead(d *schema.ResourceData, m interface{}) error {
	// filter name in the data source -> query parameter sent to Netbox
	allowed := map[string]string{
		"cluster_id":    "cluster_id",
		"cluster_group": "cluster_group",
		"device_id":     "name",
		"device":        "name",
		"name":          "name",
		"region":        "region",
		"role":          "role",
		"site":          "site",
		"tenant_id":     "tenant_id",
		"tag":           "tag",
		"status":        "status",
		"platform":      "platform",
		"platform_id":   "platform_id",
	}

	// Fetch everything and apply the regular expression before the limit
	var keep func(map[string]interface{}) bool
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		keep = func(o map[string]interface{}) bool { return r.MatchString(jStr(o, "name")) }
	}

	results, err := listFilteredWith[map[string]interface{}](d, m, "virtualization/virtual-machines/", "virtual machines", allowed, nil, keep)
	if err != nil {
		return err
	}

	var s []map[string]interface{}
	for _, v := range results {
		mapping := make(map[string]interface{})

		mapping["cluster_id"] = jID(v, "cluster")
		mapping["comments"] = jStr(v, "comments")
		mapping["description"] = jStr(v, "description")
		mapping["device_id"] = jID(v, "device")
		mapping["device_name"] = jNested(v, "device", "name")
		mapping["config_context"] = jJSON(v, "config_context")
		mapping["custom_fields"] = flattenCustomFields(v["custom_fields"])
		mapping["disk_size_mb"] = jIntP(v, "disk")
		mapping["local_context_data"] = jJSON(v, "local_context_data")
		mapping["memory_mb"] = jIntP(v, "memory")
		mapping["name"] = jStr(v, "name")
		mapping["platform_id"] = jID(v, "platform")
		mapping["platform_slug"] = jNested(v, "platform", "slug")
		mapping["platform_name"] = jNested(v, "platform", "name")
		mapping["primary_ip"] = jNested(v, "primary_ip", "address")
		mapping["primary_ip4"] = jNested(v, "primary_ip4", "address")
		mapping["primary_ip6"] = jNested(v, "primary_ip6", "address")
		mapping["role_id"] = jID(v, "role")
		mapping["site_id"] = jID(v, "site")
		mapping["status"] = jChoice(v, "status")
		mapping["tag_ids"] = jTagIDs(v)
		mapping["tenant_id"] = jID(v, "tenant")
		mapping["vcpus"] = jFloatP(v, "vcpus")
		mapping["vm_id"] = jInt(v, "id")

		// fields that are missing in the response are left out of the state
		for k, val := range mapping {
			if val == nil {
				delete(mapping, k)
			}
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("vms", s)
}
