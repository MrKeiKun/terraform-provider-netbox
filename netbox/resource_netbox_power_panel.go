package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxPowerPanel() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxPowerPanelCreate,
		Read:   resourceNetboxPowerPanelRead,
		Update: resourceNetboxPowerPanelUpdate,
		Delete: resourceNetboxPowerPanelDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/powerpanel/):

> A power panel represents the origin point in NetBox for electrical power being disseminated by one or more power feeds. In a data center environment, one power panel often serves a group of racks, with an individual power feed extending to each rack, though this is not always the case. It is common to have two sets of panels and feeds arranged in parallel to provide redundant power to each rack.`,

		Schema: map[string]*schema.Schema{
			"site_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"location_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const powerPanelAPIPath = "dcim/power-panels/"

type powerPanelRequest struct {
	Name         string             `json:"name"`
	Site         *int64             `json:"site"`
	Location     *int64             `json:"location"`
	Description  string             `json:"description,omitempty"`
	Comments     string             `json:"comments,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type powerPanelResponse struct {
	Name         string                 `json:"name"`
	Site         *netboxapi.Ref         `json:"site"`
	Location     *netboxapi.Ref         `json:"location"`
	Description  string                 `json:"description"`
	Comments     string                 `json:"comments"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func powerPanelRequestFromResourceData(api *providerState, d *schema.ResourceData) (*powerPanelRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &powerPanelRequest{
		Name:        name,
		Site:        int64ToPtr(int64(d.Get("site_id").(int))),
		Location:    getOptionalInt(d, "location_id"),
		Description: getOptionalStr(d, "description", true),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxPowerPanelCreate(d *schema.ResourceData, m interface{}) error {
	data, err := powerPanelRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, powerPanelAPIPath, data, resourceNetboxPowerPanelRead)
}

func resourceNetboxPowerPanelRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, powerPanelAPIPath, func(r *powerPanelResponse) {
		d.Set("name", r.Name)
		d.Set("site_id", netboxapi.RefID(r.Site))
		d.Set("location_id", netboxapi.RefID(r.Location))
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxPowerPanelUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := powerPanelRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, powerPanelAPIPath, data, resourceNetboxPowerPanelRead)
}

func resourceNetboxPowerPanelDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, powerPanelAPIPath)
}
