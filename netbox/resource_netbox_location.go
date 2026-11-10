package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxLocation() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxLocationCreate),
		ReadContext:   withContext(resourceNetboxLocationRead),
		UpdateContext: withContext(resourceNetboxLocationUpdate),
		DeleteContext: withContext(resourceNetboxLocationDelete),

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/sites-and-racks/#locations):

> Racks and devices can be grouped by location within a site. A location may represent a floor, room, cage, or similar organizational unit. Locations can be nested to form a hierarchy. For example, you may have floors within a site, and rooms within a floor.

Each location must have a name that is unique within its parent site and location, if any.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the location.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the location.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the location.",
				Optional:    true,
			},
			"facility": {
				Type:         schema.TypeString,
				Description:  "The facility of the location.",
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"site_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the site associated with the location.",
				Required:    true,
			},
			"parent_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the parent associated with the location.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the location.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the location.",
				Optional:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "One of [planned, staging, active, decommissioning, retired]",
			},
			tagsKey: tagsSchema,
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the location.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const locationAPIPath = "dcim/locations/"

type locationRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Facility     string             `json:"facility,omitempty"`
	Site         *int64             `json:"site"`
	Parent       *int64             `json:"parent"`
	Tenant       *int64             `json:"tenant,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Status       string             `json:"status,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type locationResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Facility     string                 `json:"facility"`
	Site         *netboxapi.Ref         `json:"site"`
	Parent       *netboxapi.Ref         `json:"parent"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Status       *netboxapi.Choice      `json:"status"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func locationRequestFromResourceData(api *providerState, d *schema.ResourceData) (*locationRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &locationRequest{
		Status:      d.Get("status").(string),
		Comments:    getOptionalStr(d, "comments", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Description: getOptionalStr(d, "description", true),
		Facility:    getOptionalStr(d, "facility", true),
		Tags:        tags,
	}
	if v, ok := d.GetOk("site_id"); ok {
		data.Site = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("parent_id"); ok {
		data.Parent = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("tenant_id"); ok {
		data.Tenant = int64ToPtr(int64(v.(int)))
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxLocationCreate(d *schema.ResourceData, m interface{}) error {
	data, err := locationRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, locationAPIPath, data, resourceNetboxLocationRead)
}

func resourceNetboxLocationRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, locationAPIPath, func(r *locationResponse) {
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
		d.Set("facility", r.Facility)
		d.Set("site_id", netboxapi.RefID(r.Site))
		d.Set("parent_id", netboxapi.RefID(r.Parent))
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxLocationUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := locationRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, locationAPIPath, data, resourceNetboxLocationRead)
}

func resourceNetboxLocationDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, locationAPIPath)
}
