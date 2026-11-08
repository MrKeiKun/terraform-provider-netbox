package netbox

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxVirtualDisk() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxVirtualDiskRead,
		Description: `:meta:subcategory:Virtualization:Retrieves all virtual disk from NetBox. Narrow the result with filter blocks and cap it with limit.`,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeList,
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
				Description:  "The name regex of the virtual disk.",
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
			"ordering": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "One or more fields to use for ordering the results. Prefix a field name with `-` to reverse the order. E.g. `id`, `-name`.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"virtual_disks": {
				Type:        schema.TypeList,
				Description: "The virtual disks of the virtual disk.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the virtual disk.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the virtual disk.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "A description of the virtual disk.",
							Computed:    true,
						},
						"size_mb": {
							Type:        schema.TypeInt,
							Description: "The size mb of the virtual disk.",
							Computed:    true,
						},
						"virtual_machine_id": {
							Type:        schema.TypeInt,
							Description: "The ID of the virtual machine associated with the virtual disk.",
							Computed:    true,
						},
						tagsKey: {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						customFieldsKey: {
							Type:     schema.TypeMap,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxVirtualDiskRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	var userLimit int64
	if limitValue, ok := d.GetOk("limit"); ok {
		userLimit = int64(limitValue.(int))
	}

	query := url.Values{}
	if ordering, ok := d.GetOk("ordering"); ok {
		var orderingStrs []string
		for _, o := range ordering.([]interface{}) {
			orderingStrs = append(orderingStrs, o.(string))
		}
		query.Set("ordering", strings.Join(orderingStrs, ","))
	}
	if filter, ok := d.GetOk("filter"); ok {
		for _, f := range filter.([]interface{}) {
			k := f.(map[string]interface{})["name"]
			vString := f.(map[string]interface{})["value"].(string)
			switch k {
			case "name":
				query.Set("name__ic", vString)
			case "virtual_machine_id":
				query.Set("virtual_machine_id", vString)
			case "tag":
				query.Add("tag", vString)
			default:
				return fmt.Errorf("'%s' is not a supported filter parameter", k)
			}
		}
	}

	// Fetch all pages, the name_regex and the limit are applied afterwards
	disks, err := netboxapi.List[map[string]interface{}](context.Background(), api.rest, "virtualization/virtual-disks/", query)
	if err != nil {
		return fmt.Errorf("failed to fetch virtual disks: %w", err)
	}

	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		var filtered []map[string]interface{}
		for _, disk := range disks {
			if r.MatchString(jStr(disk, "name")) {
				filtered = append(filtered, disk)
			}
		}
		disks = filtered
	}
	if userLimit > 0 && int64(len(disks)) > userLimit {
		disks = disks[:userLimit]
	}

	var s []map[string]interface{}
	for _, v := range disks {
		mapping := make(map[string]interface{})
		if id := jInt(v, "id"); id != 0 {
			mapping["id"] = id
		}
		mapping["name"] = jStr(v, "name")
		if description := jStr(v, "description"); description != "" {
			mapping["description"] = description
		}
		if size := jIntP(v, "size"); size != nil {
			mapping["size_mb"] = size
		}
		if vm := jID(v, "virtual_machine"); vm != nil {
			mapping["virtual_machine_id"] = vm
		}
		if v["custom_fields"] != nil {
			mapping["custom_fields"] = flattenCustomFields(v["custom_fields"])
		}
		if v["tags"] != nil {
			mapping["tags"] = jTags(v)
		}
		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("virtual_disks", s)
}
