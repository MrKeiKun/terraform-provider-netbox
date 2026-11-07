package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxWirelessLANGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxWirelessLANGroupCreate,
		Read:   resourceNetboxWirelessLANGroupRead,
		Update: resourceNetboxWirelessLANGroupUpdate,
		Delete: resourceNetboxWirelessLANGroupDelete,

		Description: `:meta:subcategory:Wireless:

> A Wireless LAN Group is used to organize wireless LANs into a recursive hierarchy.`,

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
			"parent_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const wirelessLANGroupAPIPath = "wireless/wireless-lan-groups/"

type wirelessLANGroupRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	// Parent is always sent, so a missing parent is sent as null and clears it.
	Parent       *int64             `json:"parent"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type wirelessLANGroupResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Parent       *netboxapi.Ref         `json:"parent"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func wirelessLANGroupRequestFromResourceData(api *providerState, d *schema.ResourceData) (*wirelessLANGroupRequest, error) {
	name := d.Get("name").(string)

	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}

	data := &wirelessLANGroupRequest{
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Description: getOptionalStr(d, "description", true),
		Tags:        tags,
	}
	if parentID := d.Get("parent_id").(int); parentID != 0 {
		data.Parent = int64ToPtr(int64(parentID))
	}
	if cf, ok := d.GetOk(customFieldsKey); ok {
		data.CustomFields = cf
	}
	return data, nil
}

func resourceNetboxWirelessLANGroupCreate(d *schema.ResourceData, m interface{}) error {
	data, err := wirelessLANGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, wirelessLANGroupAPIPath, data, resourceNetboxWirelessLANGroupRead)
}

func resourceNetboxWirelessLANGroupRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, wirelessLANGroupAPIPath, func(group *wirelessLANGroupResponse) {
		d.Set("name", group.Name)
		d.Set("slug", group.Slug)
		d.Set("description", group.Description)
		d.Set("parent_id", netboxapi.RefID(group.Parent))
		if cf := getCustomFields(group.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}
		api.readTagRefs(d, group.Tags)
	})
}

func resourceNetboxWirelessLANGroupUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := wirelessLANGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, wirelessLANGroupAPIPath, data, resourceNetboxWirelessLANGroupRead)
}

func resourceNetboxWirelessLANGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, wirelessLANGroupAPIPath)
}
