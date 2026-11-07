package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCoolingFeed() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCoolingFeedCreate,
		Read:   resourceNetboxCoolingFeedRead,
		Update: resourceNetboxCoolingFeedUpdate,
		Delete: resourceNetboxCoolingFeedDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/coolingfeed/):

> A cooling feed represents the supply of cooling from a cooling source to a rack or another piece of equipment.

Requires NetBox v4.7.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"cooling_source_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"rack_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [offline, active, planned, failed]",
				ValidateFunc: validation.StringInSlice([]string{"offline", "active", "planned", "failed"}, false),
			},
			"cooling_capacity": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Rated cooling capacity (kW)",
			},
			"max_flow": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"max_flow_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [lpm, m3ph, gpm]",
				ValidateFunc: validation.StringInSlice([]string{"lpm", "m3ph", "gpm"}, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
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

const coolingFeedAPIPath = "dcim/cooling-feeds/"

type coolingFeedRequest struct {
	CoolingSource   *int64             `json:"cooling_source"`
	Rack            *int64             `json:"rack"`
	Name            string             `json:"name"`
	Status          string             `json:"status,omitempty"`
	CoolingCapacity *float64           `json:"cooling_capacity"`
	MaxFlow         *float64           `json:"max_flow"`
	MaxFlowUnit     *string            `json:"max_flow_unit,omitempty"`
	Description     string             `json:"description,omitempty"`
	Tenant          *int64             `json:"tenant"`
	Owner           *int64             `json:"owner"`
	Comments        string             `json:"comments,omitempty"`
	Tags            []netboxapi.TagRef `json:"tags"`
	CustomFields    interface{}        `json:"custom_fields,omitempty"`
}

type coolingFeedResponse struct {
	CoolingSource   *netboxapi.Ref         `json:"cooling_source"`
	Rack            *netboxapi.Ref         `json:"rack"`
	Name            string                 `json:"name"`
	Status          *netboxapi.Choice      `json:"status"`
	CoolingCapacity *float64               `json:"cooling_capacity"`
	MaxFlow         *float64               `json:"max_flow"`
	MaxFlowUnit     *netboxapi.Choice      `json:"max_flow_unit"`
	Description     string                 `json:"description"`
	Tenant          *netboxapi.Ref         `json:"tenant"`
	Owner           *netboxapi.Ref         `json:"owner"`
	Comments        string                 `json:"comments"`
	Tags            []netboxapi.TagRef     `json:"tags"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
}

func coolingFeedRequestFromResourceData(api *providerState, d *schema.ResourceData) (*coolingFeedRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &coolingFeedRequest{
		CoolingSource:   int64ToPtr(int64(d.Get("cooling_source_id").(int))),
		Rack:            getOptionalInt(d, "rack_id"),
		Name:            d.Get("name").(string),
		Status:          d.Get("status").(string),
		CoolingCapacity: getOptionalFloat(d, "cooling_capacity"),
		MaxFlow:         getOptionalFloat(d, "max_flow"),
		MaxFlowUnit:     getOptionalStrClear(d, "max_flow_unit"),
		Description:     getOptionalStr(d, "description", true),
		Tenant:          getOptionalInt(d, "tenant_id"),
		Owner:           getOptionalInt(d, "owner_id"),
		Comments:        getOptionalStr(d, "comments", true),
		Tags:            tags,
	}
	if ct, ok := d.GetOk(customFieldsKey); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxCoolingFeedCreate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingFeedRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, coolingFeedAPIPath, data, resourceNetboxCoolingFeedRead)
}

func resourceNetboxCoolingFeedRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, coolingFeedAPIPath, func(r *coolingFeedResponse) {
		d.Set("cooling_source_id", netboxapi.RefID(r.CoolingSource))
		d.Set("rack_id", netboxapi.RefID(r.Rack))
		d.Set("name", r.Name)
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("cooling_capacity", r.CoolingCapacity)
		d.Set("max_flow", r.MaxFlow)
		d.Set("max_flow_unit", netboxapi.ChoiceValue(r.MaxFlowUnit))
		d.Set("description", r.Description)
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		if cf := getCustomFields(r.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}
	})
}

func resourceNetboxCoolingFeedUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingFeedRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, coolingFeedAPIPath, data, resourceNetboxCoolingFeedRead)
}

func resourceNetboxCoolingFeedDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, coolingFeedAPIPath)
}
