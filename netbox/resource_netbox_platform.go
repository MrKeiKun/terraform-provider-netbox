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
				Type:        schema.TypeString,
				Description: "The name of the platform.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the platform.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"manufacturer_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the manufacturer associated with the platform.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the platform.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the platform.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the platform.",
				Optional:    true,
			},
			"config_template_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the config template associated with the platform.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const platformAPIPath = "dcim/platforms/"

type platformRequest struct {
	Name           string             `json:"name"`
	Slug           string             `json:"slug"`
	Description    string             `json:"description,omitempty"`
	Manufacturer   *int64             `json:"manufacturer,omitempty"`
	Owner          *int64             `json:"owner"`
	Comments       string             `json:"comments,omitempty"`
	ConfigTemplate *int64             `json:"config_template"`
	CustomFields   interface{}        `json:"custom_fields,omitempty"`
	Tags           []netboxapi.TagRef `json:"tags"`
}

type platformResponse struct {
	Name           string                 `json:"name"`
	Slug           string                 `json:"slug"`
	Description    string                 `json:"description"`
	Manufacturer   *netboxapi.Ref         `json:"manufacturer"`
	Owner          *netboxapi.Ref         `json:"owner"`
	Comments       string                 `json:"comments"`
	ConfigTemplate *netboxapi.Ref         `json:"config_template"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
	Tags           []netboxapi.TagRef     `json:"tags"`
}

func platformRequestFromResourceData(api *providerState, d *schema.ResourceData) (*platformRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &platformRequest{
		CustomFields:   customFieldsRequest(d),
		ConfigTemplate: getOptionalInt(d, "config_template_id"),
		Comments:       getOptionalStr(d, "comments", true),
		Owner:          getOptionalInt(d, "owner_id"),
		Name:           name,
		Slug:           slugFromResourceData(d, name),
		Description:    d.Get("description").(string),
		Tags:           tags,
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
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("config_template_id", netboxapi.RefID(r.ConfigTemplate))
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
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
