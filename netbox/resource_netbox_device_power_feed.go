package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxPowerFeed() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxPowerFeedCreate,
		Read:   resourceNetboxPowerFeedRead,
		Update: resourceNetboxPowerFeedUpdate,
		Delete: resourceNetboxPowerFeedDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/powerfeed/):

> A power feed represents the distribution of power from a power panel to a particular device, typically a power distribution unit (PDU). The power port (inlet) on a device can be connected via a cable to a power feed. A power feed may optionally be assigned to a rack to allow more easily tracking the distribution of power among racks.`,

		Schema: map[string]*schema.Schema{
			"power_panel_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the power panel associated with the device power feed.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the device power feed.",
				Required:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "One of [offline, active, planned, failed]",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [primary, redundant]",
				ValidateFunc: validation.StringInSlice([]string{"primary", "redundant"}, false),
			},
			"supply": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [ac, dc]",
				ValidateFunc: validation.StringInSlice([]string{"ac", "dc"}, false),
			},
			"phase": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [single-phase, three-phase]",
				ValidateFunc: validation.StringInSlice([]string{"single-phase", "three-phase"}, false),
			},
			"voltage": {
				Type:        schema.TypeInt,
				Description: "The voltage of the device power feed.",
				Required:    true,
			},
			"amperage": {
				Type:        schema.TypeInt,
				Description: "The amperage of the device power feed.",
				Required:    true,
			},
			"max_percent_utilization": {
				Type:         schema.TypeInt,
				Description:  "The max percent utilization of the device power feed.",
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 100),
			},
			"rack_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the rack associated with the device power feed.",
				Optional:    true,
			},
			"mark_connected": {
				Type:        schema.TypeBool,
				Description: "Whether the device power feed is marked as connected even without a cable.",
				Default:     false,
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the device power feed.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the device power feed.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the device power feed.",
				Optional:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the device power feed.",
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

const powerFeedAPIPath = "dcim/power-feeds/"

type powerFeedRequest struct {
	Name           string             `json:"name"`
	PowerPanel     *int64             `json:"power_panel"`
	Status         string             `json:"status,omitempty"`
	Type           string             `json:"type"`
	Supply         string             `json:"supply,omitempty"`
	Phase          string             `json:"phase,omitempty"`
	Voltage        *int64             `json:"voltage,omitempty"`
	Amperage       int64              `json:"amperage,omitempty"`
	MaxUtilization int64              `json:"max_utilization,omitempty"`
	Rack           *int64             `json:"rack"`
	MarkConnected  bool               `json:"mark_connected"`
	Description    string             `json:"description,omitempty"`
	Comments       string             `json:"comments,omitempty"`
	Owner          *int64             `json:"owner"`
	Tenant         *int64             `json:"tenant"`
	CustomFields   interface{}        `json:"custom_fields,omitempty"`
	Tags           []netboxapi.TagRef `json:"tags"`
}

type powerFeedResponse struct {
	Name           string                 `json:"name"`
	PowerPanel     *netboxapi.Ref         `json:"power_panel"`
	Status         *netboxapi.Choice      `json:"status"`
	Type           *netboxapi.Choice      `json:"type"`
	Supply         *netboxapi.Choice      `json:"supply"`
	Phase          *netboxapi.Choice      `json:"phase"`
	Voltage        *int64                 `json:"voltage"`
	Amperage       int64                  `json:"amperage"`
	MaxUtilization int64                  `json:"max_utilization"`
	Rack           *netboxapi.Ref         `json:"rack"`
	MarkConnected  bool                   `json:"mark_connected"`
	Description    string                 `json:"description"`
	Comments       string                 `json:"comments"`
	Owner          *netboxapi.Ref         `json:"owner"`
	Tenant         *netboxapi.Ref         `json:"tenant"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
	Tags           []netboxapi.TagRef     `json:"tags"`
}

func powerFeedRequestFromResourceData(api *providerState, d *schema.ResourceData) (*powerFeedRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &powerFeedRequest{
		Tenant:         getOptionalInt(d, "tenant_id"),
		Owner:          getOptionalInt(d, "owner_id"),
		Name:           name,
		PowerPanel:     int64ToPtr(int64(d.Get("power_panel_id").(int))),
		Status:         d.Get("status").(string),
		Type:           d.Get("type").(string),
		Supply:         d.Get("supply").(string),
		Phase:          d.Get("phase").(string),
		Voltage:        int64ToPtr(int64(d.Get("voltage").(int))),
		Amperage:       int64(d.Get("amperage").(int)),
		MaxUtilization: int64(d.Get("max_percent_utilization").(int)),
		Rack:           getOptionalInt(d, "rack_id"),
		MarkConnected:  d.Get("mark_connected").(bool),
		Description:    getOptionalStr(d, "description", true),
		Comments:       getOptionalStr(d, "comments", true),
		Tags:           tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxPowerFeedCreate(d *schema.ResourceData, m interface{}) error {
	data, err := powerFeedRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, powerFeedAPIPath, data, resourceNetboxPowerFeedRead)
}

func resourceNetboxPowerFeedRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, powerFeedAPIPath, func(r *powerFeedResponse) {
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("power_panel_id", netboxapi.RefID(r.PowerPanel))
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("supply", netboxapi.ChoiceValue(r.Supply))
		d.Set("phase", netboxapi.ChoiceValue(r.Phase))
		d.Set("voltage", r.Voltage)
		d.Set("amperage", r.Amperage)
		d.Set("max_percent_utilization", r.MaxUtilization)
		d.Set("rack_id", netboxapi.RefID(r.Rack))
		d.Set("mark_connected", r.MarkConnected)
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxPowerFeedUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := powerFeedRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, powerFeedAPIPath, data, resourceNetboxPowerFeedRead)
}

func resourceNetboxPowerFeedDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, powerFeedAPIPath)
}
