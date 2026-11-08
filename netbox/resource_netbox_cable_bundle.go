package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxCableBundle() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCableBundleCreate,
		Read:   resourceNetboxCableBundleRead,
		Update: resourceNetboxCableBundleUpdate,
		Delete: resourceNetboxCableBundleDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/cablebundle/):

> A cable bundle groups cables that run together, for example inside the same conduit or trunk.

Requires NetBox v4.6.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the cable bundle.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cable bundle.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the cable bundle.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the cable bundle.",
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

const cableBundleAPIPath = "dcim/cable-bundles/"

type cableBundleRequest struct {
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type cableBundleResponse struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func cableBundleRequestFromResourceData(api *providerState, d *schema.ResourceData) (*cableBundleRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &cableBundleRequest{
		Name:        d.Get("name").(string),
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

func resourceNetboxCableBundleCreate(d *schema.ResourceData, m interface{}) error {
	data, err := cableBundleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, cableBundleAPIPath, data, resourceNetboxCableBundleRead)
}

func resourceNetboxCableBundleRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, cableBundleAPIPath, func(r *cableBundleResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxCableBundleUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := cableBundleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, cableBundleAPIPath, data, resourceNetboxCableBundleRead)
}

func resourceNetboxCableBundleDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, cableBundleAPIPath)
}
