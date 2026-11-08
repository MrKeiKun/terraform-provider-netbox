package netbox

import (
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxVirtualMachineType() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxVirtualMachineTypeRead,
		Description: `:meta:subcategory:Virtualization:Requires NetBox v4.6.0 or newer.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "slug"},
			},
			"default_platform_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_vcpus": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"default_memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Computed: true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

func dataSourceNetboxVirtualMachineTypeRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addStrQuery(query, d, "slug", "slug")
	r, err := findOne[map[string]interface{}](m, "virtualization/virtual-machine-types/", query, "virtual machine type")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(jInt(*r, "id"), 10))
	d.Set("name", jStr(*r, "name"))
	d.Set("slug", jStr(*r, "slug"))
	d.Set("default_platform_id", jID(*r, "default_platform"))
	d.Set("default_vcpus", jFloatP(*r, "default_vcpus"))
	d.Set("default_memory", jIntP(*r, "default_memory"))
	d.Set("description", jStr(*r, "description"))
	d.Set("owner_id", jID(*r, "owner"))
	d.Set("comments", jStr(*r, "comments"))
	d.Set(tagsKey, jTags(*r))
	return nil
}
