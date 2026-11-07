package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxPlatform() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxPlatformCreate,
		Read:   resourceNetboxPlatformRead,
		Update: resourceNetboxPlatformUpdate,
		Delete: resourceNetboxPlatformDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/devices/#platforms):

> A platform defines the type of software running on a device or virtual machine. This can be helpful to model when it is necessary to distinguish between different versions or feature sets. Note that two devices of the same type may be assigned different platforms: For example, one Juniper MX240 might run Junos 14 while another runs Junos 15.`,

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
			"manufacturer_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			tagsKey: tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const platformAPIPath = "dcim/platforms/"

type platformRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Manufacturer *int64             `json:"manufacturer,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type platformResponse struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description"`
	Manufacturer *netboxapi.Ref     `json:"manufacturer"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

func platformRequestFromResourceData(api *providerState, d *schema.ResourceData) (*platformRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &platformRequest{
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Description: d.Get("description").(string),
		Tags:        tags,
	}
	if v, ok := d.GetOk("manufacturer_id"); ok {
		data.Manufacturer = int64ToPtr(int64(v.(int)))
	}
	return data, nil
}

func resourceNetboxPlatformCreate(d *schema.ResourceData, m interface{}) error {
	data, err := platformRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, platformAPIPath, data, resourceNetboxPlatformRead)
}

func resourceNetboxPlatformRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, platformAPIPath, func(r *platformResponse) {
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
		if r.Manufacturer != nil {
			d.Set("manufacturer_id", r.Manufacturer.ID)
		}
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxPlatformUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := platformRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, platformAPIPath, data, resourceNetboxPlatformRead)
}

func resourceNetboxPlatformDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, platformAPIPath)
}
