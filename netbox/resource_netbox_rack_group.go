package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxRackGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxRackGroupCreate),
		ReadContext:   withContext(resourceNetboxRackGroupRead),
		UpdateContext: withContext(resourceNetboxRackGroupUpdate),
		DeleteContext: withContext(resourceNetboxRackGroupDelete),

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/rackgroup/):

> Rack groups can be used to organize racks. A rack can be assigned to multiple groups.

Requires NetBox v4.6.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the rack group.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the rack group.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the rack group.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the rack group.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the rack group.",
				Optional:    true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const rackGroupAPIPath = "dcim/rack-groups/"

type rackGroupRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type rackGroupResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func rackGroupRequestFromResourceData(api *providerState, d *schema.ResourceData) (*rackGroupRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &rackGroupRequest{
		Name:        d.Get("name").(string),
		Slug:        slugFromResourceData(d, name),
		Description: getOptionalStr(d, "description", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxRackGroupCreate(d *schema.ResourceData, m interface{}) error {
	data, err := rackGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, rackGroupAPIPath, data, resourceNetboxRackGroupRead)
}

func resourceNetboxRackGroupRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, rackGroupAPIPath, func(r *rackGroupResponse) {
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxRackGroupUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := rackGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, rackGroupAPIPath, data, resourceNetboxRackGroupRead)
}

func resourceNetboxRackGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, rackGroupAPIPath)
}
