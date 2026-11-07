package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxRegion() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxRegionCreate,
		Read:   resourceNetboxRegionRead,
		Update: resourceNetboxRegionUpdate,
		Delete: resourceNetboxRegionDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/sites-and-racks/#regions):

> Sites can be arranged geographically using regions. A region might represent a continent, country, city, campus, or other area depending on your use case. Regions can be nested recursively to construct a hierarchy. For example, you might define several country regions, and within each of those several state or city regions to which sites are assigned.
>
> Each region must have a name that is unique within its parent region, if any.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"parent_region_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const regionAPIPath = "dcim/regions/"

type regionRequest struct {
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description,omitempty"`
	Parent      *int64             `json:"parent,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type regionResponse struct {
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Parent      *netboxapi.Ref `json:"parent"`
}

func regionRequestFromResourceData(d *schema.ResourceData) *regionRequest {
	name := d.Get("name").(string)
	data := &regionRequest{
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Description: d.Get("description").(string),
		Tags:        []netboxapi.TagRef{},
	}
	if v, ok := d.GetOk("parent_region_id"); ok {
		data.Parent = int64ToPtr(int64(v.(int)))
	}
	return data
}

func resourceNetboxRegionCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, regionAPIPath, regionRequestFromResourceData(d), resourceNetboxRegionRead)
}

func resourceNetboxRegionRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, regionAPIPath, func(r *regionResponse) {
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
		d.Set("parent_region_id", netboxapi.RefID(r.Parent))
	})
}

func resourceNetboxRegionUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, regionAPIPath, regionRequestFromResourceData(d), resourceNetboxRegionRead)
}

func resourceNetboxRegionDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, regionAPIPath)
}
